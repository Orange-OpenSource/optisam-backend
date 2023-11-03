package v1

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"
	queuemock "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/mock"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/api/v1"
	repv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/repository/v1/mock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/repository/v1/postgres/db"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/worker"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/jsonpb"
	prodv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/thirdparty/product-service/pkg/api/v1"
	pmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/thirdparty/product-service/pkg/api/v1/mock"
)

func Test_DropReportData(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	ctx1 := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repv1.Report
	var queue workerqueue.Workerqueue

	tests := []struct {
		name    string
		r       *ReportServiceServer
		ctx     context.Context
		setup   func()
		input   *v1.DropReportDataRequest
		wantErr bool
	}{
		{
			name:    "ScopeNotFound",
			ctx:     ctx,
			setup:   func() {},
			input:   &v1.DropReportDataRequest{Scope: "Scope6"},
			wantErr: true,
		},
		{
			name:    "ClaimsNotFound",
			ctx:     context.Background(),
			setup:   func() {},
			input:   &v1.DropReportDataRequest{Scope: "Scope1"},
			wantErr: true,
		},
		{
			name:    "Not SuperAdmin",
			ctx:     ctx1,
			setup:   func() {},
			input:   &v1.DropReportDataRequest{Scope: "Scope1"},
			wantErr: true,
		},
		{
			name: "DBError",
			ctx:  ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				mockRepo.EXPECT().DeleteReportsByScope(ctx, "Scope1").Return(errors.New("DBError")).Times(1)
			},
			input:   &v1.DropReportDataRequest{Scope: "Scope1"},
			wantErr: true,
		},
		{
			name: "SuccessFullyReportDeleted",
			ctx:  ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				mockRepo.EXPECT().DeleteReportsByScope(ctx, "Scope1").Return(nil).Times(1)
			},
			input:   &v1.DropReportDataRequest{Scope: "Scope1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			r := &ReportServiceServer{
				reportRepo: rep,
				queue:      queue,
			}
			_, err := r.DropReportData(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportServiceServer.DropReportData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestReportServiceServer_SubmitReport(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	// var licenseClient ls.LicenseServiceClient
	var rep repv1.Report
	var queue workerqueue.Workerqueue
	var pclient prodv1.ProductServiceClient
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
			name: "SUCCESS - ScopeExpensesByEditorReport",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 3,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

				mockRepo.EXPECT().GetReportType(ctx, int32(3)).Return(db.ReportType{
					ReportTypeID:   3,
					ReportTypeName: "EditorExpenses",
				}, nil)
				rawJSON := json.RawMessage("[]")
				fcall :=
					mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
						Scope:          "Scope1",
						ReportTypeID:   3,
						ReportStatus:   db.ReportStatusPENDING,
						CreatedBy:      "admin@superuser.com",
						ReportMetadata: rawJSON,
					}).Return(int32(1), nil).Times(1)
				env := worker.Envelope{Type: worker.ScopeExpensesByEditorReport, Scope: "Scope1", JSON: rawJSON, ReportID: int32(1)}
				envolveData, _ := json.Marshal(env)
				mockworkerqueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData,
				}, "rw").Return(int32(1), nil).After(fcall)

			},
			want: &v1.SubmitReportResponse{
				Success: true,
			},
			wantErr: false,
		},
		{
			name: "Fail - Pushjob error",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 3,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

				mockRepo.EXPECT().GetReportType(ctx, int32(3)).Return(db.ReportType{
					ReportTypeID:   3,
					ReportTypeName: "EditorExpenses",
				}, nil)
				rawJSON := json.RawMessage("[]")
				fcall :=
					mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
						Scope:          "Scope1",
						ReportTypeID:   3,
						ReportStatus:   db.ReportStatusPENDING,
						CreatedBy:      "admin@superuser.com",
						ReportMetadata: rawJSON,
					}).Return(int32(1), nil).Times(1)
				env := worker.Envelope{Type: worker.ScopeExpensesByEditorReport, Scope: "Scope1", JSON: rawJSON, ReportID: int32(1)}
				envolveData, _ := json.Marshal(env)
				mockworkerqueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData,
				}, "rw").Return(int32(1), errors.New("some error")).After(fcall)

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "Fail - ScopeExpensesByEditorReport",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 3,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

				mockRepo.EXPECT().GetReportType(ctx, int32(3)).Return(db.ReportType{
					ReportTypeID:   3,
					ReportTypeName: "EditorExpenses",
				}, nil)
				rawJSON := json.RawMessage("[]")
				mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
					Scope:          "Scope1",
					ReportTypeID:   3,
					ReportStatus:   db.ReportStatusPENDING,
					CreatedBy:      "admin@superuser.com",
					ReportMetadata: rawJSON,
				}).Return(int32(1), errors.New("some error")).Times(1)
			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name:    "no claims",
			wantErr: true,
			args: args{
				ctx: context.Background(),
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 3,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
		},
		{
			name: "Scope not found",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scopenotfound",
					ReportTypeId: 3,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "Failiure - ProductEquipmentReport",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 2,
					ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
						ProductEquipmentsReport: &v1.ProductEquipmentsReport{
							Editor: "e1",
							//Swidtag:   []string{"p1", "p2"},
							EquipType: "partition",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				mockProductClient := pmock.NewMockProductServiceClient(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				pclient = mockProductClient
				mockProductClient.EXPECT().ListAggregationEditors(ctx, &prodv1.ListAggregationEditorsRequest{Scope: "Scope1"}).Return(&prodv1.ListAggregationEditorsResponse{Editor: []string{"editor"}}, nil).AnyTimes()
				mockRepo.EXPECT().GetReportType(ctx, int32(2)).Times(1).Return(db.ReportType{
					ReportTypeID:   2,
					ReportTypeName: "ProductEquipments",
				}, nil)

				mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
					Scope:          "Scope1",
					ReportTypeID:   2,
					ReportStatus:   db.ReportStatusPENDING,
					CreatedBy:      "admin@superuser.com",
					ReportMetadata: []byte(`{"editor":"e1","equipType":"partition"}`),
				}).Return(int32(1), nil).AnyTimes()

				mockworkerqueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","equipType":"partition"},"report_id":1}`),
				}, "rw").Return(int32(1), nil).AnyTimes()

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "Fail - report type 1",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 1,
					ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
						ProductEquipmentsReport: &v1.ProductEquipmentsReport{
							Editor: "e1",
							//Swidtag:   []string{"p1", "p2"},
							EquipType: "partition",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				mockProductClient := pmock.NewMockProductServiceClient(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				pclient = mockProductClient
				mockRepo.EXPECT().GetReportType(ctx, int32(1)).Times(1).Return(db.ReportType{
					ReportTypeID:   1,
					ReportTypeName: "ProductEquipments",
				}, nil)
				mockProductClient.EXPECT().ListAggregationEditors(ctx, &prodv1.ListAggregationEditorsRequest{Scope: "Scope1"}).Return(&prodv1.ListAggregationEditorsResponse{Editor: []string{"editor", "e1"}}, nil).AnyTimes()

				mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
					Scope:          "Scope1",
					ReportTypeID:   1,
					ReportStatus:   db.ReportStatusPENDING,
					CreatedBy:      "admin@superuser.com",
					ReportMetadata: []byte(`{"editor":"e1","equipType":"partition"}`),
				}).Return(int32(1), nil).AnyTimes()

				mockworkerqueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","equipType":"partition"},"report_id":1}`),
				}, "rw").Return(int32(1), nil).AnyTimes()

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "Fail - report type 1.1",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 1,
					ReportMetadata: &v1.SubmitReportRequest_AcqrightsReport{
						AcqrightsReport: &v1.AcqRightsReport{
							Editor: "e1",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				mockProductClient := pmock.NewMockProductServiceClient(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				pclient = mockProductClient
				mockRepo.EXPECT().GetReportType(ctx, int32(1)).Times(1).Return(db.ReportType{
					ReportTypeID:   1,
					ReportTypeName: "ProductEquipments",
				}, nil)
				mockProductClient.EXPECT().ListAggregationEditors(ctx, &prodv1.ListAggregationEditorsRequest{Scope: "Scope1"}).Return(&prodv1.ListAggregationEditorsResponse{Editor: []string{"editor", "e1"}}, nil).AnyTimes()
				var j bytes.Buffer
				marshaler := &jsonpb.Marshaler{}
				marshaler.Marshal(&j, &v1.AcqRightsReport{
					Editor: "e1",
				})
				mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
					Scope:          "Scope1",
					ReportTypeID:   1,
					ReportStatus:   db.ReportStatusPENDING,
					CreatedBy:      "admin@superuser.com",
					ReportMetadata: j.Bytes(),
				}).Times(1).Return(int32(1), nil)

				// env := worker.Envelope{Type: worker.ScopeExpensesByEditorReport, Scope: "Scope1", JSON: j.Bytes(), ReportID: int32(1)}
				// envolveData, _ := json.Marshal(env)
				mockworkerqueue.EXPECT().PushJob(ctx, gomock.Any(), gomock.Any()).Return(int32(1), errors.New("some error"))

				// mockworkerqueue.EXPECT().PushJob(ctx, job.Job{
				// 	Type:   sql.NullString{String: "rw"},
				// 	Status: job.JobStatusPENDING,
				// 	Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","equipType":"partition"},"report_id":1}`),
				// }, "rw").Times(1).Return(int32(1), nil)

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "Sucess - ProductEquipmentReport",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        "Scope1",
					ReportTypeId: 2,
					// ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
					// 	ProductEquipmentsReport: &v1.ProductEquipmentsReport{
					// 		Editor: "e1",
					// 		//Swidtag:   []string{"p1", "p2"},
					// 		EquipType: "partition",
					// 	},
					// },
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				mockProductClient := pmock.NewMockProductServiceClient(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				pclient = mockProductClient
				mockProductClient.EXPECT().ListAggregationEditors(ctx, &prodv1.ListAggregationEditorsRequest{Scope: "Scope1"}).Return(&prodv1.ListAggregationEditorsResponse{Editor: []string{"editor", "e1"}}, nil).AnyTimes()
				mockRepo.EXPECT().GetReportType(ctx, int32(2)).Times(1).Return(db.ReportType{
					ReportTypeID:   2,
					ReportTypeName: "ProductEquipments",
				}, nil)

				mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
					Scope:          "Scope1",
					ReportTypeID:   2,
					ReportStatus:   db.ReportStatusPENDING,
					CreatedBy:      "admin@superuser.com",
					ReportMetadata: []byte(`{"editor":"e1","equipType":"partition"}`),
				}).Return(int32(1), nil).AnyTimes()

				mockworkerqueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","equipType":"partition"},"report_id":1}`),
				}, "rw").Return(int32(1), nil).AnyTimes()

			},
			want: &v1.SubmitReportResponse{
				Success: false,
			},
			wantErr: true,
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
							// Swidtag: []string{"p1", "p2"},
							Editor: "e1",
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
							Editor: "e1",
							// Swidtag:   []string{"p1", "p2"},
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
							Editor: "e1",
							// Swidtag:   []string{"p1", "p2"},
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
							Editor: "e1",
							// Swidtag:   []string{"p1", "p2"},
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
			name: "FAILURE - Marshall Error in Envelope",
			args: args{
				ctx: ctx,
				req: &v1.SubmitReportRequest{
					Scope:        string('0' - 48),
					ReportTypeId: 2,
					ReportMetadata: &v1.SubmitReportRequest_ProductEquipmentsReport{
						ProductEquipmentsReport: &v1.ProductEquipmentsReport{},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepo
				queue = mockworkerqueue

				mockRepo.EXPECT().GetReportType(ctx, int32(2)).Return(db.ReportType{
					ReportTypeID:   2,
					ReportTypeName: "ProductEquipments",
				}, nil).AnyTimes()

				mockRepo.EXPECT().SubmitReport(ctx, db.SubmitReportParams{
					Scope:          "Scope1",
					ReportTypeID:   2,
					ReportStatus:   db.ReportStatusPENDING,
					CreatedBy:      "admin@superuser.com",
					ReportMetadata: []byte(`{"editor":"e1","equipType":"partition"}`),
				}).Times(1).Return(int32(1), nil).AnyTimes()

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
							Editor: "e1",
							// Swidtag:   []string{"p1", "p2"},
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
					ReportMetadata: []byte(`{"editor":"e1","equipType":"partition"}`),
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
							Editor: "e1",
							// Swidtag:   []string{"p1", "p2"},
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
					ReportMetadata: []byte(`{"editor":"e1","equipType":"partition"}`),
				}).Times(1).Return(int32(1), nil)

				mockworkerqueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","equipType":"partition"},"report_id":1}`),
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
			r := &ReportServiceServer{
				reportRepo:    rep,
				queue:         queue,
				productClient: pclient,
			}
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

func Test_ListReport(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	// ctx1 := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
	// 	UserID: "admin@superuser.com",
	// 	Role:   "Admin",
	// 	Socpes: []string{"Scope1", "Scope2", "Scope3"},
	// })
	var mockCtrl *gomock.Controller
	var rep repv1.Report
	var queue workerqueue.Workerqueue

	tests := []struct {
		name    string
		r       *ReportServiceServer
		ctx     context.Context
		setup   func()
		input   *v1.ListReportRequest
		wantErr bool
	}{
		{
			name:    "ScopeNotFound",
			ctx:     ctx,
			setup:   func() {},
			input:   &v1.ListReportRequest{Scope: "Scope6"},
			wantErr: true,
		},
		{
			name:    "ClaimsNotFound",
			ctx:     context.Background(),
			setup:   func() {},
			input:   &v1.ListReportRequest{Scope: "Scope1"},
			wantErr: true,
		},
		{
			name: "DBError",
			ctx:  ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				mockRepo.EXPECT().GetReport(ctx, gomock.Any()).Return([]db.GetReportRow{db.GetReportRow{}}, errors.New("DBError")).Times(1)
			},
			input:   &v1.ListReportRequest{Scope: "Scope1"},
			wantErr: true,
		},
		{
			name: "SuccessFullyReport",
			ctx:  ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				mockRepo.EXPECT().GetReport(ctx, gomock.Any()).Return([]db.GetReportRow{db.GetReportRow{ReportMetadata: json.RawMessage(`{"editor":"dd"}`)}}, nil).Times(1)
			},
			input:   &v1.ListReportRequest{Scope: "Scope1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			r := &ReportServiceServer{
				reportRepo: rep,
				queue:      queue,
			}
			_, err := r.ListReport(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportServiceServer.DropReportData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
func Test_DownloadReport(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	// ctx1 := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
	// 	UserID: "admin@superuser.com",
	// 	Role:   "Admin",
	// 	Socpes: []string{"Scope1", "Scope2", "Scope3"},
	// })
	var mockCtrl *gomock.Controller
	var rep repv1.Report
	var queue workerqueue.Workerqueue

	tests := []struct {
		name    string
		r       *ReportServiceServer
		ctx     context.Context
		setup   func()
		input   *v1.DownloadReportRequest
		wantErr bool
	}{
		{
			name:    "ScopeNotFound",
			ctx:     ctx,
			setup:   func() {},
			input:   &v1.DownloadReportRequest{Scope: "Scope6"},
			wantErr: true,
		},
		{
			name:    "ClaimsNotFound",
			ctx:     context.Background(),
			setup:   func() {},
			input:   &v1.DownloadReportRequest{Scope: "Scope1"},
			wantErr: true,
		},
		{
			name: "DBError",
			ctx:  ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				mockRepo.EXPECT().DownloadReport(ctx, gomock.Any()).Return(db.DownloadReportRow{}, errors.New("DBError")).Times(1)
			},
			input:   &v1.DownloadReportRequest{Scope: "Scope1"},
			wantErr: true,
		},
		{
			name: "SuccessFullyReport",
			ctx:  ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				mockRepo.EXPECT().DownloadReport(ctx, gomock.Any()).Return(db.DownloadReportRow{ReportMetadata: json.RawMessage(`{"editor":"dd"}`)}, nil).Times(1)
			},
			input:   &v1.DownloadReportRequest{Scope: "Scope1"},
			wantErr: false,
		},
		{
			name: "SuccessFullyReport ProductEquipments",
			ctx:  ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				mockRepo.EXPECT().DownloadReport(ctx, gomock.Any()).Return(db.DownloadReportRow{ReportTypeName: "ProductEquipments", ReportMetadata: json.RawMessage(`{"editor":"dd"}`)}, nil).Times(1)
			},
			input:   &v1.DownloadReportRequest{Scope: "Scope1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			r := &ReportServiceServer{
				reportRepo: rep,
				queue:      queue,
			}
			_, err := r.DownloadReport(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportServiceServer.DropReportData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
func Test_ListReportType(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	// ctx1 := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
	// 	UserID: "admin@superuser.com",
	// 	Role:   "Admin",
	// 	Socpes: []string{"Scope1", "Scope2", "Scope3"},
	// })
	var mockCtrl *gomock.Controller
	var rep repv1.Report
	var queue workerqueue.Workerqueue

	tests := []struct {
		name    string
		r       *ReportServiceServer
		ctx     context.Context
		setup   func()
		input   *v1.ListReportTypeRequest
		wantErr bool
	}{
		{
			name: "DBError",
			ctx:  ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				mockRepo.EXPECT().GetReportTypes(ctx).Return([]db.ReportType{}, errors.New("DBError")).Times(1)
			},
			input:   &v1.ListReportTypeRequest{},
			wantErr: true,
		},
		{
			name: "SuccessFullyReport",
			ctx:  ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockReport(mockCtrl)
				mockworkerqueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockworkerqueue
				mockRepo.EXPECT().GetReportTypes(ctx).Return([]db.ReportType{}, nil).Times(1)
			},
			input:   &v1.ListReportTypeRequest{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			r := &ReportServiceServer{
				reportRepo: rep,
				queue:      queue,
			}
			_, err := r.ListReportType(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportServiceServer.DropReportData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
