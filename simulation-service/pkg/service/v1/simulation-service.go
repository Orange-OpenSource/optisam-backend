package v1

import (
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/simulation-service/pkg/repository/v1"

	"google.golang.org/grpc"
)

// SimulationService is implementation of service interface
type SimulationService struct {
	repo          repo.Repository
	licenseClient v1.LicenseServiceClient
}

// NewSimulationService creates SimulationService
func NewSimulationService(rep repo.Repository, grpcServers map[string]*grpc.ClientConn) *SimulationService {
	return &SimulationService{
		repo:          rep,
		licenseClient: v1.NewLicenseServiceClient(grpcServers["license"]),
	}
}

// NewSimulationServiceForTest creates SimulationService for test
func NewSimulationServiceForTest(rep repo.Repository, licenseClient v1.LicenseServiceClient) *SimulationService {
	return &SimulationService{
		repo:          rep,
		licenseClient: licenseClient,
	}
}
