package v1

import (
	"context"
	"errors"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/simulation-service/pkg/api/v1"
	repo "optisam-backend/simulation-service/pkg/repository/v1"
	"optisam-backend/simulation-service/pkg/repository/v1/mock"
	"optisam-backend/simulation-service/pkg/repository/v1/postgres/db"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
)

type masterdataMatcher struct {
	m *repo.MasterData
	t *testing.T
}

func (d *masterdataMatcher) Matches(x interface{}) bool {
	expM, ok := x.(*repo.MasterData)
	if !ok {
		return ok
	}
	return compareMasterDataMatcher(d, expM)
}
func compareMasterDataMatcher(d *masterdataMatcher, exp *repo.MasterData) bool {
	if exp == nil {
		return false
	}
	if !assert.Equalf(d.t, d.m.Name, exp.Name, "Config Name is not same") {
		return false
	}
	if !assert.Equalf(d.t, d.m.Status, exp.Status, "Status is not same") {
		return false
	}
	if !assert.Equalf(d.t, d.m.EquipmentType, exp.EquipmentType, "Equipment Type is not same") {
		return false
	}
	if !assert.Equalf(d.t, d.m.CreatedBy, exp.CreatedBy, "Created By is not same") {
		return false
	}
	if !assert.Equalf(d.t, d.m.UpdatedBy, exp.UpdatedBy, "Updated By is not same") {
		return false
	}
	return true
}
func (d *masterdataMatcher) String() string {
	return "masterdataMatcher"
}

func TestSimulationService_CreateConfig(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Repository
	type args struct {
		ctx context.Context
		req *v1.CreateConfigRequest
	}
	tests := []struct {
		name    string
		hcs     *SimulationService
		args    args
		setup   func()
		want    *v1.CreateConfigResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.CreateConfigRequest{
					ConfigName:    "server_1",
					EquipmentType: "Server",
					Scope:         "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "1.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "2.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{}, nil).Times(1)
				mockRepository.EXPECT().CreateConfig(
					ctx,
					&masterdataMatcher{
						m: &repo.MasterData{
							Name:          "server_1",
							Status:        1,
							EquipmentType: "Server",
							CreatedBy:     "admin@superuser.com",
							CreatedOn:     time.Now().UTC(),
							UpdatedBy:     "admin@superuser.com",
							UpdatedOn:     time.Now().UTC(),
						},
						t: t,
					},
					[]*repo.ConfigData{
						{
							ConfigMetadata: &repo.Metadata{
								AttributeName:  "cpu",
								ConfigFileName: "1.csv",
							},
							ConfigValues: []*repo.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							ConfigMetadata: &repo.Metadata{
								AttributeName:  "cf",
								ConfigFileName: "2.csv",
							},
							ConfigValues: []*repo.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
					"Scope1",
				).Return(nil).Times(1)

			},
			want: &v1.CreateConfigResponse{},
		},
		{name: "FAILURE - CreateConfig - Can not get configurations",
			args: args{
				ctx: ctx,
				req: &v1.CreateConfigRequest{
					ConfigName:    "server_1",
					EquipmentType: "Server",
					Scope:         "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "1.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "2.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - CreateConfig - cannot create config data",
			args: args{
				ctx: ctx,
				req: &v1.CreateConfigRequest{
					ConfigName:    "server_1",
					EquipmentType: "Server",
					Scope:         "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "1.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "2.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{}, nil).Times(1)
				mockRepository.EXPECT().CreateConfig(
					ctx,
					&masterdataMatcher{
						m: &repo.MasterData{
							Name:          "server_1",
							Status:        1,
							EquipmentType: "Server",
							CreatedBy:     "admin@superuser.com",
							CreatedOn:     time.Now().UTC(),
							UpdatedBy:     "admin@superuser.com",
							UpdatedOn:     time.Now().UTC(),
						},
						t: t,
					},
					[]*repo.ConfigData{
						{
							ConfigMetadata: &repo.Metadata{
								AttributeName:  "cpu",
								ConfigFileName: "1.csv",
							},
							ConfigValues: []*repo.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							ConfigMetadata: &repo.Metadata{
								AttributeName:  "cf",
								ConfigFileName: "2.csv",
							},
							ConfigValues: []*repo.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
					"Scope1",
				).Return(errors.New("Internal")).Times(1)

			},
			wantErr: true,
		},
		{name: "FAILURE - CreateConfig - cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.CreateConfigRequest{
					ConfigName:    "server_1",
					EquipmentType: "Server",
					Scope:         "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "1.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "2.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - CreateConfig - user does not have access to create config data",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "User",
					Socpes: []string{"Scope1", "Scope2", "Scope3"},
				}),
				req: &v1.CreateConfigRequest{
					ConfigName:    "server_1",
					EquipmentType: "Server",
					Scope:         "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "1.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "2.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - CreateConfig - unknown role",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "abc",
					Socpes: []string{"Scope1", "Scope2", "Scope3"},
				}),
				req: &v1.CreateConfigRequest{
					ConfigName:    "server_1",
					EquipmentType: "Server",
					Scope:         "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "1.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "2.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - Config with same name already exists",
			args: args{
				ctx: ctx,
				req: &v1.CreateConfigRequest{
					ConfigName:    "server_1",
					EquipmentType: "Server",
					Scope:         "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "1.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "2.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:   1,
						Name: "server_1",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			hcs := NewSimulationService(rep, nil)
			got, err := hcs.CreateConfig(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("HardwreConfigService.CreateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimulationSrvice.CreateConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimulationService_UpdateConfig(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Repository
	type args struct {
		ctx context.Context
		req *v1.UpdateConfigRequest
	}
	tests := []struct {
		name    string
		hcs     *SimulationService
		args    args
		setup   func()
		want    *v1.UpdateConfigResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.UpdateConfigRequest{
					ConfigId:           1,
					DeletedMetadataIds: []int32{1, 2},
					Scope:              "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "3.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "4.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            int32(1),
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
					},
				}, nil).Times(1)
				mockRepository.EXPECT().GetMetadatabyConfigID(ctx, int32(1)).Return([]db.GetMetadatabyConfigIDRow{
					{
						ID:            1,
						AttributeName: "cpu1",
						EquipmentType: "Server",
					},
					{
						ID:            2,
						AttributeName: "cf",
						EquipmentType: "Server",
					},
				}, nil)
				mockRepository.EXPECT().UpdateConfig(
					ctx,
					int32(1),
					"Server",
					"admin@superuser.com",
					[]int32{1, 2},
					[]*repo.ConfigData{
						{
							ConfigMetadata: &repo.Metadata{
								AttributeName:  "cpu",
								ConfigFileName: "3.csv",
							},
							ConfigValues: []*repo.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							ConfigMetadata: &repo.Metadata{
								AttributeName:  "cf",
								ConfigFileName: "4.csv",
							},
							ConfigValues: []*repo.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
					"Scope1",
				).Return(nil).Times(1)

			},
			want: &v1.UpdateConfigResponse{},
		},
		{name: "FAILURE - UpdateConfig - Can not get configurations",
			args: args{
				ctx: ctx,
				req: &v1.UpdateConfigRequest{
					ConfigId:           1,
					DeletedMetadataIds: []int32{1, 2},
					Scope:              "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "3.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "4.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "Failure - Unable to fetch metadata",
			args: args{
				ctx: ctx,
				req: &v1.UpdateConfigRequest{
					ConfigId:           1,
					DeletedMetadataIds: []int32{1, 2},
					Scope:              "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "3.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "4.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            int32(1),
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
					},
				}, nil).Times(1)
				mockRepository.EXPECT().GetMetadatabyConfigID(ctx, int32(1)).Return(nil, errors.New("Internal"))

			},
			wantErr: true,
		},
		{name: "Failure - Already Configured attribute exist",
			args: args{
				ctx: ctx,
				req: &v1.UpdateConfigRequest{
					ConfigId:           1,
					DeletedMetadataIds: []int32{1, 3},
					Scope:              "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "3.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "4.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            int32(1),
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
					},
				}, nil).Times(1)
				mockRepository.EXPECT().GetMetadatabyConfigID(ctx, int32(1)).Return([]db.GetMetadatabyConfigIDRow{
					{
						ID:            1,
						AttributeName: "cpu1",
						EquipmentType: "Server",
					},
					{
						ID:            2,
						AttributeName: "cf",
						EquipmentType: "Server",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateConfig - cannot update config data",
			args: args{
				ctx: ctx,
				req: &v1.UpdateConfigRequest{
					ConfigId:           1,
					DeletedMetadataIds: []int32{1, 2},
					Scope:              "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "3.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "4.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            1,
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
					},
				}, nil).Times(1)
				mockRepository.EXPECT().GetMetadatabyConfigID(ctx, int32(1)).Return([]db.GetMetadatabyConfigIDRow{
					{
						ID:            1,
						EquipmentType: "Server",
					},
					{
						ID:            2,
						EquipmentType: "Server",
					},
				}, nil)
				mockRepository.EXPECT().UpdateConfig(
					ctx,
					int32(1),
					"Server",
					"admin@superuser.com",
					[]int32{1, 2},
					[]*repo.ConfigData{
						{
							ConfigMetadata: &repo.Metadata{
								AttributeName:  "cpu",
								ConfigFileName: "3.csv",
							},
							ConfigValues: []*repo.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							ConfigMetadata: &repo.Metadata{
								AttributeName:  "cf",
								ConfigFileName: "4.csv",
							},
							ConfigValues: []*repo.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
					"Scope1",
				).Return(errors.New("Internal")).Times(1)

			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateConfig - cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.UpdateConfigRequest{
					ConfigId:           1,
					DeletedMetadataIds: []int32{1, 2},
					Scope:              "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "3.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "4.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - UpdateConfig - user does not have access to update config data",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "User",
					Socpes: []string{"Scope1", "Scope2", "Scope3"},
				}),
				req: &v1.UpdateConfigRequest{
					ConfigId:           1,
					DeletedMetadataIds: []int32{1, 2},
					Scope:              "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "3.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "4.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - UpdateConfig - unknown role",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "abc",
					Socpes: []string{"Scope1", "Scope2", "Scope3"},
				}),
				req: &v1.UpdateConfigRequest{
					ConfigId:           1,
					DeletedMetadataIds: []int32{1, 2},
					Scope:              "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "3.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "4.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - Config not found",
			args: args{
				ctx: ctx,
				req: &v1.UpdateConfigRequest{
					ConfigId:           1,
					DeletedMetadataIds: []int32{1, 2},
					Scope:              "Scope1",
					Data: []*v1.Data{
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cpu",
								ConfigFilename: "3.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "xenon",
									Value: []byte(`{"cpu":"xenon","cf":"1"}`),
								},
								{
									Key:   "phenon",
									Value: []byte(`{"cpu":"phenon","cf":"2"}`),
								},
							},
						},
						{
							Metadata: &v1.Metadata{
								AttributeName:  "cf",
								ConfigFilename: "4.csv",
							},
							Values: []*v1.ConfigValue{
								{
									Key:   "1",
									Value: []byte(`{"cf":"1"}`),
								},
								{
									Key:   "2",
									Value: []byte(`{"cf":"2"}`),
								},
							},
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            2,
						EquipmentType: "Server",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			hcs := NewSimulationService(rep, nil)
			got, err := hcs.UpdateConfig(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimulationService.UpdateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimulationService.UpdateConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimulationService_DeleteConfig(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Repository
	type args struct {
		ctx context.Context
		req *v1.DeleteConfigRequest
	}
	tests := []struct {
		name    string
		hcs     *SimulationService
		args    args
		setup   func()
		want    *v1.DeleteConfigResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DeleteConfigRequest{
					ConfigId: 1,
					Scope:    "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().DeleteConfig(
					ctx,
					db.DeleteConfigParams{
						Status: 2,
						ID:     int32(1),
						Scope:  "Scope1",
					},
				).Return(nil).Times(1)
				mockRepository.EXPECT().DeleteConfigData(
					ctx,
					int32(1),
				).Return(nil).Times(1)
			},
			want: &v1.DeleteConfigResponse{},
		},
		{name: "FAILURE - DeleteConfig - cannot delete config data",
			args: args{
				ctx: ctx,
				req: &v1.DeleteConfigRequest{
					ConfigId: 1,
					Scope:    "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().DeleteConfig(
					ctx,
					db.DeleteConfigParams{
						Status: 2,
						ID:     int32(1),
						Scope:  "Scope1",
					},
				).Return(nil).Times(1)
				mockRepository.EXPECT().DeleteConfigData(ctx, int32(1)).Return(errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteConfig - cannot update master data",
			args: args{
				ctx: ctx,
				req: &v1.DeleteConfigRequest{
					ConfigId: 1,
					Scope:    "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().DeleteConfig(ctx, db.DeleteConfigParams{Status: 2, ID: int32(1), Scope: "Scope1"}).Return(errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteConfig - cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteConfigRequest{
					ConfigId: 1,
					Scope:    "Scope1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - Delete Config - user does not have access to delete config data",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "User",
					Socpes: []string{"Scope1", "Scope2", "Scope3"},
				}),
				req: &v1.DeleteConfigRequest{
					ConfigId: 1,
					Scope:    "Scope1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - deleteConfig - unknown role",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "abc",
					Socpes: []string{"Scope1", "Scope2", "Scope3"},
				}),
				req: &v1.DeleteConfigRequest{
					ConfigId: 1,
					Scope:    "Scope1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			hcs := NewSimulationService(rep, nil)
			got, err := hcs.DeleteConfig(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimulationService.DeleteConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimulationService.DeleteConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimulationService_ListConfig(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Repository
	type args struct {
		ctx context.Context
		req *v1.ListConfigRequest
	}
	tests := []struct {
		name    string
		hcs     *SimulationService
		args    args
		setup   func()
		want    *v1.ListConfigResponse
		wantErr bool
	}{
		{name: "SUCCESS - Without Equipment Type",
			args: args{
				ctx: ctx,
				req: &v1.ListConfigRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            1,
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
						CreatedBy:     "admin@superuser.com",
						CreatedOn:     time.Unix(10, 0),
					},
				}, nil).Times(1)
				gomock.InOrder(
					mockRepository.EXPECT().GetMetadatabyConfigID(ctx, int32(1)).Return([]db.GetMetadatabyConfigIDRow{
						{
							ID:             1,
							EquipmentType:  "Server",
							AttributeName:  "cpu",
							ConfigFilename: "cpu.csv",
						},
					}, nil).Times(1),
				)
			},
			want: &v1.ListConfigResponse{
				Configurations: []*v1.Configuration{
					{
						ConfigId:      1,
						ConfigName:    "server_1",
						EquipmentType: "Server",
						CreatedBy:     "admin@superuser.com",
						CreatedOn:     &tspb.Timestamp{Seconds: 10},
						ConfigAttributes: []*v1.Attribute{
							{
								AttributeId:    1,
								AttributeName:  "cpu",
								ConfigFilename: "cpu.csv",
							},
						},
					},
				},
			},
		},
		{name: "SUCCESS - With Equipment Type",
			args: args{
				ctx: ctx,
				req: &v1.ListConfigRequest{
					EquipmentType: "Server",
					Scope:         "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   true,
					EquipmentType: "Server",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            1,
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
						CreatedBy:     "admin@superuser.com",
						CreatedOn:     time.Unix(10, 0),
					},
				}, nil).Times(1)
				gomock.InOrder(
					mockRepository.EXPECT().GetMetadatabyConfigID(ctx, int32(1)).Return([]db.GetMetadatabyConfigIDRow{
						{
							ID:             1,
							EquipmentType:  "Server",
							AttributeName:  "cpu",
							ConfigFilename: "cpu.csv",
						},
					}, nil).Times(1),
				)
			},
			want: &v1.ListConfigResponse{
				Configurations: []*v1.Configuration{
					{
						ConfigId:      1,
						ConfigName:    "server_1",
						EquipmentType: "Server",
						CreatedBy:     "admin@superuser.com",
						CreatedOn:     &tspb.Timestamp{Seconds: 10},
						ConfigAttributes: []*v1.Attribute{
							{
								AttributeId:    1,
								AttributeName:  "cpu",
								ConfigFilename: "cpu.csv",
							},
						},
					},
				},
			},
		},
		{name: "FAILURE - Unable to List configurations",
			args: args{
				ctx: ctx,
				req: &v1.ListConfigRequest{
					EquipmentType: "Server",
					Scope:         "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   true,
					EquipmentType: "Server",
					Status:        1,
					Scope:         "Scope1",
				}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - Unable to List configurations",
			args: args{
				ctx: ctx,
				req: &v1.ListConfigRequest{
					EquipmentType: "Server",
					Scope:         "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   true,
					EquipmentType: "Server",
					Status:        1,
					Scope:         "Scope1",
				}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "Failure - Unable to get metadata",
			args: args{
				ctx: ctx,
				req: &v1.ListConfigRequest{
					EquipmentType: "Server",
					Scope:         "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   true,
					EquipmentType: "Server",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            1,
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
						CreatedBy:     "admin@superuser.com",
						CreatedOn:     time.Unix(10, 0),
					},
				}, nil).Times(1)
				gomock.InOrder(
					mockRepository.EXPECT().GetMetadatabyConfigID(ctx, int32(1)).Return(nil, errors.New("Internal")).Times(1),
				)
			},
			wantErr: true,
		},
		{name: "Success - No Configurations",
			args: args{
				ctx: ctx,
				req: &v1.ListConfigRequest{
					EquipmentType: "Server",
					Scope:         "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   true,
					EquipmentType: "Server",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{}, nil).Times(1)
			},
			want: &v1.ListConfigResponse{
				Configurations: []*v1.Configuration{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			hcs := NewSimulationService(rep, nil)
			got, err := hcs.ListConfig(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimulationService.ListConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimulationService.ListConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimulationService_GetConfigData(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Repository
	type args struct {
		ctx context.Context
		req *v1.GetConfigDataRequest
	}
	tests := []struct {
		name    string
		hcs     *SimulationService
		args    args
		setup   func()
		want    *v1.GetConfigDataResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetConfigDataRequest{
					ConfigId:   1,
					MetadataId: 1,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            int32(1),
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
					},
				}, nil).Times(1)
				mockRepository.EXPECT().GetMetadatabyConfigID(ctx, int32(1)).Return([]db.GetMetadatabyConfigIDRow{
					{
						ID:            1,
						AttributeName: "cf",
						EquipmentType: "Server",
					},
				}, nil)
				mockRepository.EXPECT().GetDataByMetadataID(
					ctx,
					int32(1),
				).Return([]db.GetDataByMetadataIDRow{{
					AttributeValue: "1",
					JsonData:       []byte("{\"cf\":\"1\"}"),
				},
					{
						AttributeValue: "2",
						JsonData:       []byte("{\"cf\":\"2\"}"),
					}}, nil).Times(1)

			},
			want: &v1.GetConfigDataResponse{
				Data: []byte("[{\"cf\":\"1\"},{\"cf\":\"2\"}]"),
			},
		},
		{name: "FAILURE - UpdateConfig - Can not get configurations",
			args: args{
				ctx: ctx,
				req: &v1.GetConfigDataRequest{
					ConfigId:   1,
					MetadataId: 1,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "Failure - Can not get data",
			args: args{
				ctx: ctx,
				req: &v1.GetConfigDataRequest{
					ConfigId:   1,
					MetadataId: 1,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            int32(1),
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
					},
				}, nil).Times(1)
				mockRepository.EXPECT().GetMetadatabyConfigID(ctx, int32(1)).Return([]db.GetMetadatabyConfigIDRow{
					{
						ID:            1,
						AttributeName: "cf",
						EquipmentType: "Server",
					},
				}, nil)
				mockRepository.EXPECT().GetDataByMetadataID(
					ctx,
					int32(1),
				).Return(nil, errors.New("Internal")).Times(1)

			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateConfig - Can not get configurations",
			args: args{
				ctx: ctx,
				req: &v1.GetConfigDataRequest{
					ConfigId:   1,
					MetadataId: 1,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "Failure - Unable to fetch metadata",
			args: args{
				ctx: ctx,
				req: &v1.GetConfigDataRequest{
					ConfigId:   1,
					MetadataId: 1,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            int32(1),
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
					},
				}, nil).Times(1)
				mockRepository.EXPECT().GetMetadatabyConfigID(ctx, int32(1)).Return(nil, errors.New("Internal"))

			},
			wantErr: true,
		},
		{name: "Failure - Configuration not found",
			args: args{
				ctx: ctx,
				req: &v1.GetConfigDataRequest{
					ConfigId:   2,
					MetadataId: 1,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            int32(1),
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
					},
				}, nil).Times(1)

			},
			wantErr: true,
		},
		{name: "Failure - Metadata not found",
			args: args{
				ctx: ctx,
				req: &v1.GetConfigDataRequest{
					ConfigId:   1,
					MetadataId: 1,
					Scope:      "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := mock.NewMockRepository(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListConfig(ctx, db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
					Scope:         "Scope1",
				}).Return([]db.ListConfigRow{
					{
						ID:            int32(1),
						Name:          "server_1",
						EquipmentType: "Server",
						Status:        1,
					},
				}, nil).Times(1)
				mockRepository.EXPECT().GetMetadatabyConfigID(ctx, int32(1)).Return([]db.GetMetadatabyConfigIDRow{}, nil)

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			hcs := NewSimulationService(rep, nil)
			got, err := hcs.GetConfigData(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimulationService.GetConfigData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimulationService.GetConfigData() = %v, want %v", got, tt.want)
			}
		})
	}
}
