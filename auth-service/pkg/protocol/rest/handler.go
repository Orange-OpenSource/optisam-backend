package rest

import (
	"context"
	"encoding/json"
	"net/http"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/config"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
	"gopkg.in/oauth2.v3/server"
)

type handler struct {
	service      v1.AuthService
	oauth2Server *server.Server
	cfg          config.Config
}

func newHandler(service v1.AuthService, srv *server.Server, cfg config.Config) *handler {
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
		cfg:          cfg,
	}
}

func sendResponse(code int, message string, w http.ResponseWriter) {
	w.WriteHeader(code)
	var response errorresponse
	response.Message = message
	response.Code = code
	e, _ := json.Marshal(response)
	w.Write([]byte(e))
}

type errorresponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
