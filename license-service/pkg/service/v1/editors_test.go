// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_licenseServiceServer_ListEditors(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"P1", "P2", "P3"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License

	type args struct {
		ctx context.Context
		req *v1.ListEditorsRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		want    *v1.ListEditorsResponse
		mock    func()
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
			},
			want: &v1.ListEditorsResponse{
				Editors: []*v1.Editor{
					&v1.Editor{
						ID:   "E1ID",
						Name: "e1name",
					},
					&v1.Editor{
						ID:   "E2ID",
						Name: "e2name",
					},
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ListEditors(ctx, nil, []string{"P1", "P2", "P3"}).Return([]*repo.Editor{
					&repo.Editor{
						ID:   "E1ID",
						Name: "e1name",
					},
					&repo.Editor{
						ID:   "E2ID",
						Name: "e2name",
					},
				}, nil).Times(1)
			},
			wantErr: false,
		},
		{name: "FAILURE-cannot find claims in context",
			args: args{
				ctx: context.Background(),
			},
			mock:    func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch editors",
			args: args{
				ctx: ctx,
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ListEditors(ctx, nil, []string{"P1", "P2", "P3"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			s := NewLicenseServiceServer(rep)
			got, err := s.ListEditors(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListEditors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareEditorsAll(t, "ListEditors", got, tt.want)
			}
		})
	}
}

func compareEditorsAll(t *testing.T, name string, exp *v1.ListEditorsResponse, act *v1.ListEditorsResponse) {
	for i := 0; i < len(exp.Editors)-1; i++ {
		compareEditor(t, name, exp.Editors[i], act.Editors[i])
	}
}

func compareEditor(t *testing.T, name string, exp *v1.Editor, act *v1.Editor) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID are not same", name)
	}
	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
}
