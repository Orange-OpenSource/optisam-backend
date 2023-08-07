package v1

import (
	"context"
	"encoding/json"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	repo "optisam-backend/catalog-service/pkg/repository/v1"
	logger "optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"

	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"

	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type productCatalogServer struct {
	productRepo repo.ProductCatalog
	queue       workerqueue.Workerqueue
	r           *redis.Client
}

// NewProductCatalogServer creates Product service
func NewProductCatalogServer(productRepo repo.ProductCatalog, queue workerqueue.Workerqueue, grpcServers map[string]*grpc.ClientConn, r *redis.Client) v1.ProductCatalogServer {
	return &productCatalogServer{
		productRepo: productRepo,
		queue:       queue,
		r:           r,
	}
}

// Create Product
func (p *productCatalogServer) InsertProduct(ctx context.Context, req *v1.Product) (res *v1.Product, err error) {
	// logger.Log.Info("req being processed to InsertProduct.")
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("v1/service - InsertProduct - ClaimsNotFound")
		return &v1.Product{}, status.Error(codes.Internal, "ClaimsNotFound")
	}

	if !(userClaims.Role == claims.RoleSuperAdmin || userClaims.Role == claims.RoleAdmin) {
		logger.Log.Error("v1/service - InsertProduct - ClaimsNotFound")
		return &v1.Product{}, status.Error(codes.PermissionDenied, "RoleValidationError")
	}
	res, err = p.productRepo.InsertProductTx(ctx, req)
	return res, err
}

func (p *productCatalogServer) GetProduct(ctx context.Context, req *v1.GetProductRequest) (*v1.Product, error) {
	logger.Log.Info("req being processed to GetProduct.")
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("v1/service - getProduct - ClaimsNotFound")
		return &v1.Product{}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !(userClaims.Role == claims.RoleSuperAdmin || userClaims.Role == claims.RoleAdmin) {
		logger.Log.Error("v1/service - getProduct - RoleValidationError")
		return &v1.Product{}, status.Error(codes.PermissionDenied, "RoleValidationError")
	}

	var responseObject v1.Product
	responseObject.Version = make([]*v1.Version, 0)
	responseObject.OpenSource = new(v1.OpenSource)
	productResponse, err := p.productRepo.GetProductCatalogByPrductID(ctx, req.ProdId)
	if err != nil {
		logger.Log.Error("service/v1 - GetProduct - GetProduct", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "DBError")
	}
	responseObject.Id = productResponse.ID
	responseObject.EditorID = productResponse.Editorid
	responseObject.Name = productResponse.Name
	responseObject.SwidtagProduct = productResponse.SwidTagProduct.String
	responseObject.EditorName = productResponse.EditorName

	jsonStr, err := json.Marshal(productResponse.LicencesOpensource.String)
	if err != nil {
		logger.Log.Error("service/v1 - getProduct - Marshal", zap.String("Reason: ", err.Error()))
		return nil, status.Error(codes.Internal, "Error while marshaling")
	}
	json.Unmarshal(jsonStr, &responseObject.OpenSource.OpenLicences)

	responseObject.ContracttTips = productResponse.ContractTips.String
	responseObject.GenearlInformation = productResponse.GenearlInformation.String
	responseObject.Recommendation = string(productResponse.Recommendation)
	responseObject.LocationType = string(productResponse.Location)
	responseObject.Licensing = string(productResponse.Licensing)
	responseObject.OpenSource.OpenSourceType = string(productResponse.OpensourceType)

	jsonStr, err = json.Marshal(productResponse.Metrics)
	if err != nil {
		logger.Log.Error("service/v1 - getProduct - Marshal", zap.String("Reason: ", err.Error()))
		return nil, status.Error(codes.Internal, "Error while marshaling")
	}
	json.Unmarshal(jsonStr, &responseObject.Metrics)

	jsonStr, err = json.Marshal(productResponse.SupportVendors)
	if err != nil {
		logger.Log.Error("service/v1 - getProduct - Marshal", zap.String("Reason: ", err.Error()))
		return nil, status.Error(codes.Internal, "Error while marshaling")
	}
	json.Unmarshal(jsonStr, &responseObject.SupportVendors)

	jsonStr, err = json.Marshal(productResponse.UsefulLinks)
	if err != nil {
		logger.Log.Error("service/v1 - getProduct - Marshal", zap.String("Reason: ", err.Error()))
		return nil, status.Error(codes.Internal, "Error while marshaling")
	}
	json.Unmarshal(jsonStr, &responseObject.UsefulLinks)

	createdOnObject, _ := ptypes.TimestampProto(productResponse.CreatedOn)
	responseObject.CreatedOn = createdOnObject

	updatedOnObject, _ := ptypes.TimestampProto(productResponse.UpdatedOn)
	responseObject.UpdatedOn = updatedOnObject
	versions, err := p.productRepo.GetVersionCatalogByPrductID(ctx, productResponse.ID)
	if err != nil {
		logger.Log.Error("service/v1 - GetProduct  - GetVersionCatalogByPrductID", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "Error while Getversion")

	}

	for _, ver := range versions {
		version := v1.Version{}
		endofLifeOnObject, _ := ptypes.TimestampProto(ver.EndOfLife.Time)
		version.EndOfLife = endofLifeOnObject

		endofSupportOnObject, _ := ptypes.TimestampProto(ver.EndOfSupport.Time)
		version.EndOfSupport = endofSupportOnObject
		version.SwidtagVersion = ver.SwidTagVersion.String
		version.Id = ver.ID
		version.Name = ver.Name
		version.Recommendation = ver.Recommendation.String

		responseObject.Version = append(responseObject.Version, &version)
	}

	return &responseObject, nil
}

func (s *productCatalogServer) DeleteProduct(ctx context.Context, request *v1.GetProductRequest) (*v1.DeleteResponse, error) {

	logger.Log.Info("req being processed to delete Product.")
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("v1/service - delProduct - ClaimsNotFound")
		return &v1.DeleteResponse{}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !(userClaims.Role == claims.RoleSuperAdmin || userClaims.Role == claims.RoleAdmin) {
		logger.Log.Error("v1/service - delProduct - ClaimsNotFound")
		return &v1.DeleteResponse{}, status.Error(codes.PermissionDenied, "RoleValidationError")
	}
	delPolicyErr := s.productRepo.DeleteProductCatalog(ctx, request.ProdId)
	if delPolicyErr != nil {
		logger.Log.Error("DeletePolicy- DeletePolicyByID : ", zap.String("DeletePolicyByID: ", delPolicyErr.Error()))
		return nil, status.Error(codes.Internal, "DeletePolicyByID Error.")
	}
	return &v1.DeleteResponse{Success: true}, nil
}

// ListProducts
func (p *productCatalogServer) UpdateProduct(ctx context.Context, req *v1.Product) (product *v1.Product, err error) {
	// fmt.Println("list products")
	logger.Log.Info("req being processed to UpdateProduct.")
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("v1/service - updateProduct - ClaimsNotFound")
		return &v1.Product{}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !(userClaims.Role == claims.RoleSuperAdmin || userClaims.Role == claims.RoleAdmin) {
		logger.Log.Error("v1/service - updateProduct - RoleValidationError")
		return &v1.Product{}, status.Error(codes.PermissionDenied, "RoleValidationError")
	}
	err = p.productRepo.UpdateProductTx(ctx, req)
	if err == nil {
		product, err = p.GetProduct(ctx, &v1.GetProductRequest{
			ProdId: req.Id,
		})
		if err != nil {
			logger.Log.Error("service/v1 | UpdateProduct | UpdateProduct", zap.Any("Error retriving saved record", err))
			return nil, status.Error(codes.Internal, "Error while retriving saved record")
		}
		return product, nil
	}
	return product, err
}
