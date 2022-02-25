package v1

import (
	"context"
	v1 "optisam-backend/application-service/pkg/api/v1"
	repo "optisam-backend/application-service/pkg/repository/v1"
	dbmock "optisam-backend/application-service/pkg/repository/v1/dbmock"
	"optisam-backend/application-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/application-service/pkg/repository/v1/queuemock"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_applicationServiceServer_ApplicationDomains(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	type fields struct {
		applicationRepo repo.Application
		queue           workerqueue.Workerqueue
	}
	type args struct {
		ctx context.Context
		req *v1.ApplicationDomainsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(*v1.ApplicationDomainsRequest)
		want    *v1.ApplicationDomainsResponse
		wantErr bool
	}{
		{
			name:   "SUCCESS-RIGHT-CLAIMS",
			args:   args{ctx: ctx, req: &v1.ApplicationDomainsRequest{Scope: "Scope1"}},
			fields: fields{applicationRepo: dbObj, queue: qObj},
			mock: func(input *v1.ApplicationDomainsRequest) {
				dbObj.EXPECT().GetApplicationDomains(ctx, input.Scope).Return([]string{"Payments", "Finance"}, nil).Times(1)
			},
			want:    &v1.ApplicationDomainsResponse{Domains: []string{"Payments", "Finance"}},
			wantErr: false,
		},
		{
			name:    "SUCCESS-WRONG-CLAIMS",
			args:    args{ctx: ctx, req: &v1.ApplicationDomainsRequest{Scope: "s2"}},
			fields:  fields{applicationRepo: dbObj, queue: qObj},
			mock:    func(input *v1.ApplicationDomainsRequest) {},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.args.req)
			s := NewApplicationServiceServer(dbObj, qObj, nil)
			got, err := s.ApplicationDomains(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", tt.name)
				return
			} else if (got != nil && tt.want != nil) && !assert.Equal(t, got, (tt.want)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex [%v]", tt.name, tt.want, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", tt.name))
			}
		})
	}
}

func Test_applicationServiceServer_ObsolescenceDomainCriticityMeta(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	type fields struct {
		applicationRepo repo.Application
		queue           workerqueue.Workerqueue
	}
	type args struct {
		ctx context.Context
		req *v1.DomainCriticityMetaRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(*v1.DomainCriticityMetaRequest)
		want    *v1.DomainCriticityMetaResponse
		wantErr bool
	}{
		{
			name:   "SUCCESS-RIGHT-CLAIMS",
			args:   args{ctx: ctx, req: &v1.DomainCriticityMetaRequest{}},
			fields: fields{applicationRepo: dbObj, queue: qObj},
			mock: func(input *v1.DomainCriticityMetaRequest) {
				dbObj.EXPECT().GetDomainCriticityMeta(ctx).Return([]db.DomainCriticityMetum{{
					DomainCriticID:   1,
					DomainCriticName: "Critical",
				}, {
					DomainCriticID:   2,
					DomainCriticName: "Non Critical",
				}}, nil).Times(1)
			},
			want: &v1.DomainCriticityMetaResponse{DomainCriticityMeta: []*v1.DomainCriticityMeta{{
				DomainCriticId:   1,
				DomainCriticName: "Critical",
			},
				{
					DomainCriticId:   2,
					DomainCriticName: "Non Critical",
				},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.args.req)
			s := NewApplicationServiceServer(dbObj, qObj, nil)
			got, err := s.ObsolescenceDomainCriticityMeta(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", tt.name)
				return
			} else if (got != nil && tt.want != nil) && !assert.Equal(t, got, (tt.want)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex [%v]", tt.name, tt.want, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", tt.name))
			}
		})
	}
}

func Test_applicationServiceServer_ObsolescenceMaintenanceCriticityMeta(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	type fields struct {
		applicationRepo repo.Application
		queue           workerqueue.Workerqueue
	}
	type args struct {
		ctx context.Context
		req *v1.MaintenanceCriticityMetaRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(*v1.MaintenanceCriticityMetaRequest)
		want    *v1.MaintenanceCriticityMetaResponse
		wantErr bool
	}{
		{
			name:   "SUCCESS-RIGHT_CLAIMS",
			args:   args{ctx: ctx, req: &v1.MaintenanceCriticityMetaRequest{}},
			fields: fields{applicationRepo: dbObj, queue: qObj},
			mock: func(input *v1.MaintenanceCriticityMetaRequest) {
				dbObj.EXPECT().GetMaintenanceCricityMeta(ctx).Return([]db.MaintenanceLevelMetum{
					{
						MaintenanceLevelID:   1,
						MaintenanceLevelName: "L1",
					},
					{
						MaintenanceLevelID:   2,
						MaintenanceLevelName: "L2",
					},
				}, nil).Times(1)
			},
			want: &v1.MaintenanceCriticityMetaResponse{
				MaintenanceCriticityMeta: []*v1.MaintenanceCriticityMeta{{
					MaintenanceCriticId:   1,
					MaintenanceCriticName: "L1",
				}, {
					MaintenanceCriticId:   2,
					MaintenanceCriticName: "L2",
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.args.req)
			s := NewApplicationServiceServer(dbObj, qObj, nil)
			got, err := s.ObsolescenceMaintenanceCriticityMeta(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", tt.name)
				return
			} else if (got != nil && tt.want != nil) && !assert.Equal(t, got, (tt.want)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex [%v]", tt.name, tt.want, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", tt.name))
			}
		})
	}
}

func Test_applicationServiceServer_ObsolescenceRiskMeta(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	type fields struct {
		applicationRepo repo.Application
		queue           workerqueue.Workerqueue
	}
	type args struct {
		ctx context.Context
		req *v1.RiskMetaRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(*v1.RiskMetaRequest)
		want    *v1.RiskMetaResponse
		wantErr bool
	}{
		{
			name:   "SUCCESS-RIGHT-CLAIMS",
			fields: fields{applicationRepo: dbObj, queue: qObj},
			args:   args{ctx: ctx, req: &v1.RiskMetaRequest{}},
			mock: func(input *v1.RiskMetaRequest) {
				dbObj.EXPECT().GetRiskMeta(ctx).Return([]db.RiskMetum{{
					RiskID:   1,
					RiskName: "Low",
				}, {
					RiskID:   2,
					RiskName: "High",
				},
				}, nil)
			},
			want: &v1.RiskMetaResponse{
				RiskMeta: []*v1.RiskMeta{
					{
						RiskId:   1,
						RiskName: "Low",
					},
					{
						RiskId:   2,
						RiskName: "High",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.args.req)
			s := NewApplicationServiceServer(dbObj, qObj, nil)
			got, err := s.ObsolescenceRiskMeta(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", tt.name)
				return
			} else if (got != nil && tt.want != nil) && !assert.Equal(t, got, (tt.want)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex [%v]", tt.name, tt.want, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", tt.name))
			}
		})
	}
}

func Test_applicationServiceServer_ObsolescenceDomainCriticity(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)

	type fields struct {
		applicationRepo repo.Application
		queue           workerqueue.Workerqueue
	}
	type args struct {
		ctx context.Context
		req *v1.DomainCriticityRequest
	}
	tests := []struct {
		name    string
		fields  fields
		mock    func(*v1.DomainCriticityRequest)
		args    args
		want    *v1.DomainCriticityResponse
		wantErr bool
	}{
		{
			name:   "SUCCESS-RIGHT-CLAIMS",
			fields: fields{applicationRepo: dbObj, queue: qObj},
			args:   args{ctx: ctx, req: &v1.DomainCriticityRequest{Scope: "Scope1"}},
			mock: func(input *v1.DomainCriticityRequest) {
				dbObj.EXPECT().GetDomainCriticity(ctx, input.Scope).Return([]db.GetDomainCriticityRow{{
					DomainCriticID: 1,
					Domains:        []string{"Finance", "Payment"},
				}}, nil)
			},
			want: &v1.DomainCriticityResponse{
				DomainsCriticity: []*v1.DomainCriticity{
					{
						DomainCriticId: 1,
						Domains:        []string{"Finance", "Payment"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.args.req)
			s := NewApplicationServiceServer(dbObj, qObj, nil)
			got, err := s.ObsolescenceDomainCriticity(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", tt.name)
				return
			} else if (got != nil && tt.want != nil) && !assert.Equal(t, got, (tt.want)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex [%v]", tt.name, tt.want, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", tt.name))
			}
		})
	}
}

func Test_applicationServiceServer_ObsolescenseMaintenanceCriticity(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	type fields struct {
		applicationRepo repo.Application
		queue           workerqueue.Workerqueue
	}
	type args struct {
		ctx context.Context
		req *v1.MaintenanceCriticityRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(*v1.MaintenanceCriticityRequest)
		want    *v1.MaintenanceCriticityResponse
		wantErr bool
	}{
		{
			name:   "SUCCESS-RIGHT_CLAIMS",
			fields: fields{applicationRepo: dbObj, queue: qObj},
			args:   args{ctx: ctx, req: &v1.MaintenanceCriticityRequest{Scope: "Scope1"}},
			mock: func(input *v1.MaintenanceCriticityRequest) {
				dbObj.EXPECT().GetMaintenanceTimeCriticity(ctx, input.Scope).Return([]db.MaintenanceTimeCriticity{
					{
						MaintenanceCriticID: 1,
						LevelID:             1,
						StartMonth:          1,
						EndMonth:            12,
					},
				}, nil)
			},
			want: &v1.MaintenanceCriticityResponse{
				MaintenanceCriticy: []*v1.MaintenanceCriticity{{
					MaintenanceCriticId: 1,
					MaintenanceLevelId:  1,
					StartMonth:          1,
					EndMonth:            12,
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.args.req)
			s := NewApplicationServiceServer(dbObj, qObj, nil)
			got, err := s.ObsolescenseMaintenanceCriticity(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", tt.name)
				return
			} else if (got != nil && tt.want != nil) && !assert.Equal(t, got, (tt.want)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex [%v]", tt.name, tt.want, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", tt.name))
			}
		})
	}
}

func Test_applicationServiceServer_ObsolescenseRiskMatrix(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)

	type fields struct {
		applicationRepo repo.Application
		queue           workerqueue.Workerqueue
	}
	type args struct {
		ctx context.Context
		req *v1.RiskMatrixRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(*v1.RiskMatrixRequest)
		want    *v1.RiskMatrixResponse
		wantErr bool
	}{
		{
			name:   "SUCCESS-RIGHT-CLAIMS",
			args:   args{ctx: ctx, req: &v1.RiskMatrixRequest{Scope: "Scope1"}},
			fields: fields{applicationRepo: dbObj, queue: qObj},
			mock: func(input *v1.RiskMatrixRequest) {
				dbObj.EXPECT().GetRiskMatrixConfig(ctx, input.Scope).Return([]db.GetRiskMatrixConfigRow{{
					ConfigurationID:      1,
					DomainCriticID:       1,
					DomainCriticName:     "Critical",
					MaintenanceLevelID:   1,
					MaintenanceLevelName: "L1",
					RiskID:               1,
					RiskName:             "Low",
				}}, nil)
			},
			want: &v1.RiskMatrixResponse{RiskMatrix: []*v1.RiskMatrix{{
				ConfigurationId:       1,
				DomainCriticId:        1,
				DomainCriticName:      "Critical",
				MaintenanceCriticId:   1,
				MaintenanceCriticName: "L1",
				RiskId:                1,
				RiskName:              "Low",
			}}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.args.req)
			s := NewApplicationServiceServer(dbObj, qObj, nil)
			got, err := s.ObsolescenseRiskMatrix(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", tt.name)
				return
			} else if (got != nil && tt.want != nil) && !assert.Equal(t, got, (tt.want)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex [%v]", tt.name, tt.want, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", tt.name))
			}
		})
	}
}
