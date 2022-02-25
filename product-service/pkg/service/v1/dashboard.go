package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/workerqueue/job"
	metv1 "optisam-backend/metric-service/pkg/api/v1"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *productServiceServer) CreateDashboardUpdateJob(ctx context.Context, req *v1.CreateDashboardUpdateJobRequest) (*v1.CreateDashboardUpdateJobResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.CreateDashboardUpdateJobResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return &v1.CreateDashboardUpdateJobResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	jobID, err := s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "lcalw"},
		Status: job.JobStatusPENDING,
		Data:   json.RawMessage(fmt.Sprintf(`{"updatedBy":"data_update" , "scope" :"%s"}`, req.Scope)),
	}, "lcalw")

	if err != nil {
		logger.Log.Info("Error in push job in CreateDashboardUpdateJob", zap.Error(err), zap.Any("Scope", req.Scope))
		return &v1.CreateDashboardUpdateJobResponse{Success: false}, status.Error(codes.Internal, "PushJobFailure")
	}
	logger.Log.Info("Successfully pushed job by CreateDashboardUpdateJob", zap.Int32("jobId", jobID), zap.Any("Scope", req.Scope))
	return &v1.CreateDashboardUpdateJobResponse{Success: true}, nil
}

func (s *productServiceServer) GetBanner(ctx context.Context, req *v1.GetBannerRequest) (*v1.GetBannerResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	resp := &v1.GetBannerResponse{}
	dbresp, err := s.productRepo.GetDashboardUpdates(ctx, db.GetDashboardUpdatesParams{
		Scope:   req.GetScope(),
		Column2: req.GetTimeZone(),
	})
	if err != nil {
		logger.Log.Error("Failed to get dashboard audit info", zap.Error(err), zap.Any("Scope", req.Scope))
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.Internal, "NotFound")
		}
		return nil, status.Error(codes.Internal, "DBError")
	}
	resp.UpdatedAt, resp.NextUpdateAt = dbresp.UpdatedAt.(time.Time).Format("2006-01-02 15:04"), dbresp.NextUpdateAt.(time.Time).Format("2006-01-02 15:04")
	return resp, nil
}

func (s *productServiceServer) OverviewProductQuality(ctx context.Context, req *v1.OverviewProductQualityRequest) (*v1.OverviewProductQualityResponse, error) {
	// Finding Claims of User
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.GetProductQualityOverview(ctx, req.Scope)
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.OverviewProductQualityResponse{}, nil
		}
		logger.Log.Error("service/v1 - OverviewProductQuality - db/GetDataQaulityOverview", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	notAcqPercentage, _ := dbresp.NotDeployedPercentage.Float64()
	notDeployedPercent, _ := dbresp.NotAcquiredPercentage.Float64()
	return &v1.OverviewProductQualityResponse{
		NotAcquiredProducts:           int32(dbresp.NotAcquired),
		NotDeployedProducts:           int32(dbresp.NotDeployed),
		NotAcquiredProductsPercentage: notAcqPercentage,
		NotDeployedProductsPercentage: notDeployedPercent,
	}, nil
}

func (s *productServiceServer) DashboardOverview(ctx context.Context, req *v1.DashboardOverviewRequest) (*v1.DashboardOverviewResponse, error) {
	// Finding Claims of User
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	// Convert single scope to slice of string
	var scopes []string
	scopes = append(scopes, req.Scope)

	resp := &v1.DashboardOverviewResponse{}

	// Find Total Number of Products in the System and in this scope
	products, err := s.productRepo.ListProductsView(ctx, db.ListProductsViewParams{
		Scope:    scopes,
		PageNum:  0,
		PageSize: 1,
	})
	if err != nil {
		logger.Log.Error("service/v1 - DashboardOverview - db/ListProductsView", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	if len(products) > 0 {
		resp.NumProducts = int32(products[0].Totalrecords)
	}

	// Find Total Number of Editors in the system and in this scope
	editors, err := s.productRepo.ListEditors(ctx, scopes)
	if err != nil {
		logger.Log.Error("service/v1 - DashboardOverview - db/ListEditors", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	resp.NumEditors = int32(len(editors))

	// Get the total cost and maintenance cost
	costs, err := s.productRepo.GetAcqRightsCost(ctx, scopes)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("service/v1 - DashboardOverview - db/GetAcqRightsCost", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	if !errors.Is(err, sql.ErrNoRows) {
		resp.TotalLicenseCost, _ = costs.TotalCost.Float64()
		resp.TotalMaintenanceCost, _ = costs.TotalMaintenanceCost.Float64()
	}

	// Return Results
	return resp, nil
}

func (s *productServiceServer) ProductsPerEditor(ctx context.Context, req *v1.ProductsPerEditorRequest) (*v1.ProductsPerEditorResponse, error) {
	// Finding Claims of User
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	// Convert single scope to slice of string
	var scopes []string
	scopes = append(scopes, req.Scope)

	// Find Total Number of Editors in the system and in this scope
	editors, err := s.productRepo.ListEditors(ctx, scopes)
	if err != nil {
		logger.Log.Error("service/v1 - ProductsPerEditor - db/ListEditors", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	if len(editors) == 0 {
		return &v1.ProductsPerEditorResponse{}, nil
	}

	editorProducts := make([]*v1.EditorProducts, 0)

	// Find Number of Products per Editor and Scopes
	for _, editor := range editors {
		products, err := s.productRepo.GetProductsByEditor(ctx, db.GetProductsByEditorParams{ProductEditor: editor, Scopes: scopes})
		if err != nil {
			logger.Log.Error("service/v1 - ListEditorProducts - db/GetProductsByEditor ", zap.Error(err))
			return nil, status.Error(codes.Internal, "DBError")
		}
		editorProducts = append(editorProducts, &v1.EditorProducts{
			Editor:      editor,
			NumProducts: int32(len(products)),
		})
	}

	// Return Results
	return &v1.ProductsPerEditorResponse{
		EditorsProducts: editorProducts,
	}, nil

}

func (s *productServiceServer) ProductsPerMetricType(ctx context.Context, req *v1.ProductsPerMetricTypeRequest) (*v1.ProductsPerMetricTypeResponse, error) {
	// Finding Claims of User
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	// Convert single scope to slice of string
	var scopes []string
	scopes = append(scopes, req.Scope)

	// Find Products Per Metric
	productsPerMetric, err := s.productRepo.ProductsPerMetric(ctx, scopes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "NoDataFound")
		}
		logger.Log.Error("service/v1 - ProductsPerMetricType - db/ProductsPerMetric", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	servProductsPerMetric := dbToServProductsPerMetric(productsPerMetric)
	// Return Results
	return &v1.ProductsPerMetricTypeResponse{
		MetricsProducts: servProductsPerMetric,
	}, nil
}

func (s *productServiceServer) CounterfeitedProducts(ctx context.Context, req *v1.CounterfeitedProductsRequest) (*v1.CounterfeitedProductsResponse, error) {
	// Finding Claims of User
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	var licenses []*v1.ProductsLicenses

	// Counterfeited Product Licenses
	dbLicenses, err := s.productRepo.CounterFeitedProductsLicences(ctx, db.CounterFeitedProductsLicencesParams{
		Scope:         req.Scope,
		ProductEditor: req.Editor,
	})
	if err != nil {
		logger.Log.Error("service/v1 - CounterfeitedProducts - db/CounterFeitedProductsLicences", zap.Error(err))
		// return nil, status.Error(codes.Internal, "DBError")
	}
	if len(dbLicenses) != 0 {
		licenses = dbToServCounterfeitedProductsLicenses(dbLicenses)
	}

	var costs []*v1.ProductsCosts

	// Counterfeited Product Costs
	dbCosts, err := s.productRepo.CounterFeitedProductsCosts(ctx, db.CounterFeitedProductsCostsParams{
		Scope:         req.Scope,
		ProductEditor: req.Editor,
	})
	if err != nil {
		logger.Log.Error("service/v1 - CounterfeitedProducts - db/CounterFeitedProductsCosts", zap.Error(err))
		// return nil, status.Error(codes.Internal, "DBError")
	}
	if len(dbCosts) != 0 {
		costs = dbToServCounterfeitedProductsCosts(dbCosts)
	}

	// Return Values
	return &v1.CounterfeitedProductsResponse{
		ProductsLicenses: licenses,
		ProductsCosts:    costs,
	}, nil

}

func (s *productServiceServer) OverdeployedProducts(ctx context.Context, req *v1.OverdeployedProductsRequest) (*v1.OverdeployedProductsResponse, error) {
	// Finding Claims of User
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	var licenses []*v1.ProductsLicenses

	// OverDeployed Product Licenses
	dbLicenses, err := s.productRepo.OverDeployedProductsLicences(ctx, db.OverDeployedProductsLicencesParams{
		Scope:         req.Scope,
		ProductEditor: req.Editor,
	})
	if err != nil {
		logger.Log.Error("service/v1 - OverdeployedProducts - db/OverDeployedProductsLicences", zap.Error(err))
		// return nil, status.Error(codes.Internal, "DBError")
	}
	if len(dbLicenses) != 0 {
		licenses = dbToServOverDeployedProductsLicenses(dbLicenses)
	}

	var costs []*v1.ProductsCosts

	// OverDeployed Product Costs
	dbCosts, err := s.productRepo.OverDeployedProductsCosts(ctx, db.OverDeployedProductsCostsParams{
		Scope:         req.Scope,
		ProductEditor: req.Editor,
	})
	if err != nil {
		logger.Log.Error("service/v1 - OverdeployedProducts - db/OverDeployedProductsCosts", zap.Error(err))
		// return nil, status.Error(codes.Internal, "Internal Server Error")
	}
	if len(dbCosts) != 0 {
		costs = dbToServOverDeployedProductsCosts(dbCosts)
	}

	// Return Values
	return &v1.OverdeployedProductsResponse{
		ProductsLicenses: licenses,
		ProductsCosts:    costs,
	}, nil

}

func (s *productServiceServer) ComplianceAlert(ctx context.Context, req *v1.ComplianceAlertRequest) (*v1.ComplianceAlertResponse, error) {
	// Finding Claims of User
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	metrics, err := s.metric.ListMetrices(ctx, &metv1.ListMetricRequest{
		Scopes: []string{req.Scope},
	})
	if err != nil {
		logger.Log.Error("service/v1 - ComplianceAlert - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "MetricServiceError")
	}
	if metrics == nil || len(metrics.Metrices) == 0 {
		logger.Log.Error("service/v1 - ComplianceAlert - ListMetrices - metrics are not defined")
		return &v1.ComplianceAlertResponse{}, status.Error(codes.NotFound, "MetricsNotFound")
	}
	metricNames := getMetricNames(metrics.Metrices)
	cfRow, err := s.productRepo.CounterfeitPercent(ctx, db.CounterfeitPercentParams{
		Metrics: metricNames,
		Scope:   req.Scope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "NoDataFound")
		}
		logger.Log.Error("service/v1 - ComplianceAlert - db/CounterfeitPercent", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	// check if the acqrights are not zero
	cfAcq, _ := cfRow.Acq.Float64()
	if cfAcq == 0 {
		return &v1.ComplianceAlertResponse{}, nil
	}
	cfDeltaRights, _ := cfRow.DeltaRights.Float64()

	odRow, err := s.productRepo.OverdeployPercent(ctx, db.OverdeployPercentParams{
		Metrics: metricNames,
		Scope:   req.Scope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "NoDataFound")
		}
		logger.Log.Error("service/v1 - ComplianceAlert - db/OverdeployPercent", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	// Check if the acqrights are not zero
	odAcq, _ := odRow.Acq.Float64()
	if odAcq == 0 {
		return &v1.ComplianceAlertResponse{}, nil
	}
	odDeltaRights, _ := odRow.DeltaRights.Float64()

	cfPer := (cfDeltaRights / cfAcq) * 100
	odPer := (odDeltaRights / odAcq) * 100

	cfPercent := toFixed(cfPer, 2)
	odPercent := toFixed(odPer, 2)

	return &v1.ComplianceAlertResponse{
		CounterfeitingPercentage: cfPercent,
		OverdeploymentPercentage: odPercent,
	}, nil
}

// DashboardQuality gives number of products that are not deployed and not acquired respectively
func (s *productServiceServer) DashboardQualityProducts(ctx context.Context, req *v1.DashboardQualityProductsRequest) (*v1.DashboardQualityProductsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	productsNotDeployed, err := s.productRepo.ProductsNotDeployed(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - DashboardQuality - db/ProductsNotDeployed - error in getting count of products with no deployement", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	productsNotAcquried, err := s.productRepo.ProductsNotAcquired(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - DashboardQuality - db/ProductsNotAcquired - error in getting count of products with no license", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	return &v1.DashboardQualityProductsResponse{
		ProductsNotDeployed: dbToServProductsNotDeployed(productsNotDeployed),
		ProductsNotAcquired: dbToServProductsNotAcquired(productsNotAcquried),
	}, nil
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func dbToServOverDeployedProductsCosts(dbLic []db.OverDeployedProductsCostsRow) []*v1.ProductsCosts {
	res := make([]*v1.ProductsCosts, 0)

	for _, productCost := range dbLic {
		tpc, _ := productCost.TotalPurchaseCost.Float64()
		tcc, _ := productCost.TotalComputedCost.Float64()
		delta, _ := productCost.DeltaCost.Float64()
		res = append(res, &v1.ProductsCosts{
			SwidTag:              productCost.SwidTag,
			ProductName:          productCost.ProductName,
			LicensesAcquiredCost: tpc,
			LicensesComputedCost: tcc,
			DeltaCost:            delta,
		})
	}

	return res
}

func dbToServOverDeployedProductsLicenses(dbLic []db.OverDeployedProductsLicencesRow) []*v1.ProductsLicenses {
	res := make([]*v1.ProductsLicenses, 0)

	for _, productLic := range dbLic {
		res = append(res, &v1.ProductsLicenses{
			SwidTag:             productLic.SwidTag,
			ProductName:         productLic.ProductName,
			NumLicensesAcquired: productLic.NumLicensesAcquired,
			NumLicensesComputed: int64(productLic.NumLicencesComputed),
			Delta:               int64(productLic.Delta),
		})
	}

	return res
}

func dbToServCounterfeitedProductsCosts(dbLic []db.CounterFeitedProductsCostsRow) []*v1.ProductsCosts {
	res := make([]*v1.ProductsCosts, 0)

	for _, productCost := range dbLic {
		tpc, _ := productCost.TotalPurchaseCost.Float64()
		tcc, _ := productCost.TotalComputedCost.Float64()
		delta, _ := productCost.DeltaCost.Float64()
		res = append(res, &v1.ProductsCosts{
			SwidTag:              productCost.SwidTag,
			ProductName:          productCost.ProductName,
			LicensesAcquiredCost: tpc,
			LicensesComputedCost: tcc,
			DeltaCost:            delta,
		})
	}

	return res
}

func dbToServCounterfeitedProductsLicenses(dbLic []db.CounterFeitedProductsLicencesRow) []*v1.ProductsLicenses {
	res := make([]*v1.ProductsLicenses, 0)

	for _, productLic := range dbLic {
		res = append(res, &v1.ProductsLicenses{
			SwidTag:             productLic.SwidTag,
			ProductName:         productLic.ProductName,
			NumLicensesAcquired: productLic.NumLicensesAcquired,
			NumLicensesComputed: int64(productLic.NumLicencesComputed),
			Delta:               int64(productLic.Delta),
		})
	}

	return res
}

func dbToServProductsPerMetric(prodPerMetric []db.ProductsPerMetricRow) []*v1.MetricProducts {
	var res []*v1.MetricProducts // nolint: prealloc

	for _, p := range prodPerMetric {
		res = append(res, &v1.MetricProducts{
			MetricName:  p.Metric,
			NumProducts: int32(p.NumProducts),
		})
	}

	return res
}

func dbToServProductsNotDeployed(prodNotDeployed []db.ProductsNotDeployedRow) []*v1.DashboardQualityProducts {
	res := make([]*v1.DashboardQualityProducts, 0)
	for _, p := range prodNotDeployed {
		res = append(res, &v1.DashboardQualityProducts{
			SwidTag:     p.Swidtag,
			ProductName: p.ProductName,
			Editor:      p.ProductEditor,
			Version:     p.Version,
		})
	}
	return res
}

func dbToServProductsNotAcquired(prodNotAcquried []db.ProductsNotAcquiredRow) []*v1.DashboardQualityProducts {
	res := make([]*v1.DashboardQualityProducts, 0)
	for _, p := range prodNotAcquried {
		res = append(res, &v1.DashboardQualityProducts{
			SwidTag:     p.Swidtag,
			ProductName: p.ProductName,
			Editor:      p.ProductEditor,
			Version:     p.ProductVersion,
		})
	}
	return res
}

func getMetricNames(met []*metv1.Metric) []string {
	metNames := []string{}
	for _, m := range met {
		metNames = append(metNames, m.Name)
	}
	return metNames
}
