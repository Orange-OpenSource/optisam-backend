package v1

import (
	"optisam-backend/report-service/pkg/repository/v1/postgres/db"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/report-service/pkg/repository/v1 Report

// Report interface
type Report interface {
	db.Querier
}
