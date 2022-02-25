package workerqueue

import (
	"context"
	"optisam-backend/common/optisam/workerqueue/job"
	"optisam-backend/common/optisam/workerqueue/worker"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/common/optisam/workerqueue Workerqueue

type Workerqueue interface {
	Close(ctx context.Context)
	RegisterWorker(ctx context.Context, w worker.Worker)
	PushJob(ctx context.Context, j job.Job, workerName string) (int32, error)
	ResumePendingJobs(ctx context.Context) error
	GetRetries() int32
	GetLength() int32
	GetCapacity() int32
	Shrink()
	Grow()
	PopJob() JobChan
	GetIthLength(int) int32
}
