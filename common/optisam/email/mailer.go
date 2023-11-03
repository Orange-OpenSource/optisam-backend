package email

import (
	"bytes"
	"context"
	"text/template"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"google.golang.org/grpc/codes"
)

func GenerateActivationMail(acc helper.EmailParams, ctx context.Context, activatetmpl string, resettmpl string, path string) (msg string, err error) {
	data := helper.EmailParams{}
	var tmpl *template.Template
	if acc.TokenType == "activation" {
		data.RedirectUrl = path + "/api/v1/activate_account?user=" + acc.Email + "&token=" + acc.Token
		tmpl, err = template.ParseFiles(activatetmpl)
	} else if acc.TokenType == "resetPassword" {
		data.RedirectUrl = path + "/api/v1/reset_password?user=" + acc.Email + "&token=" + acc.Token
		tmpl, err = template.ParseFiles(resettmpl)
	}
	if err != nil {
		logger.Log.Sugar().Errorw("common-mailer - GenerateActivationMail - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return "", err
	}
	data.Email = acc.Email
	data.FirstName = acc.FirstName

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		logger.Log.Sugar().Errorw("common-mailer - GenerateActivationMail - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return "", err
	}
	return buf.String(), nil
}
