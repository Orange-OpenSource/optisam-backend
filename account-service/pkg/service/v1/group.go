// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	v1 "optisam-backend/account-service/pkg/api/v1"
	repo "optisam-backend/account-service/pkg/repository/v1"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	"optisam-backend/common/optisam/token/claims"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListGroups list all the groups owned by admin user.
func (s *accountServiceServer) ListGroups(ctx context.Context, req *v1.ListGroupsRequest) (*v1.ListGroupsResponse, error) {
	claims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	logger.Log.Info("user id", zap.String("user_id", claims.UserID))
	userID := claims.UserID // get userID from context
	totalGrps, groups, err := s.accountRepo.UserOwnedGroups(ctx, userID, nil)
	if err != nil {
		logger.Log.Error("service/v1 - ListGroups - ", zap.String("reason", err.Error()), zap.String("UserID", userID))
		return nil, status.Error(codes.Unknown, "service/v1 - ListGroups - failed to get Groups")
	}

	userGroups := &v1.ListGroupsResponse{
		NumOfRecords: int32(totalGrps),
		Groups:       make([]*v1.Group, totalGrps),
	}

	for i := range groups {
		userGroups.Groups[i] = convertRepoGroupToSrvGroup(groups[i])
	}

	return userGroups, nil

}

// CreateGroup implemntsv1.AccountServiceServer CreateGroup function
func (s *accountServiceServer) CreateGroup(ctx context.Context, req *v1.Group) (*v1.Group, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	userID := userClaims.UserID
	switch userClaims.Role {
	case claims.RoleUser:
		return nil, status.Error(codes.PermissionDenied, "user doesnot have access to create group")
	case claims.RoleAdmin, claims.RoleSuperAdmin:
		_, grps, err := s.accountRepo.UserOwnedGroups(ctx, userID, nil)
		if err != nil {
			// TODO: log error
			return nil, status.Error(codes.Internal, "cannot create group - fails to fetch users owned groups")
		}
		// TODO optimize this we don't need to fetch all groups from database
		pgIdx := parentGroupIDX(req.ParentId, grps)
		if pgIdx == -1 {
			return nil, status.Error(codes.InvalidArgument, "parent cannot be found")
		}

		parentGroup, err := s.accountRepo.GroupInfo(ctx, req.ParentId)
		if err != nil {
			logger.Log.Error("service/v1 - CreateGroup - GroupInfo", zap.Error(err))
			return nil, status.Error(codes.Internal, "cannot create group - fails to fetch parent group")
		}
		if parentGroup.Name == req.Name {
			return nil, status.Error(codes.Internal, "cannot create group - parent and child can not have same name")
		}

		fullName := grps[pgIdx].FullyQualifiedName + "." + req.Name
		nameFound := ifFullQualfNameExists(fullName, grps)
		if nameFound {
			return nil, status.Error(codes.InvalidArgument, "Name Already Exists")
		}

		scopeExists := ifSubset(req.Scopes, grps[pgIdx].Scopes)
		if !scopeExists {
			return nil, status.Error(codes.InvalidArgument, "Scope Doesnt Exist")
		}

		req.FullyQualifiedName = fullName
		repoGrp := convertSrvGroupToRepoGroup(req)
		group, err := s.accountRepo.CreateGroup(ctx, userID, repoGrp)
		if err != nil {
			logger.Log.Error("service/v1 - CreateGroup - ", zap.String("reason", err.Error()), zap.String("user_id", userID))
			return nil, status.Error(codes.Unknown, "service/v1 - CreateGroup - failed to create group")
		}
		return convertRepoGroupToSrvGroup(group), nil
	default:
		logger.Log.Info("user role", zap.String("user_role", string(userClaims.Role)))
		return nil, status.Error(codes.Unknown, "unknown error")
	}
}

func (s *accountServiceServer) UpdateGroup(ctx context.Context, req *v1.UpdateGroupRequest) (*v1.Group, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	err := validateRequest(req)
	if err != nil {
		logger.Log.Error("service/v1 - UpdateGroup - validateRequest", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	switch userClaims.Role {
	case claims.RoleUser:
		return nil, status.Error(codes.PermissionDenied, "user doesnot have access to update group")
	case claims.RoleAdmin:
		return s.updateGroupName(ctx, req)
	case claims.RoleSuperAdmin:
		return s.updateGroupName(ctx, req)
	default:
		logger.Log.Info("user role", zap.String("user_role", string(userClaims.Role)))
		return nil, status.Error(codes.Unknown, "unknown error")
	}
}

func (s *accountServiceServer) DeleteGroup(ctx context.Context, req *v1.DeleteGroupRequest) (*v1.DeleteGroupResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if req.GroupId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request, GroupID can not be empty")
	}
	switch userClaims.Role {
	case claims.RoleUser:
		return nil, status.Error(codes.PermissionDenied, "user doesnot have access to delete group")
	case claims.RoleAdmin:
		return s.deleteGroup(ctx, req.GroupId)
	case claims.RoleSuperAdmin:
		return s.deleteGroup(ctx, req.GroupId)
	default:
		logger.Log.Info("user role", zap.String("user_role", string(userClaims.Role)))
		return nil, status.Error(codes.Unknown, "unknown error")
	}
}

func (s *accountServiceServer) ListChildGroups(ctx context.Context, req *v1.ListChildGroupsRequest) (*v1.ListGroupsResponse, error) {
	claims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	logger.Log.Info("user id", zap.String("user_id", claims.UserID))
	userID := claims.UserID // get userID from context
	group, err := s.accountRepo.GroupInfo(ctx, req.GroupId)
	if err != nil {
		logger.Log.Error("service/v1 - ListChildGroups - ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - ListChildGroups - failed to get Group")
	}
	groups, err := s.accountRepo.UserOwnedGroupsDirect(ctx, userID, nil)
	if err != nil {
		logger.Log.Error("service/v1 - ListChildGroups - ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - ListChildGroups - failed to get Groups")
	}
	groupValidate := false
	for j := range groups {
		if strings.HasPrefix(group.FullyQualifiedName, groups[j].FullyQualifiedName) {
			groupValidate = true
			break
		}
	}
	if !groupValidate {
		return nil, status.Error(codes.PermissionDenied, "service/v1 - ListChildGroups - user doesnot have access for the group")
	}
	grps, err := s.accountRepo.ChildGroupsDirect(ctx, req.GroupId, nil)
	if err != nil {
		logger.Log.Error("service/v1 - ListChildGroups - ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - ListChildGroups - failed to get childGroups")
	}
	return &v1.ListGroupsResponse{
		Groups: convertRepoGroupToSrvGroupAll(grps),
	}, nil

}

func (s *accountServiceServer) ListUserGroups(ctx context.Context, req *v1.ListGroupsRequest) (*v1.ListGroupsResponse, error) {
	claims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	logger.Log.Info("user id", zap.String("user_id", claims.UserID))
	userID := claims.UserID // get userID from context
	groups, err := s.accountRepo.UserOwnedGroupsDirect(ctx, userID, nil)
	if err != nil {
		logger.Log.Error("service/v1 - ListUserGroups - ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - ListUserGroups - failed to get Groups")
	}
	return &v1.ListGroupsResponse{
		Groups: convertRepoGroupToSrvGroupAll(groups),
	}, nil
}

func parentGroupIDX(parentID int64, groups []*repo.Group) int {
	for idx := range groups {
		if groups[idx].ID == parentID {
			return idx
		}
	}
	return -1
}

func ifFullQualfNameExists(name string, groups []*repo.Group) bool {
	for idx := range groups {
		if strcomp.CompareStrings(groups[idx].FullyQualifiedName, name) {
			return true
		}
	}
	return false
}

func ifSubset(scopes []string, parentScopes []string) bool {
	set := make(map[string]struct{})
	for _, value := range parentScopes {
		set[value] = struct{}{}
	}

	for _, scp := range scopes {
		if _, found := set[scp]; !found {
			return false
		}
	}

	return true
}

func convertRepoGroupToSrvGroup(grp *repo.Group) *v1.Group {
	return &v1.Group{
		ID:                 grp.ID,
		Name:               grp.Name,
		ParentId:           grp.ParentID,
		FullyQualifiedName: grp.FullyQualifiedName,
		Scopes:             grp.Scopes,
		// TODO: cover below fields in test cases
		NumOfChildGroups: grp.NumberOfGroups,
		NumOfUsers:       grp.NumberOfUsers,
	}
}

func convertSrvGroupToRepoGroup(grp *v1.Group) *repo.Group {
	return &repo.Group{
		ID:                 grp.ID,
		Name:               grp.Name,
		ParentID:           grp.ParentId,
		FullyQualifiedName: grp.FullyQualifiedName,
		Scopes:             grp.Scopes,
	}
}

func validateRequest(r *v1.UpdateGroupRequest) error {
	switch {
	case r.GroupId == 0:
		return status.Error(codes.InvalidArgument, "GroupId can not be nil")
	case r.Group == nil:
		return status.Error(codes.InvalidArgument, "Group can not be nil")
	case r.Group.Name == "":
		return status.Error(codes.InvalidArgument, "Group Name can not be empty")
	}
	return nil
}

func (s *accountServiceServer) updateGroupName(ctx context.Context, req *v1.UpdateGroupRequest) (*v1.Group, error) {
	group, err := s.accountRepo.GroupInfo(ctx, req.GroupId)
	if err != nil {
		logger.Log.Error("service/v1 - UpdateGroup - GroupInfo", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get Group")
	}
	if group.Name == req.Group.Name {
		return convertRepoGroupToSrvGroup(group), nil
	}
	fqnSlice := strings.Split(group.FullyQualifiedName, ".")
	fqnSlice = fqnSlice[:len(fqnSlice)-1]
	fqn := strings.Join(append(fqnSlice, req.Group.Name), ".")

	groupExists, err := s.accountRepo.GroupExistsByFQN(ctx, fqn)
	if err != nil {
		logger.Log.Error("service/v1 - UpdateGroup - GroupExistsByFQN", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to check GroupExistsByFQN")
	}
	if groupExists {
		return nil, status.Error(codes.InvalidArgument, "group name is not available")
	}
	if err := s.accountRepo.UpdateGroup(ctx, req.GroupId, &repo.GroupUpdate{
		Name: req.Group.Name,
	}); err != nil {
		logger.Log.Error("service/v1 - UpdateGroup - UpdateGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get update group")
	}
	group.Name = req.Group.Name
	group.FullyQualifiedName = fqn
	return convertRepoGroupToSrvGroup(group), nil
}

func (s *accountServiceServer) deleteGroup(ctx context.Context, groupID int64) (*v1.DeleteGroupResponse, error) {
	group, err := s.accountRepo.GroupInfo(ctx, groupID)
	if err != nil {
		logger.Log.Error("service/v1 - DeleteGroup - GroupInfo ", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get group")
	}
	if group.NumberOfUsers != 0 {
		return nil, status.Error(codes.PermissionDenied, "group contains users")
	}
	if group.NumberOfGroups != 0 {
		return nil, status.Error(codes.PermissionDenied, "group contains child groups")
	}
	err = s.accountRepo.DeleteGroup(ctx, groupID)
	if err != nil {
		logger.Log.Error("service/v1 - DeleteGroup - DeleteGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete Group")
	}
	return &v1.DeleteGroupResponse{
		Success: true,
	}, nil
}

func convertRepoGroupToSrvGroupAll(grps []*repo.Group) []*v1.Group {
	groups := make([]*v1.Group, len(grps))
	for i := range grps {
		groups[i] = convertRepoGroupToSrvGroup(grps[i])
	}
	return groups
}
