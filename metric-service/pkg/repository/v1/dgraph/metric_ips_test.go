package dgraph

import (
	"context"
	"errors"
	v1 "optisam-backend/metric-service/pkg/repository/v1"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

func TestMetricRepository_CreateMetricIPS(t *testing.T) {
	type args struct {
		ctx    context.Context
		mat    *v1.MetricIPS
		scopes string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func() (*v1.MetricIPS, func() error, error)
		wantErr bool
	}{
		{name: "sucess",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: "scope1",
			},
			setup: func() (retMat *v1.MetricIPS, cleanup func() error, retErr error) {

				baseID := "base"
				coreFactorAttrID := "coreFactor"
				numOfCoresAttrID := "cores"

				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{

						{
							Subject:     blankID(baseID),
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
					},
				}
				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, nil, err
				}

				baseID, ok := assigned.Uids[baseID]
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

				return &v1.MetricIPS{
						Name:             "ibm.pvu.standard",
						BaseEqTypeID:     baseID,
						CoreFactorAttrID: coreFactorAttrID,
						NumCoreAttrID:    numOfCoresAttrID,
					}, func() error {
						return deleteNodes(baseID, coreFactorAttrID, numOfCoresAttrID)
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
			gotRetMat, err := tt.l.CreateMetricIPS(tt.args.ctx, mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.CreateMetricIPS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				defer func() {
					assert.Empty(t, deleteNode(gotRetMat.ID), "error not expected in deleting metric type")
				}()
				compareMetricIPS(t, "MetricOPS", mat, gotRetMat)
			}
		})
	}
}

func TestMetricRepository_GetMetricConfigIPS(t *testing.T) {
	type args struct {
		ctx     context.Context
		metName string
		scopes  string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func() (func() error, error)
		want    *v1.MetricIPSConfig
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				metName: "ips1",
				scopes:  "scope1",
			},
			setup: func() (func() error, error) {
				ids, err := addMetricIPSConfig("ips1", "scope1")
				if err != nil {
					t.Errorf("Failed to create config of SPS metric, err : %v", err)
				}
				return func() error {
					return deleteMetricConfig(ids)
				}, nil
			},
			want: &v1.MetricIPSConfig{
				Name:           "ips1",
				NumCoreAttr:    "ips_cores",
				CoreFactorAttr: "ips_corefactor",
				BaseEqType:     "server",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup()
			if !assert.Empty(t, err, "not expecting error from setup") {
				return
			}
			defer func() {
				assert.Empty(t, cleanup(), "not expecting error in setup")
			}()
			got, err := tt.l.GetMetricConfigIPS(tt.args.ctx, tt.args.metName, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.GetMetricConfigIPS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricIPSConfig(t, "MetricRepository.GetMetricConfigIPS", tt.want, got)
			}
		})
	}
}

func TestMetricRepository_UpdateMetricIPS(t *testing.T) {
	baseID := "base"
	coreFactorAttrID := "coreFactor"
	numOfCoresAttrID := "cores"
	coreFactorAttrID1 := "corefactor1"

	mu := &api.Mutation{
		CommitNow: true,
		Set: []*api.NQuad{

			{
				Subject:     blankID(baseID),
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
				Subject:     blankID(coreFactorAttrID1),
				Predicate:   "type_name",
				ObjectValue: stringObjectValue("metadata"),
			},
		},
	}
	assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
	if err != nil {
		t.Log(err)
		return
	}

	baseID, ok := assigned.Uids[baseID]
	if !ok {
		t.Log(errors.New("baseID is not found in assigned map"))
		if err := deleteNode(baseID); err != nil {
			t.Log(err)
		}
		return
	}

	coreFactorAttrID, ok = assigned.Uids[coreFactorAttrID]
	if !ok {
		t.Log(errors.New("coreFactorAttrID is not found in assigned map"))
		if err := deleteNode(coreFactorAttrID); err != nil {
			t.Log(err)
		}
		return
	}

	numOfCoresAttrID, ok = assigned.Uids[numOfCoresAttrID]
	if !ok {
		t.Log(errors.New("numOfCoresAttrID is not found in assigned map"))
		if err := deleteNode(numOfCoresAttrID); err != nil {
			t.Log(err)
		}
		return
	}

	coreFactorAttrID1, ok = assigned.Uids[coreFactorAttrID1]
	if !ok {
		t.Log(errors.New("coreFactorAttrID1 is not found in assigned map"))
		if err := deleteNode(coreFactorAttrID1); err != nil {
			t.Log(err)
		}
		return
	}
	type args struct {
		ctx    context.Context
		met    *v1.MetricIPS
		scopes string
	}
	tests := []struct {
		name  string
		l     *MetricRepository
		args  args
		setup func(l *MetricRepository) (func() error, error)
		//checking func(l *MetricRepository) (*v1.MetricIPSConfig, error)
		wantErr bool
	}{
		{name: "sucess",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: "scope1",
				met: &v1.MetricIPS{
					Name:             "ips",
					BaseEqTypeID:     baseID,
					CoreFactorAttrID: coreFactorAttrID1,
					NumCoreAttrID:    numOfCoresAttrID,
				},
			},
			setup: func(l *MetricRepository) (cleanup func() error, retErr error) {
				_, err := l.CreateMetricIPS(context.Background(), &v1.MetricIPS{
					Name:             "ips",
					BaseEqTypeID:     baseID,
					CoreFactorAttrID: coreFactorAttrID,
					NumCoreAttrID:    numOfCoresAttrID,
				}, "scope1")
				if err != nil {
					return func() error {
						return nil
					}, errors.New("error while creating metric ips")
				}
				return func() error {
					return deleteNodes(baseID, coreFactorAttrID, numOfCoresAttrID)
				}, nil
			},
			// checking: func(l *MetricRepository) (*v1.MetricIPSConfig, error) {
			// 	actmet, err := l.GetMetricConfigIPS(context.Background(), "ips", "scope1")
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
			err = tt.l.UpdateMetricIPS(tt.args.ctx, tt.args.met, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.UpdateMetricIPS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !tt.wantErr {
			// 	got, err := tt.checking(tt.l)
			// 	if !assert.Empty(t, err, "not expecting error from checking") {
			// 		return
			// 	}
			// 	compareMetricIPS(t, "MetricRepository.UpdateMetricIPS", tt.args.met, )
			// }
		})
	}
}

func addMetricIPSConfig(metName string, scope string) (ids map[string]string, err error) {

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
				Subject:   blankID("metric"),
				Predicate: "metric.ips.base",
				ObjectId:  "_:metadata1",
			},
			{
				Subject:     blankID("metadata1"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("metadata"),
			},
			{
				Subject:     blankID("metadata1"),
				Predicate:   "metadata.equipment.type",
				ObjectValue: stringObjectValue("server"),
			},
			{
				Subject:   blankID("metric"),
				Predicate: "metric.ips.attr_core_factor",
				ObjectId:  "_:attribute1",
			},
			{
				Subject:     blankID("attribute1"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("attr"),
			},
			{
				Subject:     blankID("attribute1"),
				Predicate:   "attribute.name",
				ObjectValue: stringObjectValue("ips_corefactor"),
			},
			{
				Subject:   blankID("metric"),
				Predicate: "metric.ips.attr_num_cores",
				ObjectId:  "_:attribute3",
			},
			{
				Subject:     blankID("attribute3"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("attr"),
			},
			{
				Subject:     blankID("attribute3"),
				Predicate:   "attribute.name",
				ObjectValue: stringObjectValue("ips_cores"),
			},
			{
				Subject:     blankID("metric"),
				Predicate:   "scopes",
				ObjectValue: stringObjectValue(scope),
			},
		},
	}
	assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)

	return assigned.Uids, err
}

func compareMetricIPS(t *testing.T, name string, exp, act *v1.MetricIPS) {
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
	assert.Equalf(t, exp.BaseEqTypeID, act.BaseEqTypeID, "%s.BaseEqTypeID should be same", name)
	assert.Equalf(t, exp.CoreFactorAttrID, act.CoreFactorAttrID, "%s.CoreFactorAttrID should be same", name)
	assert.Equalf(t, exp.NumCoreAttrID, act.NumCoreAttrID, "%s.NumCoreAttrID should be same", name)
}

func compareMetricIPSConfig(t *testing.T, name string, exp, act *v1.MetricIPSConfig) {
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
	assert.Equalf(t, exp.BaseEqType, act.BaseEqType, "%s.BaseEqType should be same", name)
	assert.Equalf(t, exp.CoreFactorAttr, act.CoreFactorAttr, "%s.CoreFactorAttr should be same", name)
	assert.Equalf(t, exp.NumCoreAttr, act.NumCoreAttr, "%s.NumCoreAttr should be same", name)
}
