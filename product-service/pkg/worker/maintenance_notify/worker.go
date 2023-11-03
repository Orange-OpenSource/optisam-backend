package maintenance_notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"

	a_v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/account-service/pkg/api/v1"
	notification "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/notification-service/pkg/api/v1"

	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/config"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/worker/templates"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

	"github.com/robfig/cron/v3"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ActivationSubject = "End of maintenance Alert for Products "
)

// LicenseCalWorker ...
type MaintenanceNotifyWorker struct {
	cfg                *config.Config
	id                 string
	productRepo        repo.Product
	accountClient      a_v1.AccountServiceClient
	notificationClient notification.NotificationServiceClient
	cronTime           string
}

// NewWorker ...
func NewWorker(id string, grpcServers map[string]*grpc.ClientConn, productRepo repo.Product, cTime string, config *config.Config) *MaintenanceNotifyWorker {
	return &MaintenanceNotifyWorker{
		cfg:                config,
		id:                 id,
		notificationClient: notification.NewNotificationServiceClient(grpcServers["notification"]),
		accountClient:      a_v1.NewAccountServiceClient(grpcServers["account"]),
		productRepo:        productRepo,
		cronTime:           cTime,
	}
}

// ID ...
func (w *MaintenanceNotifyWorker) ID() string {
	return w.id
}

// DataUpdateWorker ...
type DataUpdateWorker struct {
	UpdatedBy string `json:"updatedBy"`
	Scope     string `json:"scope"`
}

// DoWork ...
func (w *MaintenanceNotifyWorker) DoWork(ctx context.Context, j *job.Job) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("Panic recovered from cron job", zap.Any("recover", r))
		}
	}()
	var data DataUpdateWorker
	var err error
	if err = json.Unmarshal(j.Data, &data); err != nil {
		logger.Log.Error("Failed to unmarshal the dashboard update job's data", zap.Error(err))
		return status.Error(codes.Internal, "Unmarshalling error")
	}
	var scopes []string
	scopeNameMap := make(map[string]string)
	if data.Scope != "" { // data update initiated
		scopes = append(scopes, data.Scope)
	} else { // cron job initiated
		scopeList, err := w.accountClient.ListScopes(ctx, &a_v1.ListScopesRequest{})
		if err != nil {
			logger.Log.Error("worker - MaintenanceNotify - ListScopes", zap.Error(err))
			return status.Error(codes.Internal, "error fetching list of scopes")
		}

		for _, s := range scopeList.Scopes {
			scopes = append(scopes, s.ScopeCode)
			scopeNameMap[s.ScopeCode] = s.ScopeName
		}
	}
	for _, scope := range scopes {
		// get admin users for this scope

		adminUsers, err := w.accountClient.GetAdminUserScope(ctx, &a_v1.GetAdminUserScopeRequest{Scopes: []string{scope}})
		if err != nil {
			logger.Log.Error("worker - MaintenanceNotify -GetAdminUser", zap.Error(err))
			return status.Error(codes.Internal, "error fetching list of adminuserforscope")
		}
		var adminEmail []string
		for _, ad := range adminUsers.AdminDetails {
			adminEmail = append(adminEmail, ad.UserName)
		}
		//expiring within month
		expiredManitenaceWithinMonth, err := w.productRepo.GetProductSkuExpiringSoonMaintenance(ctx, []string{scope})
		if err != nil {
			logger.Log.Error("worker - MaintenanceNotify - expiredManitenaceWithinMonth", zap.Error(err))
			return status.Error(codes.Internal, "error fetching expiredManitenaceWithinMonth")
		}
		if len(expiredManitenaceWithinMonth) > 0 {
			logger.Log.Sugar().Debugw("MaintenanceNotify - ExpiringSoonMaintenance - " + scope)
			//for sending email required template
			var emailData templates.Data
			emailData.Type = "expiringSoon"
			emailData.EmailTemplate = w.cfg.Emailtemplate.ExpiringSoonPath
			for _, exMain := range expiredManitenaceWithinMonth {
				var email templates.MaintennceEmailParams
				if exMain.EndOfMaintenance.Valid {
					email.EndOfMaintenance = exMain.EndOfMaintenance.Time.Format("02/01/2006")
				} else {
					email.EndOfMaintenance = ""
				}
				email.ProductName = exMain.ProductName
				email.SKU = exMain.Sku
				email.Scope = scopeNameMap[scope]
				emailData.Items = append(emailData.Items, email)

			}
			// generate email body
			emailText, err := GenerateMaintenanceMail(emailData, ctx)
			if err != nil {
				logger.Log.Sugar().Errorw("worker - MaintenanceNotify - GenerateMailBody - ExpiringSoonMaintenance - "+err.Error(),
					"status", codes.Internal,
					"reason", err.Error(),
				)
			}

			in := notification.SendMailRequest{
				MailSubject: ActivationSubject,
				MailMessage: emailText,
				MailTo:      adminEmail,
			}
			rpcres, err := w.notificationClient.SendMail(ctx, &in)
			if err != nil {
				logger.Log.Sugar().Errorw("worker - MaintenanceNotify - SendMail - ExpiringSoonMaintenance - "+err.Error(),
					"status", codes.Internal,
					"reason", err.Error(),
				)
			}

			logger.Log.Sugar().Debugw("worker - MaintenanceNotify - RPC response - ExpiringSoonMaintenance - "+rpcres.Success,
				"status", codes.OK,
				"response", rpcres,
			)
		}
		//expired maintenance
		expiredManitenace, err := w.productRepo.GetProductSkuExipredMaintenance(ctx, []string{scope})
		if err != nil {
			logger.Log.Error("worker - MaintenanceNotify - expiredMaintenance", zap.Error(err))
			return status.Error(codes.Internal, "error fetching list of expiredMaintenance")
		}
		if len(expiredManitenace) > 0 {
			logger.Log.Sugar().Debugw("MaintenanceNotify - ExpiredMaintenance - " + scope)
			//for sending email required template
			var emailDataExpiredMaintence templates.Data
			emailDataExpiredMaintence.Type = "expired"
			emailDataExpiredMaintence.EmailTemplate = w.cfg.Emailtemplate.ExpiredPath
			for _, exMain := range expiredManitenace {
				var email templates.MaintennceEmailParams
				if exMain.EndOfMaintenance.Valid {
					email.EndOfMaintenance = exMain.EndOfMaintenance.Time.Format("02/01/2006")
				} else {
					email.EndOfMaintenance = ""
				}
				email.ProductName = exMain.ProductName
				email.SKU = exMain.Sku
				email.Scope = scopeNameMap[scope]
				emailDataExpiredMaintence.Items = append(emailDataExpiredMaintence.Items, email)

			}
			// generate email body
			emailTextExpired, err := GenerateMaintenanceMail(emailDataExpiredMaintence, ctx)
			if err != nil {
				logger.Log.Sugar().Errorw("worker - MaintenanceNotify - GenerateMailBody - ExpiredMaintenance - "+err.Error(),
					"status", codes.Internal,
					"reason", err.Error(),
				)
			}

			inExpired := notification.SendMailRequest{
				MailSubject: ActivationSubject,
				MailMessage: emailTextExpired,
				MailTo:      adminEmail,
			}
			rpcres, err := w.notificationClient.SendMail(ctx, &inExpired)
			if err != nil {
				logger.Log.Sugar().Errorw("worker - MaintenanceNotify - SendMail - ExpiredMaintenance - "+err.Error(),
					"status", codes.Internal,
					"reason", err.Error(),
				)
			}
			logger.Log.Sugar().Debugw("worker - MaintenanceNotify -RPC response - ExpiredMaintenance - "+rpcres.Success,
				"status", codes.OK,
				"response", rpcres,
			)
		}
		if len(expiredManitenaceWithinMonth) > 0 || len(expiredManitenace) > 0 {
			curTime := time.Now().UTC()
			parser := cron.NewParser(cron.Descriptor)
			var nextTime time.Time
			if w.cronTime == "@midnight" {
				tmp := curTime.Minute()*60 + curTime.Hour()*3600 + curTime.Second()
				tmp = 86400 - tmp
				nextTime = curTime.Add(time.Second * time.Duration(tmp))
			} else {
				var temp cron.Schedule
				temp, err := parser.Parse(w.cronTime)
				if err != nil {
					logger.Log.Error("cron parser error", zap.Error(err), zap.Any("crontime", w.cronTime))
					return err
				}
				nextTime = temp.Next(curTime)
			}
			fmt.Println("next Time", nextTime)
		}
	}
	return nil
}

func GenerateMaintenanceMail(acc templates.Data, ctx context.Context) (msg string, err error) {
	data := templates.Data{}
	var tmpl *template.Template
	tmpl, err = template.ParseFiles(acc.EmailTemplate)
	if err != nil {
		logger.Log.Sugar().Errorw("common-mailer - GenerateActivationMail - "+err.Error(),
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return "", err
	}
	for _, acc := range acc.Items {
		var email templates.MaintennceEmailParams
		email.EndOfMaintenance = acc.EndOfMaintenance
		email.ProductName = acc.ProductName
		email.SKU = acc.SKU
		email.Scope = acc.Scope
		data.Items = append(data.Items, email)

	}

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
