// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package worker

import (
	"context"
	"optisam-backend/common/optisam/workerqueue/job"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/common/optisam/workerqueue/worker Worker
//Worker represents a worker for handling Jobs
type Worker interface {
	//DoWork is called when a worker picks up a job from the queue
	DoWork(context.Context, *job.Job) error
	//ID is a semi-unique identifier for a worker
	//it is primarily used for logging purposes
	ID() string
}
