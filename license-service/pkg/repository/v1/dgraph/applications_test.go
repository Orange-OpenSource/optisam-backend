package dgraph

import (
	"context"
	"testing"
)

func TestLicenseRepository_ProductExistsForApplication(t *testing.T) {
	type args struct {
		ctx    context.Context
		prodID string
		appID  string
		scopes string
	}
	tests := []struct {
		name    string
		lr      *LicenseRepository
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx:    context.Background(),
				prodID: "ORAC010",
				appID:  "6",
				scopes: "scope2",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lr := NewLicenseRepository(dgClient)
			got, err := lr.ProductExistsForApplication(tt.args.ctx, tt.args.prodID, tt.args.appID, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ProductExistsForApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.ProductExistsForApplication() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
func TestLicenseRepository_ProductApplicationEquipments(t *testing.T) {
	type args struct {
		ctx    context.Context
		prodID string
		appID  string
		scopes string
	}
	tests := []struct {
		name    string
		lr      *LicenseRepository
		args    args
		want    []*v1.Equipment
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx:    context.Background(),
				prodID: "WIN3",
				appID:  "A2",
				scopes: "scope3",
			},
			want: []*v1.Equipment{
				&v1.Equipment{
					ID:      "0x8b",
					EquipID: "SERV5",
				},
				&v1.Equipment{
					ID:      "0x8c",
					EquipID: "SERV6",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lr := NewLicenseRepository(dgClient)
			got, err := lr.ProductApplicationEquipments(tt.args.ctx, tt.args.prodID, tt.args.appID, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ProductApplicationEquipments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LicenseRepository.ProductApplicationEquipments() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
