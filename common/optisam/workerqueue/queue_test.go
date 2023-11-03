package workerqueue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/repository"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/repository/mock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/repository/postgres/db"
	worker "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/worker"
	workermock "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/worker/mock"

	"github.com/golang/mock/gomock"
)

func TestQueue_RegisterWorker(t *testing.T) {

	var mockCtrl *gomock.Controller
	existingmockworker := workermock.NewMockWorker(mockCtrl)

	// Mock worker
	var w worker.Worker

	// Mock Repo
	var r repository.Workerqueue

	notifier := make(chan JobChan, 100)
	ctx := context.Background()
	// // db, _, err := sqlmock.New()
	// if err != nil {
	// 	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	// }
	var wg sync.WaitGroup
	type fields struct {
		ID        string
		queueSize int
		workers   map[string][]worker.Worker
		wg        *sync.WaitGroup
		PollRate  time.Duration
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		setup             func()
		wantCountWorker   int
		wantJobChanLength int
	}{
		{
			name: "SUCCESS - Register Worker in New Worker Queue with empty notification channel",
			args: args{
				ctx: ctx,
			},
			fields: fields{
				ID:        "test-queue",
				queueSize: 1000,
				PollRate:  time.Duration(500 * time.Millisecond),
				wg:        &wg,
				workers:   make(map[string][]worker.Worker),
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockworker := workermock.NewMockWorker(mockCtrl)
				w = mockworker
				mockworker.EXPECT().ID().AnyTimes().Return("test-worker")
				mockRepo := mock.NewMockWorkerqueue(mockCtrl)
				r = mockRepo
			},
			wantCountWorker:   1,
			wantJobChanLength: 0,
		},
		{
			name: "SUCCESS - Register Worker in Worker Queue with single worker with empty notification channel",
			args: args{
				ctx: ctx,
			},
			fields: fields{
				ID:        "test-queue",
				queueSize: 1000,
				PollRate:  time.Duration(500 * time.Millisecond),
				wg:        &wg,
				workers:   map[string][]worker.Worker{"existing": {existingmockworker}},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockworker := workermock.NewMockWorker(mockCtrl)
				w = mockworker
				mockworker.EXPECT().ID().AnyTimes().Return("test-worker")
				mockRepo := mock.NewMockWorkerqueue(mockCtrl)
				r = mockRepo
			},
			wantCountWorker:   2,
			wantJobChanLength: 0,
		},
		{
			name: "SUCCESS - Register another instance of same Worker in Worker Queue with empty notification channel",
			args: args{
				ctx: ctx,
			},
			fields: fields{
				ID:        "test-queue",
				queueSize: 1000,
				PollRate:  time.Duration(500 * time.Millisecond),
				wg:        &wg,
				workers:   map[string][]worker.Worker{"existing": {existingmockworker}},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockworker := workermock.NewMockWorker(mockCtrl)
				w = mockworker
				mockworker.EXPECT().ID().AnyTimes().Return("existing")
				mockRepo := mock.NewMockWorkerqueue(mockCtrl)
				r = mockRepo
			},
			wantCountWorker:   1,
			wantJobChanLength: 0,
		},
		{
			name: "SUCCESS - Register Worker in Worker Queue with notification channel having 1 pending job",
			args: args{
				ctx: ctx,
			},
			fields: fields{
				ID:        "test-queue",
				queueSize: 1000,
				PollRate:  time.Duration(500 * time.Millisecond),
				wg:        &wg,
				workers:   make(map[string][]worker.Worker),
			},
			setup: func() {
				notifier <- JobChan{jobID: 1, workerName: "t"}
				mockCtrl = gomock.NewController(t)
				mockworker := workermock.NewMockWorker(mockCtrl)
				w = mockworker
				mockworker.EXPECT().ID().AnyTimes().Return("t")
				mockworker.EXPECT().DoWork(ctx, "123").AnyTimes().Return(nil)
				mockRepo := mock.NewMockWorkerqueue(mockCtrl)
				r = mockRepo
				ctx, _ := context.WithCancel(ctx)
				mockRepo.EXPECT().UpdateJobStatusRunning(ctx, db.UpdateJobStatusRunningParams{JobID: 1}).AnyTimes().Return(nil)
			},
			wantCountWorker:   1,
			wantJobChanLength: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("Test name:%s\n", tt.name)
			tt.setup()
			q := &Queue{
				ID:        tt.fields.ID,
				queueSize: tt.fields.queueSize,
				repo:      r,
				workers:   tt.fields.workers,
				wg:        tt.fields.wg,
				PollRate:  tt.fields.PollRate,
			}
			fmt.Printf("Queue ID:%s Worker ID:%s\n", q.ID, w.ID())
			// TODO assert on logs maybe
			q.RegisterWorker(tt.args.ctx, w)
			if len(q.workers) != tt.wantCountWorker {
				t.Errorf("Failed = got %v, want %v", len(q.workers), tt.wantCountWorker)
			}
		})
	}
}
