package dgraph

import (
	"context"
	"errors"
	"fmt"
	"log"
	v1 "optisam-backend/metric-service/pkg/repository/v1"
	"reflect"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

func TestMetricRepository_CreateMetricOPS(t *testing.T) {
	type args struct {
		ctx    context.Context
		scopes string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func() (*v1.MetricOPS, func() error, error)
		wantErr bool
	}{
		{name: "success",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: "scope1",
			},
			setup: func() (retMat *v1.MetricOPS, cleanup func() error, retErr error) {
				bottomID := "bottom"
				baseID := "base"
				aggregateID := "aggregate"
				topID := "top"
				coreFactorAttrID := "coreFactor"
				numOfCoresAttrID := "cores"
				numOfCPUsAttrID := "cpu"
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     blankID(bottomID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						{
							Subject:     blankID(baseID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						{
							Subject:     blankID(aggregateID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						{
							Subject:     blankID(topID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						{
							Subject:     blankID(coreFactorAttrID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						{
							Subject:     blankID(numOfCoresAttrID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						{
							Subject:     blankID(numOfCPUsAttrID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
					},
				}
				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, nil, err
				}

				bottomID, ok := assigned.Uids[bottomID]
				if !ok {
					return nil, nil, fmt.Errorf("bottomID is not found in assigned map: %+v", assigned.Uids)
				}

				defer func() {
					if retErr != nil {
						if err := deleteNode(bottomID); err != nil {
							t.Log(err)
						}
					}
				}()

				baseID, ok = assigned.Uids[baseID]
				if !ok {
					return nil, nil, errors.New("baseID is not found in assigned map")
				}

				defer func() {
					if retErr != nil {
						if err := deleteNode(baseID); err != nil {
							t.Log(err)
						}
					}
				}()

				aggregateID, ok = assigned.Uids[aggregateID]
				if !ok {
					return nil, nil, errors.New("aggregateID is not found in assigned map")
				}

				defer func() {
					if retErr != nil {
						if err := deleteNode(aggregateID); err != nil {
							t.Log(err)
						}
					}
				}()

				topID, ok = assigned.Uids[topID]
				if !ok {
					return nil, nil, errors.New("topID is not found in assigned map")
				}

				defer func() {
					if retErr != nil {
						if err := deleteNode(topID); err != nil {
							t.Log(err)
						}
					}
				}()

				coreFactorAttrID, ok = assigned.Uids[coreFactorAttrID]
				if !ok {
					return nil, nil, errors.New("coreFactorAttrID is not found in assigned map")
				}

				defer func() {
					if retErr != nil {
						if err := deleteNode(coreFactorAttrID); err != nil {
							t.Log(err)
						}
					}
				}()

				numOfCPUsAttrID, ok = assigned.Uids[numOfCPUsAttrID]
				if !ok {
					return nil, nil, errors.New("numOfCPUsAttrID is not found in assigned map")
				}

				defer func() {
					if retErr != nil {
						if err := deleteNode(numOfCPUsAttrID); err != nil {
							t.Log(err)
						}
					}
				}()

				numOfCoresAttrID, ok = assigned.Uids[numOfCoresAttrID]
				if !ok {
					return nil, nil, errors.New("numOfCoresAttrID is not found in assigned map")
				}

				defer func() {
					if retErr != nil {
						if err := deleteNode(numOfCoresAttrID); err != nil {
							t.Log(err)
						}
					}
				}()

				return &v1.MetricOPS{
						Name:                  "oracle.processor.standard",
						StartEqTypeID:         bottomID,
						BaseEqTypeID:          baseID,
						AggerateLevelEqTypeID: aggregateID,
						EndEqTypeID:           bottomID,
						CoreFactorAttrID:      coreFactorAttrID,
						NumCoreAttrID:         numOfCoresAttrID,
						NumCPUAttrID:          numOfCPUsAttrID,
					}, func() error {
						return deleteNodes(bottomID, baseID, aggregateID, bottomID, coreFactorAttrID, numOfCoresAttrID, numOfCPUsAttrID)
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
			gotRetMat, err := tt.l.CreateMetricOPS(tt.args.ctx, mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.CreateMetricOPS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				defer func() {
					assert.Empty(t, deleteNode(gotRetMat.ID), "error not expected in deleting metric type")
				}()
				compareMetricOPS(t, "MetricOPS", mat, gotRetMat)
			}
		})
	}
}

func TestMetricRepository_UpdateMetricOPS(t *testing.T) {
	type args struct {
		ctx    context.Context
		met    *v1.MetricOPS
		scopes string
	}
	tests := []struct {
		name  string
		l     *MetricRepository
		args  args
		setup func(l *MetricRepository) (func() error, error)
		//checking func(l *MetricRepository) (*v1.MetricSPSConfig, error)
		wantErr bool
	}{
		{name: "sucess",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: "scope1",
				met: &v1.MetricOPS{
					Name:                  "ops",
					StartEqTypeID:         "start",
					AggerateLevelEqTypeID: "Aggregate",
					BaseEqTypeID:          "zyx",
					CoreFactorAttrID:      "A2",
					EndEqTypeID:           "end",
					NumCoreAttrID:         "andd",
					NumCPUAttrID:          "abc",
				},
			},
			setup: func(l *MetricRepository) (func() error, error) {
				met, err := l.CreateMetricOPS(context.Background(), &v1.MetricOPS{
					Name:                  "ops",
					StartEqTypeID:         "st",
					AggerateLevelEqTypeID: "Agg",
					BaseEqTypeID:          "base",
					CoreFactorAttrID:      "corefactor",
					EndEqTypeID:           "last",
					NumCoreAttrID:         "core",
					NumCPUAttrID:          "cpu",
				}, "scope1")
				if err != nil {
					return func() error {
						return nil
					}, errors.New("error while creating metric ops")
				}
				return func() error {
					assert.Empty(t, deleteNode(met.ID), "error not expected in deleting metric type")
					return nil
				}, nil
			},
			// checking: func(l *MetricRepository) (*v1.MetricSPSConfig, error) {
			// 	actmet, err := l.GetMetricConfigSPS(context.Background(), "sps", "scope1")
			// 	if err != nil {
			// 		return nil, err
			// 	}

			// 	return actmet, nil
			// },
			wantErr: false,
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
			err = tt.l.UpdateMetricOPS(tt.args.ctx, tt.args.met, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.UpdateMetricOPS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !tt.wantErr {
			// 	got, err := tt.checking(tt.l)
			// 	if !assert.Empty(t, err, "not expecting error from checking") {
			// 		return
			// 	}
			// 	compareMetricOPS(t, "MetricRepository.UpdateMetricOPS", tt.args.met, )
			// }
		})
	}
}

func compareMetricOPS(t *testing.T, name string, exp, act *v1.MetricOPS) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metadata is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID should be same", name)
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Source should be same", name)
	assert.Equalf(t, exp.StartEqTypeID, act.StartEqTypeID, "%s.StartEqTypeID should be same", name)
	assert.Equalf(t, exp.BaseEqTypeID, act.BaseEqTypeID, "%s.BaseEqTypeID should be same", name)
	assert.Equalf(t, exp.AggerateLevelEqTypeID, act.AggerateLevelEqTypeID, "%s.AggerateLevelEqTypeID should be same", name)
	assert.Equalf(t, exp.EndEqTypeID, act.EndEqTypeID, "%s.EndEqTypeID should be same", name)
	assert.Equalf(t, exp.CoreFactorAttrID, act.CoreFactorAttrID, "%s.CoreFactorAttrID should be same", name)
	assert.Equalf(t, exp.NumCoreAttrID, act.NumCoreAttrID, "%s.NumCoreAttrID should be same", name)
	assert.Equalf(t, exp.NumCPUAttrID, act.NumCPUAttrID, "%s.NumCPUAttrID should be same", name)
}

func compareMetricOPSAll(t *testing.T, name string, exp, act []*v1.MetricOPS) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareMetricOPS(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func TestMetricRepository_ListMetricOPS(t *testing.T) {
	type args struct {
		ctx    context.Context
		scopes string
	}
	tests := []struct {
		name  string
		l     *MetricRepository
		args  args
		setup func() ([]*v1.MetricOPS, func() error, error)
		// want    []*v1.MetricOPS
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: "Scope1",
			},
			setup: func() (retMat []*v1.MetricOPS, cleanup func() error, retErr error) {
				retMat = []*v1.MetricOPS{}
				mat1 := &v1.MetricOPS{
					StartEqTypeID:         "start1",
					EndEqTypeID:           "end1",
					BaseEqTypeID:          "base1",
					AggerateLevelEqTypeID: "agg1",
					NumCoreAttrID:         "core1",
					NumCPUAttrID:          "cpu1",
					CoreFactorAttrID:      "cores1",
				}
				mat2 := &v1.MetricOPS{
					StartEqTypeID:         "start2",
					EndEqTypeID:           "end2",
					BaseEqTypeID:          "base2",
					AggerateLevelEqTypeID: "agg2",
					NumCoreAttrID:         "core2",
					NumCPUAttrID:          "cpu2",
					CoreFactorAttrID:      "cores2",
				}
				mat1, cleanup1, err := createMetric(mat1)
				if err != nil {
					return nil, nil, errors.New("error while creating metric 1")
				}
				mat2, cleanup2, err := createMetric(mat2)
				if err != nil {
					return nil, nil, errors.New("error while creating metric 2")
				}
				retMat = append(retMat, mat1, mat2)
				return retMat, func() error {
					err := cleanup1()
					if err != nil {
						return err
					}
					return cleanup2()
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
			got, err := tt.l.ListMetricOPS(tt.args.ctx, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.ListMetricOPS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricOPSAll(t, "MetricOPS", mat, got)
			}
		})
	}
}

func createMetric(mat *v1.MetricOPS) (retMat *v1.MetricOPS, cleanup func() error, retErr error) {
	bottomID := mat.StartEqTypeID
	baseID := mat.BaseEqTypeID
	aggregateID := mat.AggerateLevelEqTypeID
	topID := mat.EndEqTypeID
	coreFactorAttrID := mat.CoreFactorAttrID
	numOfCoresAttrID := mat.NumCoreAttrID
	numOfCPUsAttrID := mat.NumCoreAttrID
	mu := &api.Mutation{
		CommitNow: true,
		Set: []*api.NQuad{
			{
				Subject:     blankID(bottomID),
				Predicate:   "type_name",
				ObjectValue: stringObjectValue("metadata"),
			},
			{
				Subject:     blankID(baseID),
				Predicate:   "type_name",
				ObjectValue: stringObjectValue("metadata"),
			},
			{
				Subject:     blankID(aggregateID),
				Predicate:   "type_name",
				ObjectValue: stringObjectValue("metadata"),
			},
			{
				Subject:     blankID(topID),
				Predicate:   "type_name",
				ObjectValue: stringObjectValue("metadata"),
			},
			{
				Subject:     blankID(coreFactorAttrID),
				Predicate:   "type_name",
				ObjectValue: stringObjectValue("metadata"),
			},
			{
				Subject:     blankID(numOfCoresAttrID),
				Predicate:   "type_name",
				ObjectValue: stringObjectValue("metadata"),
			},
			{
				Subject:     blankID(numOfCPUsAttrID),
				Predicate:   "type_name",
				ObjectValue: stringObjectValue("metadata"),
			},
		},
	}
	assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
	if err != nil {
		return nil, nil, err
	}

	bottomID, ok := assigned.Uids[bottomID]
	if !ok {
		return nil, nil, fmt.Errorf("bottomID is not found in assigned map: %+v", assigned.Uids)
	}

	defer func() {
		if retErr != nil {
			if err := deleteNode(bottomID); err != nil {
				// t.Log(err)
			}
		}
	}()

	baseID, ok = assigned.Uids[baseID]
	if !ok {
		return nil, nil, errors.New("baseID is not found in assigned map")
	}

	defer func() {
		if retErr != nil {
			if err := deleteNode(baseID); err != nil {
				// t.Log(err)
			}
		}
	}()

	aggregateID, ok = assigned.Uids[aggregateID]
	if !ok {
		return nil, nil, errors.New("aggregateID is not found in assigned map")
	}

	defer func() {
		if retErr != nil {
			if err := deleteNode(aggregateID); err != nil {
				// t.Log(err)
			}
		}
	}()

	topID, ok = assigned.Uids[topID]
	if !ok {
		return nil, nil, errors.New("topID is not found in assigned map")
	}

	defer func() {
		if retErr != nil {
			if err := deleteNode(topID); err != nil {
				// t.Log(err)
			}
		}
	}()

	coreFactorAttrID, ok = assigned.Uids[coreFactorAttrID]
	if !ok {
		return nil, nil, errors.New("coreFactorAttrID is not found in assigned map")
	}

	defer func() {
		if retErr != nil {
			if err := deleteNode(coreFactorAttrID); err != nil {
				// t.Log(err)
			}
		}
	}()

	numOfCPUsAttrID, ok = assigned.Uids[numOfCPUsAttrID]
	if !ok {
		return nil, nil, errors.New("numOfCPUsAttrID is not found in assigned map")
	}

	defer func() {
		if retErr != nil {
			if err := deleteNode(numOfCPUsAttrID); err != nil {
				// t.Log(err)
			}
		}
	}()

	numOfCoresAttrID, ok = assigned.Uids[numOfCoresAttrID]
	if !ok {
		return nil, nil, errors.New("numOfCoresAttrID is not found in assigned map")
	}

	defer func() {
		if retErr != nil {
			if err := deleteNode(numOfCoresAttrID); err != nil {
				// t.Log(err)
			}
		}
	}()
	repo := NewMetricRepository(dgClient)
	gotRetMat, err := repo.CreateMetricOPS(context.Background(), &v1.MetricOPS{
		Name:                  "oracle.processor.standard",
		StartEqTypeID:         bottomID,
		BaseEqTypeID:          baseID,
		AggerateLevelEqTypeID: aggregateID,
		EndEqTypeID:           topID,
		CoreFactorAttrID:      coreFactorAttrID,
		NumCoreAttrID:         numOfCoresAttrID,
		NumCPUAttrID:          numOfCPUsAttrID,
	}, "Scope1")
	return gotRetMat, func() error {
		// return nil
		return deleteNodes(gotRetMat.ID, bottomID, baseID, aggregateID, bottomID, coreFactorAttrID, numOfCoresAttrID, numOfCPUsAttrID)
	}, nil
}

func TestMetricRepository_GetMetricConfigOPS(t *testing.T) {
	type args struct {
		ctx     context.Context
		metName string
		scopes  string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		want    *v1.MetricOPSConfig
		wantErr bool
		setup   func(string) (map[string]string, error)
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				metName: "dummyOps1",
				ctx:     context.Background(),
				scopes:  "scope1",
			},
			setup: func(metName string) (ids map[string]string, retErr error) {
				ids, err := addMetricConfig(metName)
				if err != nil {
					t.Errorf("Failed to create config of OPS metric, err : %v", err)
				}
				return
			},
			want: &v1.MetricOPSConfig{
				Name:                "dummyOps1",
				NumCoreAttr:         "8",
				NumCPUAttr:          "4",
				CoreFactorAttr:      "1",
				BaseEqType:          "container",
				AggerateLevelEqType: "server",
				EndEqType:           "vcenter",
				StartEqType:         "partition",
			},
			wantErr: false,
		},
		{name: "SUCCESS_WITH_NO_DATA",
			l: NewMetricRepository(dgClient),
			args: args{
				metName: "dummyOps2",
				ctx:     context.Background(),
				scopes:  "scope1",
			},
			setup:   func(metName string) (ids map[string]string, retErr error) { return },
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, _ := tt.setup("dummyOps1")
			got, err := tt.l.GetMetricConfigOPS(tt.args.ctx, tt.args.metName, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.GetMetricOPS() error = %v, wantErr %v", err, tt.wantErr)
				deleteMetricConfig(ids)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricRepository.GetMetricOPS() got = %+v, want %+v", got, tt.want)
			}
			log.Println("CALLING FOR DELETE ", ids)
			deleteMetricConfig(ids)
		})

	}
}

func addMetricConfig(metName string) (ids map[string]string, err error) {

	mu := &api.Mutation{
		CommitNow: true,
		Set: []*api.NQuad{
			{
				Subject:     blankID("metric"),
				Predicate:   "metric.name",
				ObjectValue: stringObjectValue(metName),
			},
			{
				Subject:     blankID("metric"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("Metric"),
			},
			{
				Subject:     blankID("metric"),
				Predicate:   "scopes",
				ObjectValue: stringObjectValue("scope1"),
			},
			{
				Subject:   blankID("metric"),
				Predicate: "metric.ops.bottom",
				ObjectId:  "_:metadata1",
			},
			{
				Subject:     blankID("metadata1"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("metadata"),
			},
			{
				Subject:   blankID("metric"),
				Predicate: "metric.ops.top",
				ObjectId:  "_:metadata2",
			},
			{
				Subject:     blankID("metadata2"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("metadata"),
			},
			{
				Subject:   blankID("metric"),
				Predicate: "metric.ops.aggregate",
				ObjectId:  "_:metadata3",
			},
			{
				Subject:     blankID("metadata3"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("metadata"),
			},
			{
				Subject:   blankID("metric"),
				Predicate: "metric.ops.base",
				ObjectId:  "_:metadata4",
			},
			{
				Subject:     blankID("metadata4"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("metadata"),
			},
			{
				Subject:   blankID("metric"),
				Predicate: "metric.ops.attr_core_factor",
				ObjectId:  "_:attribute1",
			},
			{
				Subject:     blankID("attribute1"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("attr"),
			},
			{
				Subject:   blankID("metric"),
				Predicate: "metric.ops.attr_num_cpu",
				ObjectId:  "_:attribute2",
			},
			{
				Subject:     blankID("attribute2"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("attr"),
			},
			{
				Subject:   blankID("metric"),
				Predicate: "metric.ops.attr_num_cores",
				ObjectId:  "_:attribute3",
			},
			{
				Subject:     blankID("attribute3"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("attr"),
			},
			{
				Subject:     blankID("metadata1"),
				Predicate:   "metadata.equipment.type",
				ObjectValue: stringObjectValue("partition"),
			},
			{
				Subject:     blankID("metadata2"),
				Predicate:   "metadata.equipment.type",
				ObjectValue: stringObjectValue("vcenter"),
			},
			{
				Subject:     blankID("metadata3"),
				Predicate:   "metadata.equipment.type",
				ObjectValue: stringObjectValue("server"),
			},
			{
				Subject:     blankID("metadata4"),
				Predicate:   "metadata.equipment.type",
				ObjectValue: stringObjectValue("container"),
			},
			{
				Subject:     blankID("attribute1"),
				Predicate:   "attribute.name",
				ObjectValue: stringObjectValue("1"),
			},
			{
				Subject:     blankID("attribute2"),
				Predicate:   "attribute.name",
				ObjectValue: stringObjectValue("4"),
			},
			{
				Subject:     blankID("attribute3"),
				Predicate:   "attribute.name",
				ObjectValue: stringObjectValue("8"),
			},
		},
	}
	assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)

	return (*assigned).Uids, err
}

func deleteMetricConfig(uids map[string]string) error {
	for _, uid := range uids {
		if err := deleteNode(uid); err != nil {
			log.Println("Failed to delete the node, Id", uid)
			return err
		}
	}
	return nil
}
