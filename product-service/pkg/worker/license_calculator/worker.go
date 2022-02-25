package licensecalculator

import (
	"context"
	"database/sql"
	"encoding/json"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	l_v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/product-service/pkg/repository/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	"strings"
	"time"

	cron "github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LicenseCalWorker ...
type LicenseCalWorker struct {
	id            string
	productRepo   repo.Product
	licenseClient l_v1.LicenseServiceClient
	cronTime      string
}

// NewWorker ...
func NewWorker(id string, grpcServers map[string]*grpc.ClientConn, productRepo repo.Product, cTime string) *LicenseCalWorker {
	return &LicenseCalWorker{id: id, licenseClient: l_v1.NewLicenseServiceClient(grpcServers["license"]), productRepo: productRepo, cronTime: cTime}
}

// ID ...
func (w *LicenseCalWorker) ID() string {
	return w.id
}

// DoWork ...
func (w *LicenseCalWorker) DoWork(ctx context.Context, j *job.Job) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("Panic recovered from cron job", zap.Any("recover", r))
		}
	}()
	type temp struct {
		UpdatedBy string `json:"updatedBy"`
		Scope     string `json:"scope"`
	}
	data := temp{UpdatedBy: "default"}
	var err error
	if err = json.Unmarshal(j.Data, &data); err != nil {
		logger.Log.Error("Failed to unmarshal the dashboard update job's data", zap.Error(err))
		return status.Error(codes.Internal, "Unmarshalling error")
	}
	if data.Scope != "" { // data update initiated
		dbresp, err := w.productRepo.ListAcqrightsProductsByScope(ctx, data.Scope)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			logger.Log.Error("worker - LicenseCalculator - ListAcqrightsProductsByScope", zap.Error(err))
			return status.Error(codes.Internal, "DBError")
		}
		for _, product := range dbresp {
			if err := w.addComputedLicencesToAcqrights(ctx, product.Swidtag, product.Scope, data.UpdatedBy); err != nil {
				logger.Log.Error("worker - LicenseCalculator - ListAcqrightsProducts", zap.Error(err))
				return status.Error(codes.Internal, "DBError")
			}
		}
	} else { // cron job initiated
		products, err := w.productRepo.ListAcqrightsProducts(ctx)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			logger.Log.Error("worker - LicenseCalculator - ListAcqrightsProducts", zap.Error(err))
			return status.Error(codes.Internal, "DBError")
		}
		for _, product := range products {
			if err := w.addComputedLicencesToAcqrights(ctx, product.Swidtag, product.Scope, data.UpdatedBy); err != nil {
				logger.Log.Error("worker - LicenseCalculator - ListAcqrightsProducts", zap.Error(err))
				return status.Error(codes.Internal, "DBError")
			}
		}
	}
	return nil
}

func (w *LicenseCalWorker) addComputedLicencesToAcqrights(ctx context.Context, swidtag, scope, updatedBy string) error {
	productAcqsResp, err := w.licenseClient.ListAcqRightsForProduct(ctx, &l_v1.ListAcquiredRightsForProductRequest{
		SwidTag: swidtag,
		Scope:   scope,
	})
	if err != nil {
		logger.Log.Error("worker - LicenseCalculator - ListAcqRightsForProduct : %s", zap.String("product", swidtag), zap.Error(err))
		return nil
	}
	if productAcqsResp == nil {
		return nil
	}
	for _, productAcq := range productAcqsResp.AcqRights {
		for _, sku := range strings.Split(productAcq.SKU, ",") {
			if err = w.productRepo.AddComputedLicenses(ctx, db.AddComputedLicensesParams{
				Sku:              sku,
				Computedlicenses: productAcq.NumCptLicences,
				Computedcost:     decimal.NewFromFloat(productAcq.AvgUnitPrice * float64(productAcq.NumCptLicences)),
				Scope:            scope,
			}); err != nil {
				logger.Log.Error("worker - LicenseCalculator - AddComputedLicenses", zap.Error(err))
				return status.Error(codes.Internal, "DBError")
			}
		}
	}
	if len(productAcqsResp.AcqRights) > 0 {
		curTime := time.Now().UTC()
		parser := cron.NewParser(cron.Descriptor)
		var nextTime time.Time
		if w.cronTime == "@midnight" {
			tmp := curTime.Minute()*60 + curTime.Hour()*3600 + curTime.Second()
			tmp = 86400 - tmp
			nextTime = curTime.Add(time.Second * time.Duration(tmp))
		} else {
			var temp cron.Schedule
			temp, err = parser.Parse(w.cronTime)
			if err != nil {
				logger.Log.Error("cron parser error", zap.Error(err), zap.Any("crontime", w.cronTime))
				return err
			}
			nextTime = temp.Next(curTime)
		}
		if err = w.productRepo.UpsertDashboardUpdates(ctx, db.UpsertDashboardUpdatesParams{
			UpdatedAt:    curTime,
			NextUpdateAt: sql.NullTime{Time: nextTime, Valid: true},
			Scope:        scope,
			UpdatedBy:    updatedBy,
		}); err != nil {
			logger.Log.Error("Failed to update dashboard Info ", zap.Error(err), zap.Any("currTime", curTime), zap.Any("nextTime", nextTime))
		}
	}
	return nil
}
