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
	metadata "google.golang.org/grpc/metadata"
)

type JobChan struct {
	jobId      int32
	workerName string
	jobData    job.Job
	ctx        context.Context
}

/* multilevel queue is conecpt of bucket filling, once
bucket is fulled , new bucket is added to system, and when
old bucket is empty and old bucket is reuesd */
type mlQueue struct {
	notifier  []chan JobChan
	pushIndex int
	popIndex  int
	total     int
	mux       sync.Mutex
}

// Queue represents a queue
type Queue struct {
	//IsWorkerRegCompleted is using for avoid concurrent map read write failure.
	IsWorkerRegCompleted bool

	//IsMultiQueue flag tells wether to move with single level or multilevel
	IsMultiQueue bool
	//ID is a unique identifier for a Queue
	ID string
	//repo represents a handle to a repo struct wrapper to *sql.DB and generated queries
	repo repoInterface.Workerqueue
	//notifier is a chan used to signal workers there is a job to begin working
	mq mlQueue
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
	q.PollRate = time.Duration(100 * time.Millisecond) //Default
	q.queueSize = 10000                                //Default
	q.retries = 3                                      //default
	q.IsMultiQueue = conf.IsMultiQueue
	q.baseDelay = time.Duration(3 * time.Second) //Default

	if conf.PollingRate > 0 {
		q.PollRate = conf.PollingRate
	}
	if conf.Qsize > 0 {
		q.queueSize = conf.Qsize
	}
	if conf.BaseDelay > 0 {
		q.baseDelay = conf.BaseDelay
	}
	if conf.Retries >= 0 {
		q.retries = conf.Retries
	}
	// Multilevel Queue/channel created
	temp := mlQueue{}
	temp.notifier = make([]chan JobChan, 1)
	temp.notifier[0] = make(chan JobChan, q.queueSize)
	temp.total = 1
	q.mq = temp

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
	for i := 0; i < q.mq.total; i++ {
		close(q.mq.notifier[i])
	}
	q.workers = nil
}

//GetRetries return queue conf retry param
func (q *Queue) GetRetries() int32 {
	return int32(q.retries)
}

//GetLength return no of msgs in queue/ith channel
func (q *Queue) GetLength() int32 {
	return int32(len(q.mq.notifier[q.mq.pushIndex]))
}

//GetCapacity return queue's msg holding capacity, default it is 10K
func (q *Queue) GetCapacity() int32 {
	return int32(cap(q.mq.notifier[q.mq.pushIndex]))
}

//RegisterWorker contains the main loop for all Workers.
func (q *Queue) RegisterWorker(ctx context.Context, w worker.Worker) {
	ctx, cancelFunc := context.WithCancel(ctx)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cancelFunc()
			q.wg.Wait()
			q.workers = nil
		}
	}()

	q.workers[w.ID()] = append(q.workers[w.ID()], w)
	q.wg.Add(1)
	//The big __main loop__ for workers.
	go func() {
		var jobC JobChan
		var ok bool
		for {
			if q.IsWorkerRegCompleted == false {
				time.Sleep(5 * time.Second)
				continue
			}
			// receive a notification from the queue chan
			select {
			case <-ctx.Done():
				q.wg.Done()
				return
			case jobC, ok = <-q.mq.notifier[q.mq.popIndex]:
				if ok == false && q.IsMultiQueue {
					q.mq.mux.Lock()
					q.Shrink()
					q.mq.mux.Unlock()
					continue
				}
				logger.Log.Debug("job is poped, queue stats", zap.Any("popIndex", q.mq.popIndex), zap.Any("queueSize", len(q.mq.notifier[q.mq.popIndex])))
				md, _ := metadata.FromIncomingContext(jobC.ctx)
				ctx = metadata.NewOutgoingContext(ctx, md)
				ctx = metadata.NewIncomingContext(ctx, md)

				lenWorker := len(q.workers[jobC.workerName])
				if lenWorker == 0 {
					continue
				}
				worker := q.workers[jobC.workerName][rand.Intn(lenWorker)]
				if jobC.jobData.JobID == 0 {
					repoJob := job.ToRepoJob(&jobC.jobData)
					jobID, err := q.repo.CreateJob(ctx, dbgen.CreateJobParams{Type: repoJob.Type, Status: repoJob.Status, Data: repoJob.Data,
						Comments: repoJob.Comments, StartTime: repoJob.StartTime, EndTime: repoJob.EndTime})
					if err != nil {
						logger.Log.Error("Unable to push job to db: %s, requeueing the job", zap.Error(err))
						q.PushJob(ctx, jobC.jobData, jobC.workerName)
						continue
					}
					jobC.jobId = jobID
					jobC.jobData.JobID = jobID
				}
				processing(ctx, q, worker, jobC)
			default:
				time.Sleep(q.PollRate)
			}
		}
	}()
}

func processing(ctx context.Context, q *Queue, worker worker.Worker, jobC JobChan) {
	for {
		if jobC.jobData.Status == job.JobStatusFAILED {
			err := q.repo.UpdateJobStatusFailed(ctx, dbgen.UpdateJobStatusFailedParams{JobID: jobC.jobId, Status: "FAILED", EndTime: sql.NullTime{Time: time.Now(), Valid: true}, Comments: jobC.jobData.Comments})
			if err != nil {
				logger.Log.Error("Error update status to failed for job: %s  requeued", zap.Error(err))
				q.PushJob(ctx, jobC.jobData, jobC.jobData.Type.String)
			}
			break
		}
		err := worker.DoWork(ctx, &jobC.jobData)
		if err != nil {
			if jobC.jobData.RetryCount.Int32 < int32(q.retries) && int32(q.retries) != 0 {
				q.repo.UpdateJobStatusRetry(ctx, dbgen.UpdateJobStatusRetryParams{
					JobID:  jobC.jobId,
					Status: "RETRY",
				})
				jobC.jobData.RetryCount.Int32++
				logger.Log.Error("Error In Retries ", zap.Error(err), zap.Int32("jobID", jobC.jobId), zap.Int32("retryCount", jobC.jobData.RetryCount.Int32))
				time.Sleep(q.baseDelay * time.Millisecond)
			} else if jobC.jobData.Type.String == "DEFER_WORKER" {
				q.PushJob(ctx, jobC.jobData, jobC.jobData.Type.String)
			} else {
				logger.Log.Error("Retries execceded for ", zap.Int32("jobId", jobC.jobData.JobID))
				err = q.repo.UpdateJobStatusFailed(ctx, dbgen.UpdateJobStatusFailedParams{JobID: jobC.jobId, Status: "FAILED", EndTime: sql.NullTime{Time: time.Now(), Valid: true}, Comments: sql.NullString{String: err.Error(), Valid: true}, RetryCount: sql.NullInt32{Int32: jobC.jobData.RetryCount.Int32, Valid: true}})
				if err != nil {
					logger.Log.Error("Error update status to failed for job: %s", zap.Error(err))
					q.PushJob(ctx, jobC.jobData, jobC.jobData.Type.String)
				}
				break
			}

		} else {
			err = q.repo.UpdateJobStatusCompleted(ctx, dbgen.UpdateJobStatusCompletedParams{JobID: jobC.jobId, Status: "COMPLETED", EndTime: sql.NullTime{Time: time.Now(), Valid: true}})
			if err != nil {
				logger.Log.Error("Failed to Update job", zap.Error(err))
			}
			break
		}
	}
}

//PushJob pushes a job to the queue and notifies workers
func (q *Queue) PushJob(ctx context.Context, j job.Job, workerName string) (int32, error) {
	q.mq.mux.Lock()
	if q.GetLength() == q.GetCapacity() && q.IsMultiQueue {
		q.Grow()
	}
	q.mq.notifier[q.mq.pushIndex] <- JobChan{j.JobID, workerName, j, ctx}
	logger.Log.Debug("queue info", zap.Any("queueno", q.mq.pushIndex), zap.Any("queueLen", q.GetLength()))
	q.mq.mux.Unlock()
	return 0, nil
}

//GetIthLength gives the current msg count in ith length
func (q *Queue) GetIthLength(i int) int32 {
	return int32(len(q.mq.notifier[i]))
}

//Shrink shrinks the queue
func (q *Queue) Shrink() {
	if q.GetIthLength(q.mq.popIndex) == 0 && q.mq.pushIndex > 0 {
		q.mq.notifier = q.mq.notifier[q.mq.popIndex+1:]
		q.mq.total--
		q.mq.pushIndex--
		//log.Printf("After shrink new mq %+v", q.mq)
		logger.Log.Error("WorkerQueue Stats", zap.Any("popindex", q.mq.popIndex), zap.Any("queueSize", len(q.mq.notifier[q.mq.popIndex])), zap.Any("totalChannelList", q.mq.total))
	}
}

//Grow dynmically changes the queue size as per msgs
func (q *Queue) Grow() {
	if q.GetLength() == q.GetCapacity() {
		temp := make(chan JobChan, q.queueSize)
		q.mq.notifier = append(q.mq.notifier, temp)
		close(q.mq.notifier[q.mq.pushIndex])
		q.mq.total++
		q.mq.pushIndex++
		//log.Printf("After grow new mq %+v", q.mq)
		logger.Log.Error("WorkerQueue Stats", zap.Any("pushindex", q.mq.pushIndex), zap.Any("queueSize", len(q.mq.notifier[q.mq.pushIndex-1])), zap.Any("totalChannelList", q.mq.total))
	}

}

//Pop get the data from queue
func (q *Queue) PopJob() JobChan {
	return <-q.mq.notifier[q.mq.popIndex]
}

//CurrentSize tells total msgs in queue
func (q *Queue) CurrentSize() int {
	return len(q.mq.notifier[q.mq.pushIndex])
}

//ResumePendingJobs loops through all pending jobs
func (q *Queue) ResumePendingJobs(ctx context.Context) error {
	jobs, err := q.repo.GetJobsForRetry(ctx)
	if err != nil {
		logger.Log.Error("Error getting jobs from DB %v", zap.Error(err))
		return err
	}
	logger.Log.Debug("Total Resume jobs ", zap.Any("jobs", len(jobs)))
	for _, j := range jobs {
		if j.RetryCount.Int32 < int32(q.retries) {
			job := *(job.FromRepoJob(&j))
			q.PushJob(ctx, job, job.Type.String)
		} else {
			logger.Log.Error("Error already retires execeeded for ", zap.Int32("jobID", j.JobID))
			err = q.repo.UpdateJobStatusCompleted(ctx, dbgen.UpdateJobStatusCompletedParams{JobID: j.JobID, Status: "FAILED", EndTime: sql.NullTime{Time: time.Now(), Valid: true}})
			if err != nil {
				logger.Log.Error("Error update status to failed for job: %s", zap.Error(err))
			}
		}
	}
	return nil
}
