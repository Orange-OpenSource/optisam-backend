package v1

import (
	"context"
	"encoding/json"
	"errors"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/equipment-service/pkg/api/v1"
	repo "optisam-backend/equipment-service/pkg/repository/v1"
	"optisam-backend/equipment-service/pkg/repository/v1/mock"
	"testing"

	"github.com/golang/mock/gomock"
)

type productQueryMatcher struct {
	q *repo.QueryProducts
	t *testing.T
}

func Test_equipmentServiceServer_ListEquipmentsForProduct(t *testing.T) {

	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Equipment

	eqTypes := []*repo.EquipmentType{
		{
			Type: "typ1",
			ID:   "1",
			Attributes: []*repo.Attribute{
				{
					ID:           "1",
					Name:         "attr1",
					Type:         repo.DataTypeString,
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					ID:           "2",
					Name:         "attr2",
					Type:         repo.DataTypeString,
					IsDisplayed:  true,
					IsSearchable: true,
				},
			},
		},
		{
			Type: "typ2",
			ID:   "2",
			Attributes: []*repo.Attribute{
				{
					ID:          "1",
					Name:        "attr1",
					Type:        repo.DataTypeString,
					IsDisplayed: true,
				},
				{
					ID:          "2",
					Name:        "attr2",
					Type:        repo.DataTypeString,
					IsDisplayed: true,
				},
			},
		},
	}
	queryParams := &repo.QueryEquipments{
		PageSize:  10,
		Offset:    90,
		SortBy:    "attr1",
		SortOrder: repo.SortDESC,
		Filter: &repo.AggregateFilter{
			Filters: []repo.Queryable{
				&repo.Filter{
					FilterKey:   "attr1",
					FilterValue: "a11",
				},
				&repo.Filter{
					FilterKey:   "attr2",
					FilterValue: "a22",
				},
			},
		},
	}
	type args struct {
		ctx context.Context
		req *v1.ListEquipmentsForProductRequest
	}
	tests := []struct {
		name    string
		s       *equipmentServiceServer
		args    args
		want    *v1.ListEquipmentsResponse
		wantErr bool
		setup   func()
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
					Scopes:       []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(eqTypes, nil)
				mockLicense.EXPECT().ProductEquipments(ctx, "P1", eqTypes[0], &productQueryMatcherEquipments{
					q: queryParams,
					t: t,
				}, []string{"A"}).Times(1).Return(int32(2), json.RawMessage(`[{ID:"1"}]`), nil)

			},
			want: &v1.ListEquipmentsResponse{
				TotalRecords: 2,
				Equipments:   json.RawMessage(`[{ID:"1"}]`),
			},
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "3",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
					Scopes:       []string{"A"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - some claims are not owned by user",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "3",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
					Scopes:       []string{"C"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE- cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
					Scopes:       []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(nil, errors.New("test error"))

			},
			wantErr: true,
		},
		{name: "FAILURE- cannot fetch equipment type with given Id",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "3",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
					Scopes:       []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE- cannot find sort by attribute",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr3",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
					Scopes:       []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE- cannot sort by attribute",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
					Scopes:       []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type: "typ1",
						ID:   "1",
						Attributes: []*repo.Attribute{
							{
								ID:           "1",
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsDisplayed:  false,
								IsSearchable: true,
							},
							{
								ID:           "2",
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsDisplayed:  false,
								IsSearchable: true,
							},
						},
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE- cannot parse equipment query param",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr3=att3",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
					Scopes:       []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE- cannot fetch product equipments",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
					Scopes:       []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(eqTypes, nil)
				mockLicense.EXPECT().ProductEquipments(ctx, "P1", eqTypes[0], &productQueryMatcherEquipments{
					q: queryParams,
					t: t,
				}, []string{"A"}).Times(1).Return(int32(2), nil, errors.New("test error"))

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewEquipmentServiceServer(rep, nil)
			got, err := s.ListEquipmentsForProduct(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("equipmentServiceServer.ListEquipmentsForProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareListEquipmentResponse(t, "ListEquipmentsForProduct", got, tt.want)
			}
		})
	}
}
