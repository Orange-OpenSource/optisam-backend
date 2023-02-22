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
			return nil, status.Error(codes.NotFound, "NotFound")
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
	productsNotDeployed, err := s.productRepo.ProductsNotDeployed(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - OverviewProductQuality - db/ProductsNotDeployed - error in getting count of products with no deployement", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	productsNotAcquried, err := s.productRepo.ProductsNotAcquired(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - OverviewProductQuality - db/ProductsNotAcquired - error in getting count of products with no license", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	productsNotDeployedCount := len(productsNotDeployed)
	productsNotAcquriedCount := len(productsNotAcquried)
	products, err := s.productRepo.ListProductsView(ctx, db.ListProductsViewParams{
		Scope:    []string{req.Scope},
		PageNum:  0,
		PageSize: 1,
	})
	if err != nil {
		logger.Log.Error("service/v1 - OverviewProductQuality - db/ListProductsView", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	var notAcqPercentage, notDeployedPercent float64
	if len(products) > 0 {
		numProducts := products[0].Totalrecords
		if numProducts != 0 {
			notAcqPercentage = float64(productsNotAcquriedCount*100) / float64(numProducts)
			notDeployedPercent = float64(productsNotDeployedCount*100) / float64(numProducts)
		}
	}
	return &v1.OverviewProductQualityResponse{
		NotAcquiredProducts:           int32(productsNotAcquriedCount),
		NotDeployedProducts:           int32(productsNotDeployedCount),
		NotAcquiredProductsPercentage: math.Round(notAcqPercentage*100) / 100,
		NotDeployedProductsPercentage: math.Round(notDeployedPercent*100) / 100,
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
	editors, err := s.productRepo.ListEditorsScope(ctx, scopes)
	if err != nil {
		logger.Log.Error("service/v1 - DashboardOverview - db/ListEditorsScope", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	resp.NumEditors = int32(len(editors))

	// Get the total cost and maintenance cost
	costs, err := s.productRepo.GetLicensesCost(ctx, scopes)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("service/v1 - DashboardOverview - db/GetLicensesCost", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	cfAmount, err := s.productRepo.GetTotalCounterfietAmount(ctx, req.Scope)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("service/v1 - DashboardOverview - db/GetTotalCounterfietAmount", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	usAmount, err := s.productRepo.GetTotalUnderusageAmount(ctx, req.Scope)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("service/v1 - DashboardOverview - db/GetTotalUnderusageAmount", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	TotalSum, err := s.productRepo.GetTotalDeltaCost(ctx, req.Scope)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("service/v1 - DashboardOverview - db/GetTotalDeltaCost", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	if TotalSum < 0 {
		cfAmount += TotalSum
	} else {
		usAmount += TotalSum
	}
	resp.TotalCounterfeitingAmount = cfAmount
	resp.TotalUnderusageAmount = usAmount

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
	editors, err := s.productRepo.ListEditorsScope(ctx, scopes)
	if err != nil {
		logger.Log.Error("service/v1 - ProductsPerEditor - db/ListEditorsScope", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	if len(editors) == 0 {
		return &v1.ProductsPerEditorResponse{}, nil
	}

	editorProducts := make([]*v1.EditorProducts, 0)

	// Find Number of Products per Editor and Scopes
	for _, editor := range editors {
		products, err := s.productRepo.GetProductsByEditorScope(ctx, db.GetProductsByEditorScopeParams{ProductEditor: editor, Scopes: scopes})
		if err != nil {
			logger.Log.Error("service/v1 - ListEditorProducts - db/GetProductsByEditorScope ", zap.Error(err))
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

	// Find Products Per Metric
	productsPerMetric, err := s.productRepo.ProductsPerMetric(ctx, req.Scope)
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
		Scope:  req.Scope,
		Editor: req.Editor,
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
		Scope:  req.Scope,
		Editor: req.Editor,
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
		Scope:  req.Scope,
		Editor: req.Editor,
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
		Scope:  req.Scope,
		Editor: req.Editor,
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

	cfRow, err := s.productRepo.CounterfeitPercent(ctx, req.Scope)
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

	odRow, err := s.productRepo.OverdeployPercent(ctx, req.Scope)
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

	totalAcq := cfAcq + odAcq
	cfPer := (cfDeltaRights / totalAcq) * 100
	odPer := (odDeltaRights / totalAcq) * 100

	cfPercent := helper.ToFixed(cfPer, 2)
	odPercent := helper.ToFixed(odPer, 2)

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

func dbToServOverDeployedProductsCosts(dbLic []db.OverDeployedProductsCostsRow) []*v1.ProductsCosts {
	res := make([]*v1.ProductsCosts, 0)

	for _, productCost := range dbLic {
		tpc, _ := productCost.PurchaseCost.Float64()
		tcc, _ := productCost.ComputedCost.Float64()
		delta, _ := productCost.DeltaCost.Float64()
		res = append(res, &v1.ProductsCosts{
			SwidTag:              productCost.SwidTags,
			ProductName:          productCost.ProductNames,
			AggregationName:      productCost.AggregationName,
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
		nla, _ := productLic.NumAcquiredLicences.Float64()
		nlc, _ := productLic.NumComputedLicences.Float64()
		delta, _ := productLic.Delta.Float64()
		res = append(res, &v1.ProductsLicenses{
			SwidTag:             productLic.SwidTags,
			ProductName:         productLic.ProductNames,
			AggregationName:     productLic.AggregationName,
			NumLicensesAcquired: int64(nla),
			NumLicensesComputed: int64(nlc),
			Delta:               int64(delta),
		})
	}

	return res
}

func dbToServCounterfeitedProductsCosts(dbLic []db.CounterFeitedProductsCostsRow) []*v1.ProductsCosts {
	res := make([]*v1.ProductsCosts, 0)

	for _, productCost := range dbLic {
		tpc, _ := productCost.PurchaseCost.Float64()
		tcc, _ := productCost.ComputedCost.Float64()
		delta, _ := productCost.DeltaCost.Float64()
		res = append(res, &v1.ProductsCosts{
			SwidTag:              productCost.SwidTags,
			ProductName:          productCost.ProductNames,
			AggregationName:      productCost.AggregationName,
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
		nla, _ := productLic.NumAcquiredLicences.Float64()
		nlc, _ := productLic.NumComputedLicences.Float64()
		delta, _ := productLic.Delta.Float64()
		res = append(res, &v1.ProductsLicenses{
			SwidTag:             productLic.SwidTags,
			ProductName:         productLic.ProductNames,
			AggregationName:     productLic.AggregationName,
			NumLicensesAcquired: int64(nla),
			NumLicensesComputed: int64(nlc),
			Delta:               int64(delta),
		})
	}

	return res
}

func dbToServProductsPerMetric(prodPerMetric []db.ProductsPerMetricRow) []*v1.MetricProducts {
	var res []*v1.MetricProducts // nolint: prealloc

	for _, p := range prodPerMetric {
		res = append(res, &v1.MetricProducts{
			MetricName:  p.Metric,
			NumProducts: int32(p.Composition),
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
			EditorId:    p.ID.String,
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
			EditorId:    p.ID.String,
		})
	}
	return res
}

/* func getMetricNames(met []*metv1.Metric) []string {
	metNames := []string{}
	for _, m := range met {
		metNames = append(metNames, m.Name)
	}
	return metNames
} */
