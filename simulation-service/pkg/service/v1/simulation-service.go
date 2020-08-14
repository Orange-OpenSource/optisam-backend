// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	repo "optisam-backend/simulation-service/pkg/repository/v1"
	v1 "optisam-backend/license-service/pkg/api/v1"

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
