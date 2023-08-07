package licensecalculator

import (
	"context"
	"database/sql"
	"encoding/json"
	a_v1 "optisam-backend/account-service/pkg/api/v1"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	l_v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/product-service/pkg/repository/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	"time"

	"github.com/robfig/cron/v3"
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
	accountClient a_v1.AccountServiceClient
	cronTime      string
}

// NewWorker ...
func NewWorker(id string, grpcServers map[string]*grpc.ClientConn, productRepo repo.Product, cTime string) *LicenseCalWorker {
	return &LicenseCalWorker{id: id, licenseClient: l_v1.NewLicenseServiceClient(grpcServers["license"]), accountClient: a_v1.NewAccountServiceClient(grpcServers["account"]), productRepo: productRepo, cronTime: cTime}
}

// ID ...
func (w *LicenseCalWorker) ID() string {
	return w.id
}

// DataUpdateWorker ...
type DataUpdateWorker struct {
	UpdatedBy string `json:"updatedBy"`
	Scope     string `json:"scope"`
}

// DoWork ...
func (w *LicenseCalWorker) DoWork(ctx context.Context, j *job.Job) error {
	logger.Log.Sugar().Infof("before sleep")
	time.Sleep(60 * time.Second)
	logger.Log.Sugar().Infof("running worker")
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("Panic recovered from cron job", zap.Any("recover", r))
		}
	}()
	var data DataUpdateWorker
	_ = json.Unmarshal(j.Data, &data)
	var err error
	if err = json.Unmarshal(j.Data, &data); err != nil {
		logger.Log.Error("Failed to unmarshal the dashboard update job's data", zap.Error(err))
		return status.Error(codes.Internal, "Unmarshalling error")
	}
	var scopes []string
	if data.Scope != "" { // data update initiated
		scopes = append(scopes, data.Scope)
	} else { // cron job initiated
		scopeList, err := w.accountClient.ListScopes(ctx, &a_v1.ListScopesRequest{})
		if err != nil {
			logger.Log.Error("worker - LicenceCalculator - ListScopes", zap.Error(err))
			return status.Error(codes.Internal, "error fetching list of scopes")
		}

		for _, s := range scopeList.Scopes {
			scopes = append(scopes, s.ScopeCode)
		}
	}
	// fmt.Println("scopes list", scopes)
	for _, scope := range scopes {
		if err := w.productRepo.DeleteOverallComputedLicensesByScope(ctx, scope); err != nil {
			logger.Log.Error("worker - licenseCalculator - DeleteOverallComputedLicensesByScope ", zap.Any("error", err), zap.Any("scope", scope))
			return status.Error(codes.Internal, "error deleting OverAllComputedLicences by scope")
		}
		editors, err := w.productRepo.ListEditorsForAggregation(ctx, []string{scope})
		if err != nil {
			logger.Log.Error("worker - LicenceCalculator - ListEditors", zap.Error(err))
			return status.Error(codes.Internal, "error fetching list of editor")
		}
		// fmt.Printf("editors list:%v for scope:%v\n", editors, scope)
		for _, editor := range editors {
			// fmt.Printf("editor:%v, scope:%v\n", editor, scope)
			resp, err := w.licenseClient.GetOverAllCompliance(ctx, &l_v1.GetOverAllComplianceRequest{
				Scope:  scope,
				Editor: editor,
			})
			if err != nil {
				logger.Log.Error("worker - licenseCalculator - GetOverAllCompliance", zap.Error(err), zap.Any("scope", scope), zap.Any("editor", editor))
				continue
				// return status.Error(codes.Internal, "error fetching GetOverAllCompliance")
			}
			if err := w.addComputedLicences(ctx, resp.AcqRights, editor, scope); err != nil {
				logger.Log.Error("worker - licenseCalculator - addComputedLicences", zap.Error(err))
				return err
			}
		}
		if len(editors) > 0 {
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
			if err := w.productRepo.UpsertDashboardUpdates(ctx, db.UpsertDashboardUpdatesParams{
				UpdatedAt:    curTime,
				NextUpdateAt: sql.NullTime{Time: nextTime, Valid: true},
				Scope:        scope,
				UpdatedBy:    data.UpdatedBy,
			}); err != nil {
				logger.Log.Error("Failed to update dashboard Info ", zap.Error(err), zap.Any("currTime", curTime), zap.Any("nextTime", nextTime))
			}
		}
	}
	return nil
}

func (w *LicenseCalWorker) addComputedLicences(ctx context.Context, resp []*l_v1.AggregationAcquiredRights, editor, scope string) error {
	for _, compLic := range resp {
		// fmt.Printf("compliance license: %v, editor: %v, scope: %v\n", compLic, editor, scope)
		if err := w.productRepo.InsertOverAllComputedLicences(ctx, db.InsertOverAllComputedLicencesParams{
			Sku:                 compLic.SKU,
			Swidtags:            compLic.SwidTags,
			Scope:               scope,
			ProductNames:        compLic.ProductNames,
			AggregationName:     compLic.AggregationName,
			Metrics:             compLic.Metric,
			NumComputedLicences: compLic.NumCptLicences,
			NumAcquiredLicences: compLic.NumAcqLicences,
			TotalCost:           decimal.NewFromFloat(helper.ToFixed(compLic.TotalCost, 2)),
			PurchaseCost:        decimal.NewFromFloat(helper.ToFixed(compLic.PurchaseCost, 2)),
			ComputedCost:        decimal.NewFromFloat(helper.ToFixed(compLic.ComputedCost, 2)),
			DeltaNumber:         compLic.DeltaNumber,
			CostOptimization:    sql.NullBool{Bool: compLic.CostOptimization, Valid: true},
			DeltaCost:           decimal.NewFromFloat(helper.ToFixed(compLic.DeltaCost, 2)),
			AvgUnitPrice:        decimal.NewFromFloat(helper.ToFixed(compLic.AvgUnitPrice, 2)),
			ComputedDetails:     compLic.ComputedDetails,
			MeticNotDefined:     sql.NullBool{Bool: compLic.MetricNotDefined, Valid: true},
			NotDeployed:         sql.NullBool{Bool: compLic.NotDeployed, Valid: true},
			Editor:              editor,
		}); err != nil {
			logger.Log.Error("worker - licenseCalculator - InsertOverAllComputedLicences ", zap.Any("error", err), zap.Any("sku", compLic.SKU), zap.Any("swidtags", compLic.SwidTags), zap.Any("scope", scope))
			return status.Error(codes.Internal, "error inserting InsertOverAllComputedLicences")
		}
	}
	return nil
}
