package v1

import (
	"context"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	prodv1 "optisam-backend/product-service/pkg/api/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// licenseServiceServer is implementation of v1.authServiceServer proto interface
type licenseServiceServer struct {
	licenseRepo   repo.License
	productClient prodv1.ProductServiceClient
}

// NewLicenseServiceServer creates License service
func NewLicenseServiceServer(licenseRepo repo.License, grpcServers map[string]*grpc.ClientConn) v1.LicenseServiceServer {
	return &licenseServiceServer{
		licenseRepo:   licenseRepo,
		productClient: prodv1.NewProductServiceClient(grpcServers["product"]),
	}
}

func (s *licenseServiceServer) GetOverAllCompliance(ctx context.Context, req *v1.GetOverAllComplianceRequest) (*v1.GetOverAllComplianceResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - GetOverAllCompliance", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	aggResp, err := s.licenseRepo.GetAggregations(ctx, req.Editor, req.Scope)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 - GetAggregations", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	logger.Log.Debug("All agg in scope or by edtior", zap.Any("agg", aggResp))
	var swidtags []string
	if !req.Simulation {
		for _, v := range aggResp {
			swidtags = append(swidtags, v.Swidtags...)
		}
	}
	acqResp, err := s.licenseRepo.GetAcqRights(ctx, swidtags, req.Editor, req.Scope)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 - GetAcqRights", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	productExists := map[string]string{}
	var agg []*v1.AggregationAcquiredRights
	for _, v := range acqResp {
		if _, ok := productExists[v.ProductName]; !ok {
			productExists[v.ProductName] = v.Swidtag
			res, err := s.ListAcqRightsForProduct(ctx, &v1.ListAcquiredRightsForProductRequest{
				SwidTag:    v.Swidtag,
				Scope:      req.Scope,
				Simulation: req.Simulation,
			})
			if err != nil {
				logger.Log.Error("service/v1 - GetOverAllCompliance: ListAcqRightsForProduct ", zap.Error(err))
				continue
			}
			temp := convertAcqRightResponseToAggregationResponse(res.AcqRights)
			agg = append(agg, temp...)
		}
	}
	for _, v := range aggResp {
		res, err := s.ListAcqRightsForAggregation(ctx, &v1.ListAcqRightsForAggregationRequest{
			Name:       v.Name,
			Scope:      req.Scope,
			Simulation: req.Simulation,
		})
		if err != nil {
			logger.Log.Error("service/v1 - GetOverAllCompliance: ListAcqRightsForAggregation ", zap.Error(err), zap.String("aggregation name", v.Name), zap.String("scope", req.Scope))
			continue
		}
		agg = append(agg, res.AcqRights...)
	}

	return &v1.GetOverAllComplianceResponse{
		AcqRights: agg,
	}, nil
}

func convertAcqRightResponseToAggregationResponse(data []*v1.ProductAcquiredRights) []*v1.AggregationAcquiredRights {
	temp := []*v1.AggregationAcquiredRights{}
	for _, v := range data {
		temp = append(temp, &v1.AggregationAcquiredRights{
			SKU:              v.SKU,
			SwidTags:         v.SwidTag,
			Metric:           v.Metric,
			NumCptLicences:   v.NumCptLicences,
			NumAcqLicences:   v.NumAcqLicences,
			TotalCost:        v.TotalCost,
			DeltaNumber:      v.DeltaNumber,
			DeltaCost:        v.DeltaCost,
			AvgUnitPrice:     v.AvgUnitPrice,
			ComputedDetails:  v.ComputedDetails,
			MetricNotDefined: v.MetricNotDefined,
			NotDeployed:      v.NotDeployed,
			ProductNames:     v.ProductName,
			PurchaseCost:     v.PurchaseCost,
			ComputedCost:     v.ComputedCost,
		})
	}
	return temp
}

// func (s *licenseServiceServer) GetProductsbyApplication(ctx context.Context, req *v1.ApplicationRequest) (*v1.ApplicationResponse, error) {

//
// 	if err != nil {
// 		return nil, status.Error(codes.Unknown, "failed to get Products information-> "+err.Error())
// 	}
// 	return res, nil
// }
