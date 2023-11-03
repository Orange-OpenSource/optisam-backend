package v1

import (
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/repository/v1/postgres/db"
)

//go:generate mockgen -destination=mock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/repository/v1 Report

// Report interface
type Report interface {
	db.Querier
}
