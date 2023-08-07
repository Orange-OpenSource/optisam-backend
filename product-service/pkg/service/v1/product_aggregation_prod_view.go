package v1

import (
	"context"
	"database/sql"
	"errors"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	metv1 "optisam-backend/metric-service/pkg/api/v1"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// nolint: gocyclo
func (s *productServiceServer) ListProductAggregationView(ctx context.Context, req *v1.ListProductAggregationViewRequest) (*v1.ListProductAggregationViewResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	dbresp, err := s.productRepo.ListProductAggregation(ctx, db.ListProductAggregationParams{
		Scope:    req.GetScopes()[0],
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListProductAggregationView - ListProductAggregation", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "DBError")
	}

	apiresp := v1.ListProductAggregationViewResponse{}

	for i := range dbresp {
		temp := &v1.ProductAggregationView{}
		temp.ID = dbresp[i].ID
		temp.AggregationName = dbresp[i].AggregationName
		temp.Swidtags = dbresp[i].Swidtags
		temp.Editor = dbresp[i].ProductEditor
		temp.NumApplications = dbresp[i].NumOfApplications
		temp.NumEquipments = dbresp[i].NumOfEquipments
		temp.EditorId = dbresp[i].EditorID.String
		individualCount, err := s.productRepo.GetIndividualProductForAggregationCount(ctx, db.GetIndividualProductForAggregationCountParams{
			Scope:    req.Scopes[0],
			Swidtags: dbresp[i].Swidtags,
		})
		if err != nil {
			logger.Log.Error("service/v1 - ListProductAggregationView - GetIndividualProductForAggregationCount", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "DBError")
		}
		if individualCount > 0 {
			temp.IndividualProductExists = true
		}
		if dbresp[i].NominativeUsers == 0 && dbresp[i].ConcurrentUsers == 0 {
			temp.UsersCount = dbresp[i].NumOfUsers
			temp.Location = "ONPREMISE"
		} else if dbresp[i].NominativeUsers > dbresp[i].ConcurrentUsers {
			temp.UsersCount = dbresp[i].NominativeUsers
			temp.Location = "SAAS"
		} else {
			temp.UsersCount = dbresp[i].ConcurrentUsers
			temp.Location = "SAAS"
		}

		apiresp.Aggregations = append(apiresp.Aggregations, temp)
	}
	apiresp.TotalRecords = int32(len(apiresp.Aggregations))
	return &apiresp, nil
}

// func (s *productServiceServer) ListProductAggregationRecords(ctx context.Context, req *v1.ListProductAggregationRecordsRequest) (*v1.ListProductAggregationRecordsResponse, error) {
// 	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
// 	if !ok {
// 		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
// 	}
// 	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
// 		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
// 	}
// 	dbresp, err := s.productRepo.ListProductsAggregationIndividual(ctx, db.ListProductsAggregationIndividualParams{
// 		AggregationName: req.AggregationName,
// 		Scope:           req.Scopes,
// 	})
// 	if err != nil {
// 		logger.Log.Error("service/v1 - ListProductAggregationRecords - db/ListProductsAggregationIndividual", zap.Error(err))
// 		return nil, status.Error(codes.Unknown, "DBError")
// 	}
// 	apiresp := v1.ListProductAggregationRecordsResponse{}
// 	apiresp.ProdAggRecord = make([]*v1.ProductAggRecord, len(dbresp))
// 	for i := range dbresp {
// 		apiresp.ProdAggRecord[i] = &v1.ProductAggRecord{}
// 		apiresp.ProdAggRecord[i].SwidTag = dbresp[i].Swidtag
// 		apiresp.ProdAggRecord[i].Name = dbresp[i].ProductName
// 		apiresp.ProdAggRecord[i].Editor = dbresp[i].ProductEditor
// 		apiresp.ProdAggRecord[i].Edition = dbresp[i].ProductEdition
// 		apiresp.ProdAggRecord[i].Version = dbresp[i].ProductVersion
// 		apiresp.ProdAggRecord[i].AggregationName = dbresp[i].AggregationName
// 		apiresp.ProdAggRecord[i].NumApplications = int32(dbresp[i].NumOfApplications)
// 		apiresp.ProdAggRecord[i].NumEquipments = int32(dbresp[i].NumOfEquipments)
// 	}
// 	return &apiresp, nil

// }

func (s *productServiceServer) AggregatedRightDetails(ctx context.Context, req *v1.AggregatedRightDetailsRequest) (*v1.AggregatedRightDetailsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.AggregatedRightDetails(ctx, db.AggregatedRightDetailsParams{
		ID:    req.GetID(),
		Scope: req.Scope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Log.Error("service/v1 - AggregatedRightDetails - db/AggregatedRightDetails - aggregation does not exist", zap.Error(err))
			return nil, status.Error(codes.NotFound, "NoContent")
		}
		logger.Log.Error("service/v1 - AggregatedRightDetails - db/AggregatedRightDetails", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	flag := false

	if dbresp.NumOfEquipments == 0 {
		flag = true
	}
	dbmetrics := dbresp.Metrics
	metrics, err := s.metric.ListMetrices(ctx, &metv1.ListMetricRequest{
		Scopes: []string{req.Scope},
	})
	if err != nil {
		logger.Log.Error("service/v1 - AggregatedRightDetails - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "ServiceError")
	}
	if metrics != nil || len(metrics.Metrices) != 0 {
		for _, met := range dbmetrics {
			exist := metricTypeOfSaasExists(metrics.Metrices, met)
			if exist == false {
				flag = false
			}
		}
	}
	return &v1.AggregatedRightDetailsResponse{
		ID:              req.ID,
		Name:            dbresp.AggregationName,
		Editor:          dbresp.ProductEditor,
		Products:        dbresp.ProductSwidtags,
		ProductNames:    dbresp.ProductNames,
		Versions:        dbresp.ProductVersions,
		NumApplications: dbresp.NumOfApplications,
		NumEquipments:   dbresp.NumOfEquipments,
		DefinedMetrics:  dbresp.Metrics,
		NotDeployed:     flag,
	}, nil
}

func (s *productServiceServer) GetAggregationProductsExpandedView(ctx context.Context, req *v1.GetAggregationProductsExpandedViewRequest) (*v1.GetAggregationProductsExpandedViewResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}

	expandedProd, err := s.productRepo.GetIndividualProductDetailByAggregation(ctx, db.GetIndividualProductDetailByAggregationParams{
		AggregationName: req.AggregationName,
		Scope:           req.Scope,
	})
	if err != nil {
		logger.Log.Error("Failed to get products on expanding aggregated products", zap.Error(err), zap.String("aggName", req.AggregationName))
		return nil, status.Error(codes.Internal, "DBError")
	}
	apiresp := &v1.GetAggregationProductsExpandedViewResponse{
		TotalRecords: int32(len(expandedProd)),
	}
	for _, v := range expandedProd {
		temp := &v1.ProductExpand{}
		temp.SwidTag = v.ProdID
		temp.Name = v.Name
		temp.Editor = v.ProductEditor
		temp.Version = v.Version
		temp.NumApplications = int32(v.NumOfApplications)
		temp.NumEquipments = int32(v.NumOfEquipments)
		temp.TotalCost, _ = v.TotalCost.Float64()
		apiresp.Products = append(apiresp.Products, temp)
	}
	return apiresp, nil
}

// func (s *productServiceServer) ProductAggregationProductViewOptions(ctx context.Context, req *v1.ProductAggregationProductViewOptionsRequest) (*v1.ProductAggregationProductViewOptionsResponse, error) {
// 	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
// 	if !ok {
// 		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
// 	}
// 	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
// 		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
// 	}

// 	dbresp, err := s.productRepo.ProductAggregationChildOptions(ctx, db.ProductAggregationChildOptionsParams{
// 		AggregationID: req.GetID(),
// 		Scope:         req.Scopes,
// 	})
// 	if err != nil {
// 		logger.Log.Error("service/v1 - ProductAggregationProductViewOptions - db/ProductAggregationChildOptions", zap.Error(err))
// 		return nil, status.Error(codes.Unknown, "DBError")
// 	}
// 	apiresp := v1.ProductAggregationProductViewOptionsResponse{}
// 	apiresp.Optioninfo = make([]*v1.OptionInfo, len(dbresp))
// 	if len(dbresp) > 0 {
// 		apiresp.NumOfOptions = int32(len(dbresp))
// 	}
// 	for i := range dbresp {
// 		apiresp.Optioninfo[i] = &v1.OptionInfo{}
// 		apiresp.Optioninfo[i].SwidTag = dbresp[i].Swidtag
// 		apiresp.Optioninfo[i].Name = dbresp[i].ProductName
// 		apiresp.Optioninfo[i].Edition = dbresp[i].ProductEdition
// 		apiresp.Optioninfo[i].Editor = dbresp[i].ProductEditor
// 		apiresp.Optioninfo[i].Version = dbresp[i].ProductVersion
// 	}
// 	return &apiresp, nil
// }
