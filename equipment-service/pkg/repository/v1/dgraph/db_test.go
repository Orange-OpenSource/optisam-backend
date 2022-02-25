package dgraph

import v1 "optisam-backend/equipment-service/pkg/repository/v1"

type database struct {
	products  []*v1.ProductData
	acqRights []*v1.AcquiredRights
}
