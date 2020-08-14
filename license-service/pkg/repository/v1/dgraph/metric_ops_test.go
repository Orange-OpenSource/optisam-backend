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
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

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

func TestLicenseRepository_ListMetricOPS(t *testing.T) {
	type args struct {
		ctx    context.Context
		scopes []string
	}
	tests := []struct {
		name  string
		l     *LicenseRepository
		args  args
		setup func() ([]*v1.MetricOPS, func() error, error)
		//want    []*v1.MetricOPS
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
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
				t.Errorf("LicenseRepository.ListMetricOPS() error = %v, wantErr %v", err, tt.wantErr)
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
				//t.Log(err)
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
				//t.Log(err)
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
				//t.Log(err)
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
				//t.Log(err)
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
				//t.Log(err)
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
				//t.Log(err)
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
				//t.Log(err)
			}
		}
	}()
	repo := NewLicenseRepository(dgClient)
	gotRetMat, err := repo.CreateMetricOPS(context.Background(), &v1.MetricOPS{
		Name:                  "oracle.processor.standard",
		StartEqTypeID:         bottomID,
		BaseEqTypeID:          baseID,
		AggerateLevelEqTypeID: aggregateID,
		EndEqTypeID:           topID,
		CoreFactorAttrID:      coreFactorAttrID,
		NumCoreAttrID:         numOfCoresAttrID,
		NumCPUAttrID:          numOfCPUsAttrID,
	}, []string{})
	return gotRetMat, func() error {
		//return nil
		return deleteNodes(gotRetMat.ID, bottomID, baseID, aggregateID, bottomID, coreFactorAttrID, numOfCoresAttrID, numOfCPUsAttrID)
	}, nil
}
