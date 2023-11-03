package postgres

import (
	"context"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/pkg/config"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/email"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/redis"
)

func (r *AccountRepository) GenerateMailBody(acc helper.EmailParams, ctx context.Context, cfg config.Config) (string, error) {
	return email.GenerateActivationMail(acc, ctx, cfg.Emailtemplate.Activationpath, cfg.Emailtemplate.Passwordresetpath, cfg.Emailtemplate.Redirecbaseurl)
}

func (r *AccountRepository) SetToken(ep helper.EmailParams, ctx context.Context, ttl int) error {
	return redis.SetToken(ep, ctx, r.r, ttl)
}
