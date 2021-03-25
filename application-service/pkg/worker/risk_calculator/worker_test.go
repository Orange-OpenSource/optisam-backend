// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package riskcalculator

import (
	"context"
	"database/sql"
	"errors"
	repo "optisam-backend/application-service/pkg/repository/v1"
	dbmock "optisam-backend/application-service/pkg/repository/v1/dbmock"
	"optisam-backend/application-service/pkg/repository/v1/postgres/db"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue/job"
	pro_v1 "optisam-backend/product-service/pkg/api/v1"
	mockacqs "optisam-backend/product-service/pkg/api/v1/mock"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRiskCalWorker_DoWork(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var productClient pro_v1.ProductServiceClient
	var rep repo.Application
	type args struct {
		ctx context.Context
		j   *job.Job
	}
	tests := []struct {
		name    string
		w       *RiskCalWorker
		args    args
		setup   func()
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				rep = mockRepo
				mockProductClient := mockacqs.NewMockProductServiceClient(mockCtrl)
				productClient = mockProductClient
				mockRepo.EXPECT().GetApplicationsDetails(ctx).Times(1).Return([]db.GetApplicationsDetailsRow{
					{
						ApplicationID:     "1",
						ApplicationDomain: "Payments",
						Scope:             "OFR",
					},
				}, nil)
				gomock.InOrder(
					mockRepo.EXPECT().GetDomainCriticityByDomain(ctx, db.GetDomainCriticityByDomainParams{
						Applicationdomain: "Payments",
						Scope:             "OFR",
					}).Times(1).Return(db.DomainCriticityMetum{
						DomainCriticID:   1,
						DomainCriticName: "Critic",
					}, nil),
					mockRepo.EXPECT().GetApplicationInstances(ctx, db.GetApplicationInstancesParams{
						ApplicationID: "1",
						Scope:         "OFR",
					}).Times(1).Return([]db.ApplicationsInstance{
						{
							ApplicationID: "1",
							InstanceID:    "Ins1",
							Products:      []string{"P1", "P2"},
						},
						{
							ApplicationID: "1",
							InstanceID:    "Ins2",
							Products:      []string{"P3", "P4"},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P1",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq1",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 2, 0)),
							},
							{
								SKU:              "acq2",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 1, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P2",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq3",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 3, 0)),
							},
							{
								SKU:              "acq4",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 4, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P3",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq5",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 5, 0)),
							},
							{
								SKU:              "acq6",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 6, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P4",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq7",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 7, 0)),
							},
							{
								SKU:              "acq8",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 8, 0)),
							},
						},
					}, nil),
					mockRepo.EXPECT().GetMaintenanceLevelByMonth(ctx, db.GetMaintenanceLevelByMonthParams{
						Calmonth: int32(1),
						Scope:    "OFR",
					}).Times(1).Return(db.MaintenanceLevelMetum{
						MaintenanceLevelID:   4,
						MaintenanceLevelName: "Level 4",
					}, nil),
					mockRepo.EXPECT().GetObsolescenceRiskForApplication(ctx, db.GetObsolescenceRiskForApplicationParams{
						Domaincriticid:     1,
						Maintenancelevelid: 4,
						Scope:              "OFR",
					}).Times(1).Return("Risky", nil),
					mockRepo.EXPECT().AddApplicationbsolescenceRisk(ctx, db.AddApplicationbsolescenceRiskParams{
						Riskvalue:     sql.NullString{String: "Risky", Valid: true},
						Applicationid: "1",
						Scope:         "OFR",
					}).Times(1).Return(nil),
				)
			},
		},
		{name: "SUCCESS - product 2 has less value",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				rep = mockRepo
				mockProductClient := mockacqs.NewMockProductServiceClient(mockCtrl)
				productClient = mockProductClient
				mockRepo.EXPECT().GetApplicationsDetails(ctx).Times(1).Return([]db.GetApplicationsDetailsRow{
					{
						ApplicationID:     "1",
						ApplicationDomain: "Payments",
						Scope:             "OFR",
					},
				}, nil)
				gomock.InOrder(
					mockRepo.EXPECT().GetDomainCriticityByDomain(ctx, db.GetDomainCriticityByDomainParams{
						Applicationdomain: "Payments",
						Scope:             "OFR",
					}).Times(1).Return(db.DomainCriticityMetum{
						DomainCriticID:   1,
						DomainCriticName: "Critic",
					}, nil),
					mockRepo.EXPECT().GetApplicationInstances(ctx, db.GetApplicationInstancesParams{
						ApplicationID: "1",
						Scope:         "OFR",
					}).Times(1).Return([]db.ApplicationsInstance{
						{
							ApplicationID: "1",
							InstanceID:    "Ins1",
							Products:      []string{"P1", "P2"},
						},
						{
							ApplicationID: "1",
							InstanceID:    "Ins2",
							Products:      []string{"P3", "P4"},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P1",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq1",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 2, 0)),
							},
							{
								SKU:              "acq2",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 3, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P2",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq3",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 1, 0)),
							},
							{
								SKU:              "acq4",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 4, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P3",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq5",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 5, 0)),
							},
							{
								SKU:              "acq6",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 6, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P4",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq7",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 7, 0)),
							},
							{
								SKU:              "acq8",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 8, 0)),
							},
						},
					}, nil),
					mockRepo.EXPECT().GetMaintenanceLevelByMonth(ctx, db.GetMaintenanceLevelByMonthParams{
						Calmonth: int32(1),
						Scope:    "OFR",
					}).Times(1).Return(db.MaintenanceLevelMetum{
						MaintenanceLevelID:   4,
						MaintenanceLevelName: "Level 4",
					}, nil),
					mockRepo.EXPECT().GetObsolescenceRiskForApplication(ctx, db.GetObsolescenceRiskForApplicationParams{
						Domaincriticid:     1,
						Maintenancelevelid: 4,
						Scope:              "OFR",
					}).Times(1).Return("Risky", nil),
					mockRepo.EXPECT().AddApplicationbsolescenceRisk(ctx, db.AddApplicationbsolescenceRiskParams{
						Riskvalue:     sql.NullString{String: "Risky", Valid: true},
						Applicationid: "1",
						Scope:         "OFR",
					}).Times(1).Return(nil),
				)
			},
		},
		{name: "SUCCESS - min month <= 0",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				rep = mockRepo
				mockProductClient := mockacqs.NewMockProductServiceClient(mockCtrl)
				productClient = mockProductClient
				mockRepo.EXPECT().GetApplicationsDetails(ctx).Times(1).Return([]db.GetApplicationsDetailsRow{
					{
						ApplicationID:     "1",
						ApplicationDomain: "Payments",
						Scope:             "OFR",
					},
				}, nil)
				gomock.InOrder(
					mockRepo.EXPECT().GetDomainCriticityByDomain(ctx, db.GetDomainCriticityByDomainParams{
						Applicationdomain: "Payments",
						Scope:             "OFR",
					}).Times(1).Return(db.DomainCriticityMetum{
						DomainCriticID:   1,
						DomainCriticName: "Critic",
					}, nil),
					mockRepo.EXPECT().GetApplicationInstances(ctx, db.GetApplicationInstancesParams{
						ApplicationID: "1",
						Scope:         "OFR",
					}).Times(1).Return([]db.ApplicationsInstance{
						{
							ApplicationID: "1",
							InstanceID:    "Ins1",
							Products:      []string{"P1", "P2"},
						},
						{
							ApplicationID: "1",
							InstanceID:    "Ins2",
							Products:      []string{"P3", "P4"},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P1",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq1",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 2, 0)),
							},
							{
								SKU:              "acq2",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now()),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P2",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						PageNum:   1,
						Scopes:    []string{"OFR"},
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq3",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 1, 0)),
							},
							{
								SKU:              "acq4",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 4, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P3",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq5",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 5, 0)),
							},
							{
								SKU:              "acq6",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 6, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P4",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq7",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 7, 0)),
							},
							{
								SKU:              "acq8",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 8, 0)),
							},
						},
					}, nil),
					mockRepo.EXPECT().GetMaintenanceLevelByMonthByName(ctx, "Level 4").Times(1).Return(db.MaintenanceLevelMetum{
						MaintenanceLevelID:   4,
						MaintenanceLevelName: "Level 4",
					}, nil),
					mockRepo.EXPECT().GetObsolescenceRiskForApplication(ctx, db.GetObsolescenceRiskForApplicationParams{
						Domaincriticid:     1,
						Maintenancelevelid: 4,
						Scope:              "OFR",
					}).Times(1).Return("Risky", nil),
					mockRepo.EXPECT().AddApplicationbsolescenceRisk(ctx, db.AddApplicationbsolescenceRiskParams{
						Riskvalue:     sql.NullString{String: "Risky", Valid: true},
						Applicationid: "1",
						Scope:         "OFR",
					}).Times(1).Return(nil),
				)
			},
		},
		{name: "FAILURE - GetApplicationsView - DBError",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				rep = mockRepo
				mockProductClient := mockacqs.NewMockProductServiceClient(mockCtrl)
				productClient = mockProductClient
				mockRepo.EXPECT().GetApplicationsDetails(ctx).Times(1).Return(nil, errors.New("DBError"))
			},
			wantErr: true,
		},
		{name: "FAILURE - GetDomainCriticityByDomain - DBError",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				rep = mockRepo
				mockProductClient := mockacqs.NewMockProductServiceClient(mockCtrl)
				productClient = mockProductClient
				mockRepo.EXPECT().GetApplicationsDetails(ctx).Times(1).Return([]db.GetApplicationsDetailsRow{
					{
						ApplicationID:     "1",
						ApplicationDomain: "Payments",
						Scope:             "OFR",
					},
				}, nil)
				gomock.InOrder(
					mockRepo.EXPECT().GetDomainCriticityByDomain(ctx, db.GetDomainCriticityByDomainParams{
						Applicationdomain: "Payments",
						Scope:             "OFR",
					}).Times(1).Return(db.DomainCriticityMetum{}, errors.New("DBError")),
				)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetApplicationInstances - DBError",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				rep = mockRepo
				mockProductClient := mockacqs.NewMockProductServiceClient(mockCtrl)
				productClient = mockProductClient
				mockRepo.EXPECT().GetApplicationsDetails(ctx).Times(1).Return([]db.GetApplicationsDetailsRow{
					{
						ApplicationID:     "1",
						ApplicationDomain: "Payments",
						Scope:             "OFR",
					},
				}, nil)
				gomock.InOrder(
					mockRepo.EXPECT().GetDomainCriticityByDomain(ctx, db.GetDomainCriticityByDomainParams{
						Applicationdomain: "Payments",
						Scope:             "OFR",
					}).Times(1).Return(db.DomainCriticityMetum{
						DomainCriticID:   1,
						DomainCriticName: "Critic",
					}, nil),
					mockRepo.EXPECT().GetApplicationInstances(ctx, db.GetApplicationInstancesParams{
						ApplicationID: "1",
						Scope:         "OFR",
					}).Times(1).Return(nil, errors.New("DBError")),
				)
			},
			wantErr: true,
		},
		{name: "FAILURE - ListAcqRights - can not fetch list acqrights",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				rep = mockRepo
				mockProductClient := mockacqs.NewMockProductServiceClient(mockCtrl)
				productClient = mockProductClient
				mockRepo.EXPECT().GetApplicationsDetails(ctx).Times(1).Return([]db.GetApplicationsDetailsRow{
					{
						ApplicationID:     "1",
						ApplicationDomain: "Payments",
						Scope:             "OFR",
					},
				}, nil)
				gomock.InOrder(
					mockRepo.EXPECT().GetDomainCriticityByDomain(ctx, db.GetDomainCriticityByDomainParams{
						Applicationdomain: "Payments",
						Scope:             "OFR",
					}).Times(1).Return(db.DomainCriticityMetum{
						DomainCriticID:   1,
						DomainCriticName: "Critic",
					}, nil),
					mockRepo.EXPECT().GetApplicationInstances(ctx, db.GetApplicationInstancesParams{
						ApplicationID: "1",
						Scope:         "OFR",
					}).Times(1).Return([]db.ApplicationsInstance{
						{
							ApplicationID: "1",
							InstanceID:    "Ins1",
							Products:      []string{"P1", "P2"},
						},
						{
							ApplicationID: "1",
							InstanceID:    "Ins2",
							Products:      []string{"P3", "P4"},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P1",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(nil, errors.New("can not fetch list acqrights")),
				)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetMaintenanceLevelByMonthByName - DBError",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				rep = mockRepo
				mockProductClient := mockacqs.NewMockProductServiceClient(mockCtrl)
				productClient = mockProductClient
				mockRepo.EXPECT().GetApplicationsDetails(ctx).Times(1).Return([]db.GetApplicationsDetailsRow{
					{
						ApplicationID:     "1",
						ApplicationDomain: "Payments",
						Scope:             "OFR",
					},
				}, nil)
				gomock.InOrder(
					mockRepo.EXPECT().GetDomainCriticityByDomain(ctx, db.GetDomainCriticityByDomainParams{
						Applicationdomain: "Payments",
						Scope:             "OFR",
					}).Times(1).Return(db.DomainCriticityMetum{
						DomainCriticID:   1,
						DomainCriticName: "Critic",
					}, nil),
					mockRepo.EXPECT().GetApplicationInstances(ctx, db.GetApplicationInstancesParams{
						ApplicationID: "1",
						Scope:         "OFR",
					}).Times(1).Return([]db.ApplicationsInstance{
						{
							ApplicationID: "1",
							InstanceID:    "Ins1",
							Products:      []string{"P1", "P2"},
						},
						{
							ApplicationID: "1",
							InstanceID:    "Ins2",
							Products:      []string{"P3", "P4"},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P1",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq1",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 2, 0)),
							},
							{
								SKU:              "acq2",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now()),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P2",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq3",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 1, 0)),
							},
							{
								SKU:              "acq4",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 4, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P3",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq5",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 5, 0)),
							},
							{
								SKU:              "acq6",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 6, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P4",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq7",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 7, 0)),
							},
							{
								SKU:              "acq8",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 8, 0)),
							},
						},
					}, nil),
					mockRepo.EXPECT().GetMaintenanceLevelByMonthByName(ctx, "Level 4").Times(1).Return(db.MaintenanceLevelMetum{}, errors.New("DBError")),
				)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetMaintenanceLevelByMonth - DBError",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				rep = mockRepo
				mockProductClient := mockacqs.NewMockProductServiceClient(mockCtrl)
				productClient = mockProductClient
				mockRepo.EXPECT().GetApplicationsDetails(ctx).Times(1).Return([]db.GetApplicationsDetailsRow{
					{
						ApplicationID:     "1",
						ApplicationDomain: "Payments",
						Scope:             "OFR",
					},
				}, nil)
				gomock.InOrder(
					mockRepo.EXPECT().GetDomainCriticityByDomain(ctx, db.GetDomainCriticityByDomainParams{
						Applicationdomain: "Payments",
						Scope:             "OFR",
					}).Times(1).Return(db.DomainCriticityMetum{
						DomainCriticID:   1,
						DomainCriticName: "Critic",
					}, nil),
					mockRepo.EXPECT().GetApplicationInstances(ctx, db.GetApplicationInstancesParams{
						ApplicationID: "1",
						Scope:         "OFR",
					}).Times(1).Return([]db.ApplicationsInstance{
						{
							ApplicationID: "1",
							InstanceID:    "Ins1",
							Products:      []string{"P1", "P2"},
						},
						{
							ApplicationID: "1",
							InstanceID:    "Ins2",
							Products:      []string{"P3", "P4"},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P1",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq1",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 1, 0)),
							},
							{
								SKU:              "acq2",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 2, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P2",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq3",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 3, 0)),
							},
							{
								SKU:              "acq4",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 4, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P3",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq5",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 5, 0)),
							},
							{
								SKU:              "acq6",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 6, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P4",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq7",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 7, 0)),
							},
							{
								SKU:              "acq8",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 8, 0)),
							},
						},
					}, nil),
					mockRepo.EXPECT().GetMaintenanceLevelByMonth(ctx, db.GetMaintenanceLevelByMonthParams{
						Calmonth: int32(1),
						Scope:    "OFR",
					}).Times(1).Return(db.MaintenanceLevelMetum{}, errors.New("DBError")),
				)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetObsolescenceRiskForApplication - DBError",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				rep = mockRepo
				mockProductClient := mockacqs.NewMockProductServiceClient(mockCtrl)
				productClient = mockProductClient
				mockRepo.EXPECT().GetApplicationsDetails(ctx).Times(1).Return([]db.GetApplicationsDetailsRow{
					{
						ApplicationID:     "1",
						ApplicationDomain: "Payments",
						Scope:             "OFR",
					},
				}, nil)
				gomock.InOrder(
					mockRepo.EXPECT().GetDomainCriticityByDomain(ctx, db.GetDomainCriticityByDomainParams{
						Applicationdomain: "Payments",
						Scope:             "OFR",
					}).Times(1).Return(db.DomainCriticityMetum{
						DomainCriticID:   1,
						DomainCriticName: "Critic",
					}, nil),
					mockRepo.EXPECT().GetApplicationInstances(ctx, db.GetApplicationInstancesParams{
						ApplicationID: "1",
						Scope:         "OFR",
					}).Times(1).Return([]db.ApplicationsInstance{
						{
							ApplicationID: "1",
							InstanceID:    "Ins1",
							Products:      []string{"P1", "P2"},
						},
						{
							ApplicationID: "1",
							InstanceID:    "Ins2",
							Products:      []string{"P3", "P4"},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P1",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq1",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 1, 0)),
							},
							{
								SKU:              "acq2",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 2, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P2",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq3",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 3, 0)),
							},
							{
								SKU:              "acq4",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 4, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P3",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq5",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 5, 0)),
							},
							{
								SKU:              "acq6",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 6, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P4",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq7",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 7, 0)),
							},
							{
								SKU:              "acq8",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 8, 0)),
							},
						},
					}, nil),
					mockRepo.EXPECT().GetMaintenanceLevelByMonth(ctx, db.GetMaintenanceLevelByMonthParams{
						Calmonth: int32(1),
						Scope:    "OFR",
					}).Times(1).Return(db.MaintenanceLevelMetum{
						MaintenanceLevelID:   4,
						MaintenanceLevelName: "Level 4",
					}, nil),
					mockRepo.EXPECT().GetObsolescenceRiskForApplication(ctx, db.GetObsolescenceRiskForApplicationParams{
						Domaincriticid:     1,
						Maintenancelevelid: 4,
						Scope:              "OFR",
					}).Times(1).Return("", errors.New("DBError")),
				)
			},
			wantErr: true,
		},
		{name: "FAILURE - AddApplicationbsolescenceRisk - DBError",
			args: args{
				ctx: ctx,
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				rep = mockRepo
				mockProductClient := mockacqs.NewMockProductServiceClient(mockCtrl)
				productClient = mockProductClient
				mockRepo.EXPECT().GetApplicationsDetails(ctx).Times(1).Return([]db.GetApplicationsDetailsRow{
					{
						ApplicationID:     "1",
						ApplicationDomain: "Payments",
						Scope:             "OFR",
					},
				}, nil)
				gomock.InOrder(
					mockRepo.EXPECT().GetDomainCriticityByDomain(ctx, db.GetDomainCriticityByDomainParams{
						Applicationdomain: "Payments",
						Scope:             "OFR",
					}).Times(1).Return(db.DomainCriticityMetum{
						DomainCriticID:   1,
						DomainCriticName: "Critic",
					}, nil),
					mockRepo.EXPECT().GetApplicationInstances(ctx, db.GetApplicationInstancesParams{
						ApplicationID: "1",
						Scope:         "OFR",
					}).Times(1).Return([]db.ApplicationsInstance{
						{
							ApplicationID: "1",
							InstanceID:    "Ins1",
							Products:      []string{"P1", "P2"},
						},
						{
							ApplicationID: "1",
							InstanceID:    "Ins2",
							Products:      []string{"P3", "P4"},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P1",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq1",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 1, 0)),
							},
							{
								SKU:              "acq2",
								ProductName:      "P1",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 2, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P2",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq3",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 3, 0)),
							},
							{
								SKU:              "acq4",
								ProductName:      "P2",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 4, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P3",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq5",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 5, 0)),
							},
							{
								SKU:              "acq6",
								ProductName:      "P3",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 6, 0)),
							},
						},
					}, nil),
					mockProductClient.EXPECT().ListAcqRights(ctx, &pro_v1.ListAcqRightsRequest{
						SearchParams: &pro_v1.AcqRightsSearchParams{
							SwidTag: &pro_v1.StringFilter{
								Filteringkey: "P4",
								FilterType:   true,
							},
						},
						SortBy:    pro_v1.ListAcqRightsRequest_END_OF_MAINTENANCE,
						SortOrder: pro_v1.SortOrder_asc,
						Scopes:    []string{"OFR"},
						PageNum:   1,
						PageSize:  50,
					}).Times(1).Return(&pro_v1.ListAcqRightsResponse{
						TotalRecords: 2,
						AcquiredRights: []*pro_v1.AcqRights{
							{
								SKU:              "acq7",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 7, 0)),
							},
							{
								SKU:              "acq8",
								ProductName:      "P4",
								EndOfMaintenance: timestamppb.New(time.Now().AddDate(0, 8, 0)),
							},
						},
					}, nil),
					mockRepo.EXPECT().GetMaintenanceLevelByMonth(ctx, db.GetMaintenanceLevelByMonthParams{
						Calmonth: int32(1),
						Scope:    "OFR",
					}).Times(1).Return(db.MaintenanceLevelMetum{
						MaintenanceLevelID:   4,
						MaintenanceLevelName: "Level 4",
					}, nil),
					mockRepo.EXPECT().GetObsolescenceRiskForApplication(ctx, db.GetObsolescenceRiskForApplicationParams{
						Domaincriticid:     1,
						Maintenancelevelid: 4,
						Scope:              "OFR",
					}).Times(1).Return("Risky", nil),
					mockRepo.EXPECT().AddApplicationbsolescenceRisk(ctx, db.AddApplicationbsolescenceRiskParams{
						Riskvalue:     sql.NullString{String: "Risky", Valid: true},
						Applicationid: "1",
						Scope:         "OFR",
					}).Times(1).Return(errors.New("DBError")),
				)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			w := &RiskCalWorker{
				id:              "ob",
				applicationRepo: rep,
				productClient:   productClient,
			}
			if err := w.DoWork(tt.args.ctx, tt.args.j); (err != nil) != tt.wantErr {
				t.Errorf("RiskCalWorker.DoWork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
