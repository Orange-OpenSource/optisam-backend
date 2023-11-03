package dgraph

import v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

type database struct {
	products  []*v1.ProductData
	acqRights []*v1.AcquiredRights
}
