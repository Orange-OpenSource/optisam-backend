package v1

import (
	"context"
	"fmt"
	v1 "optisam-backend/account-service/pkg/api/v1"
	repo "optisam-backend/account-service/pkg/repository/v1"
	"optisam-backend/account-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"time"
	"unicode"

	application "optisam-backend/application-service/pkg/api/v1"
	dps "optisam-backend/dps-service/pkg/api/v1"
	equipment "optisam-backend/equipment-service/pkg/api/v1"
	metric "optisam-backend/metric-service/pkg/api/v1"
	product "optisam-backend/product-service/pkg/api/v1"
	report "optisam-backend/report-service/pkg/api/v1"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/errgroup"

	"optisam-backend/common/optisam/token/claims"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accountServiceServer struct {
	accountRepo repo.Account
	application application.ApplicationServiceClient
	product     product.ProductServiceClient
	report      report.ReportServiceClient
	metric      metric.MetricServiceClient
	equipment   equipment.EquipmentServiceClient
	dps         dps.DpsServiceClient
}

// NewAccountServiceServer creates Auth service
func NewAccountServiceServer(accountRepo repo.Account, grpcConnections map[string]*grpc.ClientConn) v1.AccountServiceServer {
	return &accountServiceServer{
		accountRepo: accountRepo,
		application: application.NewApplicationServiceClient(grpcConnections["application"]),
		product:     product.NewProductServiceClient(grpcConnections["product"]),
		metric:      metric.NewMetricServiceClient(grpcConnections["metric"]),
		dps:         dps.NewDpsServiceClient(grpcConnections["dps"]),
		report:      report.NewReportServiceClient(grpcConnections["report"]),
		equipment:   equipment.NewEquipmentServiceClient(grpcConnections["equipment"]),
	}
}

func (s *accountServiceServer) DropScopeData(ctx context.Context, req *v1.DropScopeDataRequest) (*v1.DropScopeDataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DropScopeDataResponse{Success: false}, status.Error(codes.Internal, "ClaimsValidationFailed")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return &v1.DropScopeDataResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	if userClaims.Role != claims.RoleSuperAdmin {
		return &v1.DropScopeDataResponse{Success: false}, status.Error(codes.PermissionDenied, "RoleValidationError")
	}

	var g errgroup.Group
	// Delete application data
	g.Go(func() error {
		if _, err := s.application.DropApplicationData(ctx, &application.DropApplicationDataRequest{Scope: req.Scope}); err != nil {
			logger.Log.Error("Failed to delete resources in application service", zap.Error(err))
			return err
		}
		logger.Log.Info("Application Resource deleted successfully")
		return nil
	})

	// Delete obscolenscene data
	g.Go(func() error {
		if _, err := s.application.DropObscolenscenceData(ctx, &application.DropObscolenscenceDataRequest{Scope: req.Scope}); err != nil {
			logger.Log.Error("Failed to delete resources in application service", zap.Error(err))
			return err
		}
		logger.Log.Info("obscolenscene Resource deleted successfully")
		return nil
	})

	// Delete Product & AcquiredRights resource
	g.Go(func() error {
		if _, err := s.product.DropProductData(ctx, &product.DropProductDataRequest{Scope: req.Scope, DeletionType: product.DropProductDataRequest_FULL}); err != nil {
			logger.Log.Error("Failed to delete resources in account service", zap.Error(err))
			return err
		}
		logger.Log.Info("Application Resource deleted successfully")
		return nil
	})

	// Delete Equipment & metadata
	g.Go(func() error {
		ctx1, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*300))
		defer cancel()
		if _, err := s.equipment.DropEquipmentData(ctx1, &equipment.DropEquipmentDataRequest{Scope: req.Scope}); err != nil {
			logger.Log.Error("Failed to delete resources in equipment service", zap.Error(err))
			return err
		}
		logger.Log.Info("equipment Resource deleted successfully")
		return nil
	})

	// Delete Equipment & metadata
	g.Go(func() error {
		if _, err := s.equipment.DropMetaData(ctx, &equipment.DropMetaDataRequest{Scope: req.Scope}); err != nil {
			logger.Log.Error("Failed to delete resources in equipment service", zap.Error(err))
			return err
		}
		logger.Log.Info("metadata Resource deleted successfully")
		return nil
	})

	// Delete uploaded  files records
	g.Go(func() error {
		if _, err := s.dps.DropUploadedFileData(ctx, &dps.DropUploadedFileDataRequest{Scope: req.Scope}); err != nil {
			logger.Log.Error("Failed to delete resources in dps service", zap.Error(err))
			return err
		}
		logger.Log.Info("uploaded files records deleted successfully")
		return nil
	})

	// Delete metrics
	g.Go(func() error {
		if _, err := s.metric.DropMetricData(ctx, &metric.DropMetricDataRequest{Scope: req.Scope}); err != nil {
			logger.Log.Error("Failed to delete resources in metric service", zap.Error(err))
			return err
		}
		logger.Log.Info("metric deleted successfully")
		return nil
	})

	// Delete reports
	g.Go(func() error {
		if _, err := s.report.DropReportData(ctx, &report.DropReportDataRequest{Scope: req.Scope}); err != nil {
			logger.Log.Error("Failed to delete resources in report service", zap.Error(err))
			return err
		}
		logger.Log.Info("reports deleted successfully")
		return nil
	})

	if err := g.Wait(); err != nil {
		return &v1.DropScopeDataResponse{
			Success: false,
		}, status.Error(codes.Internal, "InternalError")
	}

	err := s.accountRepo.DropScopeTX(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("Failed to delete resources in account service", zap.Error(err))
		return &v1.DropScopeDataResponse{Success: false}, err
	}

	return &v1.DropScopeDataResponse{Success: true}, nil
}

func (s *accountServiceServer) UpdateAccount(ctx context.Context, req *v1.UpdateAccountRequest) (*v1.UpdateAccountResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.UpdateAccountResponse{
			Success: false,
		}, status.Error(codes.Internal, "cannot find claims in context")
	}
	// To check if the account exists or not
	ai, err := s.accountRepo.AccountInfo(ctx, req.Account.UserId)
	if err != nil {
		if err == repo.ErrNoData {
			logger.Log.Error("service/v1 - UpdateAccount - AccountInfo", zap.Error(err))
			return &v1.UpdateAccountResponse{
				Success: false,
			}, status.Error(codes.Internal, "user does not exist")
		}
		logger.Log.Error("service/v1 - UpdateAccount - AccountInfo", zap.Error(err))
		return &v1.UpdateAccountResponse{
			Success: false,
		}, status.Error(codes.Internal, "failed to get Account info")
	}
	// When user want to update personal information
	if userClaims.UserID == req.Account.UserId {
		updateAcc := s.updateAccFieldChk(req.Account, ai)
		if error := s.accountRepo.UpdateAccount(ctx, ai.UserID, updateAcc); error != nil {
			logger.Log.Error("service/v1 - UpdateAccount - UpdateAccount", zap.Error(error))
			return &v1.UpdateAccountResponse{
				Success: false,
			}, status.Error(codes.Internal, "failed to update account")
		}
		return &v1.UpdateAccountResponse{
			Success: true,
		}, nil
	}
	// Admin and SuperAdmin can update user's role
	switch userClaims.Role {
	case claims.RoleUser:
		return &v1.UpdateAccountResponse{
			Success: false,
		}, status.Error(codes.PermissionDenied, "user does not have the access to update other users")
	// User should belong to the group owned by admin
	case claims.RoleAdmin:
		// does user belongs to groups owned by admin and their child groups
		isGroupUser, error := s.accountRepo.UserBelongsToAdminGroup(ctx, userClaims.UserID, req.Account.UserId)
		if error != nil {
			logger.Log.Error("service/v1 - UpdateAccount - UserBelongsToAdminGroup", zap.Error(error))
			return &v1.UpdateAccountResponse{
				Success: false,
			}, status.Error(codes.Internal, "failed to check if user belongs to the admin groups")
		}
		// if not then admin does not have the permission to update role of the user
		if !isGroupUser {
			return &v1.UpdateAccountResponse{
				Success: false,
			}, status.Error(codes.PermissionDenied, "user does not belong to admin's group")
		}
	}
	updateAcc, err := s.updateUserAccFieldChk(req.Account, ai)
	if err != nil {
		logger.Log.Error("service/v1 - UpdateAccount - updateUserAccFieldChk", zap.Error(err))
		return &v1.UpdateAccountResponse{
			Success: false,
		}, status.Error(codes.InvalidArgument, "failed to validate update account request")
	}
	if err := s.accountRepo.UpdateUserAccount(ctx, ai.UserID, updateAcc); err != nil {
		logger.Log.Error("service/v1 - UpdateAccount - UpdateUserAccount", zap.Error(err))
		return &v1.UpdateAccountResponse{
			Success: false,
		}, status.Error(codes.Internal, "failed to update account")
	}
	return &v1.UpdateAccountResponse{
		Success: true,
	}, nil
}

func init() {
	// admin rights are required for this function
	adminRPCMap["/v1.AccountService/DeleteAccount"] = struct{}{}
}

// DeleteAccount update an account to be inactive if
// 1) User deleting the account should be superadmin or admin - using RBAC
// 2) Account should belong to one of the group of Admin user
// 3) Account can and cannot be associated with a group
// 4) If User is associated with a group
func (s *accountServiceServer) DeleteAccount(ctx context.Context, req *v1.DeleteAccountRequest) (*v1.DeleteAccountResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DeleteAccountResponse{Success: false}, status.Error(codes.Internal, "cannot find claims in context")
	}
	// To check if the account exists or not
	ai, err := s.accountRepo.AccountInfo(ctx, req.UserId)
	if err != nil {
		if err == repo.ErrNoData {
			logger.Log.Error("service/v1 - DeleteAccount - AccountInfo", zap.Error(err))
			return &v1.DeleteAccountResponse{
				Success: false,
			}, status.Error(codes.NotFound, "user does not exist")
		}
		logger.Log.Error("service/v1 - DeleteAccount - AccountInfo", zap.Error(err))
		return &v1.DeleteAccountResponse{
			Success: false,
		}, status.Error(codes.Internal, "failed to get Account info")
	}
	// Admin can delete user belong to one of his groups
	if userClaims.Role == claims.RoleAdmin {
		// does user belongs to groups owned by admin and their child groups
		isGroupUser, err := s.accountRepo.UserBelongsToAdminGroup(ctx, userClaims.UserID, req.UserId)
		if err != nil {
			logger.Log.Error("service/v1 - DeleteAccount - UserBelongsToAdminGroup", zap.Error(err))
			return &v1.DeleteAccountResponse{
				Success: false,
			}, status.Error(codes.Internal, "failed to check if user belongs to the admin groups")
		}
		// if not then admin does not have the permission to update role of the user
		if !isGroupUser {
			return &v1.DeleteAccountResponse{
				Success: false,
			}, status.Error(codes.PermissionDenied, "user does not belong to admin's group")
		}
	}
	if err := s.accountRepo.InsertUserAudit(ctx, db.InsertUserAuditParams{
		Username:        ai.UserID,
		FirstName:       ai.FirstName,
		LastName:        ai.LastName,
		Locale:          ai.Locale,
		Role:            ai.Role.RoleToRoleString(),
		LastLogin:       ai.LastLogin,
		ContFailedLogin: ai.ContFailedLogin,
		CreatedOn:       ai.CreatedOn,
		Operation:       db.AuditStatusDELETED,
		UpdatedBy:       userClaims.UserID,
	}); err != nil {
		logger.Log.Error("service/v1 - DeleteAccount - InsertUserAudit", zap.Error(err))
		return &v1.DeleteAccountResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	if err := s.accountRepo.DeleteUser(ctx, req.UserId); err != nil {
		logger.Log.Error("service/v1 - DeleteAccount - DeleteUser", zap.Error(err))
		return &v1.DeleteAccountResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	return &v1.DeleteAccountResponse{Success: true}, nil
}

func (s *accountServiceServer) GetAccount(ctx context.Context, req *v1.GetAccountRequest) (*v1.GetAccountResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	ai, err := s.accountRepo.AccountInfo(ctx, userClaims.UserID)
	if err != nil {
		logger.Log.Error("service/v1 - GetAccount - AccountInfo", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get Account info")
	}
	return &v1.GetAccountResponse{
		UserId:     ai.UserID,
		FirstName:  ai.FirstName,
		LastName:   ai.LastName,
		Role:       v1.ROLE(ai.Role),
		Locale:     ai.Locale,
		ProfilePic: string(ai.ProfilePic),
		FirstLogin: ai.FirstLogin,
	}, nil
}

func (s *accountServiceServer) CreateAccount(ctx context.Context, req *v1.Account) (*v1.Account, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
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
		return nil, status.Error(codes.AlreadyExists, "username already exists")
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
		return nil, status.Error(codes.PermissionDenied, "only admin and user roles are allowed")

	}
	rootGroup, err := s.accountRepo.GetRootGroup(ctx)
	if err != nil {
		logger.Log.Error("service/v1 - CreateAccount - GetRootGroup", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot get root group")
	}
	if groupBelongsToRoot(rootGroup, req.Groups) {
		logger.Log.Error("service/v1 - CreateAccount - groupBelongsToRoot", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "cannot create account with root group")
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

	acc := serviceAccountToRepoAccount(req)
	acc.Password = defaultPassHash
	if err := s.accountRepo.CreateAccount(ctx, acc); err != nil {
		logger.Log.Error("service/v1 CreateAccount - CreateAccount", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot create user account")
	}

	return req, nil
}
func init() {
	// admin rights are required for this function
	adminRPCMap["/v1.AccountService/GetUsers"] = struct{}{}
}

// GetUsers list all the users present
func (s *accountServiceServer) GetUsers(ctx context.Context, req *v1.GetUsersRequest) (*v1.ListUsersResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if userClaims.Role == claims.RoleSuperAdmin || req.UserFilter.AllUsers {
		users, err := s.accountRepo.UsersAll(ctx, userClaims.UserID)
		if err != nil {
			logger.Log.Error("service/v1 - GetUsers- UsersAll", zap.Error(err))
			return nil, status.Error(codes.Internal, "service/v1 - GetUsers - failed to get all users")
		}
		return &v1.ListUsersResponse{
			Users: s.convertRepoUserToSrvUserAll(users),
		}, nil
	}
	users, err := s.accountRepo.UsersWithUserSearchParams(ctx, userClaims.UserID, &repo.UserQueryParams{})
	if err != nil {
		logger.Log.Error("service/v1 - GetUsers- UsersWithUserSearchParams", zap.Error(err))
		return nil, status.Error(codes.Internal, "service/v1 - GetUsers - failed to get users with search params")
	}
	return &v1.ListUsersResponse{
		Users: s.convertRepoUserToSrvUserAll(users),
	}, nil
}

// GetGroupUsers list all the users present in the group
func (s *accountServiceServer) GetGroupUsers(ctx context.Context, req *v1.GetGroupUsersRequest) (*v1.ListUsersResponse, error) {
	claims, ok := grpc_middleware.RetrieveClaims(ctx)
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
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
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
		isUserOwnsGrp, error := s.accountRepo.UserOwnsGroupByID(ctx, userID, req.GroupId)
		if error != nil {
			logger.Log.Error("service/v1 - AddGroupUser - ", zap.Error(error))
			return nil, status.Error(codes.Internal, "service/v1 - AddGroupUser - failed to get UserOwnsGroupByID for user - "+userID)
		}
		if isUserOwnsGrp {
			continue
		}
		userIDS = append(userIDS, userID)
	}
	if len(userIDS) > 0 {
		if error := s.accountRepo.AddGroupUsers(ctx, req.GroupId, userIDS); error != nil {
			logger.Log.Error("service/v1 - AddGroupUser - ", zap.Error(error))
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
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
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
			admins[user.UserID] = struct{}{}
		}
	}

	for _, userID := range req.UserId {
		delete(admins, userID)
		if !userExistsInGroup(userID, users) {
			return nil, status.Error(codes.Internal, "service/v1 - DeleteGroupUser - user doesnt exist in given group - "+userID)
		}
	}

	if len(admins) == 0 {

		isGroupRoot, error := s.accountRepo.IsGroupRoot(ctx, req.GroupId)
		if error != nil {
			logger.Log.Error("service/v1 - DeleteGroupUser - IsGroupRoot ", zap.Error(error))
			return nil, status.Error(codes.Internal, "failed to get IsGroupRoot info")
		}

		if isGroupRoot {
			return nil, status.Error(codes.InvalidArgument, "service/v1 - DeleteGroupUser - cannot delete all admins of root group")
		}
	}

	if len(req.UserId) > 0 {
		if error := s.accountRepo.DeleteGroupUsers(ctx, req.GroupId, req.UserId); error != nil {
			logger.Log.Error("service/v1 - AddGroupUser - ", zap.Error(error))
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

// ChangePassword changes user's current password
func (s *accountServiceServer) ChangePassword(ctx context.Context, req *v1.ChangePasswordRequest) (*v1.ChangePasswordResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	userInfo, err := s.accountRepo.AccountInfo(ctx, userClaims.UserID)
	if err != nil {
		logger.Log.Error("service - AccountInfo", zap.Error(err))
		return nil, status.Error(codes.Internal, "unknown error occurred")
	}

	if error := bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(req.Old)); error != nil {
		return nil, status.Error(codes.InvalidArgument, "Old password is wrong")

	}
	if req.Old == req.New {
		return nil, status.Error(codes.InvalidArgument, "old and new passwords are same")
	}
	passValid, err := validatePassword(req.New)
	if !passValid {
		return nil, err
	}
	newPass, err := bcrypt.GenerateFromPassword([]byte(req.New), 11)
	if err != nil {
		logger.Log.Error("service -CheckPassword - GenerateFromPassword", zap.Error(err))
		return nil, status.Error(codes.Internal, "unknown error")
	}
	if err := s.accountRepo.ChangePassword(ctx, userClaims.UserID, string(newPass)); err != nil {
		logger.Log.Error("service/v1 - ChangePassword - ChangePassword", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to change password")
	}
	if userInfo.FirstLogin == true {
		if err := s.accountRepo.ChangeUserFirstLogin(ctx, userClaims.UserID); err != nil {
			logger.Log.Error("service/v1 - ChangePassword - ChangeUserFirstLogin", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to get change user first login status")
		}
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

const defaultPassHash = "$2a$11$Lypq8GAINiClykvfHDu2QeRzl973Xx0wrnWTy1d67vetJ.WwlMsUK" // nolint: gosec

func serviceAccountToRepoAccount(acc *v1.Account) *repo.AccountInfo {
	return &repo.AccountInfo{
		UserID:    acc.UserId,
		FirstName: acc.FirstName,
		LastName:  acc.LastName,
		Locale:    acc.Locale,
		Role:      repo.Role(acc.Role),
		Group:     acc.Groups,
	}
}

func groupBelongsToRoot(rootGroup *repo.Group, groups []int64) bool {
	for _, grp := range groups {
		if rootGroup.ID == grp {
			return true
		}
	}
	return false
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
		UserId:    user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Locale:    user.Locale,
		Groups:    user.GroupName,
		Role:      v1.ROLE(user.Role),
	}
}

func userExistsInGroup(userID string, users []*repo.AccountInfo) bool {
	for _, user := range users {
		if userID == user.UserID {
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

func (s *accountServiceServer) updateAccFieldChk(reqAcc *v1.UpdateAccount, acc *repo.AccountInfo) *repo.UpdateAccount {
	updateAcc := &repo.UpdateAccount{
		FirstName: reqAcc.FirstName,
		LastName:  reqAcc.LastName,
		Locale:    reqAcc.Locale,
	}
	if reqAcc.ProfilePic == "" {
		updateAcc.ProfilePic = acc.ProfilePic
	} else {
		updateAcc.ProfilePic = []byte(reqAcc.ProfilePic)
	}
	return updateAcc
}

func (s *accountServiceServer) updateUserAccFieldChk(reqAcc *v1.UpdateAccount, acc *repo.AccountInfo) (*repo.UpdateUserAccount, error) {
	if acc.Role == repo.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "can not update role of superadmin")
	}
	updateAcc := &repo.UpdateUserAccount{}
	switch reqAcc.Role {
	case v1.ROLE_ADMIN:
		updateAcc.Role = repo.RoleAdmin
	case v1.ROLE_USER:
		updateAcc.Role = repo.RoleUser
	case v1.ROLE_SUPER_ADMIN:
		return nil, status.Error(codes.PermissionDenied, "can not update role to superadmin")
	case v1.ROLE_UNDEFINED:
		return nil, status.Error(codes.InvalidArgument, "undefined role")
	}
	return updateAcc, nil
}
