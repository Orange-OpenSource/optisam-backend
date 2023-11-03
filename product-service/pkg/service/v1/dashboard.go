package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	accv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/account-service/pkg/api/v1"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ProductServiceServer) CreateDashboardUpdateJob(ctx context.Context, req *v1.CreateDashboardUpdateJobRequest) (*v1.CreateDashboardUpdateJobResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.CreateDashboardUpdateJobResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Sugar().Errorf("Permission Error", "Scopes", userClaims.Socpes, "Requested Scope", req.GetScope())
		return &v1.CreateDashboardUpdateJobResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	if req.Ppid == "" {
		logger.Log.Sugar().Infof("not looking for sibling jobs as ppid blank")
	} else {
		for {
			res, err := s.ProductRepo.GetJobsInExecution(ctx, sql.NullString{String: req.Ppid, Valid: true})
			if err != nil {
				logger.Log.Sugar().Errorf("DB Error", "Scopes", userClaims.Socpes, "error", err.Error())
				return &v1.CreateDashboardUpdateJobResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
			}
			if res > 0 {
				logger.Log.Sugar().Infof("jobs still under execution", "count", res)
				time.Sleep(time.Second * 1)
			} else {
				logger.Log.Sugar().Infof("All jobs of same parent jobId are executed")
				break
			}
		}
	}

	jobID, err := s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "lcalw"},
		Status: job.JobStatusPENDING,
		Data:   json.RawMessage(fmt.Sprintf(`{"updatedBy":"data_update" , "scope" :"%s"}`, req.Scope)),
		PPID:   req.Ppid,
	}, "lcalw")

	if err != nil {
		logger.Log.Sugar().Errorf("Error in push job in CreateDashboardUpdateJob", "err", err.Error(), "Scope", req.Scope)
		return &v1.CreateDashboardUpdateJobResponse{Success: false}, status.Error(codes.Internal, "PushJobFailure")
	}
	logger.Log.Sugar().Infof("Successfully pushed job by CreateDashboardUpdateJob", "jobId", jobID, "Scope", req.Scope)
	return &v1.CreateDashboardUpdateJobResponse{Success: true}, nil
}

func (s *ProductServiceServer) GetBanner(ctx context.Context, req *v1.GetBannerRequest) (*v1.GetBannerResponse, error) {
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
	dbresp, err := s.ProductRepo.GetDashboardUpdates(ctx, db.GetDashboardUpdatesParams{
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

func (s *ProductServiceServer) GetTotalSharedAmount(ctx context.Context, req *v1.GetTotalSharedAmountRequest) (*v1.GetTotalSharedAmountResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	totalSharedAmount, totalRecievedAmount := 0.00, 0.00
	sharedData, err := s.ProductRepo.GetSharedData(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetTotalSharedAmount - db/GetSharedData", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	for _, v := range sharedData {
		acq, err := s.ProductRepo.GetUnitPriceBySku(ctx, db.GetUnitPriceBySkuParams{
			Scope: v.Scope,
			Sku:   v.Sku,
		})
		if err != nil {
			logger.Log.Error("service/v1 - GetTotalSharedAmount - db/GetAcqRightBySKU", zap.Error(err))
			return nil, status.Error(codes.Internal, "DBError")
		}
		unitPrice, _ := acq.AvgUnitPrice.Float64()
		totalSharedAmount += (float64(v.SharedLicences) * unitPrice)
		totalRecievedAmount += (float64(v.RecievedLicences) * unitPrice)
	}
	return &v1.GetTotalSharedAmountResponse{
		TotalSharedAmount:   totalSharedAmount,
		TotalRecievedAmount: totalRecievedAmount,
	}, nil
}

func (s *ProductServiceServer) OverviewProductQuality(ctx context.Context, req *v1.OverviewProductQualityRequest) (*v1.OverviewProductQualityResponse, error) {
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
	productsNotDeployed, err := s.ProductRepo.ProductsNotDeployed(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - OverviewProductQuality - db/ProductsNotDeployed - error in getting count of products with no deployement", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	productsNotAcquried, err := s.ProductRepo.ProductsNotAcquired(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - OverviewProductQuality - db/ProductsNotAcquired - error in getting count of products with no license", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	productsNotDeployedCount := len(productsNotDeployed)
	productsNotAcquriedCount := len(productsNotAcquried)
	products, err := s.ProductRepo.ListProductsView(ctx, db.ListProductsViewParams{
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

func (s *ProductServiceServer) ProductMaintenancePerc(ctx context.Context, req *v1.ProductMaintenancePercRequest) (*v1.ProductMaintenancePercResponse, error) {
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
	productMaintenance, err := s.ProductRepo.ProductMaintenanceCount(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - ProductMaintenancePerc - db/ProductMaintenanceCount - error in getting maintenance of deployed products", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	var productWithMaintenancePerc, productWithoutMainPerc float64

	if productMaintenance[0].Total == 0 {
		productWithMaintenancePerc = 0
		productWithoutMainPerc = 0
	} else {
		productWithMaintenancePerc = (float64(productMaintenance[0].NumberOfSwidtag) / float64(productMaintenance[0].Total)) * 100
		productWithoutMainPerc = 100 - productWithMaintenancePerc
	}

	return &v1.ProductMaintenancePercResponse{
		ProductWithMaintenancePercentage:    productWithMaintenancePerc,
		ProductWithoutMaintenancePercentage: productWithoutMainPerc,
	}, nil
}

func (s *ProductServiceServer) ProductNoMaintenanceDetails(ctx context.Context, req *v1.ProductNoMaintenanceDetailsRequest) (*v1.ProductNoMaintenanceDetailsResponse, error) {
	// Finding Claims of User
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Sugar().Errorf("service/v1 - GetMaintenanceBySwidtag - req scope: %s, available scopes: %v", req.Scope, userClaims.Socpes)
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	productsDeployed, err := s.ProductRepo.AllNoMaintainenceProducts(ctx, req.Scope)
	if err != nil {
		logger.Log.Sugar().Errorw("service/v1 - ProductNoMaintenanceDetails - ProductMaintenancePerc - db/AllNoMaintainenceProducts - error in getting count of products deployed",
			"scope", req.Scope,
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return nil, status.Error(codes.Internal, "DBError")
	}
	productNoMaintenance, err := s.ProductRepo.ProductNoMaintenance(ctx, db.ProductNoMaintenanceParams{
		Swidtag: productsDeployed,
		Scope:   req.Scope,
	})
	if err != nil {
		logger.Log.Sugar().Errorw("service/v1 - ProductNoMaintenanceDetails - ProductNoMaintenanceDetails - db/ProductNoMaintenance - error in getting maintenance of deployed products",
			"scope", req.Scope,
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return nil, status.Error(codes.Internal, "DBError")
	}
	ProductsVer, err := s.ProductRepo.ProductCatalogVersion(ctx, db.ProductCatalogVersionParams{
		Swidtag: productsDeployed,
		Scope:   req.Scope,
	})
	if err != nil {
		logger.Log.Sugar().Errorw("service/v1 - ProductNoMaintenanceDetails - ProductNoMaintenanceDetails - db/ProductCatalogVersion - error in getting maintenance of deployed products",
			"scope", req.Scope,
			"status", codes.Internal,
			"reason", err.Error(),
		)
		return nil, status.Error(codes.Internal, "DBError")
	}
	var ProductNoMain []*v1.ProductNoMain
	countProduct := 0
	productNameMap := make(map[string]string)
	replacedString := ""
	finalString := ""
	for i := range productNoMaintenance {
		if productNoMaintenance[i].ProductName != "" {
			replacedString = strings.ReplaceAll(productNoMaintenance[i].ProductName, " ", "_")
			finalString = replacedString + "_"
			productNameMap[finalString] = productNoMaintenance[i].ProductName
		}
		for j := range productsDeployed {
			if productsDeployed[j] == productNoMaintenance[i].Swidtag {
				countProduct++
				// if productNoMaintenance[i].ProductName == "" || !strings.Contains(productNoMaintenance[i].Swidtag, finalString) {
				// 	length := 0
				// 	for key, value := range productNameMap {
				// 		l := 0
				// 		if strings.Contains(productNoMaintenance[i].Swidtag, key) {
				// 			l = len(value)
				// 			if l > length {
				// 				productNoMaintenance[i].ProductName = value
				// 				length = l
				// 			}
				// 		}
				// 	}
				// }
				ProductNoMain = append(ProductNoMain, &v1.ProductNoMain{
					ProductName: productNoMaintenance[i].ProductName,
					Swidtag:     productNoMaintenance[i].Swidtag,
				})
			}
		}
	}

	for j := range ProductsVer {
		for i := range ProductNoMain {
			if ProductsVer[j].Swidtag == ProductNoMain[i].Swidtag {
				ProductNoMain[i].Version = ProductsVer[j].Version
			}
		}
	}

	for i := range ProductNoMain {
		length := 0
		for key, value := range productNameMap {
			l := 0
			if strings.Contains(ProductNoMain[i].Swidtag, key) {
				l = len(value)
				if l > length {
					ProductNoMain[i].ProductName = value
					length = l
				}
			}
		}
	}
	return &v1.ProductNoMaintenanceDetailsResponse{
		TotalProducts: int32(countProduct),
		ProductNoMain: ProductNoMain,
	}, nil
}

func (s *ProductServiceServer) DashboardOverview(ctx context.Context, req *v1.DashboardOverviewRequest) (*v1.DashboardOverviewResponse, error) {
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
	products, err := s.ProductRepo.ListProductsView(ctx, db.ListProductsViewParams{
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
	editors, err := s.ProductRepo.ListEditorsScope(ctx, scopes)
	if err != nil {
		logger.Log.Error("service/v1 - DashboardOverview - db/ListEditorsScope", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	resp.NumEditors = int32(len(editors))

	// Get the total cost and maintenance cost
	costs, err := s.ProductRepo.GetLicensesCost(ctx, scopes)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("service/v1 - DashboardOverview - db/GetLicensesCost", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	cfAmount, err := s.ProductRepo.GetTotalCounterfietAmount(ctx, req.Scope)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("service/v1 - DashboardOverview - db/GetTotalCounterfietAmount", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	usAmount, err := s.ProductRepo.GetTotalUnderusageAmount(ctx, req.Scope)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("service/v1 - DashboardOverview - db/GetTotalUnderusageAmount", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	TotalSum, err := s.ProductRepo.GetTotalDeltaCost(ctx, req.Scope)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Error("service/v1 - DashboardOverview - db/GetTotalDeltaCost", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	computedTotal, err := s.ProductRepo.GetComputedCost(ctx, scopes)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Log.Sugar().Debug("service/v1 - DashboardOverview - db/GetComputedCost- DBError",
			"scope", req.Scope,
			"req", req,
		)
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
		resp.ComputedMaintenance, _ = computedTotal.TotalCost.Float64()
		resp.ComputedWithoutMaintenance, _ = computedTotal.PurchaseCost.Float64()
	}

	// Return Results
	return resp, nil
}

func (s *ProductServiceServer) ProductsPercOpenClosedSource(ctx context.Context, req *v1.ProductsPercOpenClosedSourceRequest) (*v1.ProductsPercOpenClosedSourceResponse, error) {
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
	totalProducts, err := s.ProductRepo.TotalProductsOfScope(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - TotalProductsOfScope - db/productsDeployed - error in getting count of total products", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	data, err := s.ProductRepo.GetOpenSourceCloseSourceData(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetOpenSourceCloseSourceData - db/ProductMaintenanceCount - error in getting opensource and closesource data of products", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	var openSource []*v1.OpenSource
	var closeSource []*v1.CloseSource
	var totalData []*v1.TotalProductData

	if totalProducts[0] != 0 {
		openSource = append(openSource, &v1.OpenSource{
			AmountOs:     int32(data[0].Oscount),
			PrecentageOs: float64((float64(data[0].Oscount) / float64(totalProducts[0])) * 100),
		})

		closeSource = append(closeSource, &v1.CloseSource{
			AmountCs:     int32(data[0].Cscount),
			PrecentageCs: float64((float64(data[0].Cscount) / float64(totalProducts[0])) * 100),
		})
	} else {
		openSource = append(openSource, &v1.OpenSource{
			AmountOs:     int32(data[0].Oscount),
			PrecentageOs: 0,
		})

		closeSource = append(closeSource, &v1.CloseSource{
			AmountCs:     int32(data[0].Cscount),
			PrecentageCs: 0,
		})
	}

	totalData = append(totalData, &v1.TotalProductData{
		Amount:     totalProducts[0] - openSource[0].AmountOs - closeSource[0].AmountCs,
		Precentage: 100 - openSource[0].PrecentageOs - closeSource[0].PrecentageCs,
	})

	return &v1.ProductsPercOpenClosedSourceResponse{
		OpenSource:   openSource,
		ClosedSource: closeSource,
		TotalAmount:  totalData,
	}, nil
}

func (s *ProductServiceServer) ProductsPerEditor(ctx context.Context, req *v1.ProductsPerEditorRequest) (*v1.ProductsPerEditorResponse, error) {
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
	editors, err := s.ProductRepo.ListEditorsScope(ctx, scopes)
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
		products, err := s.ProductRepo.GetProductsByEditorScope(ctx, db.GetProductsByEditorScopeParams{ProductEditor: editor, Scopes: scopes})
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

func (s *ProductServiceServer) ProductsPerMetricType(ctx context.Context, req *v1.ProductsPerMetricTypeRequest) (*v1.ProductsPerMetricTypeResponse, error) {
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
	productsPerMetric, err := s.ProductRepo.ProductsPerMetric(ctx, req.Scope)
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

func (s *ProductServiceServer) CounterfeitedProducts(ctx context.Context, req *v1.CounterfeitedProductsRequest) (*v1.CounterfeitedProductsResponse, error) {
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
	dbLicenses, err := s.ProductRepo.CounterFeitedProductsLicences(ctx, db.CounterFeitedProductsLicencesParams{
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
	dbCosts, err := s.ProductRepo.CounterFeitedProductsCosts(ctx, db.CounterFeitedProductsCostsParams{
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

func (s *ProductServiceServer) SoftwareExpenditureByScope(ctx context.Context, req *v1.SoftwareExpenditureByScopeRequest) (*v1.SoftwareExpenditureByScopeResponse, error) {
	// Finding claims for user
	_, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("rest - SoftwareExpenditureByScope ", zap.String("Reason: ", "ClaimsNotFoundError"))
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}

	scopes, err := s.account.ListScopes(ctx, &accv1.ListScopesRequest{})

	if err != nil {
		logger.Log.Error("service/v1 - SoftwareExpenditureByScope - ListScopes", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "ServiceError")
	}
	m := scopeNamesWithExpense(scopes.Scopes)
	var expense []*v1.SoftwareExpensePercent
	// scpoes
	dbresp, err := s.ProductRepo.TotalCostOfEachScope(ctx, req.Scope)

	if err != nil {
		logger.Log.Error("service/v1 - SoftwareExpenditureByScope - db/TotalCostOfEachScope", zap.Error(err))
		// return nil, status.Error(codes.Internal, "DBError")
	}
	var total_map map[string]float32
	var total_cost float32

	if len(dbresp) != 0 {
		total_map, total_cost = dbToServSoftwareExpenditureByScope(dbresp)
	}
	var sum float64 = 0

	for key, cost := range total_map {
		ex := m[key]
		sum += float64(ex)
		per := (ex / float64(cost)) * 100
		expense = append(expense, &v1.SoftwareExpensePercent{
			Scope:              key,
			TotalCost:          float64(cost),
			Expenditure:        float64(ex),
			ExpenditurePercent: float64(per),
		})
	}

	// // OverDeployed Product Costs
	// dbCosts, err := s.ProductRepo.OverDeployedProductsCosts(ctx, db.OverDeployedProductsCostsParams{
	// 	Scope:  req.Scope,
	// 	Editor: req.Editor,
	// })
	// if err != nil {
	// 	logger.Log.Error("service/v1 - OverdeployedProducts - db/OverDeployedProductsCosts", zap.Error(err))
	// 	// return nil, status.Error(codes.Internal, "Internal Server Error")
	// }
	// if len(dbCosts) != 0 {
	// 	costs = dbToServOverDeployedProductsCosts(dbCosts)
	// }

	// Return Values
	return &v1.SoftwareExpenditureByScopeResponse{
		ExpensePercent:   expense,
		TotalExpenditure: sum,
		TotalCost:        float64(total_cost),
	}, nil

}

func (s *ProductServiceServer) OverdeployedProducts(ctx context.Context, req *v1.OverdeployedProductsRequest) (*v1.OverdeployedProductsResponse, error) {
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
	dbLicenses, err := s.ProductRepo.OverDeployedProductsLicences(ctx, db.OverDeployedProductsLicencesParams{
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
	dbCosts, err := s.ProductRepo.OverDeployedProductsCosts(ctx, db.OverDeployedProductsCostsParams{
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

func (s *ProductServiceServer) ComplianceAlert(ctx context.Context, req *v1.ComplianceAlertRequest) (*v1.ComplianceAlertResponse, error) {
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

	cfRow, err := s.ProductRepo.CounterfeitPercent(ctx, req.Scope)
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

	odRow, err := s.ProductRepo.OverdeployPercent(ctx, req.Scope)
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
func (s *ProductServiceServer) DashboardQualityProducts(ctx context.Context, req *v1.DashboardQualityProductsRequest) (*v1.DashboardQualityProductsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	productsNotDeployed, err := s.ProductRepo.ProductsNotDeployed(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - DashboardQuality - db/ProductsNotDeployed - error in getting count of products with no deployement", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	productsNotAcquried, err := s.ProductRepo.ProductsNotAcquired(ctx, req.Scope)
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

func scopeNamesWithExpense(acc []*accv1.Scope) map[string]float64 {
	m := make(map[string]float64)
	for _, v := range acc {
		m[v.ScopeCode] = v.Expenditure
	}
	return m
}

func dbToServSoftwareExpenditureByScope(dbLic []db.TotalCostOfEachScopeRow) (map[string]float32, float32) {
	m := make(map[string]float32)
	var sum float32 = 0
	for _, softExp := range dbLic {
		s := softExp.Scope
		t, _ := softExp.TotalCost.Float64()
		m[s] = float32(t)
		sum += float32(t)
	}

	return m, sum
}
func (s *ProductServiceServer) GetProductListByEditor(ctx context.Context, req *v1.GetProductListByEditorRequest) (*v1.GetProductListByEditorResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("rest - GetProductListByEditor ", zap.String("Reason: ", "ClaimsNotFoundError"))
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		logger.Log.Error("rest - GetProductListByEditor ", zap.String("Reason: ", "user does not have access to Group Compliance Products"))
		return nil, status.Error(codes.PermissionDenied, "user does not have access to Group Compliance Products")
	}
	dbresp, err := s.ProductRepo.GetProductListByEditor(ctx, db.GetProductListByEditorParams{
		Editor: req.Editor,
		Scope:  req.GetScopes(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetProductListByEditor - db/GetProductListByEditor", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	return &v1.GetProductListByEditorResponse{Products: dbresp}, nil
}

func (s *ProductServiceServer) GroupComplianceProduct(ctx context.Context, req *v1.GroupComplianceProductRequest) (*v1.GroupComplianceProductResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("rest - GroupComplianceProduct ", zap.String("Reason: ", "ClaimsNotFoundError"))
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		logger.Log.Error("rest - GroupComplianceProduct ", zap.String("Reason: ", "user does not have access to Group Compliance Products"))
		return nil, status.Error(codes.PermissionDenied, "user does not have access to Group Compliance Products")
	}
	licenceResp, err := s.ProductRepo.GetOverallLicencesByProduct(ctx, db.GetOverallLicencesByProductParams{
		Editor:      req.Editor,
		ProductName: req.ProductName,
		Scope:       req.Scopes,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GroupComplianceProduct - db/GetOverallLicencesByProduct", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	costResp, err := s.ProductRepo.GetOverallCostByProduct(ctx, db.GetOverallCostByProductParams{
		Editor:      req.Editor,
		ProductName: req.ProductName,
		Scope:       req.Scopes,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GroupComplianceProduct - db/GetOverallCostByProduct", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	apiresp := &v1.GroupComplianceProductResponse{}
	apiresp.Licences = make([]*v1.LicencesData, len(req.Scopes))
	for i := range req.Scopes {
		apiresp.Licences[i] = &v1.LicencesData{}
		apiresp.Licences[i].Scope = req.Scopes[i]
		for j := range licenceResp {
			if req.Scopes[i] == licenceResp[j].Scope {
				acqLic, _ := licenceResp[j].AcquiredLicences.Float64()
				apiresp.Licences[i].AcquiredLicences = int32(acqLic)
				compLic, _ := licenceResp[j].ComputedLicences.Float64()
				apiresp.Licences[i].ComputedLicences = int32(compLic)
			}
		}
	}
	apiresp.Cost = make([]*v1.CostData, len(req.Scopes))
	for i := range req.Scopes {
		apiresp.Cost[i] = &v1.CostData{}
		apiresp.Cost[i].Scope = req.Scopes[i]
		for j := range costResp {
			if req.Scopes[i] == costResp[j].Scope {
				apiresp.Cost[i].CounterfeitingCost, _ = costResp[j].CounterfeitingCost.Float64()
				apiresp.Cost[i].UnderusageCost, _ = costResp[j].UnderusageCost.Float64()
				apiresp.Cost[i].TotalCost, _ = costResp[j].TotalCost.Float64()
			}
		}
	}
	return apiresp, nil
}

// GetUnderusageLicenceByEditorProduct gives number of unused licence by editor,product & scopes
func (s *ProductServiceServer) GetUnderusageLicenceByEditorProduct(ctx context.Context, req *v1.GetUnderusageByEditorRequest) (*v1.GetUnderusageByEditorResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Sugar().Errorw("Error while getting claims",
			"status", codes.Internal,
			"message", "ClaimsNotFoundError")
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	if userClaims.Role != claims.RoleSuperAdmin {
		logger.Log.Sugar().Errorw("Error while authenticating user",
			"role", userClaims.Role,
			"status", codes.PermissionDenied,
			"message", "RoleValidationError")
		return nil, status.Error(codes.PermissionDenied, "RoleValidationError")
	}

	dbReqParams := db.ListUnderusageByEditorParams{
		Scope:          req.Scopes,
		LkEditor:       req.GetEditor() != "",
		Editor:         req.GetEditor(),
		LkProductNames: req.GetProductName() != "",
		ProductNames:   req.GetProductName(),

		ScopeAsc:        strings.Contains(req.GetSortBy(), "scope") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ScopeDesc:       strings.Contains(req.GetSortBy(), "scope") && strings.Contains(req.GetSortOrder().String(), "desc"),
		MetricsAsc:      strings.Contains(req.GetSortBy(), "metrics") && strings.Contains(req.GetSortOrder().String(), "asc"),
		MetricsDesc:     strings.Contains(req.GetSortBy(), "metrics") && strings.Contains(req.GetSortOrder().String(), "desc"),
		DeltaNumberAsc:  strings.Contains(req.GetSortBy(), "delta_number") && strings.Contains(req.GetSortOrder().String(), "asc"),
		DeltaNumberDesc: strings.Contains(req.GetSortBy(), "delta_number") && strings.Contains(req.GetSortOrder().String(), "desc"),

		// API expect pagenum from 1 but the offset in DB starts
		// PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		// PageSize: req.GetPageSize(),
	}

	listUnderUageByEditor, err := s.ProductRepo.ListUnderusageByEditor(ctx, dbReqParams)
	if err != nil {
		logger.Log.Sugar().Errorw("Error while getting underusage data - service/v1 - GetUnderusageLicenceByEditorProduct - db/listUnderUageByEditor ",
			"error", err.Error(),
			"status", codes.Internal,
			"message", "DBError")

		return nil, status.Error(codes.Internal, "DBError")
	}

	if len(listUnderUageByEditor) == 0 {
		logger.Log.Sugar().Errorw("Error while getting underusage data - service/v1 - GetUnderusageLicenceByEditorProduct - db/listUnderUageByEditor ",
			"status", codes.NotFound,
			"message", "No result found")
		return nil, status.Error(codes.NotFound, "No result found")
	}

	apiresp := v1.GetUnderusageByEditorResponse{}
	//apiresp.TotalRecords = int32(listUnderUageByEditor[0].Totalrecords)
	apiresp.UnderusageByEditorData = make([]*v1.UnderusageByEditorData, len(listUnderUageByEditor))
	for i := range listUnderUageByEditor {
		apiresp.UnderusageByEditorData[i] = &v1.UnderusageByEditorData{}
		apiresp.UnderusageByEditorData[i].Scope = listUnderUageByEditor[i].Scope
		apiresp.UnderusageByEditorData[i].Metrics = listUnderUageByEditor[i].Metrics
		delta, _ := listUnderUageByEditor[i].Delta.Float64()
		apiresp.UnderusageByEditorData[i].DeltaNumber = int64(delta)
		apiresp.UnderusageByEditorData[i].ProductName = listUnderUageByEditor[i].ProductNames
		if len(listUnderUageByEditor[i].AggregationName) > 0 {
			apiresp.UnderusageByEditorData[i].IsAggregation = true
			apiresp.UnderusageByEditorData[i].ProductName = listUnderUageByEditor[i].AggregationName
		}
	}

	return &apiresp, nil
}

// ProductLocationType -
func (s *ProductServiceServer) ProductLocationType(ctx context.Context, req *v1.GetDeploymentTypeRequest) (*v1.GetDeploymentTypeResponse, error) {
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

	saasRow, err := s.ProductRepo.DeploymentPercent(ctx, db.DeploymentPercentParams{
		Scope:       req.Scope,
		ProductType: db.ProductTypeSAAS,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "NoDataFound")
		}
		logger.Log.Error("service/v1 - ComplianceAlert - db/DeploymentPercentForSAAS", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	// if saasRow == 0 {
	// 	return &v1.GetDeploymentTypeResponse{}, nil
	// }
	onPremiseRow, err := s.ProductRepo.DeploymentPercent(ctx, db.DeploymentPercentParams{
		Scope:       req.Scope,
		ProductType: db.ProductTypeONPREMISE,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "NoDataFound")
		}
		logger.Log.Error("service/v1 - ComplianceAlert - db/DeploymentPercentForOnPremise", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	// if onPremiseRow == 0 {
	// 	return &v1.GetDeploymentTypeResponse{}, nil
	// }
	totalAcq := saasRow + onPremiseRow
	if totalAcq == 0 {
		return &v1.GetDeploymentTypeResponse{}, nil
	}
	saasPer := (saasRow / totalAcq) * 100
	onPremisePer := (onPremiseRow / totalAcq) * 100

	saasPercent := helper.ToFixed(saasPer, 2)
	onPremisePercent := helper.ToFixed(onPremisePer, 2)

	return &v1.GetDeploymentTypeResponse{
		SaasPercentage:      saasPercent,
		OnPremisePercentage: onPremisePercent,
	}, nil
}

// GetWasteUpLicences - total cost of licenses in underusage - positive integer
func (s *ProductServiceServer) GetWasteUpLicences(ctx context.Context, req *v1.GetWasteUpLicencesRequest) (*v1.GetWasteUpLicencesResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.GetWasteUpLicencesResponse{}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return &v1.GetWasteUpLicencesResponse{}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	wasteCost, _ := s.ProductRepo.WasteCost(ctx, req.Scope)
	var response v1.GetWasteUpLicencesResponse

	// Map to store the editors' waste up costs
	editorMap := make(map[string]float64)
	productsMap := make(map[string][]*v1.ProductsCost)
	// Iterate over the database response and populate the response struct
	for _, entry := range wasteCost {

		// Check if the editor already exists in the map
		editorMap[entry.Editor] += entry.Cost
		productCost := &v1.ProductsCost{
			Product:         entry.ProductNames,
			ProductCost:     entry.Cost,
			AggregationName: entry.AggregationName,
		}
		productsMap[entry.Editor] = append(productsMap[entry.Editor], productCost)
		response.TotalWasteUpCost += entry.Cost

	}
	for editorkey, editorValue := range editorMap {
		editorWasteUpCost := &v1.EditorsWasteCost{
			Editor:     editorkey,
			EditorCost: editorValue,
		}
		if productValue, exists := productsMap[editorkey]; exists {
			editorWasteUpCost.ProductsWasteUpCost = productValue
		}
		response.EditorsWasteUpCost = append(response.EditorsWasteUpCost, editorWasteUpCost)
	}

	return &response, nil
}

// GetTrueUpLicences - total cost of licenses in counterfeiting-Negative integer
func (s *ProductServiceServer) GetTrueUpLicences(ctx context.Context, req *v1.GetTrueUpLicencesRequest) (*v1.GetTrueUpLicencesResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.GetTrueUpLicencesResponse{}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return &v1.GetTrueUpLicencesResponse{}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	trueCost, _ := s.ProductRepo.TrueCost(ctx, req.Scope)
	var response v1.GetTrueUpLicencesResponse

	// Map to store the editors and products' true up costs
	editorMap := make(map[string]float64)
	productsMap := make(map[string][]*v1.ProductsCost)
	// Iterate over the database response and populate the response struct
	for _, entry := range trueCost {

		// Check if the editor already exists in the map
		editorMap[entry.Editor] += entry.Cost
		productCost := &v1.ProductsCost{
			Product:         entry.ProductNames,
			ProductCost:     entry.Cost,
			AggregationName: entry.AggregationName,
		}
		productsMap[entry.Editor] = append(productsMap[entry.Editor], productCost)
		response.TotalTrueUpCost += entry.Cost

	}
	for editorkey, editorValue := range editorMap {
		editorTrueUpCost := &v1.EditorsCost{
			Editor:     editorkey,
			EditorCost: editorValue,
		}
		if productValue, exists := productsMap[editorkey]; exists {
			editorTrueUpCost.ProductsTrueUpCost = productValue
		}
		response.EditorsTrueUpCost = append(response.EditorsTrueUpCost, editorTrueUpCost)
	}

	return &response, nil
}
