package repository

import (
	db "optisam-backend/common/optisam/workerqueue/repository/postgres/db"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/common/optisam/workerqueue/repository Workerqueue
type Workerqueue interface {
	db.Querier
}
