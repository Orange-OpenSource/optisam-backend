// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_MetricNUPComputedLicenses(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     string
		mat    *v1.MetricNUPComputed
		scopes string
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
	ID1, err := getUIDForProductXID("ORAC999")
	if !assert.Empty(t, err, "error is not expected in getUIDforProductXID") {
		return
	}
	repo, err := NewLicenseRepositoryWithTemplates(dgClient)
	if !assert.Emptyf(t, err, "err is not expected from NewLicenseRepositoryWithTemplates") {
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
			l: repo,
			args: args{
				ctx: context.Background(),
				id:  ID,
				mat: &v1.MetricNUPComputed{
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
					NumOfUsers: uint32(14),
				},
			},
			want: uint64(1568),
		},
		{name: "SUCCESS - partition to server , agg-server",
			l: repo,
			args: args{
				ctx: context.Background(),
				id:  ID,
				mat: &v1.MetricNUPComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
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
					NumOfUsers: uint32(5),
				},
			},
			want: uint64(90),
		},
		{name: "SUCCESS - partition to Vcenter, agg - cluster",
			l: repo,
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricNUPComputed{
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
					NumOfUsers: uint32(1),
				},
			},
			want: uint64(88),
		},
		{name: "SUCCESS - partition to Cluster, agg-server",
			l: repo,
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricNUPComputed{
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
					NumOfUsers: uint32(1),
				},
			},
			want: uint64(34),
		},
		{name: "SUCCESS - server to Cluster, agg-server",
			l: repo,
			args: args{
				ctx: context.Background(),

				id: ID1,
				mat: &v1.MetricNUPComputed{
					EqTypeTree: []*v1.EquipmentType{

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
					NumOfUsers: uint32(1),
				},
			},
			want: uint64(2),
		},
		{name: "SUCCESS - server to server, agg-server",
			l: repo,
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricNUPComputed{
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
					NumOfUsers: uint32(1),
				},
			},
			want: uint64(6),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.MetricNUPComputedLicenses(tt.args.ctx, tt.args.id, tt.args.mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricNUPComputedLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricNUPComputedLicenses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func aggNUPSetup(metricName, productID, aggName string) (func() error, error) {
	mu := &api.Mutation{
		CommitNow: true,
		Set: []*api.NQuad{
			&api.NQuad{
				Subject:     blankID(aggName),
				Predicate:   "type_name",
				ObjectValue: stringObjectValue("product_aggreagtion"),
			},
			&api.NQuad{
				Subject:     blankID(aggName),
				Predicate:   "product_aggregation.name",
				ObjectValue: stringObjectValue(aggName),
			},
			&api.NQuad{
				Subject:   blankID(aggName),
				Predicate: "product_aggregation.products",
				ObjectId:  productID,
			},
			&api.NQuad{
				Subject:   productID,
				Predicate: "product.acqRights",
				ObjectId:  blankID("sku1"),
			},
			&api.NQuad{
				Subject:     blankID("sku1"),
				Predicate:   "acqRights.metric",
				ObjectValue: stringObjectValue(metricName),
			},
		},
	}

	assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
	if err != nil {
		return nil, err
	}

	uids := make([]string, 0, len(assigned.Uids))
	for _, uid := range assigned.Uids {
		uids = append(uids, uid)
	}

	return func() error {
		return deleteNodes(uids...)
	}, nil
}

func TestLicenseRepository_MetricNUPComputedLicensesAgg(t *testing.T) {

	type args struct {
		ctx    context.Context
		id     string
		mat    *v1.MetricNUPComputed
		scopes string
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
	ID1, err := getUIDForProductXID("ORAC999")
	if !assert.Empty(t, err, "error is not expected in getUIDforProductXID") {
		return
	}
	metric := "abc"
	aggName := "xyz"
	aggCleanup, err := aggSetup(metric, ID, aggName)
	if !assert.Empty(t, err, "error is not expected in agg setup") {
		return
	}

	defer func() {
		if !assert.Empty(t, aggCleanup(), "error is not expected in aggCleanup") {
			return
		}
	}()

	repo, err := NewLicenseRepositoryWithTemplates(dgClient)
	if !assert.Emptyf(t, err, "err is not expected from NewLicenseRepositoryWithTemplates") {
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
			l: repo,
			args: args{
				ctx: context.Background(),
				id:  ID,
				mat: &v1.MetricNUPComputed{
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
					NumOfUsers: uint32(1),
				},
			},
			want: uint64(112),
		},
		{name: "SUCCESS - partition to datacentre, agg-venter",
			l: repo,
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricNUPComputed{
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
					NumOfUsers: uint32(1),
				},
			},
			want: uint64(111),
		},
		{name: "SUCCESS - partition to Vcenter, agg - cluster",
			l: repo,
			args: args{
				ctx: context.Background(),

				id: ID,
				mat: &v1.MetricNUPComputed{
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
					NumOfUsers: uint32(1),
				},
			},
			want: uint64(88),
		},
		{name: "SUCCESS - partition to Cluster, agg-server",
			l: repo,
			args: args{
				ctx: context.Background(),

				id: ID1,
				mat: &v1.MetricNUPComputed{
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
					NumOfUsers: uint32(1),
				},
			},
			want: uint64(34),
		},
		{name: "SUCCESS - partition to Server, agg-server",
			l: repo,
			args: args{
				ctx: context.Background(),

				id: ID1,
				mat: &v1.MetricNUPComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
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
					NumOfUsers: uint32(1),
				},
			},
			want: uint64(18),
		},
		{name: "SUCCESS - Server to Cluster, agg - cluster",
			l: repo,
			args: args{
				ctx: context.Background(),

				id: ID1,
				mat: &v1.MetricNUPComputed{
					EqTypeTree: []*v1.EquipmentType{

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
					NumOfUsers: uint32(1),
				},
			},
			want: uint64(10),
		},
		{name: "SUCCESS - server to server, agg-server",
			l: repo,
			args: args{
				ctx: context.Background(),

				id: ID1,
				mat: &v1.MetricNUPComputed{
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
					NumOfUsers: uint32(1),
				},
			},
			want: uint64(6),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.MetricNUPComputedLicensesAgg(tt.args.ctx, aggName, metric, tt.args.mat, tt.args.scopes)
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
