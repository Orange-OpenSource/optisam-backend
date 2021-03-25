// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	v1 "optisam-backend/application-service/pkg/api/v1"
	gendb "optisam-backend/application-service/pkg/repository/v1/postgres/db"
)

//go:generate mockgen -destination=dbmock/mock.go -package=mock optisam-backend/application-service/pkg/repository/v1 Application
//go:generate mockgen -destination=queuemock/mock.go -package=mock optisam-backend/common/optisam/workerqueue  Workerqueue

//Application interface
type Application interface {
	gendb.Querier
	UpsertInstanceTX(ctx context.Context, req *v1.UpsertInstanceRequest) error
	DropApplicationDataTX(ctx context.Context, scope string) error
}
