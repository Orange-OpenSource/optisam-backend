package v1

import (
	"context"
	"errors"

	// "errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/config"

	"testing"

	repv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/repository/v1/mock"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"

	nmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	// "github.com/stretchr/testify/assert"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/status"
)

func Test_TokenValidation(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repv1.Repository
	var ctx context.Context
	tests := []struct {
		name    string
		s       *AuthServiceServer
		setup   func()
		wantErr bool
		ip      *v1.TokenRequest
	}{
		{name: "SUCCESS",
			s: &AuthServiceServer{},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockDB := mock.NewMockRepository(mockCtrl)
				rep = mockDB
				mockDB.EXPECT().GetToken(gomock.Nil(), helper.EmailParams{
					TokenType: "activation",
					Email:     "tets@test.com",
					Token:     "secret",
				}).Return(nil).Times(1)
			},
			wantErr: false,
			ip: &v1.TokenRequest{
				Username:  "tets@test.com",
				TokenType: "activation",
				Token:     "secret",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			var cfg config.Config
			tt.s = NewAuthServiceServer(rep, cfg, nil, nil)
			err := tt.s.TokenValidation(ctx, tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthServiceServer.UserClaims() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
func TestChangePassword(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := mock.NewMockRepository(mockCtrl)

	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
		ip      *v1.ChangePasswordRequest
	}{
		{
			name: "Valid Request",
			setup: func() {
				mockRepo.EXPECT().GetToken(ctx, gomock.Any()).Return(nil).Times(1)
				mockRepo.EXPECT().AccountInfo(ctx, "test@example.com").Return(&v1.AccountInfo{
					UserID:     "test@example.com",
					Password:   "oldPassword",
					FirstLogin: true,
				}, nil).Times(1)
				mockRepo.EXPECT().ChangePassword(ctx, "test@example.com", gomock.Any()).Return(nil).Times(1)
				mockRepo.EXPECT().ChangeUserFirstLogin(ctx, "test@example.com").Return(nil).Times(1)
				mockRepo.EXPECT().DelToken(ctx, gomock.Any()).Return(nil).Times(1)
			},
			wantErr: false,
			ip: &v1.ChangePasswordRequest{
				TokenType:            "access",
				Username:             "test@example.com",
				Token:                "validToken",
				Password:             "1@MewPassword",
				PasswordConfirmation: "1@MewPassword",
			},
		},
		{
			name: "Invalid Token",
			setup: func() {
				mockRepo.EXPECT().GetToken(ctx, gomock.Any()).Return(errors.New("token error")).Times(1)
			},
			wantErr: true,
			ip: &v1.ChangePasswordRequest{
				TokenType:            "access",
				Username:             "test@example.com",
				Token:                "invalidToken",
				Password:             "newPassword",
				PasswordConfirmation: "newPassword",
			},
		},
		{
			name: "Same Passwords",
			setup: func() {
				mockRepo.EXPECT().GetToken(ctx, gomock.Any()).Return(nil).Times(1)
				mockRepo.EXPECT().AccountInfo(ctx, "test@example.com").Return(&v1.AccountInfo{
					UserID:     "test@example.com",
					Password:   "oldPassword",
					FirstLogin: true,
				}, nil).Times(1)
			},
			wantErr: true,
			ip: &v1.ChangePasswordRequest{
				TokenType:            "access",
				Username:             "test@example.com",
				Token:                "validToken",
				Password:             "oldPassword",
				PasswordConfirmation: "oldPassword",
			},
		},
		{
			name: "Invalid Confirmation",
			setup: func() {
				mockRepo.EXPECT().GetToken(ctx, gomock.Any()).Return(nil).Times(1)
				mockRepo.EXPECT().AccountInfo(ctx, "test@example.com").Return(&v1.AccountInfo{
					UserID:     "test@example.com",
					Password:   "oldPassword",
					FirstLogin: true,
				}, nil).Times(1)
			},
			wantErr: true,
			ip: &v1.ChangePasswordRequest{
				TokenType:            "access",
				Username:             "test@example.com",
				Token:                "validToken",
				Password:             "newPassword",
				PasswordConfirmation: "differentConfirmation",
			},
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &AuthServiceServer{
				rep: mockRepo,
			}
			tt.setup()
			err := s.ChangePassword(ctx, tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestForgotPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockRepository(ctrl)

	ctrl1 := gomock.NewController(t)
	defer ctrl1.Finish()

	mockNotification := nmock.NewMockNotificationServiceClient(ctrl)
	authService := &AuthServiceServer{
		rep:          mockRepo,
		cfg:          config.Config{},
		notification: mockNotification,
	}

	ctx := context.Background()
	email := "test@example.com"
	firstName := "John"
	userID := "12345"

	t.Run("AccountInfo error", func(t *testing.T) {
		mockRepo.EXPECT().AccountInfo(ctx, email).Return(nil, errors.New("database error"))

		err := authService.ForgotPassword(ctx, email)

		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Equal(t, "database error", status.Convert(err).Message())
	})

	t.Run("User not found", func(t *testing.T) {
		mockRepo.EXPECT().AccountInfo(ctx, email).Return(nil, nil)

		err := authService.ForgotPassword(ctx, email)

		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Equal(t, "user not found", status.Convert(err).Message())
	})

	t.Run("SetToken error", func(t *testing.T) {
		mockRepo.EXPECT().AccountInfo(ctx, email).Return(&v1.AccountInfo{
			FirstName: firstName,
			UserID:    userID,
		}, nil)
		mockRepo.EXPECT().SetToken(ctx, gomock.Any(), gomock.Any()).Return(errors.New("token error"))

		err := authService.ForgotPassword(ctx, email)

		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Equal(t, "token error", status.Convert(err).Message())
	})

	t.Run("GenerateMailBody error", func(t *testing.T) {
		mockRepo.EXPECT().AccountInfo(ctx, email).Return(&v1.AccountInfo{
			FirstName: firstName,
			UserID:    userID,
		}, nil)
		mockRepo.EXPECT().SetToken(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockRepo.EXPECT().GenerateMailBody(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("mail body error"))

		err := authService.ForgotPassword(ctx, email)

		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Equal(t, "mail body error", status.Convert(err).Message())
	})

	t.Run("SendMail error", func(t *testing.T) {
		mockRepo.EXPECT().AccountInfo(ctx, email).Return(&v1.AccountInfo{
			FirstName: firstName,
			UserID:    userID,
		}, nil)
		mockRepo.EXPECT().SetToken(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockRepo.EXPECT().GenerateMailBody(ctx, gomock.Any(), gomock.Any()).Return("email body", nil)
		mockRepo.EXPECT().CreateAuthContext(gomock.Any()).Return(context.Background(), nil)
		mockNotification.EXPECT().SendMail(gomock.Any(), gomock.Any()).Return(nil, errors.New("send mail error"))

		err := authService.ForgotPassword(ctx, email)

		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Equal(t, "send mail error", status.Convert(err).Message())
	})
	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().AccountInfo(ctx, email).Return(&v1.AccountInfo{
			FirstName: firstName,
			UserID:    userID,
		}, nil)
		mockRepo.EXPECT().SetToken(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockRepo.EXPECT().GenerateMailBody(ctx, gomock.Any(), gomock.Any()).Return("email body", nil)
		mockRepo.EXPECT().CreateAuthContext(gomock.Any()).Return(context.Background(), nil)
		mockNotification.EXPECT().SendMail(gomock.Any(), gomock.Any()).Return(&v1.SendMailResponse{Success: "true"}, nil)
		err := authService.ForgotPassword(ctx, email)
		assert.NoError(t, err)
	})
	t.Run("Auth ctx err error", func(t *testing.T) {
		mockRepo.EXPECT().AccountInfo(ctx, email).Return(&v1.AccountInfo{
			FirstName: firstName,
			UserID:    userID,
		}, nil)
		mockRepo.EXPECT().SetToken(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockRepo.EXPECT().GenerateMailBody(ctx, gomock.Any(), gomock.Any()).Return("email body", nil)
		mockRepo.EXPECT().CreateAuthContext(gomock.Any()).Return(context.Background(), errors.New("error"))
		mockNotification.EXPECT().SendMail(ctx, gomock.Any()).Return(&v1.SendMailResponse{Success: "false"}, errors.New("err")).AnyTimes()

		err := authService.ForgotPassword(ctx, email)

		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Equal(t, "error", status.Convert(err).Message())
	})

}
