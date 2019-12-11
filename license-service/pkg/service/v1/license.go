// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
)

// licenseServiceServer is implementation of v1.authServiceServer proto interface
type licenseServiceServer struct {
	licenseRepo repo.License
}

// NewLicenseServiceServer creates License service
func NewLicenseServiceServer(licenseRepo repo.License) v1.LicenseServiceServer {
	return &licenseServiceServer{licenseRepo: licenseRepo}
}

// func (s *licenseServiceServer) GetProductsbyApplication(ctx context.Context, req *v1.ApplicationRequest) (*v1.ApplicationResponse, error) {

//
// 	if err != nil {
// 		return nil, status.Error(codes.Unknown, "failed to get Products information-> "+err.Error())
// 	}
// 	return res, nil
// }
