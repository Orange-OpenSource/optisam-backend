package server

import (
	"time"

	oauth2Errors "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/oauth2/errors"
	oauth2Handlers "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/oauth2/handler"

	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
)

// NewServer return a *server.Server instance configured for optisam.
func NewServer(tokenStore oauth2.TokenStore, clientStore oauth2.ClientStore, accessGen oauth2.AccessGenerate) *server.Server {
	manager := manage.NewDefaultManager()

	// Set the config for password token
	manager.SetPasswordTokenCfg(&manage.Config{AccessTokenExp: time.Hour * 2,
		// TODO: Verify this duration when we start supporting refresh grant.
		RefreshTokenExp:   time.Hour * 24 * 7,
		IsGenerateRefresh: true,
	})

	// Inject custom token store
	manager.MapTokenStorage(tokenStore)

	// Inject custom cleint store
	manager.MapClientStorage(clientStore)

	// Inject custom Access token generator
	manager.MapAccessGenerate(accessGen)

	srv := server.NewServer(server.NewConfig(), manager)

	// AllowedGrantType are only PasswordCredentials as we are supporting
	// only this grant currently
	srv.SetAllowedGrantType(oauth2.PasswordCredentials)

	srv.SetInternalErrorHandler(func(err error) *errors.Response {
		switch er := err.(type) {
		case *oauth2Errors.Error:
			return er.Response
		default:
			return nil
		}
	})

	// Set custom client info handler. We want to inject this because framework
	// will try to get the client id and secret from basic auth by default. We are not
	// supporting client id and secret currently so we need our custom handler to bypass
	// default client info handler of framework.
	srv.SetClientInfoHandler(server.ClientInfoHandler(oauth2Handlers.ClientInfoHandler))
	return srv
}
