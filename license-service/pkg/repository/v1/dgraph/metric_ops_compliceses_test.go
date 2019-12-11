// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	"encoding/json"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/loader"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func deleteForXIDs(xids []string, query string) error {
	for _, xid := range xids {
		resp, err := dgClient.NewTxn().Query(context.Background(), strings.Replace(query, "$XID", xid, -1))
		if err != nil {
			return err
		}
		type data struct {
			Data []*id
		}

		var d data
		if err := json.Unmarshal(resp.Json, &d); err != nil {
			return err
		}
		if len(d.Data) == 0 {
			continue
		}
		for _, id := range d.Data {
			if err := deleteNode(id.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteEqMetadaTypes(metas ...string) error {
	q := `
	{
		Data(func:eq( metadata.source,$XID)){
		  uid
		}
	  }
	`
	return deleteForXIDs(metas, q)
}

func deleteEquipmentTypes(eqTypes ...string) error {
	q := `
	{
		Data(func:eq(metadata.equipment.type,$XID)){
		  uid
		}
	  }
	`
	return deleteForXIDs(eqTypes, q)
}

func deleteEquipments(eqTypes ...string) error {
	q := `
	{
		Data(func:eq(equipment.type,$XID)){
		  uid
		}
	  }
	`
	return deleteForXIDs(eqTypes, q)
}

func setup() (func() error, error) {
	config := loader.NewDefaultConfig()
	config.LoadMetadata = true
	config.MasterDir = "testdata"
	config.ScopeSkeleten = "skeletonscope"
	repo := NewLicenseRepository(dgClient)
	config.Repository = repo
	config.IgnoreNew = true
	equipFiles := []string{
		"equipment_cluster.csv",
		"equipment_datacenter.csv",
		"equipment_partition.csv",
		"equipment_server.csv",
		"equipment_vcenter.csv",
	}
	config.MetadataFiles.EquipFiles = equipFiles
	loader.Load(config)

	if err := loader.LoadDefaultEquipmentTypes(repo); err != nil {
		return nil, err
	}

	config = loader.NewDefaultConfig()
	config.IgnoreNew = true
	config.MasterDir = "testdata"
	config.Scopes = []string{"scope1", "scope2"}
	config.LoadEquipments = true
	config.Repository = repo
	config.EquipmentFiles = equipFiles
	if err := loader.Load(config); err != nil {
		return nil, err
	}
	return func() error {
		if err := deleteEquipmentTypes("Server", "Partition", "Cluster", "Vcenter", "Datacenter"); err != nil {
			return err
		}
		if err := deleteEquipments("Server", "Partition", "Cluster", "Vcenter", "Datacenter"); err != nil {
			return err
		}
		equipFilesBasePath := make([]string, len(equipFiles))
		for i := range equipFiles {
			equipFilesBasePath[i] = filepath.Base(equipFiles[i])
		}
		if err := deleteEqMetadaTypes(equipFilesBasePath...); err != nil {
			return err
		}
		return nil
	}, nil
}

func TestLicenseRepository_MetricOPSComputedLicenses(t *testing.T) {

	type args struct {
		ctx    context.Context
		id     string
		mat    *v1.MetricOPSComputed
		scopes []string
	}
	cleanup, err := setup()
	if !assert.Empty(t, err, "error is not expected in setup") {
		return
	}
	defer func() {
		if !assert.Empty(t, cleanup(), "error is not expected in cleanup") {
			return
		}
	}()

	ID, err := getUIDForProductXID("ORAC099")
	if !assert.Empty(t, err, "error is not expected in getUIDforProductXID") {
		return
	}
	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		want    uint64
		wantErr bool
	}{
		{name: "SUCCESS - partition to datacentre , agg-cluster",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				id:  ID,
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
						&v1.EquipmentType{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(112),
		},
		{name: "SUCCESS - partition to datacentre, agg-venter",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
						&v1.EquipmentType{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Vcenter",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(111),
		},
		{name: "SUCCESS - partition to Vcenter, agg - cluster",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(88),
		},
		{name: "SUCCESS - partition to Cluster, agg-server",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Server",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(34),
		},
		{name: "SUCCESS - server to server, agg-server",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{

						&v1.EquipmentType{
							Type: "Server",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Server",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
			},
			want: uint64(6),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.MetricOPSComputedLicenses(tt.args.ctx, tt.args.id, tt.args.mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricOPSComputedLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricOPSComputedLicenses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getUIDForProductXID(xid string) (string, error) {
	type id struct {
		ID string
	}
	type data struct {
		IDs []*id
	}

	resp, err := dgClient.NewTxn().Query(context.Background(), `{
	        IDs(func: eq(product.swidtag,`+xid+`)){
				ID:uid
			}
	}`)
	if err != nil {
		return "", err
	}

	var d data
	if err := json.Unmarshal(resp.Json, &d); err != nil {
		return "", err
	}
	if len(d.IDs) == 0 {
		return "", v1.ErrNoData
	}
	return d.IDs[0].ID, nil
}
