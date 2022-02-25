package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	accmock "optisam-backend/account-service/pkg/api/v1/mock"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/equipment-service/pkg/api/v1"
	repo "optisam-backend/equipment-service/pkg/repository/v1"
	"optisam-backend/equipment-service/pkg/repository/v1/mock"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_CreateGenericScopeEquipmentTypes(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"A", "B"},
	})
	ctx2 := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "User",
		Socpes: []string{"A", "B"},
	})
	type args struct {
		ctx context.Context
		req *v1.CreateGenericScopeEquipmentTypesRequest
	}
	var mockCtrl *gomock.Controller
	var rep repo.Equipment
	metadata := repo.GetGenericScopeMetadata("A")
	eqType := repo.GetGenericScopeEquipmentTypes("A")
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    *v1.CreateGenericScopeEquipmentTypesResponse
		wantErr bool
	}{
		{name: "success",
			args: args{
				ctx: ctx,
				req: &v1.CreateGenericScopeEquipmentTypesRequest{
					Scope: "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UpsertMetadata(ctx, &metadata[0]).Return("1", nil).Times(1)
				eqType[metadata[0].Source].SourceID = "1"
				mockRepo.EXPECT().CreateEquipmentType(ctx, eqType[metadata[0].Source], []string{"A"}).Return(&repo.EquipmentType{}, nil).Times(1)
				mockRepo.EXPECT().UpsertMetadata(ctx, &metadata[1]).Return("2", nil).Times(1)
				eqType[metadata[1].Source].SourceID = "2"
				mockRepo.EXPECT().CreateEquipmentType(ctx, eqType[metadata[1].Source], []string{"A"}).Return(&repo.EquipmentType{}, nil).Times(1)
				mockRepo.EXPECT().UpsertMetadata(ctx, &metadata[2]).Return("3", nil).Times(1)
				eqType[metadata[2].Source].SourceID = "3"
				mockRepo.EXPECT().CreateEquipmentType(ctx, eqType[metadata[2].Source], []string{"A"}).Return(&repo.EquipmentType{}, nil).Times(1)
				mockRepo.EXPECT().UpsertMetadata(ctx, &metadata[3]).Return("4", nil).Times(1)
				eqType[metadata[3].Source].SourceID = "4"
				mockRepo.EXPECT().CreateEquipmentType(ctx, eqType[metadata[3].Source], []string{"A"}).Return(&repo.EquipmentType{}, nil).Times(1)
				// mockRepo.EXPECT().UpsertMetadata(ctx, &metadata[4]).Return("5", nil).Times(1)
				// eqType[metadata[4].Source].SourceID = "5"
				// mockRepo.EXPECT().CreateEquipmentType(ctx, eqType[metadata[4].Source], []string{"A"}).Return(&repo.EquipmentType{}, nil).Times(1)

			},
			want: &v1.CreateGenericScopeEquipmentTypesResponse{},
		},
		{name: "failure|can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.CreateGenericScopeEquipmentTypesRequest{
					Scope: "A",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure| I am not super Admin",
			args: args{
				ctx: ctx2,
				req: &v1.CreateGenericScopeEquipmentTypesRequest{
					Scope: "C",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure",
			args: args{
				ctx: ctx,
				req: &v1.CreateGenericScopeEquipmentTypesRequest{
					Scope: "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UpsertMetadata(ctx, &metadata[0]).Return("", errors.New("DgrpahError")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewEquipmentServiceServer(rep, nil)
			_, err := s.CreateGenericScopeEquipmentTypes(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("equipmentServiceServer.CreateGenericScopeEquipmentTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.setup == nil {
				mockCtrl.Finish()
			}
		})
	}
}

func Test_equipmentServiceServer_EquipmentsTypes(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	type args struct {
		ctx context.Context
		req *v1.EquipmentTypesRequest
	}
	var mockCtrl *gomock.Controller
	var rep repo.Equipment
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    *v1.EquipmentTypesResponse
		wantErr bool
	}{
		{name: "success",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentTypesRequest{
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:       "typ1",
						ID:         "1",
						SourceID:   "s1",
						ParentID:   "p1",
						ParentType: "typ_parent",
						SourceName: "equip1.csv",
						Scopes:     []string{"A"},
						Attributes: []*repo.Attribute{
							{
								ID:                 "1",
								Name:               "attr_1",
								Type:               repo.DataTypeString,
								IsIdentifier:       true,
								IsDisplayed:        true,
								IsSearchable:       true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_1",
							},
							{
								ID:                 "2",
								Name:               "attr_2",
								Type:               repo.DataTypeInt,
								IsIdentifier:       false,
								IsDisplayed:        true,
								IsSearchable:       false,
								IsParentIdentifier: true,
								MappedTo:           "mapping_2",
							},
							{
								ID:                 "3",
								Name:               "attr_3",
								Type:               repo.DataTypeFloat,
								IsIdentifier:       false,
								IsDisplayed:        false,
								IsSearchable:       false,
								IsParentIdentifier: false,
								MappedTo:           "mapping_3",
							},
						},
					},
					{
						Type:     "typ2",
						ID:       "2",
						SourceID: "s2",
						ParentID: "p2",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								ID:                 "1",
								Name:               "attr_1",
								Type:               repo.DataTypeString,
								IsIdentifier:       true,
								IsDisplayed:        true,
								IsSearchable:       true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_1",
							},
						},
					},
				}, nil)
			},
			want: &v1.EquipmentTypesResponse{
				EquipmentTypes: []*v1.EquipmentType{
					{
						ID:             "1",
						Type:           "typ1",
						ParentId:       "p1",
						MetadataId:     "s1",
						ParentType:     "typ_parent",
						MetadataSource: "equip1.csv",
						Scopes:         []string{"A"},
						Attributes: []*v1.Attribute{
							{
								ID:               "1",
								Name:             "attr_1",
								DataType:         v1.DataTypes_STRING,
								PrimaryKey:       true,
								Displayed:        true,
								Searchable:       true,
								ParentIdentifier: true,
								MappedTo:         "mapping_1",
							},
							{
								ID:               "2",
								Name:             "attr_2",
								DataType:         v1.DataTypes_INT,
								PrimaryKey:       false,
								Displayed:        true,
								Searchable:       false,
								ParentIdentifier: true,
								MappedTo:         "mapping_2",
							},
							{
								ID:               "3",
								Name:             "attr_3",
								DataType:         v1.DataTypes_FLOAT,
								PrimaryKey:       false,
								Displayed:        false,
								Searchable:       false,
								ParentIdentifier: false,
								MappedTo:         "mapping_3",
							},
						},
					},
					{
						ID:         "2",
						Type:       "typ2",
						ParentId:   "p2",
						MetadataId: "s2",
						Scopes:     []string{"A"},
						Attributes: []*v1.Attribute{
							{
								ID:               "1",
								Name:             "attr_1",
								DataType:         v1.DataTypes_STRING,
								PrimaryKey:       true,
								Displayed:        true,
								Searchable:       true,
								ParentIdentifier: true,
								MappedTo:         "mapping_1",
							},
						},
					},
				},
			},
		},
		{name: "failure|can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.EquipmentTypesRequest{
					Scopes: []string{"A"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure|scope does not belong to user",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentTypesRequest{
					Scopes: []string{"C"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentTypesRequest{
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewEquipmentServiceServer(rep, nil)
			got, err := s.EquipmentsTypes(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("equipmentServiceServer.EquipmentsTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareEquipmentTypesResponse(t, "EquipmentTypesResponse", tt.want, got)
			}
			if tt.setup == nil {
				mockCtrl.Finish()
			}
		})
	}
}

func Test_equipmentServiceServer_CreateEquipmentType(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Equipment
	var acc accv1.AccountServiceClient
	type args struct {
		ctx context.Context
		req *v1.EquipmentType
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    *v1.EquipmentType
		wantErr bool
	}{
		{name: "success",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						{
							Name:             "attr_2",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
						{
							Name:     "attr_3",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_3",
						},
						{
							Name:     "attr_4",
							DataType: v1.DataTypes_INT,
							MappedTo: "mapping_4",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
					{
						Type:     "typ3",
						ID:       "p2",
						SourceID: "s3",
						ParentID: "p1",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).
					Times(1).Return(&repo.Metadata{
					ID:     "s1",
					Source: "equip_1.csv",
					Attributes: []string{
						"mapping_1",
						"mapping_2",
						"mapping_3",
						"mapping_4",
					},
				}, nil)
				eqType := &repo.EquipmentType{
					Type:     "typ1",
					SourceID: "s1",
					ParentID: "p1",
					Scopes:   []string{"A"},
					Attributes: []*repo.Attribute{
						{
							Name:         "attr_1",
							Type:         repo.DataTypeString,
							IsIdentifier: true,
							IsDisplayed:  true,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						{
							Name:               "attr_2",
							Type:               repo.DataTypeString,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_2",
						},
						{
							Name:     "attr_3",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_3",
						},
						{
							Name:     "attr_4",
							Type:     repo.DataTypeInt,
							MappedTo: "mapping_4",
						},
					},
				}
				retEqType := &repo.EquipmentType{
					Type:     "typ1",
					ID:       "1",
					SourceID: "s1",
					ParentID: "p1",
					Scopes:   []string{"A"},
					Attributes: []*repo.Attribute{
						{
							ID:           "1",
							Name:         "attr_1",
							Type:         repo.DataTypeString,
							IsIdentifier: true,
							IsDisplayed:  true,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						{
							ID:                 "2",
							Name:               "attr_2",
							Type:               repo.DataTypeString,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_2",
						},
						{
							ID:       "3",
							Name:     "attr_3",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_3",
						},
						{
							ID:       "4",
							Name:     "attr_4",
							Type:     repo.DataTypeInt,
							MappedTo: "mapping_4",
						},
					},
				}
				mockRepo.EXPECT().CreateEquipmentType(ctx, eqType, []string{"A"}).Times(1).Return(retEqType, nil)
			},
			want: &v1.EquipmentType{
				ID:         "1",
				Type:       "typ1",
				ParentId:   "p1",
				MetadataId: "s1",
				Scopes:     []string{"A"},
				Attributes: []*v1.Attribute{
					{
						ID:               "1",
						Name:             "attr_1",
						DataType:         v1.DataTypes_STRING,
						PrimaryKey:       true,
						Displayed:        true,
						Searchable:       true,
						ParentIdentifier: false,
						MappedTo:         "mapping_1",
					},
					{
						ID:               "2",
						Name:             "attr_2",
						DataType:         v1.DataTypes_STRING,
						PrimaryKey:       false,
						Displayed:        true,
						Searchable:       false,
						ParentIdentifier: true,
						MappedTo:         "mapping_2",
					},
					{
						ID:       "3",
						Name:     "attr_3",
						DataType: v1.DataTypes_FLOAT,
						MappedTo: "mapping_3",
					},
					{
						ID:       "4",
						Name:     "attr_4",
						DataType: v1.DataTypes_INT,
						MappedTo: "mapping_4",
					},
				},
			},
		},
		{name: "failure|can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure|some claims are not owned by user",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"C"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure|unable to get scope info",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(nil, errors.New("service error"))
			},
			wantErr: true,
		},
		{name: "failure|creation not allowed on generic scope",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "GENERIC",
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|validation data source consumed",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s1",
					},
					{
						Type:     "typ3",
						ID:       "p2",
						SourceID: "s3",
						ParentID: "p1",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|type name used",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ1",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|type name used - case insensitive",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "Typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ1",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|parent not exists",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{}, nil)
			},
			wantErr: true,
		},
		{name: "failure|primary key not exits",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: false,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{}, nil)
			},
			wantErr: true,
		},
		{name: "failure|multiple primary key exits",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						{
							Name:       "attr_2",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_2",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{}, nil)
			},
			wantErr: true,
		},
		{name: "failure|parent id is not given and attribute has mapping of parent",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						{
							Name:             "attr_2",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{}, nil)
			},
			wantErr: true,
		},
		{name: "failure|multiple attribtes having parent identifier",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						{
							Name:             "attr_2",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
						{
							Name:             "attr_3",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{}, nil)
			},
			wantErr: true,
		},
		{name: "failure|multiple attribtes having same name",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						{
							Name:             "attr_1",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|attribute mapping not found",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						{
							Name:             "attr_2",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
						{
							Name:       "attr_3",
							DataType:   v1.DataTypes_STRING,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_3",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|multiple attributes have same mapping",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						{
							Name:             "attr_2",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
						{
							Name:       "attr_3",
							DataType:   v1.DataTypes_STRING,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_2",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|attributes is parent and primary key",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr_1",
							DataType:         v1.DataTypes_STRING,
							PrimaryKey:       true,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_1",
						},
						{
							Name:       "attr_3",
							DataType:   v1.DataTypes_STRING,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_3",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|primary key attribute type is not string",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_INT,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|parent key attribute type is not string",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						{
							Name:             "attr_2",
							DataType:         v1.DataTypes_FLOAT,
							ParentIdentifier: true,
							Displayed:        true,
							Searchable:       true,
							MappedTo:         "mapping_2",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|primary key should be displayable",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  false,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						{
							Name:       "attr_2",
							DataType:   v1.DataTypes_STRING,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_2",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|searchable attribute should be displayable",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						{
							Name:       "attr_2",
							DataType:   v1.DataTypes_STRING,
							Displayed:  false,
							Searchable: true,
							MappedTo:   "mapping_2",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "failure|getting equipment type",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "failure|getting metadata with id no rows",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(nil, repo.ErrNoData)
			},
			wantErr: true,
		},
		{name: "failure|getting metadata with id unknown error",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "failure|creating equipment",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentType{
					Type:       "typ1",
					ParentId:   "p1",
					MetadataId: "s1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2"},
				}, nil)

				eqType := &repo.EquipmentType{
					Type:     "typ1",
					SourceID: "s1",
					ParentID: "p1",
					Scopes:   []string{"A"},
					Attributes: []*repo.Attribute{
						{
							Name:         "attr_1",
							Type:         repo.DataTypeString,
							IsIdentifier: true,
							IsDisplayed:  true,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
					},
				}
				mockRepo.EXPECT().CreateEquipmentType(ctx, eqType, []string{"A"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &equipmentServiceServer{
				equipmentRepo: rep,
				account:       acc,
			}
			got, err := s.CreateEquipmentType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("equipmentServiceServer.CreateEquipmentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareEquipmentType(t, "EquipmentType", got, tt.want)
			}
		})
	}
}

func Test_equipmentServiceServer_ListEquipmentsMetadata(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Equipment
	type args struct {
		ctx context.Context
		req *v1.ListEquipmentMetadataRequest
	}
	tests := []struct {
		name    string
		s       *equipmentServiceServer
		args    args
		setup   func()
		want    *v1.ListEquipmentMetadataResponse
		wantErr bool
	}{
		{name: "success - ALL MetaData",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type:   v1.ListEquipmentMetadataRequest_ALL,
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						SourceID: "2",
					},
					{
						SourceID: "3",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A"}).
					Times(1).Return([]*repo.Metadata{
					{
						ID:     "1",
						Source: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
					{
						ID:     "2",
						Source: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
					{
						ID:     "3",
						Source: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
					{
						ID:     "4",
						Source: "equip_4.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
				}, nil)
			},
			want: &v1.ListEquipmentMetadataResponse{
				Metadata: []*v1.EquipmentMetadata{
					{
						ID:   "1",
						Name: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scopes: []string{"A"},
					},
					{
						ID:   "2",
						Name: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scopes: []string{"A"},
					},
					{
						ID:   "3",
						Name: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scopes: []string{"A"},
					},
					{
						ID:   "4",
						Name: "equip_4.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scopes: []string{"A"},
					},
				},
			},
			wantErr: false,
		},
		{name: "success - Mapped MetaData",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type:   v1.ListEquipmentMetadataRequest_MAPPED,
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						SourceID: "2",
					},
					{
						SourceID: "3",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A"}).
					Times(1).Return([]*repo.Metadata{
					{
						ID:     "1",
						Source: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
					{
						ID:     "2",
						Source: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
					{
						ID:     "3",
						Source: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
					{
						ID:     "4",
						Source: "equip_4.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
				}, nil)
			},
			want: &v1.ListEquipmentMetadataResponse{
				Metadata: []*v1.EquipmentMetadata{
					{
						ID:   "2",
						Name: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scopes: []string{"A"},
					},
					{
						ID:   "3",
						Name: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scopes: []string{"A"},
					},
				},
			},
			wantErr: false,
		},
		{name: "success - Un-Mapped MetaData",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type:   v1.ListEquipmentMetadataRequest_UN_MAPPED,
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						SourceID: "2",
					},
					{
						SourceID: "3",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A"}).
					Times(1).Return([]*repo.Metadata{
					{
						ID:     "1",
						Source: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
					{
						ID:     "2",
						Source: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
					{
						ID:     "3",
						Source: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
					{
						ID:     "4",
						Source: "equip_4.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scope: "A",
					},
				}, nil)
			},
			want: &v1.ListEquipmentMetadataResponse{
				Metadata: []*v1.EquipmentMetadata{
					{
						ID:   "1",
						Name: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scopes: []string{"A"},
					},
					{
						ID:   "4",
						Name: "equip_4.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
						Scopes: []string{"A"},
					},
				},
			},
			wantErr: false,
		},
		{name: "failure|can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListEquipmentMetadataRequest{
					Type:   v1.ListEquipmentMetadataRequest_UN_MAPPED,
					Scopes: []string{"A"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure|some claims are not owned by user",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type:   v1.ListEquipmentMetadataRequest_UN_MAPPED,
					Scopes: []string{"C"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure, fetching equipment type",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type:   v1.ListEquipmentMetadataRequest_UN_MAPPED,
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "failure, fetching metadata, no data",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type:   v1.ListEquipmentMetadataRequest_UN_MAPPED,
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						SourceID: "2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A"}).
					Times(1).Return(nil, repo.ErrNoData)
			},
			wantErr: true,
		},
		{name: "failure, fetching metadata",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type:   v1.ListEquipmentMetadataRequest_UN_MAPPED,
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						SourceID: "2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A"}).
					Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "failure - default query parameter",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type:   10000,
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						SourceID: "2",
					},
					{
						SourceID: "3",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A"}).
					Times(1).Return([]*repo.Metadata{
					{
						ID:     "1",
						Source: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					{
						ID:     "2",
						Source: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					{
						ID:     "3",
						Source: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					{
						ID:     "4",
						Source: "equip_4.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
				}, nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewEquipmentServiceServer(rep, nil)
			got, err := s.ListEquipmentsMetadata(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("equipmentServiceServer.ListEquipmentsMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareEquipmentMetadataResponse(t, "EquipmentMetadatas", tt.want, got)
			}
			if tt.setup == nil {
				mockCtrl.Finish()
			}
		})
	}
}

func Test_equipmentServiceServer_UpdateEquipmentType(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Equipment
	var acc accv1.AccountServiceClient
	type args struct {
		ctx context.Context
		req *v1.UpdateEquipmentTypeRequest
	}
	tests := []struct {
		name    string
		s       *equipmentServiceServer
		args    args
		setup   func()
		want    *v1.EquipmentType
		wantErr bool
	}{
		{name: "SUCCESS - parent id already exists",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "3",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:               "attr2",
								Type:               repo.DataTypeString,
								IsDisplayed:        true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().Equipments(ctx, &repo.EquipmentType{
					ID:       "1",
					Type:     "MyType",
					SourceID: "s1",
					ParentID: "3",
					Scopes:   []string{"A"},
					Attributes: []*repo.Attribute{
						{
							Name:         "attr1",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						{
							Name:               "attr2",
							Type:               repo.DataTypeString,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_2",
						},
					},
				}, &repo.QueryEquipments{
					PageSize:  50,
					Offset:    offset(50, 1),
					SortOrder: sortOrder(v1.SortOrder_ASC),
				}, []string{"A"}).Times(1).Return(int32(0), nil, repo.ErrNoData)
				mockRepo.EXPECT().UpdateEquipmentType(ctx, "1", "MyType", "3", &repo.UpdateEquipmentRequest{
					ParentID: "2",
					Attr: []*repo.Attribute{
						{
							Name:         "attr3",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_3",
						},
						{
							Name:         "attr4",
							Type:         repo.DataTypeInt,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_4",
						},
						{
							Name:     "attr5",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_5",
						},
					},
				}, []string{"A"}).Times(1).Return([]*repo.Attribute{
					{
						Name:         "attr3",
						Type:         repo.DataTypeString,
						IsSearchable: true,
						IsDisplayed:  true,
						MappedTo:     "mapping_3",
					},
					{
						Name:         "attr4",
						Type:         repo.DataTypeInt,
						IsSearchable: true,
						IsDisplayed:  true,
						MappedTo:     "mapping_4",
					},
					{
						Name:     "attr5",
						Type:     repo.DataTypeFloat,
						MappedTo: "mapping_5",
					},
				}, nil)
			},
			want: &v1.EquipmentType{
				ID:         "1",
				Type:       "MyType",
				MetadataId: "s1",
				ParentId:   "2",
				Scopes:     []string{"A"},
				Attributes: []*v1.Attribute{
					{
						Name:       "attr1",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						MappedTo:   "mapping_1",
					},
					{
						Name:             "attr2",
						DataType:         v1.DataTypes_STRING,
						Displayed:        true,
						ParentIdentifier: true,
						MappedTo:         "mapping_2",
					},
					{
						Name:       "attr3",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						Displayed:  true,
						MappedTo:   "mapping_3",
					},
					{
						Name:       "attr4",
						DataType:   v1.DataTypes_INT,
						Searchable: true,
						Displayed:  true,
						MappedTo:   "mapping_4",
					},
					{
						Name:     "attr5",
						DataType: v1.DataTypes_FLOAT,
						MappedTo: "mapping_5",
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - RoleSuperAdmin",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   claims.RoleSuperAdmin,
					Socpes: []string{"A", "B"},
				}),
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   claims.RoleSuperAdmin,
					Socpes: []string{"A", "B"},
				}), []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "3",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:               "attr2",
								Type:               repo.DataTypeString,
								IsDisplayed:        true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   claims.RoleSuperAdmin,
					Socpes: []string{"A", "B"},
				}), "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   claims.RoleSuperAdmin,
					Socpes: []string{"A", "B"},
				}), "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().Equipments(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   claims.RoleSuperAdmin,
					Socpes: []string{"A", "B"},
				}), &repo.EquipmentType{
					ID:       "1",
					Type:     "MyType",
					SourceID: "s1",
					ParentID: "3",
					Scopes:   []string{"A"},
					Attributes: []*repo.Attribute{
						{
							Name:         "attr1",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						{
							Name:               "attr2",
							Type:               repo.DataTypeString,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_2",
						},
					},
				}, &repo.QueryEquipments{
					PageSize:  50,
					Offset:    offset(50, 1),
					SortOrder: sortOrder(v1.SortOrder_ASC),
				}, []string{"A"}).Times(1).Return(int32(0), nil, repo.ErrNoData)
				mockRepo.EXPECT().UpdateEquipmentType(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   claims.RoleSuperAdmin,
					Socpes: []string{"A", "B"},
				}), "1", "MyType", "3", &repo.UpdateEquipmentRequest{
					ParentID: "2",
					Attr: []*repo.Attribute{
						{
							Name:         "attr3",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_3",
						},
						{
							Name:         "attr4",
							Type:         repo.DataTypeInt,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_4",
						},
						{
							Name:     "attr5",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_5",
						},
					},
				}, []string{"A"}).Times(1).Return([]*repo.Attribute{
					{
						Name:         "attr3",
						Type:         repo.DataTypeString,
						IsSearchable: true,
						IsDisplayed:  true,
						MappedTo:     "mapping_3",
					},
					{
						Name:         "attr4",
						Type:         repo.DataTypeInt,
						IsSearchable: true,
						IsDisplayed:  true,
						MappedTo:     "mapping_4",
					},
					{
						Name:     "attr5",
						Type:     repo.DataTypeFloat,
						MappedTo: "mapping_5",
					},
				}, nil)
			},
			want: &v1.EquipmentType{
				ID:         "1",
				Type:       "MyType",
				MetadataId: "s1",
				ParentId:   "2",
				Scopes:     []string{"A"},
				Attributes: []*v1.Attribute{
					{
						Name:       "attr1",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						MappedTo:   "mapping_1",
					},
					{
						Name:             "attr2",
						DataType:         v1.DataTypes_STRING,
						Displayed:        true,
						ParentIdentifier: true,
						MappedTo:         "mapping_2",
					},
					{
						Name:       "attr3",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						Displayed:  true,
						MappedTo:   "mapping_3",
					},
					{
						Name:       "attr4",
						DataType:   v1.DataTypes_INT,
						Searchable: true,
						Displayed:  true,
						MappedTo:   "mapping_4",
					},
					{
						Name:     "attr5",
						DataType: v1.DataTypes_FLOAT,
						MappedTo: "mapping_5",
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - parent id does not exist",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 2, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().UpdateEquipmentType(ctx, "1", "MyType", "", &repo.UpdateEquipmentRequest{
					ParentID: "2",
					Attr: []*repo.Attribute{
						{
							Name:               "attr3",
							Type:               repo.DataTypeString,
							IsSearchable:       true,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_3",
						},
						{
							Name:         "attr4",
							Type:         repo.DataTypeInt,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_4",
						},
						{
							Name:     "attr5",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_5",
						},
					},
				}, []string{"A"}).Times(1).Return([]*repo.Attribute{
					{
						Name:               "attr3",
						Type:               repo.DataTypeString,
						IsSearchable:       true,
						IsDisplayed:        true,
						IsParentIdentifier: true,
						MappedTo:           "mapping_3",
					},
					{
						Name:         "attr4",
						Type:         repo.DataTypeInt,
						IsSearchable: true,
						IsDisplayed:  true,
						MappedTo:     "mapping_4",
					},
					{
						Name:     "attr5",
						Type:     repo.DataTypeFloat,
						MappedTo: "mapping_5",
					},
				}, nil)
			},
			want: &v1.EquipmentType{
				ID:         "1",
				Type:       "MyType",
				MetadataId: "s1",
				Scopes:     []string{"A"},
				ParentId:   "2",
				Attributes: []*v1.Attribute{
					{
						Name:       "attr1",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						MappedTo:   "mapping_1",
					},
					{
						Name:       "attr2",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						MappedTo:   "mapping_2",
					},
					{
						Name:             "attr3",
						DataType:         v1.DataTypes_STRING,
						Searchable:       true,
						Displayed:        true,
						ParentIdentifier: true,
						MappedTo:         "mapping_3",
					},
					{
						Name:       "attr4",
						DataType:   v1.DataTypes_INT,
						Searchable: true,
						Displayed:  true,
						MappedTo:   "mapping_4",
					},
					{
						Name:     "attr5",
						DataType: v1.DataTypes_FLOAT,
						MappedTo: "mapping_5",
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - only attribute is added",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
							{
								Name:               "attr3",
								Type:               repo.DataTypeString,
								IsSearchable:       true,
								IsDisplayed:        true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_3",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().UpdateEquipmentType(ctx, "1", "MyType", "2", &repo.UpdateEquipmentRequest{
					ParentID: "2",
					Attr: []*repo.Attribute{
						{
							Name:         "attr4",
							Type:         repo.DataTypeInt,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_4",
						},
						{
							Name:     "attr5",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_5",
						},
					},
				}, []string{"A"}).Times(1).Return([]*repo.Attribute{
					{
						Name:         "attr4",
						Type:         repo.DataTypeInt,
						IsSearchable: true,
						IsDisplayed:  true,
						MappedTo:     "mapping_4",
					},
					{
						Name:     "attr5",
						Type:     repo.DataTypeFloat,
						MappedTo: "mapping_5",
					},
				}, nil)
			},
			want: &v1.EquipmentType{
				ID:         "1",
				Type:       "MyType",
				MetadataId: "s1",
				Scopes:     []string{"A"},
				ParentId:   "2",
				Attributes: []*v1.Attribute{
					{
						Name:       "attr1",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						MappedTo:   "mapping_1",
					},
					{
						Name:       "attr2",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						MappedTo:   "mapping_2",
					},
					{
						Name:             "attr3",
						DataType:         v1.DataTypes_STRING,
						Searchable:       true,
						Displayed:        true,
						ParentIdentifier: true,
						MappedTo:         "mapping_3",
					},
					{
						Name:       "attr4",
						DataType:   v1.DataTypes_INT,
						Searchable: true,
						Displayed:  true,
						MappedTo:   "mapping_4",
					},
					{
						Name:     "attr5",
						DataType: v1.DataTypes_FLOAT,
						MappedTo: "mapping_5",
					},
				},
			},
			wantErr: false,
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - some claims are not owned by user",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"C"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure|unable to get scope info",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   claims.RoleUser,
					Socpes: []string{"A", "B"},
				}),
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   claims.RoleUser,
					Socpes: []string{"A", "B"},
				}), &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(nil, errors.New("service error"))
			},
			wantErr: true,
		},
		{name: "failure|creation not allowed on generic scope",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   claims.RoleUser,
					Socpes: []string{"A", "B"},
				}),
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   claims.RoleUser,
					Socpes: []string{"A", "B"},
				}), &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "GENERIC",
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - repo/EquipmentTypes - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "FAILURE - equipment type does not exist",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "3",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"}},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - repo/MetadataWithID - cannot fetch meta data",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "FAILURE - repo/MetadataWithID - no meta data",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(nil, repo.ErrNoData)
			},
			wantErr: true,
		},
		{name: "FAILURE - parent not found/not exists ",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "3",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - parent cannot be same equipment",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "1",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - repo/EquipmentTypeChildren - cannot fetch equipment children",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "4",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "3",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:               "attr2",
								Type:               repo.DataTypeString,
								IsDisplayed:        true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - parent can not be any of the children",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "4",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "3",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:               "attr2",
								Type:               repo.DataTypeString,
								IsDisplayed:        true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - parent id already exists - equipment type contains equipments data",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "3",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:               "attr2",
								Type:               repo.DataTypeString,
								IsDisplayed:        true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().Equipments(ctx, &repo.EquipmentType{
					ID:       "1",
					Type:     "MyType",
					SourceID: "s1",
					ParentID: "3",
					Scopes:   []string{"A"},
					Attributes: []*repo.Attribute{
						{
							Name:         "attr1",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						{
							Name:               "attr2",
							Type:               repo.DataTypeString,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_2",
						},
					},
				}, &repo.QueryEquipments{
					PageSize:  50,
					Offset:    offset(50, 1),
					SortOrder: sortOrder(v1.SortOrder_ASC),
				}, []string{"A"}).Times(1).Return(int32(2), json.RawMessage(`{"id":1}`), nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - parent id already exists - cannot fetch equipments",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "3",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:               "attr2",
								Type:               repo.DataTypeString,
								IsDisplayed:        true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().Equipments(ctx, &repo.EquipmentType{
					ID:       "1",
					Type:     "MyType",
					SourceID: "s1",
					ParentID: "3",
					Scopes:   []string{"A"},
					Attributes: []*repo.Attribute{
						{
							Name:         "attr1",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						{
							Name:               "attr2",
							Type:               repo.DataTypeString,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_2",
						},
					},
				}, &repo.QueryEquipments{
					PageSize:  50,
					Offset:    offset(50, 1),
					SortOrder: sortOrder(v1.SortOrder_ASC),
				}, []string{"A"}).Times(1).Return(int32(0), nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - validateEquipUpdation - primary key not required",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						{
							Name:       "attr4",
							PrimaryKey: true,
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 2, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - validateEquipUpdation - only string data type is allowed for parent identifier",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr3",
							DataType:         v1.DataTypes_FLOAT,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - validateEquipUpdation - one parent identifier required",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - validateEquipUpdation - no parent identifier required when parent is already present",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "3",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:               "attr2",
								Type:               repo.DataTypeString,
								IsDisplayed:        true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().Equipments(ctx, &repo.EquipmentType{
					ID:       "1",
					Type:     "MyType",
					SourceID: "s1",
					ParentID: "3",
					Scopes:   []string{"A"},
					Attributes: []*repo.Attribute{
						{
							Name:         "attr1",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						{
							Name:               "attr2",
							Type:               repo.DataTypeString,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_2",
						},
					},
				}, &repo.QueryEquipments{
					PageSize:  50,
					Offset:    offset(50, 1),
					SortOrder: sortOrder(v1.SortOrder_ASC),
				}, []string{"A"}).Times(1).Return(int32(0), nil, repo.ErrNoData)
			},
			wantErr: true,
		},
		{name: "FAILURE - validateEquipUpdation - parent is not selected for equipment type",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:        "attr2",
								Type:        repo.DataTypeString,
								IsDisplayed: true,
								MappedTo:    "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - validateEquipUpdation - multiple parent keys are found",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						{
							Name:             "attr4",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:        "attr2",
								Type:        repo.DataTypeString,
								IsDisplayed: true,
								MappedTo:    "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - validateEquipUpdation - attribute name already exists",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr2",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:        "attr2",
								Type:        repo.DataTypeString,
								IsDisplayed: true,
								MappedTo:    "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - validateEquipUpdation - attribute mapping does not exists",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_7",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:        "attr2",
								Type:        repo.DataTypeString,
								IsDisplayed: true,
								MappedTo:    "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - validateEquipUpdation - mapping already given to some attribute",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:        "attr2",
								Type:        repo.DataTypeString,
								IsDisplayed: true,
								MappedTo:    "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - validateEquipUpdation - searchable object should always be displayable",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:        "attr2",
								Type:        repo.DataTypeString,
								IsDisplayed: true,
								MappedTo:    "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - repo/UpdateEquipmentType - cannot update equipment type",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				mockAcc := accmock.NewMockAccountServiceClient(mockCtrl)
				acc = mockAcc
				rep = mockRepo
				mockAcc.EXPECT().GetScope(ctx, &accv1.GetScopeRequest{Scope: "A"}).Times(1).Return(&accv1.Scope{
					ScopeCode:  "A",
					ScopeName:  "Scope A",
					CreatedBy:  "admin@test.com",
					CreatedOn:  &timestamppb.Timestamp{},
					GroupNames: []string{"ROOT"},
					ScopeType:  "SPECIFIC",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "MyType3",
						SourceID: "s3",
					},
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 4, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "4",
						Type:     "MyType4",
						SourceID: "s4",
					},
				}, nil)
				mockRepo.EXPECT().UpdateEquipmentType(ctx, "1", "MyType", "", &repo.UpdateEquipmentRequest{
					ParentID: "2",
					Attr: []*repo.Attribute{
						{
							Name:               "attr3",
							Type:               repo.DataTypeString,
							IsSearchable:       true,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_3",
						},
						{
							Name:         "attr4",
							Type:         repo.DataTypeInt,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_4",
						},
						{
							Name:     "attr5",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_5",
						},
					},
				}, []string{"A"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &equipmentServiceServer{
				equipmentRepo: rep,
				account:       acc,
			}
			got, err := s.UpdateEquipmentType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("equipmentServiceServer.UpdateEquipmentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareEquipmentType(t, "EquipmentType", got, tt.want)
			}
		})
	}
}

func Test_equipmentServiceServer_GetEquipmentMetadata(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Equipment
	type args struct {
		ctx context.Context
		req *v1.EquipmentMetadataRequest
	}
	tests := []struct {
		name    string
		s       *equipmentServiceServer
		args    args
		setup   func()
		want    *v1.EquipmentMetadata
		wantErr bool
	}{
		{name: "SUCCESS - no equipment exists for given metadata",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID:     "1",
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:     "s3",
					Source: "equip_1.csv",
					Attributes: []string{
						"attr_1",
						"attr_2",
					},
					Scope: "A",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
				}, nil)
			},
			want: &v1.EquipmentMetadata{
				ID:   "s3",
				Name: "equip_1.csv",
				Attributes: []string{
					"attr_1",
					"attr_2",
				},
				Scopes: []string{"A"},
			},
			wantErr: false,
		},
		{name: "SUCCESS - ALL attributes",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID:         "1",
					Attributes: v1.EquipmentMetadataRequest_All,
					Scopes:     []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:     "s1",
					Source: "equip_1.csv",
					Attributes: []string{
						"attr_1",
						"attr_2",
					},
					Scope: "A",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Scopes:   []string{"A"},
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_1",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
						},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
				}, nil)
			},
			want: &v1.EquipmentMetadata{
				ID:   "s1",
				Name: "equip_1.csv",
				Attributes: []string{
					"attr_1",
					"attr_2",
				},
				Scopes: []string{"A"},
			},
			wantErr: false,
		},
		{name: "SUCCESS - MAPPED attributes",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID:         "1",
					Attributes: v1.EquipmentMetadataRequest_Mapped,
					Scopes:     []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:     "s1",
					Source: "equip_1.csv",
					Attributes: []string{
						"attr_1",
						"attr_2",
						"attr_3",
						"attr_4",
					},
					Scope: "A",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_3",
							},
						},
						Scopes: []string{"A"},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
				}, nil)
			},
			want: &v1.EquipmentMetadata{
				ID:   "s1",
				Name: "equip_1.csv",
				Attributes: []string{
					"attr_2",
					"attr_3",
				},
				Scopes: []string{"A"},
			},
			wantErr: false,
		},
		{name: "SUCCESS - UNMAPPED attributes",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID:         "1",
					Attributes: v1.EquipmentMetadataRequest_Unmapped,
					Scopes:     []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:     "s1",
					Source: "equip_1.csv",
					Attributes: []string{
						"attr_1",
						"attr_2",
						"attr_3",
						"attr_4",
					},
					Scope: "A",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_3",
							},
						},
						Scopes: []string{"A"},
					},
					{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
				}, nil)
			},
			want: &v1.EquipmentMetadata{
				ID:   "s1",
				Name: "equip_1.csv",
				Attributes: []string{
					"attr_1",
					"attr_4",
				},
				Scopes: []string{"A"},
			},
			wantErr: false,
		},
		{name: "failure|can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.EquipmentMetadataRequest{
					ID:     "1",
					Scopes: []string{"A"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure|some claims are not owned by user",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID:     "1",
					Scopes: []string{"C"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metadata - no metadata",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID:     "1",
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A"}).Times(1).Return(nil, repo.ErrNoData)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metadata",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID:     "1",
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A"}).Times(1).Return(nil, errors.New("Test Error"))
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID:     "1",
					Scopes: []string{"A"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A"}).Times(1).Return(&repo.Metadata{
					ID:     "s3",
					Source: "equip_1.csv",
					Attributes: []string{
						"attr_1",
						"attr_2",
					},
					Scope: "A",
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewEquipmentServiceServer(rep, nil)
			got, err := s.GetEquipmentMetadata(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("equipmentServiceServer.GetEquipmentMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareEquipmentMetadata(t, "EquipmentMetadata", tt.want, got)
			}
			if tt.setup == nil {
				mockCtrl.Finish()
			}
		})
	}
}

func Test_equipmentServiceServer_DeleteEquipmentType(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Equipment
	type args struct {
		ctx context.Context
		req *v1.DeleteEquipmentTypeRequest
	}
	tests := []struct {
		name    string
		s       *equipmentServiceServer
		args    args
		setup   func()
		want    *v1.DeleteEquipmentTypeResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DeleteEquipmentTypeRequest{
					EquipType: "server",
					Scope:     "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "server",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_3",
							},
						},
						Scopes: []string{"A"},
					},
					{
						ID:       "2",
						Type:     "cluster",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 2, []string{"A"}).Return(nil, repo.ErrNoData).Times(1)
				mockRepo.EXPECT().Equipments(ctx, &repo.EquipmentType{
					ID:       "1",
					Type:     "server",
					SourceID: "s1",
					ParentID: "2",
					Attributes: []*repo.Attribute{
						{
							Name:         "attr1",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "attr_2",
						},
						{
							Name:         "attr2",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "attr_3",
						},
					},
					Scopes: []string{"A"},
				}, &repo.QueryEquipments{
					PageSize:  50,
					Offset:    offset(50, 1),
					SortOrder: sortOrder(v1.SortOrder_ASC),
				}, []string{"A"}).Return(int32(0), nil, repo.ErrNoData).Times(1)
				mockRepo.EXPECT().DeleteEquipmentType(ctx, "server", "A").Return(nil).Times(1)
			},
			want: &v1.DeleteEquipmentTypeResponse{
				Success: true,
			},
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteEquipmentTypeRequest{
					EquipType: "server",
					Scope:     "A",
				},
			},
			setup: func() {},
			want: &v1.DeleteEquipmentTypeResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - some claims are not owned by user",
			args: args{
				ctx: ctx,
				req: &v1.DeleteEquipmentTypeRequest{
					EquipType: "server",
					Scope:     "C",
				},
			},
			setup: func() {},
			want: &v1.DeleteEquipmentTypeResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.DeleteEquipmentTypeRequest{
					EquipType: "server",
					Scope:     "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Return(nil, errors.New("DBError")).Times(1)
			},
			want: &v1.DeleteEquipmentTypeResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - equipment type does not exist",
			args: args{
				ctx: ctx,
				req: &v1.DeleteEquipmentTypeRequest{
					EquipType: "server",
					Scope:     "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Return([]*repo.EquipmentType{
					{
						ID:       "2",
						Type:     "cluster",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "3",
						Type:     "vcenter",
						SourceID: "s3",
						Scopes:   []string{"A"},
					},
				}, nil).Times(1)
			},
			want: &v1.DeleteEquipmentTypeResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipment type children",
			args: args{
				ctx: ctx,
				req: &v1.DeleteEquipmentTypeRequest{
					EquipType: "server",
					Scope:     "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "server",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_3",
							},
						},
						Scopes: []string{"A"},
					},
					{
						ID:       "2",
						Type:     "cluster",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 2, []string{"A"}).Return(nil, errors.New("DBError")).Times(1)
			},
			want: &v1.DeleteEquipmentTypeResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - equipment type has children",
			args: args{
				ctx: ctx,
				req: &v1.DeleteEquipmentTypeRequest{
					EquipType: "server",
					Scope:     "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "partition",
						SourceID: "s2",
						ParentID: "2",
						Scopes:   []string{"A"},
					},
					{
						ID:       "2",
						Type:     "server",
						SourceID: "s1",
						ParentID: "3",
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_3",
							},
						},
						Scopes: []string{"A"},
					},
					{
						ID:       "3",
						Type:     "cluster",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "2", 3, []string{"A"}).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "partition",
						SourceID: "s2",
						ParentID: "2",
						Scopes:   []string{"A"},
					},
				}, nil).Times(1)
			},
			want: &v1.DeleteEquipmentTypeResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipments",
			args: args{
				ctx: ctx,
				req: &v1.DeleteEquipmentTypeRequest{
					EquipType: "server",
					Scope:     "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "server",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_3",
							},
						},
						Scopes: []string{"A"},
					},
					{
						ID:       "2",
						Type:     "cluster",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 2, []string{"A"}).Return(nil, repo.ErrNoData).Times(1)
				mockRepo.EXPECT().Equipments(ctx, &repo.EquipmentType{
					ID:       "1",
					Type:     "server",
					SourceID: "s1",
					ParentID: "2",
					Attributes: []*repo.Attribute{
						{
							Name:         "attr1",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "attr_2",
						},
						{
							Name:         "attr2",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "attr_3",
						},
					},
					Scopes: []string{"A"},
				}, &repo.QueryEquipments{
					PageSize:  50,
					Offset:    offset(50, 1),
					SortOrder: sortOrder(v1.SortOrder_ASC),
				}, []string{"A"}).Return(int32(0), nil, errors.New("DBError")).Times(1)
			},
			want: &v1.DeleteEquipmentTypeResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - equipment type contains equipments data",
			args: args{
				ctx: ctx,
				req: &v1.DeleteEquipmentTypeRequest{
					EquipType: "server",
					Scope:     "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "server",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_3",
							},
						},
						Scopes: []string{"A"},
					},
					{
						ID:       "2",
						Type:     "cluster",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 2, []string{"A"}).Return(nil, repo.ErrNoData).Times(1)
				mockRepo.EXPECT().Equipments(ctx, &repo.EquipmentType{
					ID:       "1",
					Type:     "server",
					SourceID: "s1",
					ParentID: "2",
					Attributes: []*repo.Attribute{
						{
							Name:         "attr1",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "attr_2",
						},
						{
							Name:         "attr2",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "attr_3",
						},
					},
					Scopes: []string{"A"},
				}, &repo.QueryEquipments{
					PageSize:  50,
					Offset:    offset(50, 1),
					SortOrder: sortOrder(v1.SortOrder_ASC),
				}, []string{"A"}).Return(int32(2), json.RawMessage(`[{ID:"1"}]`), nil).Times(1)
			},
			want: &v1.DeleteEquipmentTypeResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot delete equipment type",
			args: args{
				ctx: ctx,
				req: &v1.DeleteEquipmentTypeRequest{
					EquipType: "server",
					Scope:     "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockEquipment(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Return([]*repo.EquipmentType{
					{
						ID:       "1",
						Type:     "server",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
							{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_3",
							},
						},
						Scopes: []string{"A"},
					},
					{
						ID:       "2",
						Type:     "cluster",
						SourceID: "s2",
						Scopes:   []string{"A"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().EquipmentTypeChildren(ctx, "1", 2, []string{"A"}).Return(nil, repo.ErrNoData).Times(1)
				mockRepo.EXPECT().Equipments(ctx, &repo.EquipmentType{
					ID:       "1",
					Type:     "server",
					SourceID: "s1",
					ParentID: "2",
					Attributes: []*repo.Attribute{
						{
							Name:         "attr1",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "attr_2",
						},
						{
							Name:         "attr2",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							MappedTo:     "attr_3",
						},
					},
					Scopes: []string{"A"},
				}, &repo.QueryEquipments{
					PageSize:  50,
					Offset:    offset(50, 1),
					SortOrder: sortOrder(v1.SortOrder_ASC),
				}, []string{"A"}).Return(int32(0), nil, repo.ErrNoData).Times(1)
				mockRepo.EXPECT().DeleteEquipmentType(ctx, "server", "A").Return(errors.New("DBError")).Times(1)
			},
			want: &v1.DeleteEquipmentTypeResponse{
				Success: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewEquipmentServiceServer(rep, nil)
			got, err := s.DeleteEquipmentType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("equipmentServiceServer.DeleteEquipmentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("equipmentServiceServer.DeleteEquipmentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareEquipmentTypesResponse(t *testing.T, name string, exp *v1.EquipmentTypesResponse, act *v1.EquipmentTypesResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	compareEquipmentTypeAll(t, name+".EquipmentTypes", exp.EquipmentTypes, act.EquipmentTypes)
}

func compareEquipmentTypeAll(t *testing.T, name string, exp []*v1.EquipmentType, act []*v1.EquipmentType) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareEquipmentType(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareEquipmentType(t *testing.T, name string, exp *v1.EquipmentType, act *v1.EquipmentType) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID are not same", name)
	}

	assert.Equalf(t, exp.Type, act.Type, "%s.Type are not same", name)
	assert.Equalf(t, exp.ParentId, act.ParentId, "%s.ParentId are not same", name)
	assert.Equalf(t, exp.MetadataId, act.MetadataId, "%s.MetadataId are not same", name)
	assert.Equalf(t, exp.ParentType, act.ParentType, "%s.MetadataId are not same", name)
	assert.Equalf(t, exp.MetadataSource, act.MetadataSource, "%s.MetadataId are not same", name)
	assert.Equalf(t, exp.Scopes, act.Scopes, "%s.Scope are not same", name)
	compareAttributeAll(t, fmt.Sprintf("%s.Attributes are not same", name), exp.Attributes, act.Attributes)
}

func compareAttributeAll(t *testing.T, name string, exp []*v1.Attribute, act []*v1.Attribute) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareAttribute(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareAttribute(t *testing.T, name string, exp *v1.Attribute, act *v1.Attribute) {
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
	assert.Equalf(t, exp.DataType, act.DataType, "%s.DataType are not same", name)
	assert.Equalf(t, exp.PrimaryKey, act.PrimaryKey, "%s.PrimaryKey are not same", name)
	assert.Equalf(t, exp.Displayed, act.Displayed, "%s.Displayed are not same", name)
	assert.Equalf(t, exp.Searchable, act.Searchable, "%s.Searchable are not same", name)
	assert.Equalf(t, exp.ParentIdentifier, act.ParentIdentifier, "%s.ParentIdentifier are not same", name)
	assert.Equalf(t, exp.MappedTo, act.MappedTo, "%s.MappedTo are not same", name)
}

func compareEquipmentMetadataResponse(t *testing.T, name string, exp *v1.ListEquipmentMetadataResponse, act *v1.ListEquipmentMetadataResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	compareEquipmentMetadataAll(t, name+".Metadata", exp.Metadata, act.Metadata)
}

func compareEquipmentMetadataAll(t *testing.T, name string, exp []*v1.EquipmentMetadata, act []*v1.EquipmentMetadata) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareEquipmentMetadata(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareEquipmentMetadata(t *testing.T, name string, exp *v1.EquipmentMetadata, act *v1.EquipmentMetadata) {
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
	assert.ElementsMatchf(t, exp.Attributes, act.Attributes, "%s.Attributes are not same", name)
}
