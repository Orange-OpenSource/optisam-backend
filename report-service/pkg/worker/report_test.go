package worker

import (
	"context"
	"database/sql"
	"errors"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue/job"
	ls "optisam-backend/license-service/pkg/api/v1"
	mockls "optisam-backend/license-service/pkg/api/v1/mock"
	repv1 "optisam-backend/report-service/pkg/repository/v1"
	dmock "optisam-backend/report-service/pkg/repository/v1/dmock"
	"optisam-backend/report-service/pkg/repository/v1/mock"
	"optisam-backend/report-service/pkg/repository/v1/postgres/db"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestWorker_DoWork(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var licenseClient ls.LicenseServiceClient
	var rep repv1.Report
	var drep repv1.DgraphReport
	type args struct {
		ctx context.Context
		j   *job.Job
	}
	tests := []struct {
		name    string
		w       *Worker
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "SUCCESS - ProductEquipmentsReport",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				dmockRepo.EXPECT().EquipmentTypeParents(ctx, "partition", "Scope1").Times(1).Return([]string{"server", "cluster"}, nil)
				dmockRepo.EXPECT().EquipmentTypeAttrs(ctx, "partition", "Scope1").Times(1).Return([]*repv1.EquipmentAttributes{
					{
						AttributeName:       "ap1",
						AttributeIdentifier: true,
						ParentIdentifier:    false,
					},
					{
						AttributeName:       "ap2",
						AttributeIdentifier: false,
						ParentIdentifier:    true,
					},
				}, nil)

				gomock.InOrder(
					dmockRepo.EXPECT().ProductEquipments(ctx, "p1", "Scope1", "partition").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "e11",
							EquipmentType: "partition",
						},
					}, nil),
					dmockRepo.EXPECT().EquipmentAttributes(ctx, "e11", "partition", []*repv1.EquipmentAttributes{
						{
							AttributeName:       "ap1",
							AttributeIdentifier: true,
							ParentIdentifier:    false,
						},
						{
							AttributeName:       "ap2",
							AttributeIdentifier: false,
							ParentIdentifier:    true,
						},
					}, "Scope1").Return([]byte(`{"ap1":"apv1","ap2":"apv2"}`), nil),
					dmockRepo.EXPECT().EquipmentParents(ctx, "e11", "partition", "Scope1").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "pp1",
							EquipmentType: "server",
						},
						{
							EquipmentID:   "pp2",
							EquipmentType: "cluster",
						},
					}, nil),
					dmockRepo.EXPECT().ProductEquipments(ctx, "p2", "Scope1", "partition").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "e2",
							EquipmentType: "partition",
						},
					}, nil),
					dmockRepo.EXPECT().EquipmentAttributes(ctx, "e2", "partition", []*repv1.EquipmentAttributes{
						{
							AttributeName:       "ap1",
							AttributeIdentifier: true,
							ParentIdentifier:    false,
						},
						{
							AttributeName:       "ap2",
							AttributeIdentifier: false,
							ParentIdentifier:    true,
						},
					}, "Scope1").Return([]byte(`{"ap1":"apv3","ap2":"apv4"}`), nil),
					dmockRepo.EXPECT().EquipmentParents(ctx, "e2", "partition", "Scope1").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "pp3",
							EquipmentType: "server",
						},
						{
							EquipmentID:   "pp4",
							EquipmentType: "cluster",
						},
					}, nil),
				)
				finalJson := []byte(`[{"swidtag":"p1","editor":"e1","partition":"e11","ap1":"apv1","ap2":"apv2","server":"pp1","cluster":"pp2"},{"swidtag":"p2","editor":"e1","partition":"e2","ap1":"apv3","ap2":"apv4","server":"pp3","cluster":"pp4"}]`)
				mockrepo.EXPECT().InsertReportData(ctx, db.InsertReportDataParams{
					ReportID:       int32(1),
					ReportDataJson: finalJson,
				}).Return(nil).Times(1)
				mockrepo.EXPECT().UpdateReportStatus(ctx, db.UpdateReportStatusParams{
					ReportStatus: db.ReportStatusCOMPLETED,
					ReportID:     int32(1),
				}).Times(1).Return(nil)

			},
		},
		{
			name: "SUCCESS - Acquired Right Reports",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"AcqRightsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"]},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				gomock.InOrder(
					mocklicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &ls.ListAcquiredRightsForProductRequest{SwidTag: "p1", Scope: "Scope1"}).Times(1).Return(&ls.ListAcquiredRightsForProductResponse{
						AcqRights: []*ls.ProductAcquiredRights{
							{
								SKU:            "sku1",
								SwidTag:        "p1",
								Metric:         "metric1",
								NumCptLicences: int32(1000),
								NumAcqLicences: int32(10000),
								TotalCost:      float64(104.5),
								DeltaNumber:    int32(9000),
								DeltaCost:      float64(100.00),
								AvgUnitPrice:   float64(2.5),
							},
						},
					}, nil),
					mocklicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &ls.ListAcquiredRightsForProductRequest{SwidTag: "p2", Scope: "Scope1"}).Times(1).Return(&ls.ListAcquiredRightsForProductResponse{
						AcqRights: []*ls.ProductAcquiredRights{
							{
								SKU:            "sku2",
								SwidTag:        "p2",
								Metric:         "metric2",
								NumCptLicences: int32(1001),
								NumAcqLicences: int32(10001),
								TotalCost:      float64(104.6),
								DeltaNumber:    int32(9001),
								DeltaCost:      float64(100.01),
								AvgUnitPrice:   float64(2.6),
							},
						},
					}, nil),
				)

				finaljson := []byte(`[{"sku":"sku1","swidtag":"p1","editor":"e1","metric":"metric1","computedLicenses":1000,"acquiredLicenses":10000,"delta(number)":9000,"delta(cost)":100,"totalcost":104.5,"avgunitprice":2.5},{"sku":"sku2","swidtag":"p2","editor":"e1","metric":"metric2","computedLicenses":1001,"acquiredLicenses":10001,"delta(number)":9001,"delta(cost)":100.01,"totalcost":104.6,"avgunitprice":2.6}]`)
				mockrepo.EXPECT().InsertReportData(ctx, db.InsertReportDataParams{
					ReportID:       int32(1),
					ReportDataJson: finaljson,
				}).Return(nil).Times(1)
				mockrepo.EXPECT().UpdateReportStatus(ctx, db.UpdateReportStatusParams{
					ReportStatus: db.ReportStatusCOMPLETED,
					ReportID:     int32(1),
				}).Times(1).Return(nil)

			},
		},
		{
			name: "SUCCESS - with No Acquired Right of one product",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"AcqRightsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"]},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				gomock.InOrder(
					mocklicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &ls.ListAcquiredRightsForProductRequest{SwidTag: "p1", Scope: "Scope1"}).Times(1).Return(&ls.ListAcquiredRightsForProductResponse{
						AcqRights: []*ls.ProductAcquiredRights{
							{
								SKU:            "sku1",
								SwidTag:        "p1",
								Metric:         "metric1",
								NumCptLicences: int32(1000),
								NumAcqLicences: int32(10000),
								TotalCost:      float64(104.5),
								DeltaNumber:    int32(9000),
								DeltaCost:      float64(100.00),
								AvgUnitPrice:   float64(2.5),
							},
						},
					}, nil),
					mocklicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &ls.ListAcquiredRightsForProductRequest{SwidTag: "p2", Scope: "Scope1"}).Times(1).Return(&ls.ListAcquiredRightsForProductResponse{
						AcqRights: nil,
					}, nil),
				)

				finaljson := []byte(`[{"sku":"sku1","swidtag":"p1","editor":"e1","metric":"metric1","computedLicenses":1000,"acquiredLicenses":10000,"delta(number)":9000,"delta(cost)":100,"totalcost":104.5,"avgunitprice":2.5}]`)
				mockrepo.EXPECT().InsertReportData(ctx, db.InsertReportDataParams{
					ReportID:       int32(1),
					ReportDataJson: finaljson,
				}).Return(nil).Times(1)
				mockrepo.EXPECT().UpdateReportStatus(ctx, db.UpdateReportStatusParams{
					ReportStatus: db.ReportStatusCOMPLETED,
					ReportID:     int32(1),
				}).Times(1).Return(nil)

			},
		},
		{
			name: "FAILURE - Error in db/InsertReportData",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"AcqRightsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"]},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				gomock.InOrder(
					mocklicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &ls.ListAcquiredRightsForProductRequest{SwidTag: "p1", Scope: "Scope1"}).Times(1).Return(&ls.ListAcquiredRightsForProductResponse{
						AcqRights: []*ls.ProductAcquiredRights{
							{
								SKU:            "sku1",
								SwidTag:        "p1",
								Metric:         "metric1",
								NumCptLicences: int32(1000),
								NumAcqLicences: int32(10000),
								TotalCost:      float64(104.5),
								DeltaNumber:    int32(9000),
								DeltaCost:      float64(100.00),
								AvgUnitPrice:   float64(2.5),
							},
						},
					}, nil),
					mocklicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &ls.ListAcquiredRightsForProductRequest{SwidTag: "p2", Scope: "Scope1"}).Times(1).Return(&ls.ListAcquiredRightsForProductResponse{
						AcqRights: []*ls.ProductAcquiredRights{
							{
								SKU:            "sku2",
								SwidTag:        "p2",
								Metric:         "metric2",
								NumCptLicences: int32(1001),
								NumAcqLicences: int32(10001),
								TotalCost:      float64(104.6),
								DeltaNumber:    int32(9001),
								DeltaCost:      float64(100.01),
								AvgUnitPrice:   float64(2.6),
							},
						},
					}, nil),
				)

				finaljson := []byte(`[{"sku":"sku1","swidtag":"p1","editor":"e1","metric":"metric1","computedLicenses":1000,"acquiredLicenses":10000,"delta(number)":9000,"delta(cost)":100,"totalcost":104.5,"avgunitprice":2.5},{"sku":"sku2","swidtag":"p2","editor":"e1","metric":"metric2","computedLicenses":1001,"acquiredLicenses":10001,"delta(number)":9001,"delta(cost)":100.01,"totalcost":104.6,"avgunitprice":2.6}]`)
				mockrepo.EXPECT().InsertReportData(ctx, db.InsertReportDataParams{
					ReportID:       int32(1),
					ReportDataJson: finaljson,
				}).Return(errors.New("Internal Error")).Times(1)

			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in db/updateReportStatus",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"AcqRightsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"]},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				gomock.InOrder(
					mocklicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &ls.ListAcquiredRightsForProductRequest{SwidTag: "p1", Scope: "Scope1"}).Times(1).Return(&ls.ListAcquiredRightsForProductResponse{
						AcqRights: []*ls.ProductAcquiredRights{
							{
								SKU:            "sku1",
								SwidTag:        "p1",
								Metric:         "metric1",
								NumCptLicences: int32(1000),
								NumAcqLicences: int32(10000),
								TotalCost:      float64(104.5),
								DeltaNumber:    int32(9000),
								DeltaCost:      float64(100.00),
								AvgUnitPrice:   float64(2.5),
							},
						},
					}, nil),
					mocklicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &ls.ListAcquiredRightsForProductRequest{SwidTag: "p2", Scope: "Scope1"}).Times(1).Return(&ls.ListAcquiredRightsForProductResponse{
						AcqRights: []*ls.ProductAcquiredRights{
							{
								SKU:            "sku2",
								SwidTag:        "p2",
								Metric:         "metric2",
								NumCptLicences: int32(1001),
								NumAcqLicences: int32(10001),
								TotalCost:      float64(104.6),
								DeltaNumber:    int32(9001),
								DeltaCost:      float64(100.01),
								AvgUnitPrice:   float64(2.6),
							},
						},
					}, nil),
				)

				finaljson := []byte(`[{"sku":"sku1","swidtag":"p1","editor":"e1","metric":"metric1","computedLicenses":1000,"acquiredLicenses":10000,"delta(number)":9000,"delta(cost)":100,"totalcost":104.5,"avgunitprice":2.5},{"sku":"sku2","swidtag":"p2","editor":"e1","metric":"metric2","computedLicenses":1001,"acquiredLicenses":10001,"delta(number)":9001,"delta(cost)":100.01,"totalcost":104.6,"avgunitprice":2.6}]`)
				mockrepo.EXPECT().InsertReportData(ctx, db.InsertReportDataParams{
					ReportID:       int32(1),
					ReportDataJson: finaljson,
				}).Return(nil).Times(1)
				mockrepo.EXPECT().UpdateReportStatus(ctx, db.UpdateReportStatusParams{
					ReportStatus: db.ReportStatusCOMPLETED,
					ReportID:     int32(1),
				}).Times(1).Return(errors.New("Internal Error"))

			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in licenseService/ListAcqRightsForProduct",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"AcqRightsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"]},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				gomock.InOrder(
					mocklicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &ls.ListAcquiredRightsForProductRequest{SwidTag: "p1", Scope: "Scope1"}).Return(nil, errors.New("Internal Error")),
				)

			},
			wantErr: true,
		},
		{
			name: "SUCCESS with parent missing",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				// mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				// licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				dmockRepo.EXPECT().EquipmentTypeParents(ctx, "partition", "Scope1").Times(1).Return([]string{"server", "cluster"}, nil)
				dmockRepo.EXPECT().EquipmentTypeAttrs(ctx, "partition", "Scope1").Times(1).Return([]*repv1.EquipmentAttributes{
					{
						AttributeName:       "ap1",
						AttributeIdentifier: true,
						ParentIdentifier:    false,
					},
					{
						AttributeName:       "ap2",
						AttributeIdentifier: false,
						ParentIdentifier:    true,
					},
				}, nil)

				gomock.InOrder(
					dmockRepo.EXPECT().ProductEquipments(ctx, "p1", "Scope1", "partition").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "e11",
							EquipmentType: "partition",
						},
					}, nil),
					dmockRepo.EXPECT().EquipmentAttributes(ctx, "e11", "partition", []*repv1.EquipmentAttributes{
						{
							AttributeName:       "ap1",
							AttributeIdentifier: true,
							ParentIdentifier:    false,
						},
						{
							AttributeName:       "ap2",
							AttributeIdentifier: false,
							ParentIdentifier:    true,
						},
					}, "Scope1").Return([]byte(`{"ap1":"apv1","ap2":"apv2"}`), nil),
					dmockRepo.EXPECT().EquipmentParents(ctx, "e11", "partition", "Scope1").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "pp1",
							EquipmentType: "server",
						},
					}, nil),
					dmockRepo.EXPECT().ProductEquipments(ctx, "p2", "Scope1", "partition").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "e2",
							EquipmentType: "partition",
						},
					}, nil),
					dmockRepo.EXPECT().EquipmentAttributes(ctx, "e2", "partition", []*repv1.EquipmentAttributes{
						{
							AttributeName:       "ap1",
							AttributeIdentifier: true,
							ParentIdentifier:    false,
						},
						{
							AttributeName:       "ap2",
							AttributeIdentifier: false,
							ParentIdentifier:    true,
						},
					}, "Scope1").Return([]byte(`{"ap1":"apv3","ap2":"apv4"}`), nil),
					dmockRepo.EXPECT().EquipmentParents(ctx, "e2", "partition", "Scope1").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "pp3",
							EquipmentType: "server",
						},
						{
							EquipmentID:   "pp4",
							EquipmentType: "cluster",
						},
					}, nil),
				)
				finalJson := []byte(`[{"swidtag":"p1","editor":"e1","partition":"e11","ap1":"apv1","ap2":"apv2","server":"pp1","cluster":""},{"swidtag":"p2","editor":"e1","partition":"e2","ap1":"apv3","ap2":"apv4","server":"pp3","cluster":"pp4"}]`)
				mockrepo.EXPECT().InsertReportData(ctx, db.InsertReportDataParams{
					ReportID:       int32(1),
					ReportDataJson: finalJson,
				}).Return(nil).Times(1)
				mockrepo.EXPECT().UpdateReportStatus(ctx, db.UpdateReportStatusParams{
					ReportStatus: db.ReportStatusCOMPLETED,
					ReportID:     int32(1),
				}).Times(1).Return(nil)

			},
		},
		{
			name: "FAILURE - Json Marshal Error	- Envelope",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				},
			},
			setup: func() {
			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in db/EquipmentTypeParents",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				// mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				// licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				dmockRepo.EXPECT().EquipmentTypeParents(ctx, "partition", "Scope1").Times(1).Return(nil, errors.New("Internal Error"))

			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in db/EquipmentTypeAttrs",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				// mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				// licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				dmockRepo.EXPECT().EquipmentTypeParents(ctx, "partition", "Scope1").Times(1).Return([]string{"server", "cluster"}, nil)
				dmockRepo.EXPECT().EquipmentTypeAttrs(ctx, "partition", "Scope1").Times(1).Return(nil, errors.New("Internal Error"))
			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in db/ProductEquipments",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				// mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				// licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				dmockRepo.EXPECT().EquipmentTypeParents(ctx, "partition", "Scope1").Times(1).Return([]string{"server", "cluster"}, nil)
				dmockRepo.EXPECT().EquipmentTypeAttrs(ctx, "partition", "Scope1").Times(1).Return([]*repv1.EquipmentAttributes{
					{
						AttributeName:       "ap1",
						AttributeIdentifier: true,
						ParentIdentifier:    false,
					},
					{
						AttributeName:       "ap2",
						AttributeIdentifier: false,
						ParentIdentifier:    true,
					},
				}, nil)

				gomock.InOrder(
					dmockRepo.EXPECT().ProductEquipments(ctx, "p1", "Scope1", "partition").Return(nil, errors.New("Internal Error")),
				)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in db/EquipmentAttributes",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				// mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				// licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				dmockRepo.EXPECT().EquipmentTypeParents(ctx, "partition", "Scope1").Times(1).Return([]string{"server", "cluster"}, nil)
				dmockRepo.EXPECT().EquipmentTypeAttrs(ctx, "partition", "Scope1").Times(1).Return([]*repv1.EquipmentAttributes{
					{
						AttributeName:       "ap1",
						AttributeIdentifier: true,
						ParentIdentifier:    false,
					},
					{
						AttributeName:       "ap2",
						AttributeIdentifier: false,
						ParentIdentifier:    true,
					},
				}, nil)

				gomock.InOrder(
					dmockRepo.EXPECT().ProductEquipments(ctx, "p1", "Scope1", "partition").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "e11",
							EquipmentType: "partition",
						},
					}, nil),
					dmockRepo.EXPECT().EquipmentAttributes(ctx, "e11", "partition", []*repv1.EquipmentAttributes{
						{
							AttributeName:       "ap1",
							AttributeIdentifier: true,
							ParentIdentifier:    false,
						},
						{
							AttributeName:       "ap2",
							AttributeIdentifier: false,
							ParentIdentifier:    true,
						},
					}, "Scope1").Return(nil, errors.New("Internal Error")),
				)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in db/EquipmentParents",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				// mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				// licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				dmockRepo.EXPECT().EquipmentTypeParents(ctx, "partition", "Scope1").Times(1).Return([]string{"server", "cluster"}, nil)
				dmockRepo.EXPECT().EquipmentTypeAttrs(ctx, "partition", "Scope1").Times(1).Return([]*repv1.EquipmentAttributes{
					{
						AttributeName:       "ap1",
						AttributeIdentifier: true,
						ParentIdentifier:    false,
					},
					{
						AttributeName:       "ap2",
						AttributeIdentifier: false,
						ParentIdentifier:    true,
					},
				}, nil)

				gomock.InOrder(
					dmockRepo.EXPECT().ProductEquipments(ctx, "p1", "Scope1", "partition").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "e11",
							EquipmentType: "partition",
						},
					}, nil),
					dmockRepo.EXPECT().EquipmentAttributes(ctx, "e11", "partition", []*repv1.EquipmentAttributes{
						{
							AttributeName:       "ap1",
							AttributeIdentifier: true,
							ParentIdentifier:    false,
						},
						{
							AttributeName:       "ap2",
							AttributeIdentifier: false,
							ParentIdentifier:    true,
						},
					}, "Scope1").Return([]byte(`{"ap1":"apv1","ap2":"apv2"}`), nil),
					dmockRepo.EXPECT().EquipmentParents(ctx, "e11", "partition", "Scope1").Return(nil, errors.New("Internal Error")),
				)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in db/InsertReportData",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				// mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				// licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				dmockRepo.EXPECT().EquipmentTypeParents(ctx, "partition", "Scope1").Times(1).Return([]string{"server", "cluster"}, nil)
				dmockRepo.EXPECT().EquipmentTypeAttrs(ctx, "partition", "Scope1").Times(1).Return([]*repv1.EquipmentAttributes{
					{
						AttributeName:       "ap1",
						AttributeIdentifier: true,
						ParentIdentifier:    false,
					},
					{
						AttributeName:       "ap2",
						AttributeIdentifier: false,
						ParentIdentifier:    true,
					},
				}, nil)

				gomock.InOrder(
					dmockRepo.EXPECT().ProductEquipments(ctx, "p1", "Scope1", "partition").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "e11",
							EquipmentType: "partition",
						},
					}, nil),
					dmockRepo.EXPECT().EquipmentAttributes(ctx, "e11", "partition", []*repv1.EquipmentAttributes{
						{
							AttributeName:       "ap1",
							AttributeIdentifier: true,
							ParentIdentifier:    false,
						},
						{
							AttributeName:       "ap2",
							AttributeIdentifier: false,
							ParentIdentifier:    true,
						},
					}, "Scope1").Return([]byte(`{"ap1":"apv1","ap2":"apv2"}`), nil),
					dmockRepo.EXPECT().EquipmentParents(ctx, "e11", "partition", "Scope1").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "pp1",
							EquipmentType: "server",
						},
						{
							EquipmentID:   "pp2",
							EquipmentType: "cluster",
						},
					}, nil),
					dmockRepo.EXPECT().ProductEquipments(ctx, "p2", "Scope1", "partition").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "e2",
							EquipmentType: "partition",
						},
					}, nil),
					dmockRepo.EXPECT().EquipmentAttributes(ctx, "e2", "partition", []*repv1.EquipmentAttributes{
						{
							AttributeName:       "ap1",
							AttributeIdentifier: true,
							ParentIdentifier:    false,
						},
						{
							AttributeName:       "ap2",
							AttributeIdentifier: false,
							ParentIdentifier:    true,
						},
					}, "Scope1").Return([]byte(`{"ap1":"apv3","ap2":"apv4"}`), nil),
					dmockRepo.EXPECT().EquipmentParents(ctx, "e2", "partition", "Scope1").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "pp3",
							EquipmentType: "server",
						},
						{
							EquipmentID:   "pp4",
							EquipmentType: "cluster",
						},
					}, nil),
				)
				finalJson := []byte(`[{"swidtag":"p1","editor":"e1","partition":"e11","ap1":"apv1","ap2":"apv2","server":"pp1","cluster":"pp2"},{"swidtag":"p2","editor":"e1","partition":"e2","ap1":"apv3","ap2":"apv4","server":"pp3","cluster":"pp4"}]`)
				mockrepo.EXPECT().InsertReportData(ctx, db.InsertReportDataParams{
					ReportID:       int32(1),
					ReportDataJson: finalJson,
				}).Return(errors.New("Internal Error")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in db/UpdateReportStatus",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "rw"},
					Status: job.JobStatusPENDING,
					Data:   []byte(`{"report_type":"ProductEquipmentsReport","scope":"Scope1","json":{"editor":"e1","swidtag":["p1","p2"],"equipType":"partition"},"report_id":1}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				// mocklicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				// licenseClient = mocklicenseClient
				mockrepo := mock.NewMockReport(mockCtrl)
				rep = mockrepo
				dmockRepo := dmock.NewMockDgraphReport(mockCtrl)
				drep = dmockRepo

				dmockRepo.EXPECT().EquipmentTypeParents(ctx, "partition", "Scope1").Times(1).Return([]string{"server", "cluster"}, nil)
				dmockRepo.EXPECT().EquipmentTypeAttrs(ctx, "partition", "Scope1").Times(1).Return([]*repv1.EquipmentAttributes{
					{
						AttributeName:       "ap1",
						AttributeIdentifier: true,
						ParentIdentifier:    false,
					},
					{
						AttributeName:       "ap2",
						AttributeIdentifier: false,
						ParentIdentifier:    true,
					},
				}, nil)

				gomock.InOrder(
					dmockRepo.EXPECT().ProductEquipments(ctx, "p1", "Scope1", "partition").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "e11",
							EquipmentType: "partition",
						},
					}, nil),
					dmockRepo.EXPECT().EquipmentAttributes(ctx, "e11", "partition", []*repv1.EquipmentAttributes{
						{
							AttributeName:       "ap1",
							AttributeIdentifier: true,
							ParentIdentifier:    false,
						},
						{
							AttributeName:       "ap2",
							AttributeIdentifier: false,
							ParentIdentifier:    true,
						},
					}, "Scope1").Return([]byte(`{"ap1":"apv1","ap2":"apv2"}`), nil),
					dmockRepo.EXPECT().EquipmentParents(ctx, "e11", "partition", "Scope1").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "pp1",
							EquipmentType: "server",
						},
						{
							EquipmentID:   "pp2",
							EquipmentType: "cluster",
						},
					}, nil),
					dmockRepo.EXPECT().ProductEquipments(ctx, "p2", "Scope1", "partition").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "e2",
							EquipmentType: "partition",
						},
					}, nil),
					dmockRepo.EXPECT().EquipmentAttributes(ctx, "e2", "partition", []*repv1.EquipmentAttributes{
						{
							AttributeName:       "ap1",
							AttributeIdentifier: true,
							ParentIdentifier:    false,
						},
						{
							AttributeName:       "ap2",
							AttributeIdentifier: false,
							ParentIdentifier:    true,
						},
					}, "Scope1").Return([]byte(`{"ap1":"apv3","ap2":"apv4"}`), nil),
					dmockRepo.EXPECT().EquipmentParents(ctx, "e2", "partition", "Scope1").Return([]*repv1.ProductEquipment{
						{
							EquipmentID:   "pp3",
							EquipmentType: "server",
						},
						{
							EquipmentID:   "pp4",
							EquipmentType: "cluster",
						},
					}, nil),
				)
				finalJson := []byte(`[{"swidtag":"p1","editor":"e1","partition":"e11","ap1":"apv1","ap2":"apv2","server":"pp1","cluster":"pp2"},{"swidtag":"p2","editor":"e1","partition":"e2","ap1":"apv3","ap2":"apv4","server":"pp3","cluster":"pp4"}]`)
				mockrepo.EXPECT().InsertReportData(ctx, db.InsertReportDataParams{
					ReportID:       int32(1),
					ReportDataJson: finalJson,
				}).Return(nil).Times(1)
				mockrepo.EXPECT().UpdateReportStatus(ctx, db.UpdateReportStatusParams{
					ReportStatus: db.ReportStatusCOMPLETED,
					ReportID:     int32(1),
				}).Times(1).Return(errors.New("Internal Error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			w := &Worker{
				id:            "rw",
				reportRepo:    rep,
				dgraphRepo:    drep,
				licenseClient: licenseClient,
			}
			if err := w.DoWork(tt.args.ctx, tt.args.j); (err != nil) != tt.wantErr {
				t.Errorf("Worker.DoWork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
