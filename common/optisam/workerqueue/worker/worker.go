package worker

import (
	"context"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"
)

// Worker represents a worker for handling Jobs
//
//go:generate mockgen -destination=mock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/worker Worker
type Worker interface {
	// DoWork is called when a worker picks up a job from the queue
	DoWork(context.Context, *job.Job) error
	// ID is a semi-unique identifier for a worker
	// it is primarily used for logging purposes
	ID() string
}
