// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package rest

import (
	"context"
	"optisam-backend/auth-service/pkg/api/v1"
	"optisam-backend/common/optisam/logger"

	"go.uber.org/zap"
	"gopkg.in/oauth2.v3/server"
)

type handler struct {
	service      v1.AuthService
	oauth2Server *server.Server
}

func newHandler(service v1.AuthService, srv *server.Server) *handler {
	// In PasswordCredentials framework relies on us for validating user's credential so
	// we inject our custom handler for verifying the identity of user.
	srv.SetPasswordAuthorizationHandler(func(username, password string) (string, error) {
		resp, err := service.Login(context.Background(), &v1.LoginRequest{
			Username: username,
			Password: password,
		})
		if err != nil {
			logger.Log.Error("failed to login user", zap.String("reason", err.Error()))
			return "", err
		}
		return resp.UserID, nil
	})

	return &handler{
		service:      service,
		oauth2Server: srv,
	}
}
