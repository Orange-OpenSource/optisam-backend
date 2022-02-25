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
