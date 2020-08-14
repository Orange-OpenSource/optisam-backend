// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package workerqueue

import (
	"context"
	"database/sql"
	"math/rand"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	repoInterface "optisam-backend/common/optisam/workerqueue/repository"
	repo "optisam-backend/common/optisam/workerqueue/repository/postgres"
	dbgen "optisam-backend/common/optisam/workerqueue/repository/postgres/db"
	"optisam-backend/common/optisam/workerqueue/worker"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.uber.org/zap"
)

type jobChan struct {
	jobId      int32
	workerName string
}

// Queue represents a queue
type Queue struct {
	//ID is a unique identifier for a Queue
	ID string
	//repo represents a handle to a repo struct wrapper to *sql.DB and generated queries
	repo repoInterface.Workerqueue
	//notifier is a chan used to signal workers there is a job to begin working
	notifier chan jobChan
	//queueSize is the size of notification channel
	queueSize int

	//no of max attempt on failed job in queue
	retries int

	//exponential backoff for retires
	baseDelay time.Duration
	//workers is a list of *Workers
	workers map[string][]worker.Worker

	//wg is used to help gracefully shutdown workers
	wg *sync.WaitGroup

	//PollRate the duration to Sleep each worker before checking the queue for jobs again
	//queue for jobs again.
	PollRate time.Duration
}

//NewQueue creates a connection to the internal database and initializes the Queue type
func NewQueue(ctx context.Context, queueID string, db *sql.DB, conf QueueConfig) (*Queue, error) {
	q := &Queue{ID: queueID}
	q.repo = repo.NewRepository(db)
	q.PollRate = time.Duration(100 * time.Millisecond)  //Default
	q.queueSize = 1000                                  //Default
	q.retries = 3                                       //default
	q.baseDelay = time.Duration(100 * time.Millisecond) //Default

	if conf.PollingRate > 0 {
		q.PollRate = conf.PollingRate
	}
	if conf.Qsize > 0 {
		q.queueSize = conf.Qsize
	}
	if conf.BaseDelay > 0 {
		q.baseDelay = conf.BaseDelay
	}
	if conf.Retries > 0 {
		q.retries = conf.Retries
	}
	// Make notification channels
	c := make(chan jobChan, q.queueSize) //TODO: channel probably isn't the best way to handle the queue buffer
	q.notifier = c
	m := make(map[string][]worker.Worker)
	q.workers = m
	var wg sync.WaitGroup
	q.wg = &wg
	//resume stopped jobs
	err := q.ResumePendingJobs(ctx)
	if err != nil {
		logger.Log.Error("Unable to resume jobs from bucket: %s", zap.Error(err))
		//Don't fail out, this isn't really fatal. But maybe it should be?
	}
	return q, nil
}

//Close attempts to gracefull shutdown all workers in a queue and shutdown the db connection
func (q *Queue) Close(ctx context.Context) {
	q.wg.Wait()
	close(q.notifier)
	q.workers = nil
}

//GetRetries return queue conf retry param
func (q *Queue) GetRetries() int32 {
	return int32(q.retries)
}

//RegisterWorker contains the main loop for all Workers.
func (q *Queue) RegisterWorker(ctx context.Context, w worker.Worker) {
	logger.Log.Info("", zap.String("Registering worker with ID", w.ID()))
	ctx, cancelFunc := context.WithCancel(ctx)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			logger.Log.Info("Shutdown Signal Receieved - QUEUE")
			cancelFunc()
			q.wg.Wait()
			q.workers = nil
		}
	}()

	q.workers[w.ID()] = append(q.workers[w.ID()], w)
	q.wg.Add(1)
	//The big __main loop__ for workers.
	go func() {
		logger.Log.Info("Starting up new worker...")
		var jobC jobChan
		for {
			// receive a notification from the queue chan
			select {
			case <-ctx.Done():
				logger.Log.Info("Received signal to shutdown worker. Exiting.")
				q.wg.Done()
				return
			case jobC = <-q.notifier:
				lenWorker := len(q.workers[jobC.workerName])
				if lenWorker == 0 {
					return
				}
				worker := q.workers[jobC.workerName][rand.Intn(lenWorker)]
				logger.Log.Info("", zap.Int32("Received Job", jobC.jobId), zap.String("Picked By worker", worker.ID()))
				// err := q.repo.UpdateJobStatusRunning(ctx, dbgen.UpdateJobStatusRunningParams{JobID: jobC.jobId, Status: "RUNNING", StartTime: sql.NullTime{Time: time.Now(), Valid: true}})
				// if err != nil {
				// 	logger.Log.Error("Unable to update job status: %s", zap.Error(err))
				// 	continue
				// }
				//If subsequent calls to updateJobStatus fail, the whole thing is probably hosed and
				//it should probably do something more drastic for error handling.
				j, err := q.repo.GetJob(ctx, jobC.jobId)
				if err != nil {
					logger.Log.Error("Error processing job: %s", zap.Error(err))
					q.repo.UpdateJobStatusCompleted(ctx, dbgen.UpdateJobStatusCompletedParams{JobID: jobC.jobId, Status: "FAILED", EndTime: sql.NullTime{Time: time.Now(), Valid: true}})
					if err != nil {
						logger.Log.Error("Error update status to failed for job: %s", zap.Error(err))
					}
					continue
				}
				// Call the worker func handling this job
				// go func() {

				err = worker.DoWork(ctx, job.FromRepoJob(&j))
				if err != nil {
					if j.RetryCount.Int32 < int32(q.retries) {
						logger.Log.Error("Retry error received from worker retrying ", zap.Error(err), zap.Int32("jobID", j.JobID), zap.Int32("retryCount", j.RetryCount.Int32+1))
						err = q.repo.UpdateJobStatusRetry(ctx, dbgen.UpdateJobStatusRetryParams{JobID: jobC.jobId, Status: "RETRY"})
						if err != nil {
							logger.Log.Error("Failed to Update job", zap.Error(err))
						}
						time.Sleep(q.baseDelay * time.Duration(j.RetryCount.Int32) * time.Millisecond)
						q.notifier <- jobChan{jobC.jobId, jobC.workerName}
					} else {
						logger.Log.Error("Retries execceded for ", zap.Int32("jobId", j.JobID))
					}

				} else {
					logger.Log.Info("Worker", zap.Int32("Job Processed", jobC.jobId))
					err = q.repo.UpdateJobStatusCompleted(ctx, dbgen.UpdateJobStatusCompletedParams{JobID: jobC.jobId, Status: "COMPLETED", EndTime: sql.NullTime{Time: time.Now(), Valid: true}})
					if err != nil {
						logger.Log.Error("Failed to Update job", zap.Error(err))
					}
				}
				// }()
				logger.Log.Info("Finished processing job ", zap.Int32("jobID", jobC.jobId))
			default:
				//logger.Log.Info("Worker: %s. No message to queue. Sleeping 500ms", w.ID())
				//logger.Log.Info("Worker: %s. No message to queue. Sleeping 1s", w.ID())
				time.Sleep(q.PollRate)
			}
		}
	}()
	// time.Sleep(100 * time.Millisecond)
}

//PushJob pushes a job to the queue and notifies workers
func (q *Queue) PushJob(ctx context.Context, j job.Job, workerName string) (int32, error) {
	repoJob := job.ToRepoJob(&j)
	jobID, err := q.repo.CreateJob(ctx, dbgen.CreateJobParams{Type: repoJob.Type, Status: repoJob.Status, Data: repoJob.Data,
		Comments: repoJob.Comments, StartTime: repoJob.StartTime, EndTime: repoJob.EndTime})
	if err != nil {
		logger.Log.Error("Unable to push job to queue: %s", zap.Error(err))
		return 0, err
	}
	q.notifier <- jobChan{jobID, workerName}
	return jobID, nil
}

/*
//PushJob pushes a job to the queue and notifies workers
func (q *Queue) PushJobs(ctx context.Context, j []job.Job, workerName string) (int, error) {
	query := getBulkJobQuery(j)
	jobIDs, err := q.repo.CreateJobs(ctx, query)
	if err != nil {
		logger.Log.Error("Failed To push jobs in bulk err %v", zap.Error(err))
		return 0, err
	}
	for _, jobID := range jobIDs {
		logger.Log.Info("Job pushed ", zap.Int("jobID", jobID))
		q.notifier <- jobChan{int32(jobID), workerName}
	}
	return len(jobIDs), nil
}

func getBulkJobQuery(j []job.Job) (query string) {
	query = "insert into jobs (type,status,data,comments) values  "
	for _, data := range j {
		query += fmt.Sprintf("('%s','%s','%s','%s') ,", data.Type.String, data.Status, string(data.Data), data.Comments.String)
	}
	query = query[0 : len(query)-1]
	query += " returning job_id ;"
	return
}

//CurrentSize tells total msgs in queue
func (q *Queue) CurrentSize() int {
	return len(q.notifier)
}
*/

//ResumePendingJobs loops through all pending jobs
func (q *Queue) ResumePendingJobs(ctx context.Context) error {
	jobs, err := q.repo.GetJobs(ctx)
	if err != nil {
		logger.Log.Error("Error getting jobs from DB %v", zap.Error(err))
		return err
	}
	for _, j := range jobs {
		if j.Status == "PENDING" || j.Status == "RETRY" || j.Status == "RUNNING" {
			if j.RetryCount.Int32 < int32(q.retries) {
				logger.Log.Info("Job not processed. Retrying...", zap.Int32("jobID", j.JobID))
				q.notifier <- jobChan{j.JobID, j.Type}
			} else {
				logger.Log.Error("Error already retires execeeded for ", zap.Int32("jobID", j.JobID))
			}
		}
	}
	return nil
}
