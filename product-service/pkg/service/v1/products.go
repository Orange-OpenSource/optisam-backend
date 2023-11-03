package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	appv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/application-service/pkg/api/v1"

	accv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/account-service/pkg/api/v1"

	metv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	dgo "github.com/dgraph-io/dgo/v2"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
	dgworker "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/worker/dgraph"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// productServiceServer is implementation of v1.authServiceServer proto interface
type ProductServiceServer struct {
	ProductRepo           repo.Product
	queue                 workerqueue.Workerqueue
	metric                metv1.MetricServiceClient
	application           appv1.ApplicationServiceClient
	account               accv1.AccountServiceClient
	dashboardTimeLocation string
	KafkaProducer         *kafka.Producer
	Dg                    *dgo.Dgraph
	Cfg                   *config.Config
}

// NewProductServiceServer creates Product service
func NewProductServiceServer(ProductRepo repo.Product, queue workerqueue.Workerqueue, grpcServers map[string]*grpc.ClientConn, zone string, kafkaProducer *kafka.Producer, dgraph *dgo.Dgraph, cfg *config.Config) *ProductServiceServer {
	return &ProductServiceServer{
		ProductRepo:           ProductRepo,
		queue:                 queue,
		metric:                metv1.NewMetricServiceClient(grpcServers["metric"]),
		application:           appv1.NewApplicationServiceClient(grpcServers["application"]),
		account:               accv1.NewAccountServiceClient(grpcServers["account"]),
		dashboardTimeLocation: zone,
		KafkaProducer:         kafkaProducer,
		Dg:                    dgraph,
		Cfg:                   cfg,
	}
}

func (s *ProductServiceServer) GetEditorProductExpensesByScope(ctx context.Context, req *v1.EditorProductsExpensesByScopeRequest) (*v1.EditorProductExpensesByScopeResponse, error) {
	logger.Log.Sugar().Debug("service/v1 - GetEditorProductExpensesByScope - GetEditorProductExpensesByScope",
		"scope", req.Scope,
		"req", req,
	)
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Sugar().Errorw("product-v1 - GetEditorProductExpensesByScope - wrong userClaims",
			"status", codes.Unknown,
			"reason", "ClaimsNotFoundError",
		)
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Sugar().Errorw("product-v1 - GetEditorProductExpensesByScope - ScopeValidationError",
			"status", codes.Internal,
			"reason", "ScopeValidationError",
		)
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	editor := req.Editor
	if editor == "" {
		logger.Log.Sugar().Errorw("product-v1 - GetEditorProductExpensesByScope - Editor Not Found",
			"status", codes.Internal,
			"reason", "Editor Not Found",
		)
		return nil, status.Error(codes.PermissionDenied, "Editor Not Found")
	}
	dbresp, err := s.ProductRepo.GetEditorProductExpensesByScopeData(ctx, db.GetEditorProductExpensesByScopeDataParams{
		Scope:     []string{req.Scope},
		Reqeditor: editor,
	})
	if err != nil {
		logger.Log.Sugar().Errorw("service/v1 - GetEditorProductExpensesByScope - GetEditorProductExpensesByScope",
			"error", err.Error(),
			"scope", req.Scope,
			"status", codes.Internal,
		)
		return nil, status.Error(codes.Internal, "DBError")
	}
	dbrespocl, err := s.ProductRepo.GetComputedCostEditorProducts(ctx, db.GetComputedCostEditorProductsParams{
		Scope:  []string{req.Scope},
		Editor: editor,
	})
	if err != nil {
		logger.Log.Sugar().Errorw("service/v1 - GetEditorProductExpensesByScope - GetEditorProductExpensesByScope",
			"error", err.Error(),
			"scope", req.Scope,
			"status", codes.Internal,
		)
		return nil, status.Error(codes.Internal, "DBError")
	}

	apiresp := v1.EditorProductExpensesByScopeResponse{}
	apiresp.EditorProductExpensesByScope = make([]*v1.EditorProductExpensesByScopeData, len(dbresp))
	for i := range dbresp {
		apiresp.EditorProductExpensesByScope[i] = &v1.EditorProductExpensesByScopeData{}
		apiresp.EditorProductExpensesByScope[i].Name = dbresp[i].Name
		apiresp.EditorProductExpensesByScope[i].TotalPurchaseCost = dbresp[i].TotalPurchaseCost
		apiresp.EditorProductExpensesByScope[i].TotalMaintenanceCost = dbresp[i].TotalMaintenanceCost
		apiresp.EditorProductExpensesByScope[i].TotalCost = dbresp[i].TotalCost
	}
	for _, oclrow := range dbrespocl {
		for i, apires := range apiresp.EditorProductExpensesByScope {
			if oclrow.ProductNames == apires.Name || apires.Name == oclrow.AggregationName {
				apiresp.EditorProductExpensesByScope[i].TotalComputedCost = oclrow.Cost
			}
		}
	}
	return &apiresp, nil
}

func (s *ProductServiceServer) UpsertAllocatedMetricEquipment(ctx context.Context, req *v1.UpsertAllocateMetricEquipementRequest) (*v1.UpsertAllocateMetricEquipementResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	err := s.ProductRepo.UpsertProductEquipments(ctx, db.UpsertProductEquipmentsParams{
		Swidtag: req.Swidtag, EquipmentID: req.EquipmentId, NumOfUsers: sql.NullInt32{Int32: req.AllocatedUsers,
			Valid: true}, Scope: req.Scope,
		AllocatedMetric: req.AllocatedMetrics,
	})

	if err != nil {
		logger.Log.Error("UpsertProductUpsertAllocatedMetricEquipment UpsertProductEquipments Failed", zap.Error(err))
		return &v1.UpsertAllocateMetricEquipementResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}

	return &v1.UpsertAllocateMetricEquipementResponse{Success: true}, nil

}

func (s *ProductServiceServer) DeleteAllocatedMetricEquipment(ctx context.Context, req *v1.DropAllocateMetricEquipementRequest) (*v1.UpsertAllocateMetricEquipementResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	err := s.ProductRepo.DropAllocatedMetricFromEquipment(ctx, db.DropAllocatedMetricFromEquipmentParams{
		Swidtag:         req.Swidtag,
		EquipmentID:     req.EquipmentId,
		Scope:           req.Scope,
		AllocatedMetric: req.AllocatedMetrics,
	})

	if err != nil {
		logger.Log.Error("DeleteAllocatedMetricEquipment DeleteAllocatedMetricEquipment Failed", zap.Error(err))
		return &v1.UpsertAllocateMetricEquipementResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}

	return &v1.UpsertAllocateMetricEquipementResponse{Success: true}, nil

}

func (s *ProductServiceServer) UpsertProduct(ctx context.Context, req *v1.UpsertProductRequest) (*v1.UpsertProductResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	err := s.ProductRepo.UpsertProductTx(ctx, req, userClaims.UserID)
	if err != nil {
		logger.Log.Error("UpsertProduct Failed", zap.Error(err))
		return &v1.UpsertProductResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}

	// For Worker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.UpsertProductRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	_, err = s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
		PPID:   req.Ppid,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	return &v1.UpsertProductResponse{Success: true}, nil
}

func (s *ProductServiceServer) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	var apiresp *v1.ListProductsResponse
	var err error
	// nolint: gocritic
	if req.GetSearchParams().GetApplicationId().GetFilteringkey() != "" {
		apiresp, err = s.listProductViewInApplication(ctx, req, req.Scopes)
	} else if req.GetSearchParams().GetEquipmentId().GetFilteringkey() != "" {
		apiresp, err = s.listProductViewInEquipment(ctx, req, req.Scopes)
	} else {
		apiresp, err = s.listProductView(ctx, req, req.Scopes)
	}
	if err != nil {
		return nil, err
	}
	return apiresp, nil
}

// nolint: gocyclo
func (s *ProductServiceServer) listProductView(ctx context.Context, req *v1.ListProductsRequest, scopes []string) (*v1.ListProductsResponse, error) {
	dbresp, err := s.ProductRepo.ListProductsView(ctx, db.ListProductsViewParams{
		Scope:                 scopes,
		Swidtag:               req.GetSearchParams().GetSwidTag().GetFilteringkey(),
		IsSwidtag:             req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		LkSwidtag:             !req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		ProductName:           req.GetSearchParams().GetName().GetFilteringkey(),
		IsProductName:         req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkProductName:         !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		ProductEditor:         req.GetSearchParams().GetEditor().GetFilteringkey(),
		IsProductEditor:       req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		LkProductEditor:       !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		ProductNameAsc:        strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductNameDesc:       strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		SwidtagAsc:            strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "asc"),
		SwidtagDesc:           strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductVersionAsc:     strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductVersionDesc:    strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditionAsc:     strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditionDesc:    strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductCategoryAsc:    strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductCategoryDesc:   strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:      strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc:     strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfApplicationsAsc:  strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfApplicationsDesc: strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfEquipmentsAsc:    strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfEquipmentsDesc:   strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "desc"),
		UsersAsc:              strings.Contains(req.GetSortBy(), "numofUsers") && strings.Contains(req.GetSortOrder().String(), "asc"),
		UsersDesc:             strings.Contains(req.GetSortBy(), "numofUsers") && strings.Contains(req.GetSortOrder().String(), "desc"),

		CostAsc:         strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "asc"),
		CostDesc:        strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "desc"),
		IsProductType:   !req.GetSearchParams().GetLocation().GetFilterType() && req.GetSearchParams().GetLocation().GetFilteringkey() != "",
		ProductType:     req.GetSearchParams().GetLocation().GetFilteringkey(),
		ProductTypeAsc:  strings.Contains(req.GetSortBy(), "location") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductTypeDesc: strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "desc"),
		// API expect pagenum from 1 but the offset in DB starts
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - listProductView - db/ListProductsView", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	apiresp := v1.ListProductsResponse{}
	apiresp.Products = make([]*v1.Product, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.Products[i] = &v1.Product{}
		apiresp.Products[i].SwidTag = dbresp[i].Swidtag
		apiresp.Products[i].Name = dbresp[i].ProductName
		apiresp.Products[i].Edition = dbresp[i].ProductEdition
		apiresp.Products[i].Editor = dbresp[i].ProductEditor
		apiresp.Products[i].Version = dbresp[i].ProductVersion
		apiresp.Products[i].Category = dbresp[i].ProductCategory
		apiresp.Products[i].NumOfApplications = dbresp[i].NumOfApplications
		apiresp.Products[i].NumofEquipments = dbresp[i].NumOfEquipments
		apiresp.Products[i].TotalCost = dbresp[i].Cost
		if dbresp[i].PProductType == "SAAS" {
			if dbresp[i].ConcurrentUsers > dbresp[i].NominativeUsers {
				apiresp.Products[i].NumofUsers = dbresp[i].ConcurrentUsers
			} else {
				apiresp.Products[i].NumofUsers = dbresp[i].NominativeUsers
			}
		} else {
			apiresp.Products[i].NumofUsers = dbresp[i].EquipmentUsers
		}
		apiresp.Products[i].EditorId = dbresp[i].EditorID.String
		apiresp.Products[i].ProductSwidTag = dbresp[i].ProductSwidTag.String
		apiresp.Products[i].VersionSwidTag = dbresp[i].VersionSwidTag.String
		apiresp.Products[i].ProductId = dbresp[i].ProductID.String

		apiresp.Products[i].Location = dbresp[i].PProductType
	}
	return &apiresp, nil
}

func (s *ProductServiceServer) GetProductCountByApp(ctx context.Context, req *v1.GetProductCountByAppRequest) (*v1.GetProductCountByAppResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("GetProductCountByApp - Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.ProductRepo.GetProductCount(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - GetProductCountByApp - error from repo/GetProductCountByApp", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	apiresp := v1.GetProductCountByAppResponse{}
	apiresp.AppData = make([]*v1.GetProductCountByAppResponseApplications, len(dbresp))

	for i := range dbresp {
		apiresp.AppData[i] = &v1.GetProductCountByAppResponseApplications{}
		apiresp.AppData[i].ApplicationId = dbresp[i].ApplicationID
		apiresp.AppData[i].NumOfProducts = dbresp[i].NumOfProducts
	}
	return &apiresp, nil
}

func (s *ProductServiceServer) GetApplicationsByProduct(ctx context.Context, req *v1.GetApplicationsByProductRequest) (*v1.GetApplicationsByProductResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("GetApplicationsByProduct - Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	app, err := s.ProductRepo.GetApplicationsByProductID(ctx, db.GetApplicationsByProductIDParams{
		Scope:   req.Scope,
		Swidtag: req.Swidtag,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetApplicationsByProduct - error from repo/GetApplicationsByProductID", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	return &v1.GetApplicationsByProductResponse{ApplicationId: app}, nil
}

func (s *ProductServiceServer) GetEquipmentsByProduct(ctx context.Context, req *v1.GetEquipmentsByProductRequest) (*v1.GetEquipmentsByProductResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("GetEquipmentsByProduct - Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	equipments, err := s.ProductRepo.GetEquipmentsBySwidtag(ctx, db.GetEquipmentsBySwidtagParams{
		Scope:   req.Scope,
		Swidtag: req.SwidTag,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetEquipmentsByProduct - error from repo/GetEquipmentsBySwidtag", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	return &v1.GetEquipmentsByProductResponse{EquipmentId: equipments}, nil
}

// nolint: gocyclo
func (s *ProductServiceServer) listProductViewInApplication(ctx context.Context, req *v1.ListProductsRequest, scopes []string) (*v1.ListProductsResponse, error) {
	products, err := s.ProductRepo.GetProductsByApplicationID(ctx, db.GetProductsByApplicationIDParams{
		Scope:         scopes[0],
		ApplicationID: req.SearchParams.ApplicationId.Filteringkey,
	})
	if err != nil {
		logger.Log.Error("service/v1 - listProductViewInApplicationInstance - application/GetProductsByApplicationInstance", zap.Error(err))
		return nil, status.Error(codes.Internal, "ServiceError")
	}
	prodFilter := []string{}
	if products != nil {
		prodFilter = products
	}
	dbresp, err := s.ProductRepo.ListProductsByApplication(ctx, db.ListProductsByApplicationParams{
		Scope:               req.Scopes,
		Swidtag:             prodFilter,
		ProductName:         req.GetSearchParams().GetName().GetFilteringkey(),
		IsProductName:       req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkProductName:       !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		ProductEditor:       req.GetSearchParams().GetEditor().GetFilteringkey(),
		IsProductEditor:     req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		LkProductEditor:     !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		ProductNameAsc:      strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductNameDesc:     strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		SwidtagAsc:          strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "asc"),
		SwidtagDesc:         strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductVersionAsc:   strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductVersionDesc:  strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditionAsc:   strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditionDesc:  strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductCategoryAsc:  strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductCategoryDesc: strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:    strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc:   strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TotalCostAsc:        strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TotalCostDesc:       strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "desc"),
		PageNum:             req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize:            req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - listProductViewInApplication - db/ListProductsByApplication", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	apiresp := v1.ListProductsResponse{}
	apiresp.Products = make([]*v1.Product, len(dbresp))
	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.Products[i] = &v1.Product{}
		apiresp.Products[i].SwidTag = dbresp[i].Swidtag
		apiresp.Products[i].Name = dbresp[i].ProductName
		apiresp.Products[i].Editor = dbresp[i].ProductEditor
		apiresp.Products[i].Version = dbresp[i].ProductVersion
		apiresp.Products[i].TotalCost = dbresp[i].TotalCost
		apiresp.Products[i].NumofEquipments = dbresp[i].NumOfEquipments

	}
	return &apiresp, nil
}

// nolint: gocyclo
// func (s *ProductServiceServer) listProductViewInApplication(ctx context.Context, req *v1.ListProductsRequest, scopes []string) (*v1.ListProductsResponse, error) {
// 	appEquipments, err := s.application.GetEquipmentsByApplication(ctx, &appv1.GetEquipmentsByApplicationRequest{
// 		Scope:         scopes[0],
// 		ApplicationId: req.SearchParams.ApplicationId.Filteringkey,
// 	})
// 	if err != nil {
// 		logger.Log.Error("service/v1 - listProductViewInApplication - application/GetEquipmentsByApplication", zap.Error(err))
// 		return nil, status.Error(codes.Internal, "ServiceError")
// 	}
// 	equipmentFilter := []string{}
// 	if appEquipments != nil {
// 		equipmentFilter = appEquipments.EquipmentId
// 	}
// 	dbresp, err := s.ProductRepo.ListProductsViewRedirectedApplication(ctx, db.ListProductsViewRedirectedApplicationParams{
// 		Scope:                 scopes,
// 		Swidtag:               req.GetSearchParams().GetSwidTag().GetFilteringkey(),
// 		IsSwidtag:             req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
// 		LkSwidtag:             !req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
// 		ProductName:           req.GetSearchParams().GetName().GetFilteringkey(),
// 		IsProductName:         req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
// 		LkProductName:         !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
// 		ProductEditor:         req.GetSearchParams().GetEditor().GetFilteringkey(),
// 		IsProductEditor:       req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
// 		LkProductEditor:       !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
// 		ApplicationID:         req.GetSearchParams().GetApplicationId().GetFilteringkey(),
// 		IsApplicationID:       req.GetSearchParams().GetApplicationId().GetFilterType() && req.GetSearchParams().GetApplicationId().GetFilteringkey() != "",
// 		EquipmentIds:          equipmentFilter,
// 		IsEquipmentID:         true,
// 		ProductNameAsc:        strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "asc"),
// 		ProductNameDesc:       strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "desc"),
// 		SwidtagAsc:            strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "asc"),
// 		SwidtagDesc:           strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "desc"),
// 		ProductVersionAsc:     strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "asc"),
// 		ProductVersionDesc:    strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "desc"),
// 		ProductEditionAsc:     strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "asc"),
// 		ProductEditionDesc:    strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "desc"),
// 		ProductCategoryAsc:    strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "asc"),
// 		ProductCategoryDesc:   strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "desc"),
// 		ProductEditorAsc:      strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
// 		ProductEditorDesc:     strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "desc"),
// 		NumOfApplicationsAsc:  strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "asc"),
// 		NumOfApplicationsDesc: strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "desc"),
// 		NumOfEquipmentsAsc:    strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "asc"),
// 		NumOfEquipmentsDesc:   strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "desc"),
// 		CostAsc:               strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "asc"),
// 		CostDesc:              strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "desc"),
// 		// API expect pagenum from 1 but the offset in DB starts
// 		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
// 		PageSize: req.GetPageSize(),
// 	})
// 	if err != nil {
// 		logger.Log.Error("service/v1 - listProductViewInApplication - db/ListProductsViewRedirectedApplication", zap.Error(err))
// 		return nil, status.Error(codes.Internal, "DBError")
// 	}

// 	apiresp := v1.ListProductsResponse{}
// 	apiresp.Products = make([]*v1.Product, len(dbresp))

// 	if len(dbresp) > 0 {
// 		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
// 	}

// 	for i := range dbresp {
// 		apiresp.Products[i] = &v1.Product{}
// 		apiresp.Products[i].SwidTag = dbresp[i].Swidtag
// 		apiresp.Products[i].Name = dbresp[i].ProductName
// 		apiresp.Products[i].Edition = dbresp[i].ProductEdition
// 		apiresp.Products[i].Editor = dbresp[i].ProductEditor
// 		apiresp.Products[i].Version = dbresp[i].ProductVersion
// 		apiresp.Products[i].Category = dbresp[i].ProductCategory
// 		apiresp.Products[i].NumOfApplications = dbresp[i].NumOfApplications
// 		apiresp.Products[i].NumofEquipments = dbresp[i].NumOfEquipments
// 		apiresp.Products[i].TotalCost = dbresp[i].Cost
// 	}
// 	return &apiresp, nil
// }

// nolint: gocyclo
func (s *ProductServiceServer) listProductViewInEquipment(ctx context.Context, req *v1.ListProductsRequest, scopes []string) (*v1.ListProductsResponse, error) {
	dbresp, err := s.ProductRepo.ListProductsViewRedirectedEquipment(ctx, db.ListProductsViewRedirectedEquipmentParams{
		Scope:                 scopes,
		Swidtag:               req.GetSearchParams().GetSwidTag().GetFilteringkey(),
		IsSwidtag:             req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		LkSwidtag:             !req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		ProductName:           req.GetSearchParams().GetName().GetFilteringkey(),
		IsProductName:         req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkProductName:         !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		ProductEditor:         req.GetSearchParams().GetEditor().GetFilteringkey(),
		IsProductEditor:       req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		LkProductEditor:       !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		ApplicationID:         req.GetSearchParams().GetApplicationId().GetFilteringkey(),
		IsApplicationID:       req.GetSearchParams().GetApplicationId().GetFilterType() && req.GetSearchParams().GetApplicationId().GetFilteringkey() != "",
		EquipmentID:           req.GetSearchParams().GetEquipmentId().GetFilteringkey(),
		IsEquipmentID:         req.GetSearchParams().GetEquipmentId().GetFilterType() && req.GetSearchParams().GetEquipmentId().GetFilteringkey() != "",
		ProductNameAsc:        strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductNameDesc:       strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		SwidtagAsc:            strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "asc"),
		SwidtagDesc:           strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductVersionAsc:     strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductVersionDesc:    strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditionAsc:     strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditionDesc:    strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductCategoryAsc:    strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductCategoryDesc:   strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:      strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc:     strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfApplicationsAsc:  strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfApplicationsDesc: strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfEquipmentsAsc:    strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfEquipmentsDesc:   strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "desc"),
		CostAsc:               strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "asc"),
		CostDesc:              strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "desc"),
		// API expect pagenum from 1 but the offset in DB starts
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - listProductViewInEquipment - db/ListProductsViewRedirectedEquipment", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	apiresp := v1.ListProductsResponse{}
	apiresp.Products = make([]*v1.Product, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.Products[i] = &v1.Product{}
		apiresp.Products[i].SwidTag = dbresp[i].Swidtag
		apiresp.Products[i].Name = dbresp[i].ProductName
		apiresp.Products[i].Edition = dbresp[i].ProductEdition
		apiresp.Products[i].Editor = dbresp[i].ProductEditor
		apiresp.Products[i].Version = dbresp[i].ProductVersion
		apiresp.Products[i].Category = dbresp[i].ProductCategory
		apiresp.Products[i].NumOfApplications = dbresp[i].NumOfApplications
		apiresp.Products[i].NumofEquipments = dbresp[i].NumOfEquipments
		apiresp.Products[i].TotalCost = dbresp[i].Cost
		apiresp.Products[i].AllocatedUser = dbresp[i].EquipmentUsers
		apiresp.Products[i].AllocatedMetric = dbresp[i].AllocatedMetric
	}
	return &apiresp, nil
}

func (s *ProductServiceServer) GetProductDetail(ctx context.Context, req *v1.ProductRequest) (*v1.ProductResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.ProductRepo.GetProductInformation(ctx, db.GetProductInformationParams{
		Swidtag: req.SwidTag,
		Scope:   req.Scope,
	})
	var dbmetrics []string
	apiresp := v1.ProductResponse{}
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Log.Error("service/v1 - GetProductDetail - db/GetProductInformation", zap.Error(err))
			return nil, status.Error(codes.Internal, "DBError")
		}
		logger.Log.Error("service/v1 - GetProductDetail - db/GetProductInformation - product does not exist", zap.Error(err))
		dbresp, err1 := s.ProductRepo.GetProductInformationFromAcqright(ctx, db.GetProductInformationFromAcqrightParams{
			Swidtag: req.SwidTag,
			Scope:   req.Scope,
		})
		if err1 != nil {
			if errors.Is(err1, sql.ErrNoRows) {
				logger.Log.Error("service/v1 - GetProductDetail - db/GetProductInformationFromAcqright - product does not exist", zap.Error(err1))
				return nil, status.Error(codes.NotFound, "NoContent")
			}
			logger.Log.Error("service/v1 - GetProductDetail - db/GetProductInformationFromAcqright", zap.Error(err1))
			return nil, status.Error(codes.Internal, "DBError")
		}
		apiresp.SwidTag = dbresp.Swidtag
		apiresp.ProductName = dbresp.ProductName
		apiresp.Editor = dbresp.ProductEditor
		apiresp.Version = dbresp.Version
		dbmetrics = dbresp.Metrics
		apiresp.ProductSwidTag = dbresp.ProductSwidTag.String
		apiresp.VersionSwidTag = dbresp.VersionSwidTag.String
		if apiresp.Version == "" || strings.ToLower(apiresp.Version) == "all" {
			productsResponse, _ := s.ProductRepo.GetProductByNameEditor(ctx, db.GetProductByNameEditorParams{
				ProductName:   []string{dbresp.ProductName},
				ProductEditor: []string{dbresp.ProductEditor},
			})
			if len(productsResponse) == 0 {
				apiresp.NotDeployed = true
			}
		}
	} else {
		apiresp.SwidTag = dbresp.Swidtag
		apiresp.ProductName = dbresp.ProductName
		apiresp.Editor = dbresp.ProductEditor
		apiresp.Version = dbresp.ProductVersion
		apiresp.NumApplications = dbresp.NumOfApplications
		apiresp.NumEquipments = dbresp.NumOfEquipments
		apiresp.NotDeployed = false
		if len(dbresp.Metrics) > 0 {
			dbmetrics = dbresp.Metrics
		} else {
			pName := removeSpecialChars(dbresp.ProductName)
			pEditor := removeSpecialChars(dbresp.ProductEditor)
			swittagwithoutversion := strings.ReplaceAll(strings.ReplaceAll(strings.Join([]string{pName, pEditor}, "_"), " ", "_"), "-", "_")
			dbrespAcq, err1 := s.ProductRepo.GetProductInformationFromAcqright(ctx, db.GetProductInformationFromAcqrightParams{
				Swidtag: swittagwithoutversion,
				Scope:   req.Scope,
			})
			if err1 == nil || errors.Is(err1, sql.ErrNoRows) {
				if len(dbrespAcq.Metrics) > 0 {
					dbmetrics = dbrespAcq.Metrics
				} else {
					swittagwithoutversion = strings.ReplaceAll(strings.ReplaceAll(strings.Join([]string{pName, pEditor, "all"}, "_"), " ", "_"), "-", "_")
					dbrespAcqAll, err2 := s.ProductRepo.GetProductInformationFromAcqrightForAll(ctx, db.GetProductInformationFromAcqrightForAllParams{
						Swidtag: strings.ToLower(swittagwithoutversion),
						Scope:   req.Scope,
					})
					if err2 == nil || errors.Is(err2, sql.ErrNoRows) {
						if len(dbrespAcqAll.Metrics) > 0 {
							dbmetrics = dbrespAcqAll.Metrics
						}
					}
				}

			}

		}
		if apiresp.NumEquipments == 0 && ((len(apiresp.Version) != 0 && strings.ToLower(apiresp.Version) != "all") || dbresp.ProductType != "SAAS") {
			apiresp.NotDeployed = true
		}
	}

	metrics, err := s.metric.ListMetrices(ctx, &metv1.ListMetricRequest{
		Scopes: []string{req.Scope},
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetProductDetail - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "ServiceError")
	}
	if metrics != nil || len(metrics.Metrices) != 0 {
		for _, met := range dbmetrics {
			if idx := metricExists(metrics.Metrices, met); idx != -1 {
				apiresp.DefinedMetrics = append(apiresp.DefinedMetrics, met)
			}
			flag := metricTypeOfSaasExists(metrics.Metrices, met)
			if flag == false {
				apiresp.NotDeployed = false
			}
		}
	}
	return &apiresp, nil

}

func (s *ProductServiceServer) GetProductOptions(ctx context.Context, req *v1.ProductRequest) (*v1.ProductOptionsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.ProductRepo.GetProductOptions(ctx, db.GetProductOptionsParams{
		Swidtag: req.GetSwidTag(),
		Scope:   req.Scope,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetProductOptions - db/GetProductOptions", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	apiresp := v1.ProductOptionsResponse{}
	apiresp.Optioninfo = make([]*v1.OptionInfo, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.NumOfOptions = int32(len(dbresp))
	}
	for i := range dbresp {
		apiresp.Optioninfo[i] = &v1.OptionInfo{}
		apiresp.Optioninfo[i].SwidTag = dbresp[i].Swidtag
		apiresp.Optioninfo[i].Name = dbresp[i].ProductName
		apiresp.Optioninfo[i].Edition = dbresp[i].ProductEdition
		apiresp.Optioninfo[i].Version = dbresp[i].ProductVersion
		apiresp.Optioninfo[i].Editor = dbresp[i].ProductEditor
	}
	return &apiresp, nil
}

func (s *ProductServiceServer) DropProductData(ctx context.Context, req *v1.DropProductDataRequest) (*v1.DropProductDataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DropProductDataResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return &v1.DropProductDataResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	if err := s.ProductRepo.DropProductDataTx(ctx, req.Scope, req.DeletionType); err != nil {
		return &v1.DropProductDataResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	// For dgworker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.DropProductDataRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	_, err = s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
		PPID:   req.Ppid,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	return &v1.DropProductDataResponse{Success: true}, nil
}

func (s *ProductServiceServer) DropAggregationData(ctx context.Context, req *v1.DropAggregationDataRequest) (*v1.DropAggregationDataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DropAggregationDataResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return &v1.DropAggregationDataResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	if err := s.ProductRepo.DeleteAggregationByScope(ctx, req.Scope); err != nil {
		return &v1.DropAggregationDataResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	// For dgworker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.DropAggregationData, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	_, err = s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	return &v1.DropAggregationDataResponse{Success: true}, nil
}

func metricTypeOfSaasExists(metrics []*metv1.Metric, name string) bool {
	flag := true
	for _, met := range metrics {
		if met.Name == name {
			if met.Type == "saas.nominative.standard" || met.Type == "saas.concurrent.standard" {
				flag = false
			}
		}
	}
	return flag
}

func (s *ProductServiceServer) GetProductInformationBySwidTag(ctx context.Context, req *v1.GetProductInformationBySwidTagRequest) (*v1.GetProductInformationBySwidTagResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.ProductRepo.GetProductInformation(ctx, db.GetProductInformationParams{
		Swidtag: req.SwidTag,
		Scope:   req.Scope,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.GetProductInformationBySwidTagResponse{}, nil
		}
		logger.Log.Error("service/v1 - GetProductInformationBySwidTag - GetProductInformation", zap.String("reason", err.Error()))
		return &v1.GetProductInformationBySwidTagResponse{}, status.Error(codes.Internal, "DBError")
	}
	return &v1.GetProductInformationBySwidTagResponse{
		Swidtag:           dbresp.Swidtag,
		ProductName:       dbresp.ProductName,
		ProductEditor:     dbresp.ProductEditor,
		ProductVersion:    dbresp.ProductVersion,
		Metrics:           dbresp.Metrics,
		NumOfApplications: dbresp.NumOfApplications,
		NumOfEquipments:   dbresp.NumOfEquipments,
	}, nil
}
