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
	"fmt"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_licenseServiceServer_EquipmentsTypes(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	type args struct {
		ctx context.Context
		req *v1.EquipmentTypesRequest
	}
	var mockCtrl *gomock.Controller
	var rep repo.License
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
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:       "typ1",
						ID:         "1",
						SourceID:   "s1",
						ParentID:   "p1",
						ParentType: "typ_parent",
						SourceName: "equip1.csv",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:                 "1",
								Name:               "attr_1",
								Type:               repo.DataTypeString,
								IsIdentifier:       true,
								IsDisplayed:        true,
								IsSearchable:       true,
								IsParentIdentifier: true,
								MappedTo:           "mapping_1",
							},
							&repo.Attribute{
								ID:                 "2",
								Name:               "attr_2",
								Type:               repo.DataTypeInt,
								IsIdentifier:       false,
								IsDisplayed:        true,
								IsSearchable:       false,
								IsParentIdentifier: true,
								MappedTo:           "mapping_2",
							},
							&repo.Attribute{
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
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "2",
						SourceID: "s2",
						ParentID: "p2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
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
					&v1.EquipmentType{
						ID:             "1",
						Type:           "typ1",
						ParentId:       "p1",
						MetadataId:     "s1",
						ParentType:     "typ_parent",
						MetadataSource: "equip1.csv",
						Attributes: []*v1.Attribute{
							&v1.Attribute{
								ID:               "1",
								Name:             "attr_1",
								DataType:         v1.DataTypes_STRING,
								PrimaryKey:       true,
								Displayed:        true,
								Searchable:       true,
								ParentIdentifier: true,
								MappedTo:         "mapping_1",
							},
							&v1.Attribute{
								ID:               "2",
								Name:             "attr_2",
								DataType:         v1.DataTypes_INT,
								PrimaryKey:       false,
								Displayed:        true,
								Searchable:       false,
								ParentIdentifier: true,
								MappedTo:         "mapping_2",
							},
							&v1.Attribute{
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
					&v1.EquipmentType{
						ID:         "2",
						Type:       "typ2",
						ParentId:   "p2",
						MetadataId: "s2",
						Attributes: []*v1.Attribute{
							&v1.Attribute{
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
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.EquipmentsTypes(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.EquipmentsTypes() error = %v, wantErr %v", err, tt.wantErr)
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

func Test_licenseServiceServer_CreateEquipmentType(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.License
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						&v1.Attribute{
							Name:             "attr_2",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
						&v1.Attribute{
							Name:     "attr_3",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_3",
						},
						&v1.Attribute{
							Name:     "attr_4",
							DataType: v1.DataTypes_INT,
							MappedTo: "mapping_4",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
					&repo.EquipmentType{
						Type:     "typ3",
						ID:       "p2",
						SourceID: "s3",
						ParentID: "p1",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).
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
					Attributes: []*repo.Attribute{
						&repo.Attribute{
							Name:         "attr_1",
							Type:         repo.DataTypeString,
							IsIdentifier: true,
							IsDisplayed:  true,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						&repo.Attribute{
							Name:               "attr_2",
							Type:               repo.DataTypeString,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_2",
						},
						&repo.Attribute{
							Name:     "attr_3",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_3",
						},
						&repo.Attribute{
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
					Attributes: []*repo.Attribute{
						&repo.Attribute{
							ID:           "1",
							Name:         "attr_1",
							Type:         repo.DataTypeString,
							IsIdentifier: true,
							IsDisplayed:  true,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
						&repo.Attribute{
							ID:                 "2",
							Name:               "attr_2",
							Type:               repo.DataTypeString,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_2",
						},
						&repo.Attribute{
							ID:       "3",
							Name:     "attr_3",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_3",
						},
						&repo.Attribute{
							ID:       "4",
							Name:     "attr_4",
							Type:     repo.DataTypeInt,
							MappedTo: "mapping_4",
						},
					},
				}
				mockRepo.EXPECT().CreateEquipmentType(ctx, eqType, []string{"A", "B"}).Times(1).Return(retEqType, nil)
			},
			want: &v1.EquipmentType{
				ID:         "1",
				Type:       "typ1",
				ParentId:   "p1",
				MetadataId: "s1",
				Attributes: []*v1.Attribute{
					&v1.Attribute{
						ID:               "1",
						Name:             "attr_1",
						DataType:         v1.DataTypes_STRING,
						PrimaryKey:       true,
						Displayed:        true,
						Searchable:       true,
						ParentIdentifier: false,
						MappedTo:         "mapping_1",
					},
					&v1.Attribute{
						ID:               "2",
						Name:             "attr_2",
						DataType:         v1.DataTypes_STRING,
						PrimaryKey:       false,
						Displayed:        true,
						Searchable:       false,
						ParentIdentifier: true,
						MappedTo:         "mapping_2",
					},
					&v1.Attribute{
						ID:       "3",
						Name:     "attr_3",
						DataType: v1.DataTypes_FLOAT,
						MappedTo: "mapping_3",
					},
					&v1.Attribute{
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
				},
			},
			setup:   func() {},
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s1",
					},
					&repo.EquipmentType{
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{}, nil)
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: false,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{}, nil)
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						&v1.Attribute{
							Name:       "attr_2",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_2",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{}, nil)
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						&v1.Attribute{
							Name:             "attr_2",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{}, nil)
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						&v1.Attribute{
							Name:             "attr_2",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
						&v1.Attribute{
							Name:             "attr_3",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{}, nil)
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						&v1.Attribute{
							Name:             "attr_1",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						&v1.Attribute{
							Name:             "attr_2",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
						&v1.Attribute{
							Name:       "attr_3",
							DataType:   v1.DataTypes_STRING,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_3",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						&v1.Attribute{
							Name:             "attr_2",
							DataType:         v1.DataTypes_STRING,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_2",
						},
						&v1.Attribute{
							Name:       "attr_3",
							DataType:   v1.DataTypes_STRING,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_2",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
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
						&v1.Attribute{
							Name:             "attr_1",
							DataType:         v1.DataTypes_STRING,
							PrimaryKey:       true,
							Displayed:        true,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_1",
						},
						&v1.Attribute{
							Name:       "attr_3",
							DataType:   v1.DataTypes_STRING,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_3",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_INT,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						&v1.Attribute{
							Name:             "attr_2",
							DataType:         v1.DataTypes_FLOAT,
							ParentIdentifier: true,
							Displayed:        true,
							Searchable:       true,
							MappedTo:         "mapping_2",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  false,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						&v1.Attribute{
							Name:       "attr_2",
							DataType:   v1.DataTypes_STRING,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_2",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
						&v1.Attribute{
							Name:       "attr_2",
							DataType:   v1.DataTypes_STRING,
							Displayed:  false,
							Searchable: true,
							MappedTo:   "mapping_2",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(nil, repo.ErrNoData)
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))
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
						&v1.Attribute{
							Name:       "attr_1",
							DataType:   v1.DataTypes_STRING,
							PrimaryKey: true,
							Displayed:  true,
							Searchable: true,
							MappedTo:   "mapping_1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type:     "typ2",
						ID:       "p1",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2"},
				}, nil)

				eqType := &repo.EquipmentType{
					Type:     "typ1",
					SourceID: "s1",
					ParentID: "p1",
					Attributes: []*repo.Attribute{
						&repo.Attribute{
							Name:         "attr_1",
							Type:         repo.DataTypeString,
							IsIdentifier: true,
							IsDisplayed:  true,
							IsSearchable: true,
							MappedTo:     "mapping_1",
						},
					},
				}
				mockRepo.EXPECT().CreateEquipmentType(ctx, eqType, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.CreateEquipmentType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.CreateEquipmentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareEquipmentType(t, "EquipmentType", got, tt.want)
			}
		})
	}
}

func Test_licenseServiceServer_ListEquipmentsMetadata(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License
	type args struct {
		ctx context.Context
		req *v1.ListEquipmentMetadataRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.ListEquipmentMetadataResponse
		wantErr bool
	}{
		{name: "success - ALL MetaData",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type: v1.ListEquipmentMetadataRequest_ALL,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						SourceID: "2",
					},
					&repo.EquipmentType{
						SourceID: "3",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A", "B"}).
					Times(1).Return([]*repo.Metadata{
					&repo.Metadata{
						ID:     "1",
						Source: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
						ID:     "2",
						Source: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
						ID:     "3",
						Source: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
						ID:     "4",
						Source: "equip_4.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
				}, nil)
			},
			want: &v1.ListEquipmentMetadataResponse{
				Metadata: []*v1.EquipmentMetadata{
					&v1.EquipmentMetadata{
						ID:   "1",
						Name: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&v1.EquipmentMetadata{
						ID:   "2",
						Name: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&v1.EquipmentMetadata{
						ID:   "3",
						Name: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&v1.EquipmentMetadata{
						ID:   "4",
						Name: "equip_4.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
				},
			},
			wantErr: false,
		},
		{name: "success - Mapped MetaData",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type: v1.ListEquipmentMetadataRequest_MAPPED,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						SourceID: "2",
					},
					&repo.EquipmentType{
						SourceID: "3",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A", "B"}).
					Times(1).Return([]*repo.Metadata{
					&repo.Metadata{
						ID:     "1",
						Source: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
						ID:     "2",
						Source: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
						ID:     "3",
						Source: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
						ID:     "4",
						Source: "equip_4.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
				}, nil)
			},
			want: &v1.ListEquipmentMetadataResponse{
				Metadata: []*v1.EquipmentMetadata{
					&v1.EquipmentMetadata{
						ID:   "2",
						Name: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&v1.EquipmentMetadata{
						ID:   "3",
						Name: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
				},
			},
			wantErr: false,
		},
		{name: "success - Un-Mapped MetaData",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type: v1.ListEquipmentMetadataRequest_UN_MAPPED,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						SourceID: "2",
					},
					&repo.EquipmentType{
						SourceID: "3",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A", "B"}).
					Times(1).Return([]*repo.Metadata{
					&repo.Metadata{
						ID:     "1",
						Source: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
						ID:     "2",
						Source: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
						ID:     "3",
						Source: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
						ID:     "4",
						Source: "equip_4.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
				}, nil)
			},
			want: &v1.ListEquipmentMetadataResponse{
				Metadata: []*v1.EquipmentMetadata{
					&v1.EquipmentMetadata{
						ID:   "1",
						Name: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&v1.EquipmentMetadata{
						ID:   "4",
						Name: "equip_4.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
				},
			},
			wantErr: false,
		},
		{name: "failure|can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListEquipmentMetadataRequest{
					Type: v1.ListEquipmentMetadataRequest_UN_MAPPED,
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure, fetching equipment type",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type: v1.ListEquipmentMetadataRequest_UN_MAPPED,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "failure, fetching metadata, no data",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type: v1.ListEquipmentMetadataRequest_UN_MAPPED,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						SourceID: "2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A", "B"}).
					Times(1).Return(nil, repo.ErrNoData)
			},
			wantErr: true,
		},
		{name: "failure, fetching metadata",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type: v1.ListEquipmentMetadataRequest_UN_MAPPED,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						SourceID: "2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A", "B"}).
					Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "failure - default query parameter",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentMetadataRequest{
					Type: 10000,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						SourceID: "2",
					},
					&repo.EquipmentType{
						SourceID: "3",
					},
				}, nil)
				mockRepo.EXPECT().MetadataAllWithType(ctx, repo.MetadataTypeEquipment, []string{"A", "B"}).
					Times(1).Return([]*repo.Metadata{
					&repo.Metadata{
						ID:     "1",
						Source: "equip_1.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
						ID:     "2",
						Source: "equip_2.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
						ID:     "3",
						Source: "equip_3.csv",
						Attributes: []string{
							"attr_1",
							"attr_2",
						},
					},
					&repo.Metadata{
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
			t.Log(tt.name)
			s := NewLicenseServiceServer(rep)
			got, err := s.ListEquipmentsMetadata(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListEquipmentsMetadata() error = %v, wantErr %v", err, tt.wantErr)
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

func Test_licenseServiceServer_UpdateEquipmentType(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.License
	type args struct {
		ctx context.Context
		req *v1.UpdateEquipmentTypeRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.EquipmentType
		wantErr bool
	}{
		{name: "SUCCESS - parent id already exists",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)

				mockRepo.EXPECT().UpdateEquipmentType(ctx, "1", "MyType", &repo.UpdateEquipmentRequest{
					Attr: []*repo.Attribute{
						&repo.Attribute{
							Name:         "attr3",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_3",
						},
						&repo.Attribute{
							Name:         "attr4",
							Type:         repo.DataTypeInt,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_4",
						},
						&repo.Attribute{
							Name:     "attr5",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_5",
						},
					},
				}, []string{"A", "B"}).Times(1).Return([]*repo.Attribute{
					&repo.Attribute{
						Name:         "attr3",
						Type:         repo.DataTypeString,
						IsSearchable: true,
						IsDisplayed:  true,
						MappedTo:     "mapping_3",
					},
					&repo.Attribute{
						Name:         "attr4",
						Type:         repo.DataTypeInt,
						IsSearchable: true,
						IsDisplayed:  true,
						MappedTo:     "mapping_4",
					},
					&repo.Attribute{
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
				Attributes: []*v1.Attribute{
					&v1.Attribute{
						Name:       "attr1",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						MappedTo:   "mapping_1",
					},
					&v1.Attribute{
						Name:       "attr2",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						MappedTo:   "mapping_2",
					},
					&v1.Attribute{
						Name:       "attr3",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						Displayed:  true,
						MappedTo:   "mapping_3",
					},
					&v1.Attribute{
						Name:       "attr4",
						DataType:   v1.DataTypes_INT,
						Searchable: true,
						Displayed:  true,
						MappedTo:   "mapping_4",
					},
					&v1.Attribute{
						Name:     "attr5",
						DataType: v1.DataTypes_FLOAT,
						MappedTo: "mapping_5",
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - new parent created",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)

				mockRepo.EXPECT().UpdateEquipmentType(ctx, "1", "MyType", &repo.UpdateEquipmentRequest{
					ParentID: "2",
					Attr: []*repo.Attribute{
						&repo.Attribute{
							Name:               "attr3",
							Type:               repo.DataTypeString,
							IsSearchable:       true,
							IsDisplayed:        true,
							IsParentIdentifier: true,
							MappedTo:           "mapping_3",
						},
						&repo.Attribute{
							Name:         "attr4",
							Type:         repo.DataTypeInt,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_4",
						},
						&repo.Attribute{
							Name:     "attr5",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_5",
						},
					},
				}, []string{"A", "B"}).Times(1).Return([]*repo.Attribute{
					&repo.Attribute{
						Name:               "attr3",
						Type:               repo.DataTypeString,
						IsSearchable:       true,
						IsDisplayed:        true,
						IsParentIdentifier: true,
						MappedTo:           "mapping_3",
					},
					&repo.Attribute{
						Name:         "attr4",
						Type:         repo.DataTypeInt,
						IsSearchable: true,
						IsDisplayed:  true,
						MappedTo:     "mapping_4",
					},
					&repo.Attribute{
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
				Attributes: []*v1.Attribute{
					&v1.Attribute{
						Name:       "attr1",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						MappedTo:   "mapping_1",
					},
					&v1.Attribute{
						Name:       "attr2",
						DataType:   v1.DataTypes_STRING,
						Searchable: true,
						MappedTo:   "mapping_2",
					},
					&v1.Attribute{
						Name:             "attr3",
						DataType:         v1.DataTypes_STRING,
						Searchable:       true,
						Displayed:        true,
						ParentIdentifier: true,
						MappedTo:         "mapping_3",
					},
					&v1.Attribute{
						Name:       "attr4",
						DataType:   v1.DataTypes_INT,
						Searchable: true,
						Displayed:  true,
						MappedTo:   "mapping_4",
					},
					&v1.Attribute{
						Name:     "attr5",
						DataType: v1.DataTypes_FLOAT,
						MappedTo: "mapping_5",
					},
				},
			},
			wantErr: false,
		},
		{name: "failure|can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
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
		{name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))

			},
			wantErr: true,
		},
		{name: "FAILURE - equipment not exists",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "3",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch meta data",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "FAILURE - no meta data",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(nil, repo.ErrNoData)
			},
			wantErr: true,
		},
		{name: "FAILURE - parent id already exists but new parent creation requested ",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "3",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
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
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
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
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - parent not selected for identifier ",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - primary key given in argument",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							PrimaryKey: true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - parent id exists already but parent identifier given",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - multiple parent keys ",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:             "attr3",
							DataType:         v1.DataTypes_STRING,
							Searchable:       true,
							Displayed:        true,
							ParentIdentifier: true,
							MappedTo:         "mapping_3",
						},
						&v1.Attribute{
							Name:             "attr4",
							DataType:         v1.DataTypes_INT,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - only string data type is allowed for parent key ",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:             "attr4",
							DataType:         v1.DataTypes_INT,
							Searchable:       true,
							ParentIdentifier: true,
							MappedTo:         "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - attribute name already exists ",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr2",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - attribute name already exists - case insensitive",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "Attr2",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - mapping already taken ",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_2",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - mapping not found ",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_6",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - searchable attribute should be dispayable ",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id:       "1",
					ParentId: "2",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							Displayed:  false,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot update equipment",
			args: args{
				ctx: ctx,
				req: &v1.UpdateEquipmentTypeRequest{
					Id: "1",
					Attributes: []*v1.Attribute{
						&v1.Attribute{
							Name:       "attr3",
							DataType:   v1.DataTypes_STRING,
							Searchable: true,
							Displayed:  true,
							MappedTo:   "mapping_3",
						},
						&v1.Attribute{
							Name:       "attr4",
							DataType:   v1.DataTypes_INT,
							Searchable: true,
							MappedTo:   "mapping_4",
						},
						&v1.Attribute{
							Name:     "attr5",
							DataType: v1.DataTypes_FLOAT,
							MappedTo: "mapping_5",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "mapping_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
					},
				}, nil)
				mockRepo.EXPECT().MetadataWithID(ctx, "s1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:         "s1",
					Source:     "test.csv",
					Attributes: []string{"mapping_1", "mapping_2", "mapping_3", "mapping_4", "mapping_5"},
				}, nil)

				mockRepo.EXPECT().UpdateEquipmentType(ctx, "1", "MyType", &repo.UpdateEquipmentRequest{
					Attr: []*repo.Attribute{
						&repo.Attribute{
							Name:         "attr3",
							Type:         repo.DataTypeString,
							IsSearchable: true,
							IsDisplayed:  true,
							MappedTo:     "mapping_3",
						},
						&repo.Attribute{
							Name:         "attr4",
							Type:         repo.DataTypeInt,
							IsSearchable: true,
							MappedTo:     "mapping_4",
						},
						&repo.Attribute{
							Name:     "attr5",
							Type:     repo.DataTypeFloat,
							MappedTo: "mapping_5",
						},
					},
				}, []string{"A", "B"}).Return(nil, errors.New("test error")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.UpdateEquipmentType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.UpdateEquipmentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareEquipmentType(t, "EquipmentType", got, tt.want)
			}
		})
	}
}

func Test_licenseServiceServer_GetEquipmentMetadata(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.License
	type args struct {
		ctx context.Context
		req *v1.EquipmentMetadataRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.EquipmentMetadata
		wantErr bool
	}{
		{name: "SUCCESS - no equipment exists for given metadata",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID: "1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:     "s3",
					Source: "equip_1.csv",
					Attributes: []string{
						"attr_1",
						"attr_2",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
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
			},
			wantErr: false,
		},
		{name: "SUCCESS - ALL attributes",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID:         "1",
					Attributes: v1.EquipmentMetadataRequest_All,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:     "s1",
					Source: "equip_1.csv",
					Attributes: []string{
						"attr_1",
						"attr_2",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_1",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
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
			},
			wantErr: false,
		},
		{name: "SUCCESS - MAPPED attributes",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID:         "1",
					Attributes: v1.EquipmentMetadataRequest_Mapped,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:     "s1",
					Source: "equip_1.csv",
					Attributes: []string{
						"attr_1",
						"attr_2",
						"attr_3",
						"attr_4",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_3",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
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
			},
			wantErr: false,
		},
		{name: "SUCCESS - UNMAPPED attributes",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID:         "1",
					Attributes: v1.EquipmentMetadataRequest_Unmapped,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:     "s1",
					Source: "equip_1.csv",
					Attributes: []string{
						"attr_1",
						"attr_2",
						"attr_3",
						"attr_4",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						ID:       "1",
						Type:     "MyType",
						SourceID: "s1",
						ParentID: "2",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_2",
							},
							&repo.Attribute{
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsSearchable: true,
								MappedTo:     "attr_3",
							},
						},
					},
					&repo.EquipmentType{
						ID:       "2",
						Type:     "MyType2",
						SourceID: "s2",
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
			},
			wantErr: false,
		},
		{name: "failure|can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.EquipmentMetadataRequest{
					ID: "1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metadata - no metadata",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID: "1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A", "B"}).Times(1).Return(nil, repo.ErrNoData)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metadata",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID: "1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A", "B"}).Times(1).Return(nil, errors.New("Test Error"))
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.EquipmentMetadataRequest{
					ID: "1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().MetadataWithID(ctx, "1", []string{"A", "B"}).Times(1).Return(&repo.Metadata{
					ID:     "s3",
					Source: "equip_1.csv",
					Attributes: []string{
						"attr_1",
						"attr_2",
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.GetEquipmentMetadata(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.GetEquipmentMetadata() error = %v, wantErr %v", err, tt.wantErr)
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
