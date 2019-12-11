// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"fmt"
	v1 "optisam-backend/account-service/pkg/api/v1"
	repo "optisam-backend/account-service/pkg/repository/v1"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"unicode"

	"go.uber.org/zap"

	"optisam-backend/common/optisam/token/claims"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accountServiceServer struct {
	accountRepo repo.Account
}

// NewAccountServiceServer creates Auth service
func NewAccountServiceServer(accountRepo repo.Account) v1.AccountServiceServer {
	return &accountServiceServer{accountRepo: accountRepo}
}

func (s *accountServiceServer) UpdateAccount(ctx context.Context, req *v1.UpdateAccountRequest) (*v1.UpdateAccountResponse, error) {
	ai, err := s.accountRepo.AccountInfo(ctx, req.GetAccount().GetUserId())
	if err != nil {
		return &v1.UpdateAccountResponse{
			Success: false,
		}, status.Error(codes.Unknown, "failed to get Account info-> "+err.Error())
	}

	updated := false
	updateAcc := &repo.UpdateAccount{}
	// Populate update with the data that we read
	updateAcc.Locale = ai.Locale

	for _, path := range req.GetUpdateMask().GetPaths() {
		switch path {
		case "Locale":
			if updateAcc.Locale != req.Account.Locale {
				updateAcc.Locale = req.Account.Locale
				updated = true
			}
		}
	}

	if !updated {
		return &v1.UpdateAccountResponse{
			Success: true,
		}, nil
	}

	if err := s.accountRepo.UpdateAccount(ctx, req.GetAccount().GetUserId(), updateAcc); err != nil {
		return &v1.UpdateAccountResponse{
			Success: false,
		}, status.Error(codes.Unknown, "failed to update Account-> "+err.Error())
	}
	return &v1.UpdateAccountResponse{
		Success: true,
	}, nil
}

func (s *accountServiceServer) GetAccount(ctx context.Context, req *v1.GetAccountRequest) (*v1.GetAccountResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	ai, err := s.accountRepo.AccountInfo(ctx, userClaims.UserID)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Account info-> "+err.Error())
	}

	return &v1.GetAccountResponse{
		UserId: ai.UserId,
		Role:   v1.ROLE(ai.Role),
		Locale: ai.Locale,
	}, nil
}

func (s *accountServiceServer) CreateAccount(ctx context.Context, req *v1.Account) (*v1.Account, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}

	if userClaims.Role != claims.RoleAdmin && userClaims.Role != claims.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "only admin users can create users")
	}

	userExists, err := s.accountRepo.UserExistsByID(ctx, req.UserId)
	if err != nil {
		logger.Log.Error("service/v1 - CreateAccount - ", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot find user by ID")
	}
	if userExists {
		return nil, status.Error(codes.InvalidArgument, "username already exists")
	}

	if req.FirstName == "" {
		return nil, status.Error(codes.InvalidArgument, "first name should be non-empty")
	}
	if req.LastName == "" {
		return nil, status.Error(codes.InvalidArgument, "last name should be non-empty")
	}

	if req.Locale == "" {
		return nil, status.Error(codes.InvalidArgument, "Locale should be non-empty")
	}

	if req.Role == v1.ROLE_UNDEFINED || req.Role == v1.ROLE_SUPER_ADMIN {
		return nil, status.Error(codes.InvalidArgument, "only admin and user roles are allowed")

	}

	grps, err := s.highestAscendants(ctx, req.Groups)
	if err != nil {
		logger.Log.Error("service/v1 - CreateAccount - highestAscendants", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "cannot create account")
	}

	// assign most permissive groups to request groups
	req.Groups = grps

	_, userGroups, err := s.accountRepo.UserOwnedGroups(ctx, userClaims.UserID, nil)
	if err != nil {
		logger.Log.Error("service/v1 CreateAccount - UserOwnedGroups", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot create user account")
	}

	for _, grp := range req.Groups {
		if !groupExists(grp, userGroups) {
			return nil, status.Errorf(codes.PermissionDenied, "cannot create user account group: %d not owned by user", grp)
		}
	}

	if err := s.accountRepo.CreateAccount(ctx, serviceAccountToRepoAccount(req)); err != nil {
		logger.Log.Error("service/v1 CreateAccount - CreateAccount", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot create user account")
	}

	return req, nil
}

// GetUsers list all the users present
func (s *accountServiceServer) GetUsers(ctx context.Context, req *v1.GetUsersRequest) (*v1.ListUsersResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if (userClaims.Role != claims.RoleAdmin) && (userClaims.Role != claims.RoleSuperAdmin) {
		return nil, status.Error(codes.PermissionDenied, "user doesnot have access to fetch all users")
	}
	users, err := s.accountRepo.UsersAll(ctx)
	if err != nil {
		logger.Log.Error("service/v1 - GetGroupUsers- ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - GetGroupUsers - failed to get users")
	}
	return &v1.ListUsersResponse{
		Users: s.convertRepoUserToSrvUserAll(users),
	}, nil
}

// GetGroupUsers list all the users present in the group
func (s *accountServiceServer) GetGroupUsers(ctx context.Context, req *v1.GetGroupUsersRequest) (*v1.ListUsersResponse, error) {
	claims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	_, grps, err := s.accountRepo.UserOwnedGroups(ctx, claims.UserID, nil)
	if err != nil {
		logger.Log.Error("service/v1 - GetGroupUsers- ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - GetGroupUsers - failed to get groups")
	}
	userOwnsGroup := false
	for i := range grps {
		if grps[i].ID == req.GroupId {
			userOwnsGroup = true
			break
		}
	}
	if userOwnsGroup == false {
		return nil, status.Error(codes.Internal, "service/v1 - GetGroupUsers - user does not have access to group")
	}
	users, err := s.accountRepo.GroupUsers(ctx, req.GroupId)
	if err != nil {
		logger.Log.Error("service/v1 - GetGroupUsers- ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - GetGroupUsers - failed to get users")
	}
	return &v1.ListUsersResponse{
		Users: s.convertRepoUserToSrvUserAll(users),
	}, nil
}

// AddGroupUser adds user to the group
func (s *accountServiceServer) AddGroupUser(ctx context.Context, req *v1.AddGroupUsersRequest) (*v1.ListUsersResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}

	if userClaims.Role != claims.RoleAdmin && userClaims.Role != claims.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "user doesnot have access to add users")
	}

	isUserOwnsGroup, err := s.accountRepo.UserOwnsGroupByID(ctx, userClaims.UserID, req.GroupId)
	if err != nil {
		logger.Log.Error("service/v1 - AddGroupUser - ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - AddGroupUser - failed to get UserOwnsGroupByID")
	}

	if !isUserOwnsGroup {
		return nil, status.Error(codes.Internal, "service/v1 - AddGroupUser - user doesnt own the given group")
	}

	userIDS := []string{}
	for _, userID := range req.UserId {
		isUserOwnsGrp, err := s.accountRepo.UserOwnsGroupByID(ctx, userID, req.GroupId)
		if err != nil {
			logger.Log.Error("service/v1 - AddGroupUser - ", zap.Error(err))
			return nil, status.Error(codes.Internal, "service/v1 - AddGroupUser - failed to get UserOwnsGroupByID for user - "+userID)
		}
		if isUserOwnsGrp {
			continue
		}
		userIDS = append(userIDS, userID)
	}
	if len(userIDS) > 0 {
		if err := s.accountRepo.AddGroupUsers(ctx, req.GroupId, userIDS); err != nil {
			logger.Log.Error("service/v1 - AddGroupUser - ", zap.Error(err))
			return nil, status.Error(codes.Internal, "service/v1 - AddGroupUser - failed to add user")
		}
	}
	users, err := s.accountRepo.GroupUsers(ctx, req.GroupId)
	if err != nil {
		logger.Log.Error("service/v1 - AddGroupUser- ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - GetGroupUsers - failed to get users")
	}

	return &v1.ListUsersResponse{
		Users: s.convertRepoUserToSrvUserAll(users),
	}, nil
}

// DeleteGroupUser deletes users from the group
func (s *accountServiceServer) DeleteGroupUser(ctx context.Context, req *v1.DeleteGroupUsersRequest) (*v1.ListUsersResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}

	if userClaims.Role != claims.RoleAdmin && userClaims.Role != claims.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "user doesnot have access to delete users")
	}

	isUserOwnsGroup, err := s.accountRepo.UserOwnsGroupByID(ctx, userClaims.UserID, req.GroupId)
	if err != nil {
		logger.Log.Error("service/v1 - DeleteGroupUser - ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 -  DeleteGroupUser - failed to get UserOwnsGroupByID")
	}

	if !isUserOwnsGroup {
		return nil, status.Error(codes.Internal, "service/v1 -  DeleteGroupUser - user doesnt owns the given group")
	}

	users, err := s.accountRepo.GroupUsers(ctx, req.GroupId)
	if err != nil {
		logger.Log.Error("service/v1 - DeleteGroupUsers- ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - DeleteGroupUser - failed to get users")
	}

	admins := make(map[string]struct{})
	for _, user := range users {
		if user.Role == repo.RoleAdmin || user.Role == repo.RoleSuperAdmin {
			admins[user.UserId] = struct{}{}
		}
	}

	for _, userID := range req.UserId {
		delete(admins, userID)
		if !userExistsInGroup(userID, users) {
			return nil, status.Error(codes.Internal, "service/v1 - DeleteGroupUser - user doesnt exist in given group - "+userID)
		}
	}

	if len(admins) == 0 {

		isGroupRoot, err := s.accountRepo.IsGroupRoot(ctx, req.GroupId)
		if err != nil {
			logger.Log.Error("service/v1 - DeleteGroupUser - IsGroupRoot ", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to get IsGroupRoot info")
		}

		if isGroupRoot {
			return nil, status.Error(codes.InvalidArgument, "service/v1 - DeleteGroupUser - cannot delete all admins of root group")
		}
	}

	if len(req.UserId) > 0 {
		if err := s.accountRepo.DeleteGroupUsers(ctx, req.GroupId, req.UserId); err != nil {
			logger.Log.Error("service/v1 - AddGroupUser - ", zap.Error(err))
			return nil, status.Error(codes.Internal, "service/v1 - AddGroupUser - failed to add user")
		}
	}

	users, err = s.accountRepo.GroupUsers(ctx, req.GroupId)
	if err != nil {
		logger.Log.Error("service/v1 - AddGroupUser- ", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - GetGroupUsers - failed to get users")
	}

	return &v1.ListUsersResponse{
		Users: s.convertRepoUserToSrvUserAll(users),
	}, nil

}

//ChangePassword changes user's current password
func (s *accountServiceServer) ChangePassword(ctx context.Context, req *v1.ChangePasswordRequest) (*v1.ChangePasswordResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	passCorrect, err := s.accountRepo.CheckPassword(ctx, userClaims.UserID, req.Old)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check password")
	}
	if !passCorrect {
		return nil, status.Error(codes.Unauthenticated, "password does not exists in database")
	}
	if req.Old == req.New {
		return nil, status.Error(codes.InvalidArgument, "old and new passwords are same")
	}
	passValid, err := validatePassword(req.New)
	if !passValid {
		return nil, err
	}
	if err := s.accountRepo.ChangePassword(ctx, userClaims.UserID, req.New); err != nil {
		return nil, status.Error(codes.Internal, "failed to change password")
	}
	return &v1.ChangePasswordResponse{
		Success: true,
	}, nil
}

func groupExists(groupID int64, groups []*repo.Group) bool {
	for _, group := range groups {
		if group.ID == groupID {
			return true
		}
	}
	return false
}

func serviceAccountToRepoAccount(acc *v1.Account) *repo.AccountInfo {
	return &repo.AccountInfo{
		UserId:    acc.UserId,
		FirstName: acc.FirstName,
		LastName:  acc.LastName,
		Locale:    acc.Locale,
		Role:      repo.Role(acc.Role),
		Group:     acc.Groups,
	}
}

func (s *accountServiceServer) highestAscendants(ctx context.Context, groups []int64) ([]int64, error) {
	grps := make(map[int64]struct{})
	for _, grp := range groups {
		grps[grp] = struct{}{}
	}
	for _, grp := range groups {
		if _, ok := grps[grp]; !ok {
			// We already covered this group
			continue
		}
		childGrps, err := s.accountRepo.ChildGroupsAll(ctx, grp, &repo.GroupQueryParams{})
		if err != nil {
			return nil, err
		}
		for _, subGrp := range childGrps {
			_, ok := grps[subGrp.ID]
			if ok {
				delete(grps, subGrp.ID)
			}
		}
	}
	parentGroups := make([]int64, 0, len(grps))
	for key := range grps {
		parentGroups = append(parentGroups, key)
	}
	return parentGroups, nil
}

func (s *accountServiceServer) convertRepoUserToSrvUserAll(users []*repo.AccountInfo) []*v1.User {
	usrs := make([]*v1.User, len(users))
	for i := range users {
		usrs[i] = s.convertRepoUserToSrvUser(users[i])
	}
	return usrs
}

func (s *accountServiceServer) convertRepoUserToSrvUser(user *repo.AccountInfo) *v1.User {
	return &v1.User{
		UserId:    user.UserId,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Locale:    user.Locale,
		Role:      v1.ROLE(user.Role),
	}
}

func userExistsInGroup(userID string, users []*repo.AccountInfo) bool {
	for _, user := range users {
		if userID == user.UserId {
			return true
		}
	}
	return false
}

func validatePassword(s string) (bool, error) {
	var number, upper, lower, special bool
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsLower(c):
			lower = true
		case specialCharacter(c):
			special = true
		}
	}
	if !number {
		return false, status.Error(codes.InvalidArgument, "password must contain at least one number")
	}
	if !upper {
		return false, status.Error(codes.InvalidArgument, "password must contain at least one upper case letter")
	}
	if !lower {
		return false, status.Error(codes.InvalidArgument, "password must contain at least one lower case letter")
	}
	if !special {
		return false, status.Error(codes.InvalidArgument, "password must contain at least one special character(./@/#/$/&/*/_/,)")
	}
	return true, nil
}

func specialCharacter(c rune) bool {
	s := fmt.Sprintf("%c", c)
	specialList := []string{".", "@", "#", "$", "&", "*", "_", ","}
	for _, a := range specialList {
		if a == s {
			return true
		}
	}
	return false
}
