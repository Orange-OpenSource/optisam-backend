package postgres

import (
	"context"
	"database/sql"
	"fmt"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/config"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/ctxmanage"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/email"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/iam"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/redis"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	selectAccountInfo = `
	SELECT
	username,
	password,
	first_name,
	last_name,
	locale,
	profile_pic,
	cont_failed_login,
	created_on,
	first_login
	FROM users
	WHERE username = $1`

	changeUserFirstLoginQuery = "UPDATE users SET first_login = FALSE, account_status = 'Active'  WHERE username = $1"
	changePasswordQuery       = "UPDATE users SET password = $2 where username =$1" // nolint: gosec
)

func (d *Default) GetToken(ctx context.Context, emaiReq helper.EmailParams) error {
	return redis.GetToken(emaiReq, ctx, d.r)
}

func (d *Default) SetToken(ctx context.Context, acc helper.EmailParams, ttl int) error {
	return redis.SetToken(acc, ctx, d.r, ttl)
}
func (d *Default) DelToken(ctx context.Context, acc helper.EmailParams) error {
	return redis.DelToken(acc, ctx, d.r)
}

// AccountInfo implements v1.Account's AccountInfo function.
func (r *Default) AccountInfo(ctx context.Context, userID string) (*v1.AccountInfo, error) {
	ai := &v1.AccountInfo{}
	err := r.db.QueryRowContext(ctx, selectAccountInfo, userID).
		Scan(&ai.UserID, &ai.Password, &ai.FirstName, &ai.LastName, &ai.Locale, &ai.ProfilePic, &ai.ContFailedLogin, &ai.CreatedOn, &ai.FirstLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Sugar().Errorw("unable to get account info ", err.Error())
			return nil, fmt.Errorf("no records found")
		}
		return nil, err
	}
	// roleUserDB, err := postgresRoleToDBRole(rUser)
	// if err != nil {
	// 	return nil, err
	// }
	// ai.Role = roleUserDB
	return ai, nil
}

// ChangeUserFirstLogin implements Account ChangeUserFirstLogin function
func (r *Default) ChangeUserFirstLogin(ctx context.Context, userID string) error {
	result, err := r.db.ExecContext(ctx, changeUserFirstLoginQuery, userID)
	if err != nil {
		logger.Log.Sugar().Errorw("repo/postgres - ChangeUserFirstLogin - failed to execute query", zap.String("reason", err.Error()))
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		logger.Log.Sugar().Errorw("repo/postgres - ChangeUserFirstLogin - failed to get number of rows affected", zap.String("reason", err.Error()))
		return err
	}
	if n != 1 {
		logger.Log.Sugar().Errorw("repo/postgres - ChangeUserFirstLogin - ", zap.String("reason", fmt.Errorf("repo/postgres - ChangeUserFirstLogin - expected one row to be affected,actual affected rows: %v", n).Error()))
		return fmt.Errorf("repo/postgres - ChangeUserFirstLogin - expected one row to be affected,actual affected rows: %v", n)
	}

	return nil
}

// ChangePassword ..
func (r *Default) ChangePassword(ctx context.Context, userID, password string) error {
	result, err := r.db.ExecContext(ctx, changePasswordQuery, userID, password)
	if err != nil {
		logger.Log.Sugar().Errorw(" ChangePassword - failed to change password", zap.String("reason", err.Error()))
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		logger.Log.Sugar().Errorw(" ChangePassword - failed to change password", zap.String("reason", err.Error()))
		return err
	}
	if n != 1 {
		logger.Log.Sugar().Errorw("repo/postgres - ChangePassword - ", zap.String("reason", fmt.Errorf("repo/postgres - ChangePassword - expected one row to be affected,actual affected rows: %v", n).Error()))
		return fmt.Errorf("repo/postgres - ChangePassword - expected one row to be affected,actual affected rows: %v", n)
	}
	return nil
}

func (r *Default) GenerateMailBody(ctx context.Context, acc helper.EmailParams, cfg config.Config) (string, error) {
	return email.GenerateActivationMail(acc, ctx, cfg.Emailtemplate.Activationpath, cfg.Emailtemplate.Passwordresetpath, cfg.Emailtemplate.Redirecbaseurl)
}

func (r *Default) CreateAuthContext(cfg config.Config) (context.Context, error) {
	verifyKey, err := iam.GetVerifyKey(cfg.IAM)
	if err != nil {
		logger.Log.Sugar().Errorw("auth service - forgot password - couldnt fetch verify key", zap.Any("error", err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	apikey := cfg.IAM.APIKey
	cronCtx, err := ctxmanage.CreateSharedContext(cfg.Application.UserNameSuperAdmin, cfg.Application.PasswordSuperAdmin, cfg.HTTPServers.Address["auth"])
	if err != nil {
		logger.Log.Sugar().Errorw("auth service - forgot password - couldnt fetch token", zap.Any("error", err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	cronAPIKeyCtx, err := grpc.AddClaimsInContext(*cronCtx, verifyKey, apikey)
	return cronAPIKeyCtx, err
}
