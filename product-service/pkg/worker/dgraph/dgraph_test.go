package dgraph

import (
	"context"
	"encoding/json"
	"math/rand"
	"optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

var countTotal int32
var coutFialed int32

func Test_worker_DoWorkAsync(t *testing.T) {
	t.Skip()
	defer func() {
		t.Logf("Total txn:      %d\n", countTotal)
		t.Logf("Failed txn:     %d\n", coutFialed)
		t.Logf("Failed Percent: %.2f\n", (float64(coutFialed)/float64(countTotal))*100)
	}()
	wg := &sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(iter int) {
			do(t, iter)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func Test_worker_DoWorkSync(t *testing.T) {
	t.Skip()
	//	wg := &sync.WaitGroup{}
	defer func() {
		t.Logf("Total txn:      %d\n", countTotal)
		t.Logf("Failed txn:     %d\n", coutFialed)
		t.Logf("Failed Percent: %.2f\n", (float64(coutFialed)/float64(countTotal))*100)
	}()
	for i := 0; i < 1000; i++ {
		// wg.Add(1)
		func(iter int) {
			do(t, iter)
			// wg.Done()
		}(i)
	}
	//	wg.Wait()
}

func do(t *testing.T, i int) {
	is := String(10)
	upr := &v1.UpsertProductRequest{
		SwidTag:  "sw" + is,
		Name:     "p1",
		Category: "c1",
		Edition:  "e1",
		Editor:   "edt1",
		Version:  "v1",
		OptionOf: "sw0" + is,
		Scope:    "x",
		Applications: &v1.UpsertProductRequestApplication{
			Operation:     "add",
			ApplicationId: []string{"app1" + is, "app2" + is},
		},
		Equipments: &v1.UpsertProductRequestEquipment{
			Operation: "add",
			Equipmentusers: []*v1.UpsertProductRequestEquipmentEquipmentuser{
				{
					EquipmentId:    "1" + is,
					AllocatedUsers: 10,
				},
				{
					EquipmentId:    "2" + is,
					AllocatedUsers: 10,
				},
			},
		},
	}

	data, err := json.Marshal(upr)
	if err != nil {
		t.Fatal(err)
		return
	}

	e := &Envelope{
		Type: UpsertProductRequest,
		JSON: json.RawMessage(data),
	}

	data, err = json.Marshal(e)
	if err != nil {
		t.Fatal(err)
		return
	}

	type args struct {
		ctx context.Context
		j   *job.Job
	}
	tests := []struct {
		name    string
		w       *Worker
		args    args
		wantErr bool
	}{
		{name: "check what is wrong",
			w: &Worker{
				dg: dg,
			},
			args: args{
				ctx: context.Background(),
				j: &job.Job{
					Data: json.RawMessage(data),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//	var err error
			for {
				atomic.AddInt32(&countTotal, 1)
				if err := tt.w.DoWork(tt.args.ctx, tt.args.j); (err != nil) != tt.wantErr {
					atomic.AddInt32(&coutFialed, 1)
					//	t.Errorf("worker.DoWork() error = %v, wantErr %v", err, tt.wantErr)
					if err == errRetry {
						continue
					}
				}
				break
			}
		})
	}
}
