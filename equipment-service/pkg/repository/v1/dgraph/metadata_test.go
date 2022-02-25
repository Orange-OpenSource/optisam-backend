package dgraph

import (
	"context"
	"errors"
	"fmt"
	v1 "optisam-backend/equipment-service/pkg/repository/v1"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

func TestEquipmentRepository_MetadataAllWithType(t *testing.T) {
	type args struct {
		ctx    context.Context
		typ    v1.MetadataType
		scopes []string
	}
	tests := []struct {
		name    string
		lr      *EquipmentRepository
		args    args
		setup   func() (func() error, error)
		want    []*v1.Metadata
		wantErr bool
	}{
		{name: "success",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				typ:    v1.MetadataTypeEquipment,
				scopes: []string{"scope1"},
			},
			setup: func() (func() error, error) {
				id := blankID("source")
				id1 := blankID("source1")
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     id,
							Predicate:   "metadata.source",
							ObjectValue: stringObjectValue("equip_1.csv"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.type",
							ObjectValue: stringObjectValue("equipment"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_1"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_2"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_3"),
						},
						{
							Subject:     id,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("scope1"),
						},
						{
							Subject:     id1,
							Predicate:   "metadata.source",
							ObjectValue: stringObjectValue("equip_2.csv"),
						},
						{
							Subject:     id1,
							Predicate:   "metadata.type",
							ObjectValue: stringObjectValue("equipment"),
						},
						{
							Subject:     id1,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_1"),
						},
						{
							Subject:     id1,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_2"),
						},
						{
							Subject:     id1,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_3"),
						},
						{
							Subject:     id1,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("scope2"),
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
				{
					Source: "equip_1.csv",
					Attributes: []string{
						"col_1",
						"col_2",
						"col_3",
					},
					Scope: "scope1",
				},
			},
		},
		{name: "failure unsupported type",
			lr: NewEquipmentRepository(dgClient),
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
				t.Errorf("EquipmentRepository.Metadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareMetadataAll(t, "metadatas", tt.want, got)
			}
		})
	}
}

func TestEquipmentRepository_MetadataWithID(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     string
		scopes []string
	}
	tests := []struct {
		name    string
		lr      *EquipmentRepository
		args    args
		setup   func() (string, func() error, error)
		want    *v1.Metadata
		wantErr bool
	}{
		{name: "success",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				id:     "source",
				scopes: []string{"scope1"},
			},
			setup: func() (string, func() error, error) {
				id := blankID("source")
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     id,
							Predicate:   "metadata.source",
							ObjectValue: stringObjectValue("equip_3.csv"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.type",
							ObjectValue: stringObjectValue("equipment"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_1"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_2"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_3"),
						},
						{
							Subject:     id,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("scope1"),
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
				Scope: "scope1",
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
				t.Errorf("EquipmentRepository.Metadata() error = %v, wantErr %v", err, tt.wantErr)
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
	assert.Equalf(t, exp.Scope, act.Scope, "%s.Scope should be same", name)
	assert.ElementsMatchf(t, exp.Attributes, act.Attributes, "%s.Attributes should be same", name)
}

func TestEquipmentRepository_UpsertMetadata(t *testing.T) {
	type args struct {
		ctx      context.Context
		metadata *v1.Metadata
	}
	tests := []struct {
		name    string
		lr      *EquipmentRepository
		setup   func() (func() error, error)
		args    args
		wantErr bool
	}{
		{name: "success - does not exists",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx: context.Background(),
				metadata: &v1.Metadata{
					Type:   v1.MetadataTypeEquipment,
					Source: "source1",
					Scope:  "scope1",
					Attributes: []string{
						"col_1",
						"col_2",
						"col_3",
					},
				},
			},
			setup: func() (func() error, error) {
				id := blankID("source")
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     id,
							Predicate:   "metadata.source",
							ObjectValue: stringObjectValue("equip_1.csv"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.type",
							ObjectValue: stringObjectValue("equipment"),
						},
						{
							Subject:     id,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("scope1"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_1"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_2"),
						},
						{
							Subject:     id,
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
				return func() error {
					return deleteNodes(id)
				}, nil
			},
		},
		{name: "success - exists",
			lr: NewEquipmentRepository(dgClient),
			args: args{
				ctx: context.Background(),
				metadata: &v1.Metadata{
					Type:   v1.MetadataTypeEquipment,
					Source: "source",
					Scope:  "scope1",
					Attributes: []string{
						"col_1",
						"col_2",
						"col_3",
					},
				},
			},
			setup: func() (func() error, error) {
				id := blankID("source")
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     id,
							Predicate:   "metadata.source",
							ObjectValue: stringObjectValue("equip_1.csv"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.type",
							ObjectValue: stringObjectValue("equipment"),
						},
						{
							Subject:     id,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("scope1"),
						},
						{
							Subject:     id,
							Predicate:   "metadata.attributes",
							ObjectValue: stringObjectValue("col_1"),
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
				return func() error {
					return deleteNodes(id)
				}, nil
			},
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
			if _, err := tt.lr.UpsertMetadata(tt.args.ctx, tt.args.metadata); (err != nil) != tt.wantErr {
				t.Errorf("EquipmentRepository.UpsertMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
