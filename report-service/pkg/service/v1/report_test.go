// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"database/sql"
	"errors"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	queuemock "optisam-backend/common/optisam/workerqueue/mock"
	v1 "optisam-backend/report-service/pkg/api/v1"
	repv1 "optisam-backend/report-service/pkg/repository/v1"
	"optisam-backend/report-service/pkg/repository/v1/mock"
	"optisam-backend/report-service/pkg/repository/v1/postgres/db"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestReportServiceServer_SubmitReport(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	// var licenseClient ls.LicenseServiceClient
	var rep repv1.Report
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.SubmitReportRequest
	}
	tests := []struct {
		name    string
		r       *ReportServiceServer
		args    args
		setup   func()
		want    *v1.SubmitReportResponse
		wantErr bool
	}{
		{
			name: "SUCCESS - ProductEquipmentReport",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 2,
					ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
						ProductEquipmentsReport: &v1.ProductEquipmentsReport{
							Editor:    "e1",
							Swidtag:   []string{"p1", "p2"},
							EquipType: "partition",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

				mockRepo.EXPECT().GetReportType(ctx, int32(2)).Times(1).Return(db.ReportType{
					ReportTypeID:   2,
					ReportTypeName: "ProductEquipments",
				}, nil)

				mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
					Scope:          "Scope1",
					ReportTypeID:   2,
					ReportStatus:   db.ReportStatusPENDING,
					CreatedBy:      "admin@superuser.com",
					ReportMetadata: []byte(`{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"}`),
				}).Times(1).Return(int32(1), nil)

				mockworkerqueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				}, "rw").Times(1).Return(int32(1), nil)

			},
			want: &v1.SubmitReportResponse{
				Success: true,
			},
		},
		{
			name: "FAILURE - Error in report_type",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 2,
					ReportMetadata: &v1.SubmitReportRequest_AcqrightsReport{
						AcqrightsReport: &v1.AcqRightsReport{
							Swidtag: []string{"p1", "p2"},
							Editor:  "e1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

				mockRepo.EXPECT().GetReportType(ctx, int32(2)).Times(1).Return(db.ReportType{
					ReportTypeID:   2,
					ReportTypeName: "ProductEquipments",
				}, nil)

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in db/GetReportType",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 2,
					ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
						ProductEquipmentsReport: &v1.ProductEquipmentsReport{
							Editor:    "e1",
							Swidtag:   []string{"p1", "p2"},
							EquipType: "partition",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

				mockRepo.EXPECT().GetReportType(ctx, int32(2)).Times(1).Return(db.ReportType{}, errors.New("Internal Error"))
			},
			wantErr: true,
			want: &v1.SubmitReportResponse{
				Success: false,
			},
		},
		{
			name: "FAILURE - Cannot Find claim in context",
			args: args{
				ctx: context.Background(),
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 2,
					ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
						ProductEquipmentsReport: &v1.ProductEquipmentsReport{
							Editor:    "e1",
							Swidtag:   []string{"p1", "p2"},
							EquipType: "partition",
						},
					},
				},
			},
			setup: func() {
			},
			wantErr: true,
			want: &v1.SubmitReportResponse{
				Success: false,
			},
		},
		{
			name: "FAILURE - Cannot Find Scope",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "OFR",
					ReportTypeId: 2,
					ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
						ProductEquipmentsReport: &v1.ProductEquipmentsReport{
							Editor:    "e1",
							Swidtag:   []string{"p1", "p2"},
							EquipType: "partition",
						},
					},
				},
			},
			setup: func() {
			},
			wantErr: true,
			want: &v1.SubmitReportResponse{
				Success: false,
			},
		},
		{
			name: "FAILURE - Marshall Error",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 2,
					ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
						ProductEquipmentsReport: nil,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

				mockRepo.EXPECT().GetReportType(ctx, int32(2)).Times(1).Return(db.ReportType{
					ReportTypeID:   2,
					ReportTypeName: "ProductEquipments",
				}, nil)

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAILURE - Marshall Error in Envelope",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        string('0' - 48),
					ReportTypeId: 2,
					ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
						ProductEquipmentsReport: &v1.ProductEquipmentsReport{
							Editor:    "e1",
							Swidtag:   []string{"p1", "p2"},
							EquipType: "partition",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

				mockRepo.EXPECT().GetReportType(ctx, int32(2)).Times(1).Return(db.ReportType{
					ReportTypeID:   2,
					ReportTypeName: "ProductEquipments",
				}, nil)

				mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
					Scope:          "Scope1",
					ReportTypeID:   2,
					ReportStatus:   db.ReportStatusPENDING,
					CreatedBy:      "admin@superuser.com",
					ReportMetadata: []byte(`{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"}`),
				}).Times(1).Return(int32(1), nil)

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in submit report",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 2,
					ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
						ProductEquipmentsReport: &v1.ProductEquipmentsReport{
							Editor:    "e1",
							Swidtag:   []string{"p1", "p2"},
							EquipType: "partition",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

				mockRepo.EXPECT().GetReportType(ctx, int32(2)).Times(1).Return(db.ReportType{
					ReportTypeID:   2,
					ReportTypeName: "ProductEquipments",
				}, nil)

				mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
					Scope:          "Scope1",
					ReportTypeID:   2,
					ReportStatus:   db.ReportStatusPENDING,
					CreatedBy:      "admin@superuser.com",
					ReportMetadata: []byte(`{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"}`),
				}).Times(1).Return(int32(0), errors.New("Internal Error"))

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in pushJob",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 2,
					ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
						ProductEquipmentsReport: &v1.ProductEquipmentsReport{
							Editor:    "e1",
							Swidtag:   []string{"p1", "p2"},
							EquipType: "partition",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

				mockRepo.EXPECT().GetReportType(ctx, int32(2)).Times(1).Return(db.ReportType{
					ReportTypeID:   2,
					ReportTypeName: "ProductEquipments",
				}, nil)

				mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
					Scope:          "Scope1",
					ReportTypeID:   2,
					ReportStatus:   db.ReportStatusPENDING,
					CreatedBy:      "admin@superuser.com",
					ReportMetadata: []byte(`{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"}`),
				}).Times(1).Return(int32(1), nil)

				mockworkerqueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				}, "rw").Times(1).Return(int32(0), errors.New("Internal Error"))

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			r := NewReportServiceServer(rep, queue)
			got, err := r.SubmitReport(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportServiceServer.SubmitReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReportServiceServer.SubmitReport() = %v, want %v", got, tt.want)
			}
		})
	}
}
