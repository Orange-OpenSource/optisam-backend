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
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func Test_equipmentServiceServer_EquipmentsPerEquipmentType(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Equipment
	type args struct {
		ctx context.Context
		req *v1.EquipmentsPerEquipmentTypeRequest
	}
	tests := []struct {
		name    string
		s       *equipmentServiceServer
		args    args
		setup   func()
		want    *v1.EquipmentsPerEquipmentTypeResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentsPerEquipmentTypeRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockEquipment(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type: "Server",
					},
				}, nil)
				gomock.InOrder(
					mockRepository.EXPECT().Equipments(ctx, &repo.EquipmentType{
						Type: "Server",
					}, &repo.QueryEquipments{}, []string{"Scope1"}).Times(1).Return(int32(10), json.RawMessage(`Hello`), nil),
				)
			},
			want: &v1.EquipmentsPerEquipmentTypeResponse{
				TypesEquipments: []*v1.TypeEquipments{
					{
						EquipType:     "Server",
						NumEquipments: int32(10),
					},
				},
			},
		},
		{
			name: "FAILURE: Error in db/EquipmentTypes",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentsPerEquipmentTypeRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockEquipment(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{
			name: "FAILURE: Error in db/Equipments",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentsPerEquipmentTypeRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockEquipment(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type: "Server",
					},
				}, nil)
				gomock.InOrder(
					mockRepository.EXPECT().Equipments(ctx, &repo.EquipmentType{
						Type: "Server",
					}, &repo.QueryEquipments{}, []string{"Scope1"}).Times(1).Return(int32(0), nil, errors.New("Internal")),
				)
			},
			wantErr: true,
		},
		{
			name: "FAILURE: User Claims not found",
			args: args{
				ctx: context.Background(),
				req: &v1.EquipmentsPerEquipmentTypeRequest{
					Scope: "Scope1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE: User do not have access to scopes",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentsPerEquipmentTypeRequest{
					Scope: "Scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewEquipmentServiceServer(rep, nil)
			got, err := s.EquipmentsPerEquipmentType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("equipmentServiceServer.EquipmentsPerEquipmentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("equipmentServiceServer.EquipmentsPerEquipmentType() = %v, want %v", got, tt.want)
			}
		})
	}
}
