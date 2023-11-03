package v1

import (
	"context"
	"encoding/json"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"

	//"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/iam"

	//"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ActivationSubject      = "Welcome to OPTISAM: Activate your account to get started"
	ForgotPasswordSubject  = "Password reset requested: Follow these instructions to access your account"
	failresmap             = "failed to get respmap data"
	TopicEmailNotification = "email_notification"
)

func (s *AuthServiceServer) TokenValidation(ctx context.Context, req *v1.TokenRequest) error {
	acc := helper.EmailParams{
		TokenType: req.TokenType,
		Email:     req.Username,
		Token:     req.Token,
	}
	err := s.rep.GetToken(ctx, acc)
	return err
}

func (s *AuthServiceServer) ChangePassword(ctx context.Context, req *v1.ChangePasswordRequest) error {
	acc := helper.EmailParams{
		TokenType: req.TokenType,
		Email:     req.Username,
		Token:     req.Token,
	}
	err := s.rep.GetToken(ctx, acc)
	if err != nil {
		return err
	}
	userInfo, err := s.rep.AccountInfo(ctx, req.Username)
	if err != nil {

		logger.Log.Sugar().Errorw("service - AccountInfo", zap.Error(err))
		return status.Error(codes.Internal, "unknown error occurred")
	}

	if req.Password == userInfo.Password {
		return status.Error(codes.InvalidArgument, "old and new passwords are same")
	}
	if req.Password != req.PasswordConfirmation {
		return status.Error(codes.InvalidArgument, "password_confirmation and passwords are not same")
	}
	passValid, err := helper.ValidatePassword(req.Password)
	if !passValid {
		return err
	}
	newPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), 11)
	if err != nil {
		logger.Log.Sugar().Errorw("service -CheckPassword - GenerateFromPassword", zap.Error(err))
		return status.Error(codes.Internal, "unknown error")
	}
	if err := s.rep.ChangePassword(ctx, req.Username, string(newPass)); err != nil {
		logger.Log.Sugar().Errorw("service/v1 - ChangePassword - ChangePassword", zap.Error(err))
		return status.Error(codes.Internal, "failed to change password")
	}
	if userInfo.FirstLogin {
		if err := s.rep.ChangeUserFirstLogin(ctx, req.Username); err != nil {
			logger.Log.Sugar().Errorw("service/v1 - ChangePassword - ChangeUserFirstLogin", zap.Error(err))
			return status.Error(codes.Internal, "failed to get change user first login status")
		}
	}
	err = s.rep.DelToken(ctx, acc)
	if err != nil {
		return err
	}

	return nil
}
func (s *AuthServiceServer) ForgotPassword(ctx context.Context, email string) error {
	userInfo, err := s.rep.AccountInfo(ctx, email)
	if err != nil {
		logger.Log.Sugar().Errorw("service - AccountInfo", zap.Error(err))
		return status.Error(codes.Internal, err.Error())
	}

	if userInfo == nil {
		logger.Log.Sugar().Errorw("service - AccountInfo", "user not found")
		return status.Error(codes.InvalidArgument, "user not found")
	}
	emailParams := helper.EmailParams{
		FirstName: userInfo.FirstName,
		Email:     userInfo.UserID,
		TokenType: "resetPassword",
		Token:     helper.CreateToken(),
	}

	err = s.rep.SetToken(ctx, emailParams, s.cfg.Forgotpasstimeout)
	if err != nil {
		logger.Log.Sugar().Errorw("service/v1 CreateAccount - Set Token", zap.Error(err))
		return status.Error(codes.Internal, err.Error())
	}
	// generate email body
	emailText, err := s.rep.GenerateMailBody(ctx, emailParams, s.cfg)
	if err != nil {
		logger.Log.Sugar().Errorw("accountservice/v1 - Create Account- GenerateMailBody - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return status.Error(codes.Internal, err.Error())
	}
	in := v1.SendMailRequest{
		MailSubject: ActivationSubject,
		MailMessage: emailText,
		MailTo:      []string{email},
	}
	// verifyKey, err := iam.GetVerifyKey(s.cfg.IAM)
	// if err != nil {
	// 	logger.Log.Sugar().Errorw("auth service - forgot password - couldnt fetch verify key", zap.Any("error", err))
	// 	return status.Error(codes.Internal, err.Error())
	// }
	// apikey := s.cfg.IAM.APIKey
	// cronCtx, err := createSharedContext(s.cfg)
	// if err != nil {
	// 	logger.Log.Sugar().Errorw("auth service - forgot password - couldnt fetch token", zap.Any("error", err))
	// 	return status.Error(codes.Internal, err.Error())
	// }
	// cronAPIKeyCtx, err := grpc.AddClaimsInContext(*cronCtx, verifyKey, apikey)
	//if err == nil {
	notificationReq, _ := json.Marshal(in)
	t := TopicEmailNotification
	err = s.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &t, Partition: kafka.PartitionAny},
		Value:          []byte(notificationReq),
	}, nil)
	//rpcres, err := s.notification.SendMail(cronAPIKeyCtx, &in)
	if err != nil {
		logger.Log.Sugar().Errorw("auth service - forgot password - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return status.Error(codes.Internal, err.Error())
	} else {
		logger.Log.Sugar().Debug("successfully produced event")
		return nil
	}
	// } else {
	// 	logger.Log.Sugar().Errorw("auth service - forgot password "+err.Error(),
	// 		"status", codes.Internal,
	// 		"reason", err.Error(),
	// 	)
	// 	return status.Error(codes.Internal, err.Error())
	// }
}
