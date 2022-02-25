package worker

import (
	"context"
	"optisam-backend/common/optisam/workerqueue/job"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/common/optisam/workerqueue/worker Worker
// Worker represents a worker for handling Jobs
type Worker interface {
	// DoWork is called when a worker picks up a job from the queue
	DoWork(context.Context, *job.Job) error
	// ID is a semi-unique identifier for a worker
	// it is primarily used for logging purposes
	ID() string
}
