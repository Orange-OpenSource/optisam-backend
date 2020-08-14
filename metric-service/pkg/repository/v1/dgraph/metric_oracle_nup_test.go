// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"errors"
	"fmt"
	v1 "optisam-backend/metric-service/pkg/repository/v1"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

func TestMetricRepository_CreateMetricOracleNUPStandard(t *testing.T) {
	type args struct {
		ctx    context.Context
		scopes []string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func() (*v1.MetricNUPOracle, func() error, error)
		wantErr bool
	}{
		{name: "success",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (retMat *v1.MetricNUPOracle, cleanup func() error, retErr error) {
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
						&api.NQuad{
							Subject:     blankID(bottomID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						&api.NQuad{
							Subject:     blankID(baseID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						&api.NQuad{
							Subject:     blankID(aggregateID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						&api.NQuad{
							Subject:     blankID(topID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						&api.NQuad{
							Subject:     blankID(coreFactorAttrID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						&api.NQuad{
							Subject:     blankID(numOfCoresAttrID),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metadata"),
						},
						&api.NQuad{
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

				return &v1.MetricNUPOracle{
						Name:                  "oracle.nup.standard",
						StartEqTypeID:         bottomID,
						BaseEqTypeID:          baseID,
						AggerateLevelEqTypeID: aggregateID,
						EndEqTypeID:           bottomID,
						CoreFactorAttrID:      coreFactorAttrID,
						NumCoreAttrID:         numOfCoresAttrID,
						NumCPUAttrID:          numOfCPUsAttrID,
						NumberOfUsers:         25,
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
			gotRetMat, err := tt.l.CreateMetricOracleNUPStandard(tt.args.ctx, mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.CreateMetricOracleNUPStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				defer func() {
					assert.Empty(t, deleteNode(gotRetMat.ID), "error not expected in deleting metric type")
				}()
				compareMetricOracleNUP(t, "MetricOracleNUP", mat, gotRetMat)
			}
		})
	}
}

func TestMetricRepository_GetMetricConfigNUP(t *testing.T) {
	type args struct {
		ctx     context.Context
		metName string
		scopes  []string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func() (func() error, error)
		want    *v1.MetricNUPConfig
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				metName: "nup1",
			},
			setup: func() (func() error, error) {
				ids, err := addMetricNUPConfig("nup1")
				if err != nil {
					t.Errorf("Failed to create config of OPS metric, err : %v", err)
				}
				return func() error {
					return deleteMetricConfig(ids)
				}, nil
			},
			want: &v1.MetricNUPConfig{
				Name:                "nup1",
				NumCoreAttr:         "nup_cores",
				NumCPUAttr:          "nup_cpu",
				CoreFactorAttr:      "nup_corefactor",
				BaseEqType:          "server",
				AggerateLevelEqType: "vcenter",
				EndEqType:           "cluster",
				StartEqType:         "partition",
				NumberOfUsers:       10,
			},
			wantErr: false,
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
			got, err := tt.l.GetMetricConfigNUP(tt.args.ctx, tt.args.metName, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.GetMetricConfigNUP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricOracleNUPConfig(t, "MetricRepository.GetMetricConfigNUP", tt.want, got)
			}
		})
	}
}

func addMetricNUPConfig(metName string) (ids map[string]string, err error) {

	mu := &api.Mutation{
		CommitNow: true,
		Set: []*api.NQuad{
			&api.NQuad{
				Subject:     blankID("metric"),
				Predicate:   "metric.name",
				ObjectValue: stringObjectValue(metName),
			},
			&api.NQuad{
				Subject:     blankID("metric"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("Metric"),
			},
			&api.NQuad{
				Subject:   blankID("metric"),
				Predicate: "metric.oracle_nup.bottom",
				ObjectId:  "_:metadata1",
			},
			&api.NQuad{
				Subject:     blankID("metadata1"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("metadata"),
			},
			&api.NQuad{
				Subject:     blankID("metadata1"),
				Predicate:   "metadata.equipment.type",
				ObjectValue: stringObjectValue("partition"),
			},
			&api.NQuad{
				Subject:   blankID("metric"),
				Predicate: "metric.oracle_nup.top",
				ObjectId:  "_:metadata2",
			},
			&api.NQuad{
				Subject:     blankID("metadata2"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("metadata"),
			},
			&api.NQuad{
				Subject:     blankID("metadata2"),
				Predicate:   "metadata.equipment.type",
				ObjectValue: stringObjectValue("cluster"),
			},
			&api.NQuad{
				Subject:   blankID("metric"),
				Predicate: "metric.oracle_nup.aggregate",
				ObjectId:  "_:metadata3",
			},
			&api.NQuad{
				Subject:     blankID("metadata3"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("metadata"),
			},
			&api.NQuad{
				Subject:     blankID("metadata3"),
				Predicate:   "metadata.equipment.type",
				ObjectValue: stringObjectValue("vcenter"),
			},
			&api.NQuad{
				Subject:   blankID("metric"),
				Predicate: "metric.oracle_nup.base",
				ObjectId:  "_:metadata4",
			},
			&api.NQuad{
				Subject:     blankID("metadata4"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("metadata"),
			},
			&api.NQuad{
				Subject:     blankID("metadata4"),
				Predicate:   "metadata.equipment.type",
				ObjectValue: stringObjectValue("server"),
			},
			&api.NQuad{
				Subject:   blankID("metric"),
				Predicate: "metric.oracle_nup.attr_core_factor",
				ObjectId:  "_:attribute1",
			},
			&api.NQuad{
				Subject:     blankID("attribute1"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("attr"),
			},
			&api.NQuad{
				Subject:     blankID("attribute1"),
				Predicate:   "attribute.name",
				ObjectValue: stringObjectValue("nup_corefactor"),
			},
			&api.NQuad{
				Subject:   blankID("metric"),
				Predicate: "metric.oracle_nup.attr_num_cpu",
				ObjectId:  "_:attribute2",
			},
			&api.NQuad{
				Subject:     blankID("attribute2"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("attr"),
			},
			&api.NQuad{
				Subject:     blankID("attribute2"),
				Predicate:   "attribute.name",
				ObjectValue: stringObjectValue("nup_cpu"),
			},
			&api.NQuad{
				Subject:   blankID("metric"),
				Predicate: "metric.oracle_nup.attr_num_cores",
				ObjectId:  "_:attribute3",
			},
			&api.NQuad{
				Subject:     blankID("attribute3"),
				Predicate:   "dgraph.type",
				ObjectValue: stringObjectValue("attr"),
			},
			&api.NQuad{
				Subject:     blankID("attribute3"),
				Predicate:   "attribute.name",
				ObjectValue: stringObjectValue("nup_cores"),
			},
			&api.NQuad{
				Subject:     blankID("metric"),
				Predicate:   "metric.oracle_nup.num_users",
				ObjectValue: stringObjectValue("10"),
			},
		},
	}
	assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)

	return assigned.Uids, err
}

func compareMetricOracleNUP(t *testing.T, name string, exp, act *v1.MetricNUPOracle) {
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
	assert.Equalf(t, exp.NumberOfUsers, act.NumberOfUsers, "%s.NumUsersAttrID should be same", name)
}

func compareMetricOracleNUPConfig(t *testing.T, name string, exp, act *v1.MetricNUPConfig) {
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
	assert.Equalf(t, exp.StartEqType, act.StartEqType, "%s.StartEqType should be same", name)
	assert.Equalf(t, exp.BaseEqType, act.BaseEqType, "%s.BaseEqType should be same", name)
	assert.Equalf(t, exp.AggerateLevelEqType, act.AggerateLevelEqType, "%s.AggerateLevelEqType should be same", name)
	assert.Equalf(t, exp.EndEqType, act.EndEqType, "%s.EndEqType should be same", name)
	assert.Equalf(t, exp.CoreFactorAttr, act.CoreFactorAttr, "%s.CoreFactorAttr should be same", name)
	assert.Equalf(t, exp.NumCoreAttr, act.NumCoreAttr, "%s.NumCoreAttr should be same", name)
	assert.Equalf(t, exp.NumCPUAttr, act.NumCPUAttr, "%s.NumCPUAttr should be same", name)
	assert.Equalf(t, exp.NumberOfUsers, act.NumberOfUsers, "%s.NumUsers should be same", name)
}
