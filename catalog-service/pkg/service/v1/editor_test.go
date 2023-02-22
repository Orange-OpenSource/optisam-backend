package v1

import (
	"context"
	"database/sql"
	queuemock "optisam-backend/application-service/pkg/repository/v1/queuemock"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	repo "optisam-backend/catalog-service/pkg/repository/v1"
	"optisam-backend/catalog-service/pkg/repository/v1/mock"
	"optisam-backend/catalog-service/pkg/repository/v1/postgres/db"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

func Test_productCatalogServer_CreateEditor(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.ProductCatalog
	var queue workerqueue.Workerqueue
	timenow := time.Now()
	createdOnObject, _ := ptypes.TimestampProto(timenow)
	updatedOnObject, _ := ptypes.TimestampProto(timenow)
	uid := uuid.New().String()
	type args struct {
		ctx context.Context
		req *v1.CreateEditorRequest
	}
	tests := []struct {
		name    string
		s       *productCatalogServer
		args    args
		setup   func()
		want    *v1.Editor
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.CreateEditorRequest{
					Name:               "Editor1",
					GenearlInformation: "generalInfo",
					Vendors:            nil,
					Audits:             nil,
					PartnerManagers:    nil,

					// Vendors: []*v1.Vendors{
					// 	{
					// 		Name: "string",
					// 	},
					// },
					// PartnerManagers: []*v1.PartnerManagers{
					// 	{
					// 		Email: "string",
					// 		Name:  "string",
					// 	},
					// },
					// Audits: []*v1.Audits{
					// 	{
					// 		Entity: "string",
					// 	},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				defer mockCtrl.Finish()
				mockRepository := mock.NewMockProductCatalog(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				rep = mockRepository
				_ = db.InsertEditorCatalogParams{
					ID:                 uid,
					Name:               "Editor1",
					GeneralInformation: sql.NullString{String: "generalInfo", Valid: true},
					PartnerManagers:    nil,
					Audits:             nil,
					Vendors:            nil,
					CreatedOn:          timenow,
					UpdatedOn:          timenow,
				}
				mockRepository.EXPECT().InsertEditorCatalog(ctx, gomock.Any()).Return(nil).AnyTimes()
			},
			want: &v1.Editor{
				Id:                 uid,
				Name:               "Editor1",
				GenearlInformation: "generalInfo",
				PartnerManagers:    nil,
				Audits:             nil,
				Vendors:            nil,
				CreatedOn:          createdOnObject,
				UpdatedOn:          updatedOnObject,
			},
			s: NewProductCatalogServer(rep, queue, nil).(*productCatalogServer),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			s := NewProductCatalogServer(rep, queue, nil)
			got, err := s.CreateEditor(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("catalogServiceServer.catalogsPercatalogType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("catalogServiceServer.catalogsPercatalogType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productCatalogServer_UpdateEditor(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.ProductCatalog
	var queue workerqueue.Workerqueue
	timenow := time.Now()
	createdOnObject, _ := ptypes.TimestampProto(timenow)
	updatedOnObject, _ := ptypes.TimestampProto(timenow)
	uid := uuid.New().String()
	type args struct {
		ctx context.Context
		req *v1.Editor
	}
	tests := []struct {
		name    string
		s       *productCatalogServer
		args    args
		setup   func()
		want    *v1.Editor
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.Editor{
					Id:                 uuid.New().String(),
					Name:               "Editor1",
					GenearlInformation: "generalInfo",
					Vendors:            nil,
					Audits:             nil,
					PartnerManagers:    nil,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				defer mockCtrl.Finish()
				mockRepository := mock.NewMockProductCatalog(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				rep = mockRepository
				ue := db.UpdateEditorCatalogParams{
					ID:                 uid,
					GeneralInformation: sql.NullString{String: "generalInfo", Valid: true},
					PartnerManagers:    nil,
					Audits:             nil,
					Vendors:            nil,
					UpdatedOn:          timenow,
				}
				edit := db.EditorCatalog{
					ID:                 uuid.New().String(),
					Name:               "Editor1",
					GeneralInformation: sql.NullString{String: "generalInfo", Valid: true},
					Vendors:            nil,
					Audits:             nil,
					PartnerManagers:    nil,
				}
				first := mockRepository.EXPECT().UpdateEditorCatalog(ctx, gomock.Any()).Return(nil).AnyTimes()
				second := mockRepository.EXPECT().UpdateEditorCatalog(ctx, ue).Return(nil).AnyTimes().After(first)
				mockRepository.EXPECT().GetEditorCatalog(ctx, gomock.Any()).Return(edit, nil).AnyTimes().After(second)
			},
			want: &v1.Editor{
				Id:                 uid,
				Name:               "Editor1",
				GenearlInformation: "generalInfo",
				PartnerManagers:    nil,
				Audits:             nil,
				Vendors:            nil,
				CreatedOn:          createdOnObject,
				UpdatedOn:          updatedOnObject,
			},
			s: NewProductCatalogServer(rep, queue, nil).(*productCatalogServer),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			s := NewProductCatalogServer(rep, queue, nil)
			got, err := s.UpdateEditor(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("catalogServiceServer.catalogsPercatalogType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("catalogServiceServer.catalogsPercatalogType() = %v, want %v", got, tt.want)
			}
		})
	}
}
