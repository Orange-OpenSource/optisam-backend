package v1

import (
	"context"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/application-service/pkg/api/v1"
	gendb "gitlab.tech.orange/optisam/optisam-it/optisam-services/application-service/pkg/repository/v1/postgres/db"
)

//go:generate mockgen -destination=dbmock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/application-service/pkg/repository/v1 Application
//go:generate mockgen -destination=queuemock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue  Workerqueue

// Application interface
type Application interface {
	gendb.Querier
	UpsertApplicationEquipTx(ctx context.Context, req *v1.UpsertApplicationEquipRequest) error
	UpsertInstanceTX(ctx context.Context, req *v1.UpsertInstanceRequest) error
	DropApplicationDataTX(ctx context.Context, scope string) error
	DropObscolenscenceDataTX(ctx context.Context, scope string) error
}
