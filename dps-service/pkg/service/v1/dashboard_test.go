package v1

import (
	"context"
	"errors"
	"fmt"
	"testing"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1"
	dbmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1/dbmock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1/postgres/db"
	queuemock "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1/queuemock"
	equipV1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/equipment-service/pkg/api/v1"
	metv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/metric-service/pkg/api/v1"
	prodV1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/product-service/pkg/api/v1"

	"github.com/golang/mock/gomock"
)

var (
	ctx = grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3", "AAK"},
	})
)

func Test_DashboardQualityOverview(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var rep repo.Dps
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	var met metv1.MetricServiceClient
	var equip equipV1.EquipmentServiceClient
	var prod prodV1.ProductServiceClient

	tests := []struct {
		name    string
		ctx     context.Context
		input   *v1.DashboardQualityOverviewRequest
		setup   func(*v1.DashboardQualityOverviewRequest)
		output  *v1.DashboardQualityOverviewResponse
		wantErr bool
		s       *dpsServiceServer
	}{
		{
			name: "claims Not found",
			ctx:  context.Background(),
			input: &v1.DashboardQualityOverviewRequest{
				Scope:          "Scope1",
				NoOfDataPoints: int32(10),
			},
			setup:   func(*v1.DashboardQualityOverviewRequest) {},
			output:  &v1.DashboardQualityOverviewResponse{},
			wantErr: true,
		},
		{
			name: "Scope Not found",
			ctx:  ctx,
			input: &v1.DashboardQualityOverviewRequest{
				Scope:          "Scope10",
				NoOfDataPoints: int32(10),
			},
			setup:   func(*v1.DashboardQualityOverviewRequest) {},
			output:  &v1.DashboardQualityOverviewResponse{},
			wantErr: true,
		},

		{
			name: "correct sheet",
			ctx:  ctx,
			input: &v1.DashboardQualityOverviewRequest{
				Scope:          "AAK",
				NoOfDataPoints: int32(10),
			},
			setup: func(*v1.DashboardQualityOverviewRequest) {
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().GetEntityMonthWise(gomock.Any(), gomock.Any()).Return([]db.GetEntityMonthWiseRow{
					{Filename: "aak_applications.csv", Sum: int64(10)},
					{Filename: "aak_products.csv", Sum: int64(10)},
					{Filename: "aak_products_acquiredrights.csv", Sum: int64(10)},
					{Filename: "aak_equipments.csv", Sum: int64(10)},
				}, nil)
			},
			output:  &v1.DashboardQualityOverviewResponse{},
			wantErr: false,
		},
		{
			name: "correct db err",
			ctx:  ctx,
			input: &v1.DashboardQualityOverviewRequest{
				Scope:          "AAK",
				NoOfDataPoints: int32(10),
			},
			setup: func(*v1.DashboardQualityOverviewRequest) {
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().GetEntityMonthWise(gomock.Any(), gomock.Any()).Return([]db.GetEntityMonthWiseRow{
					{Filename: "aak_applications.csv", Sum: int64(10)},
					{Filename: "aak_products.csv", Sum: int64(10)},
					{Filename: "aak_products_acquiredrights.csv", Sum: int64(10)},
					{Filename: "aak_equipments.csv", Sum: int64(10)},
				}, errors.New("err"))
			},
			output:  &v1.DashboardQualityOverviewResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			tt.s = &dpsServiceServer{
				dpsRepo:   rep,
				queue:     qObj,
				equipment: equip,
				metric:    met,
				product:   prod,
			}
			_, err := tt.s.DashboardQualityOverview(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.DashboardQualityOverview() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("Test passed ", tt.name)
		})
	}
}

// func Test_ListFailureReasons(t *testing.T) {
// 	var mockCtrl *gomock.Controller
// 	var rep repo.Dps
// 	var queue workerqueue.Queue
// 	tests := []struct {
// 		name    string
// 		ctx     context.Context
// 		input   *v1.ListFailureReasonRequest
// 		setup   func(*v1.ListFailureReasonRequest)
// 		output  *v1.ListFailureReasonResponse
// 		wantErr bool
// 	}{

// 		{
// 			name: " Failed Record Present",
// 			ctx:  ctx,
// 			input: &v1.ListFailureReasonRequest{
// 				Scope: "Scope1",
// 			},
// 			setup: func(req *v1.ListFailureReasonRequest) {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepository := dbmock.NewMockDps(mockCtrl)
// 				rep = mockRepository
// 				qYear, qMon, qDay := time.Now().Add(time.Hour * 24 * -(30)).Date()
// 				mockRepository.EXPECT().GetFailureReasons(ctx, db.GetFailureReasonsParams{
// 					Year:  int32(qYear),
// 					Month: int32(qMon),
// 					Day:   int32(qDay),
// 					Data:  json.RawMessage(fmt.Sprintf("%s", req.GetScope()))}).Times(1).Return([]db.GetFailureReasonsRow{
// 					{
// 						FailedRecords: int64(10),
// 						Comments:      sql.NullString{String: "InvalidFileName", Valid: true},
// 					},
// 					{
// 						FailedRecords: int64(10),
// 						Comments:      sql.NullString{String: "FileNotSupported", Valid: true},
// 					},
// 					{
// 						FailedRecords: int64(30),
// 						Comments:      sql.NullString{String: "BadFile", Valid: true},
// 					},
// 					{
// 						FailedRecords: int64(40),
// 						Comments:      sql.NullString{String: "NoDataInFile", Valid: true},
// 					},
// 					{
// 						FailedRecords: int64(50),
// 						Comments:      sql.NullString{String: "HeadersMissing", Valid: true},
// 					},
// 					{
// 						FailedRecords: int64(60),
// 						Comments:      sql.NullString{String: "InsufficentData", Valid: true},
// 					},
// 				}, nil)

// 			},
// 			output: &v1.ListFailureReasonResponse{
// 				FailureReasons: map[string]float32{
// 					"InvalidFileName":  float32(5),
// 					"FileNotSupported": float32(5),
// 					"BadFile":          float32(15),
// 					"NoDataInFile":     float32(20),
// 					"HeadersMissing":   float32(25),
// 					"InsufficentData":  float32(30),
// 				},
// 			},
// 		},
// 		{
// 			name: "Zero Failure Causes",
// 			ctx:  ctx,
// 			input: &v1.ListFailureReasonRequest{
// 				Scope: "Scope1",
// 			},
// 			setup: func(req *v1.ListFailureReasonRequest) {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepository := dbmock.NewMockDps(mockCtrl)
// 				rep = mockRepository
// 				qYear, qMon, qDay := time.Now().Add(time.Hour * 24 * -(30)).Date()
// 				mockRepository.EXPECT().GetFailureReasons(ctx, db.GetFailureReasonsParams{
// 					Year:  int32(qYear),
// 					Month: int32(qMon),
// 					Day:   int32(qDay),
// 					Data:  json.RawMessage(fmt.Sprintf("%s", req.GetScope()))}).Times(1).Return([]db.GetFailureReasonsRow{}, nil)

// 			},
// 			output: &v1.ListFailureReasonResponse{
// 				FailureReasons: map[string]float32{},
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Context not found ",
// 			ctx:  context.Background(),
// 			input: &v1.ListFailureReasonRequest{
// 				Scope: "Scope1",
// 			},
// 			setup:   func(req *v1.ListFailureReasonRequest) {},
// 			wantErr: true,
// 		},
// 		{
// 			name: "scope out of context ",
// 			ctx:  context.Background(),
// 			input: &v1.ListFailureReasonRequest{
// 				Scope: "Scope5",
// 			},
// 			setup:   func(req *v1.ListFailureReasonRequest) {},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.setup(tt.input)
// 			obj := NewDpsServiceServer(rep, &queue, nil)
// 			got, err := obj.ListFailureReasonsRatio(tt.ctx, tt.input)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("dpsServiceServer.ListFailureReasonRequest() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.output) {
// 				t.Errorf("dpsServiceServer.ListFailureReasonRequest() = %v, want %v", got, tt.output)
// 			}
// 			log.Println("Test Passed : ", tt.name)
// 		})
// 	}
// }

// func Test_DashboardDataFailureRate(t *testing.T) {
// 	var mockCtrl *gomock.Controller
// 	var rep repo.Dps
// 	var queue workerqueue.Queue
// 	tests := []struct {
// 		name    string
// 		ctx     context.Context
// 		input   *v1.DataFailureRateRequest
// 		setup   func(*v1.DataFailureRateRequest)
// 		output  *v1.DataFailureRateResponse
// 		wantErr bool
// 	}{

// 		{
// 			name: "total And Failed Record Present",
// 			ctx:  ctx,
// 			input: &v1.DataFailureRateRequest{
// 				Scope: "Scope1",
// 			},
// 			setup: func(req *v1.DataFailureRateRequest) {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepository := dbmock.NewMockDps(mockCtrl)
// 				rep = mockRepository
// 				prevYear, PrevMon, prevDay := time.Now().Add(time.Hour * 24 * -(30)).Date()
// 				mockRepository.EXPECT().GetDataFileRecords(ctx, db.GetDataFileRecordsParams{
// 					Year:          int32(prevYear),
// 					Month:         int32(PrevMon),
// 					Day:           int32(prevDay),
// 					Scope:         req.Scope,
// 					SimilarEscape: fmt.Sprintf("%s_(applications|products|instance|products_acquiredRights|equipment%%)%%.csv", req.Scope),
// 				}).Times(1).Return(db.GetDataFileRecordsRow{
// 					TotalRecords:  int64(100),
// 					FailedRecords: int64(20),
// 				}, nil)

// 			},
// 			output: &v1.DataFailureRateResponse{
// 				FailureRate: float32(20),
// 			},
// 		},
// 		{
// 			name: "total And Failed Record Are Zero",
// 			ctx:  ctx,
// 			input: &v1.DataFailureRateRequest{
// 				Scope: "Scope1",
// 			},
// 			setup: func(req *v1.DataFailureRateRequest) {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepository := dbmock.NewMockDps(mockCtrl)
// 				rep = mockRepository
// 				prevYear, PrevMon, prevDay := time.Now().Add(time.Hour * 24 * -(30)).Date()
// 				mockRepository.EXPECT().GetDataFileRecords(ctx, db.GetDataFileRecordsParams{
// 					Year:          int32(prevYear),
// 					Month:         int32(PrevMon),
// 					Day:           int32(prevDay),
// 					Scope:         req.Scope,
// 					SimilarEscape: fmt.Sprintf("%s_(applications|products|instance|products_acquiredRights|equipment%%)%%.csv", req.Scope),
// 				}).Times(1).Return(db.GetDataFileRecordsRow{
// 					TotalRecords:  int64(0),
// 					FailedRecords: int64(0),
// 				}, nil)

// 			},
// 			output: &v1.DataFailureRateResponse{
// 				FailureRate: float32(0),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Failed Records Are Zero",
// 			ctx:  ctx,
// 			input: &v1.DataFailureRateRequest{
// 				Scope: "Scope1",
// 			},
// 			setup: func(req *v1.DataFailureRateRequest) {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepository := dbmock.NewMockDps(mockCtrl)
// 				rep = mockRepository
// 				prevYear, PrevMon, prevDay := time.Now().Add(time.Hour * 24 * -(30)).Date()
// 				mockRepository.EXPECT().GetDataFileRecords(ctx, db.GetDataFileRecordsParams{
// 					Year:          int32(prevYear),
// 					Month:         int32(PrevMon),
// 					Day:           int32(prevDay),
// 					Scope:         req.Scope,
// 					SimilarEscape: fmt.Sprintf("%s_(applications|products|instance|products_acquiredRights|equipment%%)%%.csv", req.Scope),
// 				}).Times(1).Return(db.GetDataFileRecordsRow{
// 					TotalRecords:  int64(100),
// 					FailedRecords: int64(0),
// 				}, nil)

// 			},
// 			output: &v1.DataFailureRateResponse{
// 				FailureRate: float32(0),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Context not found ",
// 			ctx:  context.Background(),
// 			input: &v1.DataFailureRateRequest{
// 				Scope: "Scope1",
// 			},
// 			setup:   func(req *v1.DataFailureRateRequest) {},
// 			wantErr: true,
// 		},
// 		{
// 			name: "scope out of context ",
// 			ctx:  context.Background(),
// 			input: &v1.DataFailureRateRequest{
// 				Scope: "Scope5",
// 			},
// 			setup:   func(req *v1.DataFailureRateRequest) {},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.setup(tt.input)
// 			obj := NewDpsServiceServer(rep, &queue, nil)
// 			got, err := obj.DashboardDataFailureRate(tt.ctx, tt.input)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("dpsServiceServer.DashboardDataFailureRate() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.output) {
// 				t.Errorf("dpsServiceServer.DashboardDataFailureRate() got = %v, want = %v", got, tt.output)
// 			}
// 			log.Println("Test Passed : ", tt.name)
// 		})
// 	}
// }
