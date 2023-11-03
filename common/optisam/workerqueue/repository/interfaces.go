package repository

import (
	db "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/repository/postgres/db"
)

//go:generate mockgen -destination=mock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/repository Workerqueue
type Workerqueue interface {
	db.Querier
}
