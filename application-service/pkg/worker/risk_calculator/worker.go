// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package riskcalculator

import (
	"context"
	"database/sql"
	"math"
	repo "optisam-backend/application-service/pkg/repository/v1"
	"optisam-backend/application-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	pro_v1 "optisam-backend/product-service/pkg/api/v1"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//RiskCalWorker ...
type RiskCalWorker struct {
	id              string
	applicationRepo repo.Application
	productClient   pro_v1.ProductServiceClient
}

//NewWorker ...
func NewWorker(id string, grpcServers map[string]*grpc.ClientConn, applicationRepo repo.Application) *RiskCalWorker {
	return &RiskCalWorker{id: id, productClient: pro_v1.NewProductServiceClient(grpcServers["product"]), applicationRepo: applicationRepo}
}

//ID ...
func (w *RiskCalWorker) ID() string {
	return w.id
}

//DoWork ...
func (w *RiskCalWorker) DoWork(ctx context.Context, j *job.Job) error {
	apps, err := w.applicationRepo.GetApplicationsDetails(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		logger.Log.Error("worker - RiskCalculator - GetApplicationsView", zap.Error(err))
		return status.Error(codes.Internal, "DBError")
	}
	for _, app := range apps {
		domainCriticity, err := w.applicationRepo.GetDomainCriticityByDomain(ctx, db.GetDomainCriticityByDomainParams{
			Applicationdomain: app.ApplicationDomain,
			Scope:             app.Scope,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				if err := w.applicationRepo.AddApplicationbsolescenceRisk(ctx, db.AddApplicationbsolescenceRiskParams{
					Riskvalue:     sql.NullString{Valid: false},
					Applicationid: app.ApplicationID,
					Scope:         app.Scope,
				}); err != nil {
					logger.Log.Error("worker - RiskCalculator - AddApplicationbsolescenceRisk", zap.Error(err))
					return status.Error(codes.Internal, "DBError")
				}
				continue
			}
			logger.Log.Error("worker - RiskCalculator - GetDomainCriticityByDomain - application: %s", zap.String("application", app.ApplicationID), zap.Error(err))
			return status.Error(codes.Internal, "DBError")
		}
		appIns, err := w.applicationRepo.GetApplicationInstances(ctx, db.GetApplicationInstancesParams{
			ApplicationID: app.ApplicationID,
			Scope:         app.Scope,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			logger.Log.Error("worker - RiskCalculator - GetApplicationInstances", zap.Error(err))
			return status.Error(codes.Internal, "DBError")
		}
		maintenanceCriticity := db.MaintenanceLevelMetum{}
		var calMonthDiff, calMonthDiffPro int
		endDateExists := false
		for i, ins := range appIns {
			for i, pro := range ins.Products {
				acqRights, err := w.productClient.ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
					SearchParams: &pro_v1.AcqRightsSearchParams{
						SwidTag: &pro_v1.StringFilter{
							Filteringkey: pro,
							FilterType:   true,
						},
					},
					SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
					SortOrder: pro_v1.SortOrder_asc,
					Scopes:    []string{app.Scope},
					PageSize:  50,
					PageNum:   1,
				})

				if err != nil {
					if err == sql.ErrNoRows {
						continue
					}
					logger.Log.Error("worker - RiskCalculator - ListAcqRights", zap.Error(err))
					return status.Error(codes.Internal, "can not fetch list acqrights")
				}
				if acqRights == nil || len(acqRights.AcquiredRights) == 0 {

					continue
				}
				if acqRights.AcquiredRights[0].EndOfMaintenance == nil {

					continue
				}
				endDateExists = true
				calMonthAcqDiff := roundTime(time.Until(acqRights.AcquiredRights[0].EndOfMaintenance.AsTime()).Seconds() / 2600640)
				for i := range acqRights.AcquiredRights {
					monthAcqDiff := roundTime(time.Until(acqRights.AcquiredRights[i].EndOfMaintenance.AsTime()).Seconds() / 2600640)
					if monthAcqDiff <= calMonthAcqDiff {
						calMonthAcqDiff = monthAcqDiff
					}
				}
				if i == 0 {
					calMonthDiffPro = calMonthAcqDiff
				} else if calMonthAcqDiff < calMonthDiffPro {
					calMonthDiffPro = calMonthAcqDiff
				}
			}
			if i == 0 {
				calMonthDiff = calMonthDiffPro
			} else if calMonthDiffPro < calMonthDiff {
				calMonthDiff = calMonthDiffPro
			}
		}
		if !endDateExists {
			logger.Log.Info("worker - RiskCalculator - No End Date Found for application", zap.String("AppName", app.ApplicationName))
			continue
		}
		if calMonthDiff <= 0 {
			maintenanceCriticity, err = w.applicationRepo.GetMaintenanceLevelByMonthByName(ctx, "Level 4")
			if err != nil {
				logger.Log.Error("worker - RiskCalculator - GetMaintenanceLevel", zap.Error(err))
				return status.Error(codes.Internal, "DBError")
			}
		} else {
			maintenanceCriticity, err = w.applicationRepo.GetMaintenanceLevelByMonth(ctx, db.GetMaintenanceLevelByMonthParams{
				Calmonth: int32(calMonthDiff),
				Scope:    app.Scope,
			})
			if err != nil {
				logger.Log.Error("worker - RiskCalculator - GetMaintenanceLevel", zap.Error(err))
				return status.Error(codes.Internal, "DBError")
			}
		}
		obsolescenceRisk, err := w.applicationRepo.GetObsolescenceRiskForApplication(ctx, db.GetObsolescenceRiskForApplicationParams{
			Domaincriticid:     domainCriticity.DomainCriticID,
			Maintenancelevelid: maintenanceCriticity.MaintenanceLevelID,
			Scope:              app.Scope,
		})
		if err != nil && err != sql.ErrNoRows {
			logger.Log.Error("worker - RiskCalculator - GetRiskMatrixNameForApplication", zap.Error(err))
			return status.Error(codes.Internal, "DBError")
		}
		riskValue := sql.NullString{Valid: false}
		if obsolescenceRisk != "" {
			riskValue = sql.NullString{String: obsolescenceRisk, Valid: true}
		}
		if err := w.applicationRepo.AddApplicationbsolescenceRisk(ctx, db.AddApplicationbsolescenceRiskParams{
			Riskvalue:     riskValue,
			Applicationid: app.ApplicationID,
			Scope:         app.Scope,
		}); err != nil {
			logger.Log.Error("worker - RiskCalculator - AddApplicationbsolescenceRisk", zap.Error(err))
			return status.Error(codes.Internal, "DBError")
		}
	}
	logger.Log.Info("Cron Job is Successfull")

	return nil
}

func roundTime(input float64) int {
	var result float64

	if input < 0 {
		result = math.Ceil(input - 0.5)
	} else {
		result = math.Floor(input + 0.5)
	}

	// only interested in integer, ignore fractional
	i, _ := math.Modf(result)

	return int(i)
}
