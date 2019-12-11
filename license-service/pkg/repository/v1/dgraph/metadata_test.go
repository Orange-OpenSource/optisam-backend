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
	"fmt"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/dgraph-io/dgo/protos/api"
	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_Metadata(t *testing.T) {
	type args struct {
		ctx    context.Context
		typ    v1.MetadataType
		scopes []string
	}
	tests := []struct {
		name    string
		lr      *LicenseRepository
		args    args
		setup   func() (func() error, error)
		want    []*v1.Metadata
		wantErr bool
	}{
		{name: "success",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				typ: v1.MetadataTypeEquipment,
			},
			setup: func() (func() error, error) {
				id := blankID("source")
				id1 := blankID("source1")
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						&api.NQuad{
							Subject:     id,
							Predicate:   "metadata.source",
							ObjectValue: stringObjectValue("equip_1.csv"),
						},
						&api.NQuad{
							Subject:     id,
							Predicate:   "metadata.type",
							ObjectValue: stringObjectValue("equipment"),
						},
						&api.NQuad{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_1"),
						},
						&api.NQuad{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_2"),
						},
						&api.NQuad{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_3"),
						},
						&api.NQuad{
							Subject:     id1,
							Predicate:   "metadata.source",
							ObjectValue: stringObjectValue("equip_2.csv"),
						},
						&api.NQuad{
							Subject:     id1,
							Predicate:   "metadata.type",
							ObjectValue: stringObjectValue("equipment"),
						},
						&api.NQuad{
							Subject:     id1,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_1"),
						},
						&api.NQuad{
							Subject:     id1,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_2"),
						},
						&api.NQuad{
							Subject:     id1,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_3"),
						},
					},
				}
				res, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, err
				}
				id, ok := res.Uids["source"]
				if !ok {
					return nil, errors.New("no id can be found for mutation")
				}
				id1, ok = res.Uids["source1"]
				if !ok {
					return nil, errors.New("no id can be found for mutation")
				}
				return func() error {
					return deleteNodes(id1, id)
				}, nil
			},
			want: []*v1.Metadata{
				&v1.Metadata{
					Source: "equip_1.csv",
					Attributes: []string{
						"col_1",
						"col_2",
						"col_3",
					},
				},
				&v1.Metadata{
					Source: "equip_2.csv",
					Attributes: []string{
						"col_1",
						"col_2",
						"col_3",
					},
				},
			},
		},
		{name: "failure unsupported type",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				typ: v1.MetadataType(255), //some unsupported type
			},
			setup: func() (func() error, error) {
				return func() error {
					return nil
				}, nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected from setup") {
				return
			}
			defer func() {
				assert.Empty(t, cleanup(), "error is not expected from cleanup")
			}()
			got, err := tt.lr.MetadataAllWithType(tt.args.ctx, tt.args.typ, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.Metadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareMetadataAll(t, "metadatas", tt.want, got)
			}
		})
	}
}

func TestLicenseRepository_MetadataWithID(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     string
		scopes []string
	}
	tests := []struct {
		name    string
		lr      *LicenseRepository
		args    args
		setup   func() (string, func() error, error)
		want    *v1.Metadata
		wantErr bool
	}{
		{name: "success",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (string, func() error, error) {
				id := blankID("source")
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						&api.NQuad{
							Subject:     id,
							Predicate:   "metadata.source",
							ObjectValue: stringObjectValue("equip_3.csv"),
						},
						&api.NQuad{
							Subject:     id,
							Predicate:   "metadata.type",
							ObjectValue: stringObjectValue("equipment"),
						},
						&api.NQuad{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_1"),
						},
						&api.NQuad{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_2"),
						},
						&api.NQuad{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_3"),
						},
					},
				}
				res, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return "", nil, err
				}
				id, ok := res.Uids["source"]
				if !ok {
					return "", nil, errors.New("no id can be found for mutation")
				}
				return id, func() error {
					return deleteNode(id)
				}, nil
			},
			want: &v1.Metadata{
				Source: "equip_3.csv",
				Attributes: []string{
					"col_1",
					"col_2",
					"col_3",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, cleanup, err := tt.setup()
			if !assert.Empty(t, err, "no error is expected from cleanup") {
				return
			}
			defer func() {
				assert.Empty(t, cleanup(), "error is not expected from cleanup")
			}()
			got, err := tt.lr.MetadataWithID(tt.args.ctx, id, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.Metadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareMetadata(t, "metadatas", tt.want, got)
			}
		})
	}
}

func compareMetadataAll(t *testing.T, name string, exp []*v1.Metadata, act []*v1.Metadata) {
	if !assert.Lenf(t, act, len(exp), "expected number of metdata is: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareMetadata(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareMetadata(t *testing.T, name string, exp *v1.Metadata, act *v1.Metadata) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metadata is expected to be nil")
	}

	if exp.ID != "" {
		assert.Emptyf(t, act.ID, "%s.ID is expected to be nil", name)
	}
	assert.Equalf(t, exp.Source, act.Source, "%s.Source should be same", name)
	assert.ElementsMatchf(t, exp.Attributes, act.Attributes, "%s.Attributes should be same", name)
}
