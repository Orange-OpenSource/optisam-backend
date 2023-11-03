package dgraph

import (
	"context"
	"errors"
	"fmt"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/repository/v1"
	"reflect"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

func TestReportRepository_EquipmentTypeParents(t *testing.T) {
	type args struct {
		ctx       context.Context
		equipType string
		scope     string
	}
	tests := []struct {
		name    string
		r       *ReportRepository
		args    args
		setup   func() (func() error, error)
		want    []string
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx:       context.Background(),
				equipType: "partition",
				scope:     "scope1",
			},
			setup: func() (func() error, error) {
				id := blankID("partition")
				id1 := blankID("server")
				id2 := blankID("cluster")
				sID := blankID("source")
				attrID := blankID("attrName")
				attrID1 := blankID("attr1Name")
				attrID2 := blankID("attr2Name")
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     id,
							Predicate:   "metadata.equipment.type",
							ObjectValue: stringObjectValue("partition"),
						},
						{
							Subject:     id,
							Predicate:   "dgraph.type",
							ObjectValue: stringObjectValue("Equipment"),
						},
						{
							Subject:   id,
							Predicate: "metadata.equipment.parent",
							ObjectId:  id1,
						},
						{
							Subject:   id,
							Predicate: "metadata.equipment.source",
							ObjectId:  sID,
						},
						{
							Subject:   id,
							Predicate: "metadata.equipment.attribute",
							ObjectId:  attrID,
						},
						{
							Subject:     attrID,
							Predicate:   "attribute.name",
							ObjectValue: stringObjectValue("attrName"),
						},
						{
							Subject:     attrID,
							Predicate:   "attribute.type",
							ObjectValue: intObjectValue(0),
						},
						{
							Subject:     attrID,
							Predicate:   "attribute.parentIdentifier",
							ObjectValue: boolObjectValue(true),
						},
						{
							Subject:     id,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("scope1"),
						},
						{
							Subject:     id1,
							Predicate:   "metadata.equipment.type",
							ObjectValue: stringObjectValue("server"),
						},
						{
							Subject:     id1,
							Predicate:   "dgraph.type",
							ObjectValue: stringObjectValue("Equipment"),
						},
						{
							Subject:   id1,
							Predicate: "metadata.equipment.parent",
							ObjectId:  id2,
						},
						{
							Subject:   id1,
							Predicate: "metadata.equipment.source",
							ObjectId:  sID,
						},
						{
							Subject:   id1,
							Predicate: "metadata.equipment.attribute",
							ObjectId:  attrID1,
						},
						{
							Subject:     attrID1,
							Predicate:   "attribute.name",
							ObjectValue: stringObjectValue("attrName"),
						},
						{
							Subject:     attrID1,
							Predicate:   "attribute.type",
							ObjectValue: intObjectValue(0),
						},
						{
							Subject:     attrID1,
							Predicate:   "attribute.parentIdentifier",
							ObjectValue: boolObjectValue(true),
						},
						{
							Subject:     id1,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("scope1"),
						},
						{
							Subject:     id2,
							Predicate:   "metadata.equipment.type",
							ObjectValue: stringObjectValue("cluster"),
						},
						{
							Subject:     id2,
							Predicate:   "dgraph.type",
							ObjectValue: stringObjectValue("Equipment"),
						},
						{
							Subject:   id2,
							Predicate: "metadata.equipment.source",
							ObjectId:  sID,
						},
						{
							Subject:   id2,
							Predicate: "metadata.equipment.attribute",
							ObjectId:  attrID2,
						},
						{
							Subject:     attrID2,
							Predicate:   "attribute.name",
							ObjectValue: stringObjectValue("attrName"),
						},
						{
							Subject:     attrID2,
							Predicate:   "attribute.type",
							ObjectValue: intObjectValue(0),
						},
						{
							Subject:     attrID2,
							Predicate:   "attribute.parentIdentifier",
							ObjectValue: boolObjectValue(false),
						},
						{
							Subject:     id2,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("scope1"),
						},
					},
				}
				res, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, err
				}
				sourceid, ok := res.Uids["source"]
				if !ok {
					return nil, errors.New("no id can be found for mutation")
				}
				cAttrid, ok := res.Uids["attr2Name"]
				if !ok {
					return nil, errors.New("attr2Name, no id can be found for mutation")
				}
				sAttrid, ok := res.Uids["attr1Name"]
				if !ok {
					return nil, errors.New("attr1Name, no id can be found for mutation")
				}
				pAttrid, ok := res.Uids["attrName"]
				if !ok {
					return nil, errors.New("attrName, no id can be found for mutation")
				}
				clusterid, ok := res.Uids["cluster"]
				if !ok {
					return nil, errors.New("cluster, no id can be found for mutation")
				}
				serverid, ok := res.Uids["server"]
				if !ok {
					return nil, errors.New("server, no id can be found for mutation")
				}
				partitonid, ok := res.Uids["partition"]
				if !ok {
					return nil, errors.New("partition, no id can be found for mutation")
				}
				return func() error {
					return deleteNodes(cAttrid, sAttrid, pAttrid, sourceid, clusterid, serverid, partitonid)
				}, nil
			},
			want: []string{"server", "cluster"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup()
			if !assert.Empty(t, err, "error is not expect in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expect in cleanup")
			}()
			r := NewReportRepository(dgClient)
			got, err := r.EquipmentTypeParents(tt.args.ctx, tt.args.equipType, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportRepository.EquipmentTypeParents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReportRepository.EquipmentTypeParents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReportRepository_EquipmentTypeAttrs(t *testing.T) {
	type args struct {
		ctx    context.Context
		eqtype string
		scope  string
	}
	tests := []struct {
		name    string
		r       *ReportRepository
		args    args
		setup   func() (func() error, error)
		want    []*repo.EquipmentAttributes
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx:    context.Background(),
				eqtype: "partition",
				scope:  "scope1",
			},
			setup: func() (func() error, error) {
				id := blankID("partition")
				id1 := blankID("server")
				sID := blankID("source")
				attrID := blankID("partition_code")
				attr1ID := blankID("parent_id")
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     id,
							Predicate:   "metadata.equipment.type",
							ObjectValue: stringObjectValue("partition"),
						},
						{
							Subject:     id,
							Predicate:   "dgraph.type",
							ObjectValue: stringObjectValue("Equipment"),
						},
						{
							Subject:   id,
							Predicate: "metadata.equipment.parent",
							ObjectId:  id1,
						},
						{
							Subject:   id,
							Predicate: "metadata.equipment.source",
							ObjectId:  sID,
						},
						{
							Subject:   id,
							Predicate: "metadata.equipment.attribute",
							ObjectId:  attrID,
						},
						{
							Subject:     attrID,
							Predicate:   "attribute.name",
							ObjectValue: stringObjectValue("partition_code"),
						},
						{
							Subject:     attrID,
							Predicate:   "attribute.type",
							ObjectValue: intObjectValue(1),
						},
						{
							Subject:     attrID,
							Predicate:   "attribute.identifier",
							ObjectValue: boolObjectValue(true),
						},
						{
							Subject:     attrID,
							Predicate:   "attribute.parentIdentifier",
							ObjectValue: boolObjectValue(false),
						},
						{
							Subject:     attrID,
							Predicate:   "attribute.searchable",
							ObjectValue: boolObjectValue(true),
						},
						{
							Subject:     attrID,
							Predicate:   "attribute.displayed",
							ObjectValue: boolObjectValue(true),
						},
						{
							Subject:     attrID,
							Predicate:   "attribute.mapped_to",
							ObjectValue: stringObjectValue("partiton_code"),
						},
						{
							Subject:   id,
							Predicate: "metadata.equipment.attribute",
							ObjectId:  attr1ID,
						},
						{
							Subject:     attr1ID,
							Predicate:   "attribute.name",
							ObjectValue: stringObjectValue("parent_id"),
						},
						{
							Subject:     attr1ID,
							Predicate:   "attribute.type",
							ObjectValue: intObjectValue(1),
						},
						{
							Subject:     attr1ID,
							Predicate:   "attribute.identifier",
							ObjectValue: boolObjectValue(false),
						},
						{
							Subject:     attr1ID,
							Predicate:   "attribute.parentIdentifier",
							ObjectValue: boolObjectValue(true),
						},
						{
							Subject:     attr1ID,
							Predicate:   "attribute.searchable",
							ObjectValue: boolObjectValue(true),
						},
						{
							Subject:     attr1ID,
							Predicate:   "attribute.displayed",
							ObjectValue: boolObjectValue(true),
						},
						{
							Subject:     attrID,
							Predicate:   "attribute.mapped_to",
							ObjectValue: stringObjectValue("partiton_code"),
						},
						{
							Subject:     attr1ID,
							Predicate:   "attribute.mapped_to",
							ObjectValue: stringObjectValue("parent_id"),
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
					return nil, err
				}
				sourceid, ok := res.Uids["source"]
				if !ok {
					return nil, errors.New("no id can be found for mutation")
				}
				pAttrid, ok := res.Uids["partition_code"]
				if !ok {
					return nil, errors.New("attrName, no id can be found for mutation")
				}
				pAttr1id, ok := res.Uids["parent_id"]
				if !ok {
					return nil, errors.New("attrName, no id can be found for mutation")
				}
				partitonid, ok := res.Uids["partition"]
				if !ok {
					return nil, errors.New("partition, no id can be found for mutation")
				}
				return func() error {
					return deleteNodes(pAttr1id, pAttrid, sourceid, partitonid)
				}, nil
			},
			want: []*repo.EquipmentAttributes{
				{
					AttributeName:       "partition_code",
					AttributeIdentifier: true,
					ParentIdentifier:    false,
				},
				{
					AttributeName:       "parent_id",
					AttributeIdentifier: false,
					ParentIdentifier:    true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup()
			if !assert.Empty(t, err, "error is not expect in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expect in cleanup")
			}()
			r := NewReportRepository(dgClient)
			got, err := r.EquipmentTypeAttrs(tt.args.ctx, tt.args.eqtype, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportRepository.EquipmentTypeAttrs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareEquipmentAttributeAll(t, "ReportRepository.EquipmentTypeAttrs", tt.want, got)
			}
		})
	}
}

// func TestReportRepository_ProductEquipments(t *testing.T) {
// 	type args struct {
// 		ctx     context.Context
// 		swidTag string
// 		scope   string
// 		eqtype  string
// 	}
// 	tests := []struct {
// 		name    string
// 		r       *ReportRepository
// 		args    args
// 		want    []*repo.ProductEquipment
// 		wantErr bool
// 	}{
// 		{
// 			name: "SUCCESS",
// 			args: args{
// 				ctx:     context.Background(),
// 				swidTag: "Oracle_Database_11g_Enterprise_Edition_10.3",
// 				scope:   "TST",
// 				eqtype:  "server",
// 			},
// 			want: []*repo.ProductEquipment{
// 				{
// 					EquipmentID:   "31353337-3135-5a43-3334-34394a4a4635",
// 					EquipmentType: "server",
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := NewReportRepository(dgClient)
// 			got, err := r.ProductEquipments(tt.args.ctx, tt.args.swidTag, tt.args.scope, tt.args.eqtype)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ReportRepository.ProductEquipments() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			fmt.Println(got)
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ReportRepository.ProductEquipments() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestReportRepository_EquipmentParents(t *testing.T) {
// 	type args struct {
// 		ctx       context.Context
// 		equipID   string
// 		equipType string
// 		scope     string
// 	}
// 	tests := []struct {
// 		name    string
// 		r       *ReportRepository
// 		args    args
// 		want    []*repo.ProductEquipment
// 		wantErr bool
// 	}{
// 		{
// 			name: "SUCCESS",
// 			args: args{
// 				ctx:       context.Background(),
// 				equipID:   "31353337-3135-5a43-3334-34394a4a4635",
// 				equipType: "server",
// 				scope:     "TST",
// 			},
// 			want: []*repo.ProductEquipment{
// 				{
// 					EquipmentID:   "EXIT1WND1028",
// 					EquipmentType: "cluster",
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := NewReportRepository(dgClient)
// 			got, err := r.EquipmentParents(tt.args.ctx, tt.args.equipID, tt.args.equipType, tt.args.scope)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ReportRepository.EquipmentParents() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			for _, equip := range got {
// 				fmt.Println(equip.EquipmentID, equip.EquipmentType)
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ReportRepository.EquipmentParents() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestReportRepository_EquipmentAttributes(t *testing.T) {
// 	type args struct {
// 		ctx       context.Context
// 		equipID   string
// 		equipType string
// 		attrs     []*repo.EquipmentAttributes
// 		scope     string
// 	}
// 	tests := []struct {
// 		name    string
// 		r       *ReportRepository
// 		args    args
// 		want    json.RawMessage
// 		wantErr bool
// 	}{
// 		{
// 			name: "SUCCESS",
// 			args: args{
// 				ctx: context.Background(),
// 				attrs: []*repo.EquipmentAttributes{
// 					{
// 						AttributeName:       "partition_code",
// 						AttributeIdentifier: true,
// 						ParentIdentifier:    false,
// 					},
// 					{
// 						AttributeName:       "partition_hostname",
// 						AttributeIdentifier: false,
// 						ParentIdentifier:    false,
// 					},
// 					{
// 						AttributeName:       "VirtualCores_VCPU",
// 						AttributeIdentifier: false,
// 						ParentIdentifier:    false,
// 					},
// 					{
// 						AttributeName:       "parent_id",
// 						AttributeIdentifier: false,
// 						ParentIdentifier:    true,
// 					},
// 				},
// 				equipID:   "619625",
// 				equipType: "partition",
// 				scope:     "TST",
// 			},
// 			want: []byte(`{"partition_code":"619625","partition_hostname":"optvo01cc04","VirtualCores_VCPU":0.000000}`),
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := NewReportRepository(dgClient)
// 			got, err := r.EquipmentAttributes(tt.args.ctx, tt.args.equipID, tt.args.equipType, tt.args.attrs, tt.args.scope)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ReportRepository.EquipmentAttributes() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ReportRepository.EquipmentAttributes() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func compareEquipmentAttributeAll(t *testing.T, name string, exp, act []*repo.EquipmentAttributes) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareEquipmentAttribute(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareEquipmentAttribute(t *testing.T, name string, exp, act *repo.EquipmentAttributes) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.AttributeName, act.AttributeName, "%s.AttributeName are not same", name)
	assert.Equalf(t, exp.AttributeIdentifier, act.AttributeIdentifier, "%s.AttributeIdentifier are not same", name)
	assert.Equalf(t, exp.AttributeValue, act.AttributeValue, "%s.AttributeValue are not same", name)
	assert.Equalf(t, exp.ParentIdentifier, act.ParentIdentifier, "%s.ParentIdentifier are not same", name)
}

func stringObjectValue(val string) *api.Value {
	return &api.Value{
		Val: &api.Value_StrVal{
			StrVal: val,
		},
	}
}

func boolObjectValue(val bool) *api.Value {
	return &api.Value{
		Val: &api.Value_BoolVal{
			BoolVal: val,
		},
	}
}

func intObjectValue(val int64) *api.Value {
	return &api.Value{
		Val: &api.Value_IntVal{
			IntVal: val,
		},
	}
}

func blankID(id string) string {
	return "_:" + id
}

func deleteNodes(ids ...string) error {

	for _, id := range ids {
		if err := deleteNode(id); err != nil {
			return err
		}
	}

	return nil
}

func deleteNode(id string) error {
	mu := &api.Mutation{
		CommitNow:  true,
		DeleteJson: []byte(`{"uid": "` + id + `"}`),
		// Del: []*api.NQuad{
		// 	&api.NQuad{
		// 		Subject:     id,
		// 		Predicate:   "*",
		// 		ObjectValue: deleteAll,
		// 	},
	}

	// delete all the data
	_, err := dgClient.NewTxn().Mutate(context.Background(), mu)
	if err != nil {
		return err
	}

	return nil
}
