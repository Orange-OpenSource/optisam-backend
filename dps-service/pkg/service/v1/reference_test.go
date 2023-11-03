package v1

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1"
	dbmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1/dbmock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1/postgres/db"
	queuemock "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1/queuemock"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
)

func Test_StoreCoreFactor(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var rep repo.Dps
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	ctx2 := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "User",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	tests := []struct {
		name    string
		ctx     context.Context
		input   *v1.StoreReferenceDataRequest
		setup   func(*v1.StoreReferenceDataRequest)
		output  *v1.StoreReferenceDataResponse
		wantErr bool
	}{
		{
			name: "claims Not found",
			ctx:  context.Background(),
			input: &v1.StoreReferenceDataRequest{
				ReferenceData: []byte(`{"a":{"b":"1"}}`),
			},
			setup:   func(*v1.StoreReferenceDataRequest) {},
			output:  &v1.StoreReferenceDataResponse{Success: false},
			wantErr: true,
		},
		{
			name: "UnauthorisedUser",
			ctx:  ctx2,
			input: &v1.StoreReferenceDataRequest{
				ReferenceData: []byte(`{"a":{"b":"1"}}`),
			},
			setup:   func(*v1.StoreReferenceDataRequest) {},
			output:  &v1.StoreReferenceDataResponse{Success: false},
			wantErr: true,
		},
		{
			name: "Success save",
			ctx:  ctx,
			input: &v1.StoreReferenceDataRequest{
				ReferenceData: []byte(`{"a":{"b":"1"}}`),
				Filename:      "temp.xlsx",
			},
			setup: func(data *v1.StoreReferenceDataRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				temp := make(map[string]map[string]string)
				json.Unmarshal(data.ReferenceData, &temp)
				mockRepository.EXPECT().DeleteCoreFactorReference(ctx).Return(nil).Times(1)
				mockRepository.EXPECT().StoreCoreFactorReferences(ctx, temp).Return(nil).Times(1)
				mockRepository.EXPECT().LogCoreFactor(ctx, data.Filename).Return(nil).Times(1)
			},
			output:  &v1.StoreReferenceDataResponse{Success: false},
			wantErr: false,
		},
		{
			name: "DBError",
			ctx:  ctx,
			input: &v1.StoreReferenceDataRequest{
				ReferenceData: []byte(`{"a":{"b":"1"}}`),
			},
			setup: func(data *v1.StoreReferenceDataRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				temp := make(map[string]map[string]string)
				json.Unmarshal(data.ReferenceData, &temp)
				mockRepository.EXPECT().DeleteCoreFactorReference(ctx).Return(errors.New("DBError")).Times(1)
			},
			output:  &v1.StoreReferenceDataResponse{Success: false},
			wantErr: true,
		},
		{
			name: "Success db err 2",
			ctx:  ctx,
			input: &v1.StoreReferenceDataRequest{
				ReferenceData: []byte(`{"a":{"b":"1"}}`),
				Filename:      "temp.xlsx",
			},
			setup: func(data *v1.StoreReferenceDataRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				temp := make(map[string]map[string]string)
				json.Unmarshal(data.ReferenceData, &temp)
				mockRepository.EXPECT().DeleteCoreFactorReference(ctx).Return(errors.New("DBError")).Times(1)
				mockRepository.EXPECT().StoreCoreFactorReferences(ctx, temp).Return(nil).AnyTimes()
			},
			output:  &v1.StoreReferenceDataResponse{Success: false},
			wantErr: true,
		},
		{
			name: "Success db err3",
			ctx:  ctx,
			input: &v1.StoreReferenceDataRequest{
				ReferenceData: []byte(`{"a":{"b":"1"}}`),
				Filename:      "temp.xlsx",
			},
			setup: func(data *v1.StoreReferenceDataRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				temp := make(map[string]map[string]string)
				json.Unmarshal(data.ReferenceData, &temp)
				mockRepository.EXPECT().DeleteCoreFactorReference(ctx).Return(nil).Times(1)
				mockRepository.EXPECT().StoreCoreFactorReferences(ctx, temp).AnyTimes().Return(nil)
				mockRepository.EXPECT().LogCoreFactor(ctx, data.Filename).Return(errors.New("DBError")).Times(1)
			},
			output:  &v1.StoreReferenceDataResponse{Success: false},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			obj := NewDpsServiceServer(rep, qObj, nil)
			_, err := obj.StoreCoreFactorReference(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.StoreCoreFactorReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_ViewCoreFactor(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	ctx2 := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "User",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	mockCtrl := gomock.NewController(t)
	var rep repo.Dps
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	tests := []struct {
		name    string
		ctx     context.Context
		input   *v1.ViewReferenceDataRequest
		setup   func(*v1.ViewReferenceDataRequest)
		output  *v1.ViewReferenceDataResponse
		wantErr bool
	}{
		{
			name: "claims Not found",
			ctx:  context.Background(),
			input: &v1.ViewReferenceDataRequest{
				PageNo:   int32(1),
				PageSize: int32(10),
			},
			setup:   func(*v1.ViewReferenceDataRequest) {},
			wantErr: true,
		},
		{
			name: "unauthoriseUser",
			ctx:  ctx2,
			input: &v1.ViewReferenceDataRequest{
				PageNo:   int32(1),
				PageSize: int32(10),
			},
			setup:   func(*v1.ViewReferenceDataRequest) {},
			wantErr: true,
		},
		{
			name: "DBError",
			ctx:  ctx,
			input: &v1.ViewReferenceDataRequest{
				PageNo:   int32(1),
				PageSize: int32(10),
			},
			setup: func(data *v1.ViewReferenceDataRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().GetCoreFactorReferences(ctx, db.GetCoreFactorReferencesParams{
					Limit:  int32(10),
					Offset: int32(0),
				}).Return(nil, errors.New("DBERROR")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "success",
			ctx:  ctx,
			input: &v1.ViewReferenceDataRequest{
				PageNo:   int32(1),
				PageSize: int32(10),
			},
			setup: func(data *v1.ViewReferenceDataRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				data.PageNo = 1
				data.PageSize = 10
				mockRepository.EXPECT().GetCoreFactorReferences(ctx, db.GetCoreFactorReferencesParams{
					Limit:  int32(10),
					Offset: int32(0),
				}).Return([]db.GetCoreFactorReferencesRow{
					{
						ID:           int32(1),
						Manufacturer: "x",
						Model:        "y",
						CoreFactor:   "1",
						TotalRecords: int64(1),
					},
				}, nil).Times(1)
			},
			output: &v1.ViewReferenceDataResponse{
				References: []*v1.CoreFactorReference{
					{
						Manufacturer: "x",
						Model:        "y",
						Corefactor:   "1",
					},
				},
				TotalRecord: int32(1),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			obj := NewDpsServiceServer(rep, qObj, nil)
			got, err := obj.ViewFactorReference(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.StoreCoreFactorReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !compareCorfactorReferences(tt.output, got) {
				t.Errorf("dpsServiceServer.StoreCoreFactorReference() got = %v, want %v", got, tt.output)
				return
			}
		})
	}

}
func compareCorfactorReferences(x, y *v1.ViewReferenceDataResponse) bool {
	if x.TotalRecord != y.TotalRecord || len(x.References) != len(y.References) {
		return false
	}
	for i, v := range x.References {
		if v.Corefactor != y.References[i].Corefactor || v.Manufacturer != y.References[i].Manufacturer || v.Model != y.References[i].Model {
			return false
		}
	}
	return true
}

func Test_ViewCoreFactorLog(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	ut := time.Now()
	gt, _ := ptypes.TimestampProto(ut)
	ctx2 := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "User",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	mockCtrl := gomock.NewController(t)
	var rep repo.Dps
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	tests := []struct {
		name    string
		ctx     context.Context
		input   *v1.ViewCoreFactorLogsRequest
		setup   func()
		output  *v1.ViewCoreFactorLogsResponse
		wantErr bool
	}{
		{
			name:    "claims Not found",
			ctx:     context.Background(),
			input:   &v1.ViewCoreFactorLogsRequest{},
			setup:   func() {},
			wantErr: true,
		},
		{
			name:    "unauthoriseUser",
			ctx:     ctx2,
			input:   &v1.ViewCoreFactorLogsRequest{},
			setup:   func() {},
			wantErr: true,
		},
		{
			name:  "DBError",
			ctx:   ctx,
			input: &v1.ViewCoreFactorLogsRequest{},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().GetCoreFactorLogs(ctx).Return(nil, errors.New("DBERROR")).Times(1)
			},
			wantErr: true,
		},
		{
			name:  "success",
			ctx:   ctx,
			input: &v1.ViewCoreFactorLogsRequest{},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().GetCoreFactorLogs(ctx).Return([]db.CoreFactorLog{
					{
						FileName:   "temp.xlsx",
						UploadedOn: ut,
					},
				}, nil).Times(1)
			},
			output: &v1.ViewCoreFactorLogsResponse{
				Corefactorlogs: []*v1.CoreFactorlogs{
					{
						Filename:   "temp.xlsx",
						UploadedOn: gt,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			obj := NewDpsServiceServer(rep, qObj, nil)
			got, err := obj.ViewCoreFactorLogs(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.ViewCoreFactorLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(tt.output.Corefactorlogs) != len(got.Corefactorlogs) && tt.output.Corefactorlogs[0].Filename != got.Corefactorlogs[0].Filename && tt.output.Corefactorlogs[0].UploadedOn != got.Corefactorlogs[0].UploadedOn {
				t.Errorf("dpsServiceServer.ViewCoreFactorLogs() got = %v, want %v", got, tt.output)
				return
			}
		})
	}

}
