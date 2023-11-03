package dgraph

import (
	"context"
	"errors"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

// CreateMetricDataCenterForScope handles datacenter scope metric creation
// func TestMetricRepository_CreateMetricWindowServerStandard(t *testing.T) {
// 	type args struct {
// 		ctx context.Context
// 		met *v1.MetricWSS
// 	}
// 	tests := []struct {
// 		name       string
// 		l          *MetricRepository
// 		args       args
// 		wantRetmet *v1.MetricWSS
// 		wantErr    bool
// 	}{
// 		{
// 			name: "sucess",
// 			l:    NewMetricRepository(dgClient),
// 			args: args{
// 				ctx: context.Background(),
// 				met: &v1.MetricWSS{
// 					MetricType: "windows.server.datacenter",
// 					MetricName: "windows.server.datacenter.2016",
// 					Reference:  "server",
// 					Core:       "cores_per_processor",
// 					CPU:        "server_processors_numbers",
// 					Default:    false,
// 					Scope:      "scope1",
// 				},
// 			},
// 			wantRetmet: &v1.MetricWSS{
// 				MetricType: "windows.server.datacenter",
// 				MetricName: "windows.server.datacenter.2016",
// 				Reference:  "server",
// 				Core:       "cores_per_processor",
// 				CPU:        "server_processors_numbers",
// 				Default:    false,
// 				Scope:      "scope1",
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotRetmet, err := tt.l.CreateMetricWindowServerStandard(tt.args.ctx, tt.args.met)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("MetricRepository.CreateMetricUSS() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !tt.wantErr {
// 				defer func() {
// 					assert.Empty(t, deleteNode(gotRetmet.ID), "error not expected in deleting metric type")
// 				}()
// 			}
// 		})
// 	}
// }

// CreateMetricDataCenterForScope handles datacenter scope metric creation
func TestMetricRepository_CreateMetricWindowServerStandard(t *testing.T) {
	type args struct {
		ctx    context.Context
		mat    *v1.MetricWSS
		scopes string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func() (*v1.MetricWSS, func() error, error)
		wantErr bool
	}{
		{name: "sucess",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: "scope1",
			},
			setup: func() (retMat *v1.MetricWSS, cleanup func() error, retErr error) {

				reference := "base"
				CPU := "cpu"
				core := "cores"

				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{

						{
							Subject:     blankID(reference),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},

						{
							Subject:     blankID(CPU),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						{
							Subject:     blankID(core),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
					},
				}
				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, nil, err
				}

				reference, ok := assigned.Uids[reference]
				if !ok {
					return nil, nil, errors.New("reference is not found in assigned map")
				}

				defer func() {
					if retErr != nil {
						if err := deleteNode(reference); err != nil {
							t.Log(err)
						}
					}
				}()

				CPU, ok = assigned.Uids[CPU]
				if !ok {
					return nil, nil, errors.New("CPU is not found in assigned map")
				}

				defer func() {
					if retErr != nil {
						if err := deleteNode(CPU); err != nil {
							t.Log(err)
						}
					}
				}()

				core, ok = assigned.Uids[core]
				if !ok {
					return nil, nil, errors.New("core is not found in assigned map")
				}

				defer func() {
					if retErr != nil {
						if err := deleteNode(core); err != nil {
							t.Log(err)
						}
					}
				}()

				return &v1.MetricWSS{
						MetricName: "ibm.pvu.standard",
						Reference:  reference,
						CPU:        CPU,
						Core:       core,
					}, func() error {
						return deleteNodes(reference, CPU, core)
					}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mat, cleanup, err := tt.setup()
			if !assert.Empty(t, err, "not expecting error from setup") {
				return
			}
			defer func() {
				assert.Empty(t, cleanup(), "not expecting error in setup")
			}()
			gotRetMat, err := tt.l.CreateMetricWindowServerStandard(tt.args.ctx, mat)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.CreateMetricWindowServerStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				defer func() {
					assert.Empty(t, deleteNode(gotRetMat.ID), "error not expected in deleting metric type")
				}()
				compareMetricWSS(t, "MetricWSS", mat, gotRetMat)
			}
		})
	}
}

func compareMetricWSS(t *testing.T, name string, exp, act *v1.MetricWSS) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metadata is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID should be same", name)
	}

	assert.Equalf(t, exp.MetricName, act.MetricName, "%s.Source should be same", name)
	assert.Equalf(t, exp.Reference, act.Reference, "%s.Reference should be same", name)
	assert.Equalf(t, exp.Core, act.Core, "%s.core should be same", name)
	assert.Equalf(t, exp.CPU, act.CPU, "%s.CPU should be same", name)
}

func TestMetricRepository_GetMetricConfigWindowServerStandard(t *testing.T) {
	type args struct {
		ctx     context.Context
		metName string
		scopes  string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func(l *MetricRepository) (func() error, error)
		want    *v1.MetricWSS
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				metName: "windows.server.standard.2016",
				scopes:  "scope1",
			},
			setup: func(l *MetricRepository) (func() error, error) {
				met1, err := l.CreateMetricWindowServerStandard(context.Background(), &v1.MetricWSS{
					MetricName: "windows.server.standard.2016",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scope:      "scope1",
					Default:    true,
				})
				if err != nil {
					return func() error {
						return nil
					}, errors.New("error while creating metric 1")
				}
				return func() error {
					assert.Empty(t, deleteNode(met1.ID), "error not expected in deleting metric type")
					return nil
				}, nil
			},
			want: &v1.MetricWSS{
				MetricName: "windows.server.standard.2016",
				Reference:  "server",
				Core:       "cores_per_processor",
				CPU:        "server_processors_numbers",
				Scope:      "scope1",
				Default:    true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup(tt.l)
			if !assert.Empty(t, err, "not expecting error from setup") {
				return
			}
			defer func() {
				assert.Empty(t, cleanup(), "not expecting error in setup")
			}()
			got, err := tt.l.GetMetricConfigWindowServerStandard(tt.args.ctx, tt.args.metName, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.GetMetricConfigWindowServerStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricWSS(t, "MetricRepository.GetMetricConfigWindowServerStandard", tt.want, got)
			}
		})
	}
}
