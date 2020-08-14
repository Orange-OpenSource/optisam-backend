// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	repo "optisam-backend/report-service/pkg/repository/v1"
	"reflect"
	"testing"
)

func TestReportRepository_EquipmentTypeParents(t *testing.T) {
	type args struct {
		ctx       context.Context
		equipType string
	}
	tests := []struct {
		name    string
		r       *ReportRepository
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx:       context.Background(),
				equipType: "partition",
			},
			want: []string{"server", "cluster", "vcenter", "datacenter"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReportRepository(dgClient)
			got, err := r.EquipmentTypeParents(tt.args.ctx, tt.args.equipType)
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
	}
	tests := []struct {
		name    string
		r       *ReportRepository
		args    args
		want    []*repo.EquipmentAttributes
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx:    context.Background(),
				eqtype: "partition",
			},
			want: []*repo.EquipmentAttributes{
				&repo.EquipmentAttributes{
					AttributeName:       "partition_code",
					AttributeIdentifier: true,
					ParentIdentifier:    false,
				},
				&repo.EquipmentAttributes{
					AttributeName:       "partition_hostname",
					AttributeIdentifier: false,
					ParentIdentifier:    false,
				},
				&repo.EquipmentAttributes{
					AttributeName:       "VirtualCores_VCPU",
					AttributeIdentifier: false,
					ParentIdentifier:    false,
				},
				&repo.EquipmentAttributes{
					AttributeName:       "parent_id",
					AttributeIdentifier: false,
					ParentIdentifier:    true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReportRepository(dgClient)
			got, err := r.EquipmentTypeAttrs(tt.args.ctx, tt.args.eqtype)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportRepository.EquipmentTypeAttrs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReportRepository.EquipmentTypeAttrs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReportRepository_ProductEquipments(t *testing.T) {
	type args struct {
		ctx     context.Context
		swidTag string
		scope   string
		eqtype  string
	}
	tests := []struct {
		name    string
		r       *ReportRepository
		args    args
		want    []*repo.ProductEquipment
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx:     context.Background(),
				swidTag: "Oracle_Database_11g_Enterprise_Edition_10.3",
				scope:   "TST",
				eqtype:  "server",
			},
			want: []*repo.ProductEquipment{
				&repo.ProductEquipment{
					EquipmentID:   "31353337-3135-5a43-3334-34394a4a4635",
					EquipmentType: "server",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReportRepository(dgClient)
			got, err := r.ProductEquipments(tt.args.ctx, tt.args.swidTag, tt.args.scope, tt.args.eqtype)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportRepository.ProductEquipments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReportRepository.ProductEquipments() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReportRepository_EquipmentParents(t *testing.T) {
	type args struct {
		ctx       context.Context
		equipID   string
		equipType string
		scope     string
	}
	tests := []struct {
		name    string
		r       *ReportRepository
		args    args
		want    []*repo.ProductEquipment
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx:       context.Background(),
				equipID:   "31353337-3135-5a43-3334-34394a4a4635",
				equipType: "server",
				scope:     "TST",
			},
			want: []*repo.ProductEquipment{
				&repo.ProductEquipment{
					EquipmentID:   "EXIT1WND1028",
					EquipmentType: "cluster",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReportRepository(dgClient)
			got, err := r.EquipmentParents(tt.args.ctx, tt.args.equipID, tt.args.equipType, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportRepository.EquipmentParents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, equip := range got {
				fmt.Println(equip.EquipmentID, equip.EquipmentType)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReportRepository.EquipmentParents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReportRepository_EquipmentAttributes(t *testing.T) {
	type args struct {
		ctx       context.Context
		equipID   string
		equipType string
		attrs     []*repo.EquipmentAttributes
	}
	tests := []struct {
		name    string
		r       *ReportRepository
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: context.Background(),
				attrs: []*repo.EquipmentAttributes{
					&repo.EquipmentAttributes{
						AttributeName:       "partition_code",
						AttributeIdentifier: true,
						ParentIdentifier:    false,
					},
					&repo.EquipmentAttributes{
						AttributeName:       "partition_hostname",
						AttributeIdentifier: false,
						ParentIdentifier:    false,
					},
					&repo.EquipmentAttributes{
						AttributeName:       "VirtualCores_VCPU",
						AttributeIdentifier: false,
						ParentIdentifier:    false,
					},
					&repo.EquipmentAttributes{
						AttributeName:       "parent_id",
						AttributeIdentifier: false,
						ParentIdentifier:    true,
					},
				},
				equipID:   "619625",
				equipType: "partition",
			},
			want: []byte(`{"partition_code":"619625","partition_hostname":"optvo01cc04","VirtualCores_VCPU":0.000000}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReportRepository(dgClient)
			got, err := r.EquipmentAttributes(tt.args.ctx, tt.args.equipID, tt.args.equipType, tt.args.attrs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportRepository.EquipmentAttributes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReportRepository.EquipmentAttributes() = %v, want %v", got, tt.want)
			}
		})
	}
}
