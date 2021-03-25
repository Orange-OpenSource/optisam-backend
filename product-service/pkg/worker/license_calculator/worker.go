// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package licensecalculator

import (
	"context"
	"database/sql"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	l_v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/product-service/pkg/repository/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//LicenseCalWorker ...
type LicenseCalWorker struct {
	id            string
	productRepo   repo.Product
	licenseClient l_v1.LicenseServiceClient
}

//NewWorker ...
func NewWorker(id string, grpcServers map[string]*grpc.ClientConn, productRepo repo.Product) *LicenseCalWorker {
	return &LicenseCalWorker{id: id, licenseClient: l_v1.NewLicenseServiceClient(grpcServers["license"]), productRepo: productRepo}
}

//ID ...
func (w *LicenseCalWorker) ID() string {
	return w.id
}

//DoWork ...
func (w *LicenseCalWorker) DoWork(ctx context.Context, j *job.Job) error {
	products, err := w.productRepo.ListAcqrightsProducts(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		logger.Log.Error("worker - LicenseCalculator - ListAcqrightsProducts", zap.Error(err))
		return status.Error(codes.Internal, "DBError")
	}
	for _, product := range products {
		productAcqsResp, err := w.licenseClient.ListAcqRightsForProduct(ctx, &l_v1.ListAcquiredRightsForProductRequest{
			SwidTag: product.Swidtag,
			Scope:   product.Scope,
		})
		if err != nil {
			logger.Log.Error("worker - LicenseCalculator - ListAcqRightsForProduct : %s", zap.String("product", product.Swidtag), zap.Error(err))
			continue
		}
		if productAcqsResp == nil {
			continue
		}
		for _, productAcq := range productAcqsResp.AcqRights {
			if err := w.productRepo.AddComputedLicenses(ctx, db.AddComputedLicensesParams{
				Sku:              productAcq.SKU,
				Computedlicenses: productAcq.NumCptLicences,
				Computedcost:     decimal.NewFromFloat(productAcq.AvgUnitPrice * float64(productAcq.NumCptLicences)),
				Scope:            product.Scope,
			}); err != nil {
				logger.Log.Error("worker - LicenseCalculator - AddComputedLicenses", zap.Error(err))
				return status.Error(codes.Internal, "DBError")
			}
		}
	}
	return nil
}
