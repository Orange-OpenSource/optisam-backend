// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	"errors"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/dgraph-io/dgo/protos/api"
	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_CreateMetricIPS(t *testing.T) {
	type args struct {
		ctx    context.Context
		mat    *v1.MetricIPS
		scopes []string
	}
	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		setup   func() (*v1.MetricIPS, func() error, error)
		wantErr bool
	}{
		{name: "sucess",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (retMat *v1.MetricIPS, cleanup func() error, retErr error) {

				baseID := "base"
				coreFactorAttrID := "coreFactor"
				numOfCoresAttrID := "cores"

				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{

						&api.NQuad{
							Subject:     blankID(baseID),
							Predicate:   "type",
							ObjectValue: stringObjectValue("metadata"),
						},

						&api.NQuad{
							Subject:     blankID(coreFactorAttrID),
							Predicate:   "type",
							ObjectValue: stringObjectValue("metadata"),
						},
						&api.NQuad{
							Subject:     blankID(numOfCoresAttrID),
							Predicate:   "type",
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
				t.Errorf("LicenseRepository.CreateMetricIPS() error = %v, wantErr %v", err, tt.wantErr)
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
