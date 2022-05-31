package v1

// func Test_licenseServiceServer_ListAcqRightsForApplicationsProduct(t *testing.T) {
// 	t.Skip()
// 	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
// 		UserID: "admin@superuser.com",
// 		Role:   "Admin",
// 		Socpes: []string{"A"},
// 	})
// 	var mockCtrl *gomock.Controller
// 	var rep repo.License
// 	type args struct {
// 		ctx context.Context
// 		req *v1.ListAcqRightsForApplicationsProductRequest
// 	}
// 	tests := []struct {
// 		name    string
// 		s       *licenseServiceServer
// 		args    args
// 		setup   func()
// 		want    *v1.ListAcqRightsForApplicationsProductResponse
// 		wantErr bool
// 	}{
// 		{
// 			name: "SUCCESS",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.ListAcqRightsForApplicationsProductRequest{
// 					AppId:  "a1",
// 					ProdId: "p1",
// 					Scope:  "A",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockLicense(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().ProductExistsForApplication(ctx, "p1", "a1", []string{"A"}).Times(1).Return(true, nil)
// 				mockRepo.EXPECT().ProductAcquiredRights(ctx, "p1", false, []string{"A"}).Times(1).Return("p1", []*repo.ProductAcquiredRight{
// 					{
// 						SKU:          "s1",
// 						Metric:       "OPS",
// 						AcqLicenses:  5,
// 						TotalCost:    20,
// 						AvgUnitPrice: 4,
// 					},
// 				}, nil)
// 				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
// 					{
// 						Name: "OPS",
// 						Type: repo.MetricOPSOracleProcessorStandard,
// 					},
// 					{
// 						Name: "WS",
// 						Type: repo.MetricOPSOracleProcessorStandard,
// 					},
// 				}, nil)

// 				cores := &repo.Attribute{
// 					ID:   "cores",
// 					Type: repo.DataTypeInt,
// 				}
// 				cpu := &repo.Attribute{
// 					ID:   "cpus",
// 					Type: repo.DataTypeInt,
// 				}
// 				corefactor := &repo.Attribute{
// 					ID:   "corefactor",
// 					Type: repo.DataTypeInt,
// 				}

// 				base := &repo.EquipmentType{
// 					ID:         "e2",
// 					ParentID:   "e3",
// 					Type:       "server",
// 					Attributes: []*repo.Attribute{cores, cpu, corefactor},
// 				}
// 				start := &repo.EquipmentType{
// 					ID:       "e1",
// 					Type:     "partition",
// 					ParentID: "e2",
// 				}
// 				agg := &repo.EquipmentType{
// 					ID:       "e3",
// 					Type:     "cluster",
// 					ParentID: "e4",
// 				}
// 				end := &repo.EquipmentType{
// 					ID:       "e4",
// 					Type:     "vcenter",
// 					ParentID: "e5",
// 				}
// 				endP := &repo.EquipmentType{
// 					ID:   "e5",
// 					Type: "datacenter",
// 				}

// 				mat := &repo.MetricOPSComputed{
// 					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
// 					BaseType:       base,
// 					AggregateLevel: agg,
// 					NumCoresAttr:   cores,
// 					NumCPUAttr:     cpu,
// 					CoreFactorAttr: corefactor,
// 				}

// 				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
// 				mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).Times(1).Return(
// 					[]*repo.Equipment{
// 						{
// 							ID:      "ue1",
// 							EquipID: "ee1",
// 							Type:    "partition",
// 						},
// 					}, nil)

// 				mockRepo.EXPECT().ListMetricOPS(ctx, "A").Times(1).Return([]*repo.MetricOPS{
// 					{
// 						Name:                  "OPS",
// 						NumCoreAttrID:         "cores",
// 						NumCPUAttrID:          "cpus",
// 						CoreFactorAttrID:      "corefactor",
// 						BaseEqTypeID:          "e2",
// 						AggerateLevelEqTypeID: "e3",
// 						StartEqTypeID:         "e1",
// 						EndEqTypeID:           "e4",
// 					},
// 				}, nil)

// 				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, "p1", mat, []string{"A"}).Times(1).Return(uint64(8), nil)

// 			},
// 			want: &v1.ListAcqRightsForApplicationsProductResponse{
// 				AcqRights: []*v1.ProductAcquiredRights{
// 					{
// 						SKU:            "s1",
// 						SwidTag:        "p1",
// 						Metric:         "OPS",
// 						NumCptLicences: 8,
// 						NumAcqLicences: 5,
// 						TotalCost:      20,
// 						DeltaNumber:    -3,
// 						DeltaCost:      -12,
// 						AvgUnitPrice:   4,
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "FAILURE: Can not find user claims",
// 			args: args{
// 				ctx: context.Background(),
// 				req: &v1.ListAcqRightsForApplicationsProductRequest{
// 					AppId:  "a1",
// 					ProdId: "p1",
// 				},
// 			},
// 			setup:   func() {},
// 			wantErr: true,
// 		},
// 		{
// 			name: "FAILURE: Error in repo/ProductExistsForApplication",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.ListAcqRightsForApplicationsProductRequest{
// 					AppId:  "a1",
// 					ProdId: "p1",
// 					Scope:  "A",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockLicense(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().ProductExistsForApplication(ctx, "p1", "a1", []string{"A"}).Times(1).Return(false, errors.New("Internal Error"))
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "FAILURE: Product is not linked with application",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.ListAcqRightsForApplicationsProductRequest{
// 					AppId:  "a1",
// 					ProdId: "p1",
// 					Scope:  "A",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockLicense(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().ProductExistsForApplication(ctx, "p1", "a1", []string{"A"}).Times(1).Return(false, nil)

// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "FAILURE: error in db/ProductAcquiredRights",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.ListAcqRightsForApplicationsProductRequest{
// 					AppId:  "a1",
// 					ProdId: "p1",
// 					Scope:  "A",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockLicense(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().ProductExistsForApplication(ctx, "p1", "a1", []string{"A"}).Times(1).Return(true, nil)
// 				mockRepo.EXPECT().ProductAcquiredRights(ctx, "p1", false, []string{"A"}).Times(1).Return("", nil, errors.New("Internal Error"))
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "FAILURE: Error in db/ListMetrices",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.ListAcqRightsForApplicationsProductRequest{
// 					AppId:  "a1",
// 					ProdId: "p1",
// 					Scope:  "A",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockLicense(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().ProductExistsForApplication(ctx, "p1", "a1", []string{"A"}).Times(1).Return(true, nil)
// 				mockRepo.EXPECT().ProductAcquiredRights(ctx, "p1", []string{"A"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
// 					{
// 						SKU:          "s1",
// 						Metric:       "OPS",
// 						AcqLicenses:  5,
// 						TotalCost:    20,
// 						AvgUnitPrice: 4,
// 					},
// 					{
// 						SKU:          "s2",
// 						Metric:       "WS",
// 						AcqLicenses:  10,
// 						TotalCost:    50,
// 						AvgUnitPrice: 5,
// 					},
// 				}, nil)
// 				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return(nil, errors.New("Internal Error"))
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "FAILURE: No metric type exists in system",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.ListAcqRightsForApplicationsProductRequest{
// 					AppId:  "a1",
// 					ProdId: "p1",
// 					Scope:  "A",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockLicense(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().ProductExistsForApplication(ctx, "p1", "a1", []string{"A"}).Times(1).Return(true, nil)
// 				mockRepo.EXPECT().ProductAcquiredRights(ctx, "p1", []string{"A"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
// 					{
// 						SKU:          "s1",
// 						Metric:       "OPS",
// 						AcqLicenses:  5,
// 						TotalCost:    20,
// 						AvgUnitPrice: 4,
// 					},
// 					{
// 						SKU:          "s2",
// 						Metric:       "WS",
// 						AcqLicenses:  10,
// 						TotalCost:    50,
// 						AvgUnitPrice: 5,
// 					},
// 				}, nil)
// 				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return(nil, repo.ErrNoData)
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "FAILURE: Error in db/EquipmentTypes",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.ListAcqRightsForApplicationsProductRequest{
// 					AppId:  "a1",
// 					ProdId: "p1",
// 					Scope:  "A",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockLicense(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().ProductExistsForApplication(ctx, "p1", "a1", []string{"A"}).Times(1).Return(true, nil)
// 				mockRepo.EXPECT().ProductAcquiredRights(ctx, "p1", []string{"A"}).Times(1).Return("p1", []*repo.ProductAcquiredRight{
// 					{
// 						SKU:          "s1",
// 						Metric:       "OPS",
// 						AcqLicenses:  5,
// 						TotalCost:    20,
// 						AvgUnitPrice: 4,
// 					},
// 					{
// 						SKU:          "s2",
// 						Metric:       "WS",
// 						AcqLicenses:  10,
// 						TotalCost:    50,
// 						AvgUnitPrice: 5,
// 					},
// 				}, nil)
// 				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
// 					{
// 						Name: "OPS",
// 						Type: repo.MetricOPSOracleProcessorStandard,
// 					},
// 					{
// 						Name: "WS",
// 						Type: repo.MetricOPSOracleProcessorStandard,
// 					},
// 				}, nil)

// 				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(nil, errors.New("Internal Error"))

// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "FAILURE: Error in db/ProductApplicationEquipments",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.ListAcqRightsForApplicationsProductRequest{
// 					AppId:  "a1",
// 					ProdId: "p1",
// 					Scope:  "A",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockLicense(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().ProductExistsForApplication(ctx, "p1", "a1", []string{"A"}).Times(1).Return(true, nil)
// 				mockRepo.EXPECT().ProductAcquiredRights(ctx, "p1", []string{"A"}).Times(1).Return("p1", []*repo.ProductAcquiredRight{
// 					{
// 						SKU:          "s1",
// 						Metric:       "OPS",
// 						AcqLicenses:  5,
// 						TotalCost:    20,
// 						AvgUnitPrice: 4,
// 					},
// 					{
// 						SKU:          "s2",
// 						Metric:       "WS",
// 						AcqLicenses:  10,
// 						TotalCost:    50,
// 						AvgUnitPrice: 5,
// 					},
// 				}, nil)
// 				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
// 					{
// 						Name: "OPS",
// 						Type: repo.MetricOPSOracleProcessorStandard,
// 					},
// 					{
// 						Name: "WS",
// 						Type: repo.MetricOPSOracleProcessorStandard,
// 					},
// 				}, nil)

// 				cores := &repo.Attribute{
// 					ID:   "cores",
// 					Type: repo.DataTypeInt,
// 				}
// 				cpu := &repo.Attribute{
// 					ID:   "cpus",
// 					Type: repo.DataTypeInt,
// 				}
// 				corefactor := &repo.Attribute{
// 					ID:   "corefactor",
// 					Type: repo.DataTypeInt,
// 				}

// 				base := &repo.EquipmentType{
// 					ID:         "e2",
// 					ParentID:   "e3",
// 					Type:       "server",
// 					Attributes: []*repo.Attribute{cores, cpu, corefactor},
// 				}
// 				start := &repo.EquipmentType{
// 					ID:       "e1",
// 					Type:     "partition",
// 					ParentID: "e2",
// 				}
// 				agg := &repo.EquipmentType{
// 					ID:       "e3",
// 					Type:     "cluster",
// 					ParentID: "e4",
// 				}
// 				end := &repo.EquipmentType{
// 					ID:       "e4",
// 					Type:     "vcenter",
// 					ParentID: "e5",
// 				}
// 				endP := &repo.EquipmentType{
// 					ID:   "e5",
// 					Type: "datacenter",
// 				}
// 				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
// 				mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).Times(1).Return(nil, errors.New("Internal Error"))

// 			},
// 			wantErr: true,
// 		},

// 		{
// 			name: "SUCCESS: No Equipment is linked with product and application",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.ListAcqRightsForApplicationsProductRequest{
// 					AppId:  "a1",
// 					ProdId: "p1",
// 					Scope:  "A",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockLicense(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().ProductExistsForApplication(ctx, "p1", "a1", []string{"A"}).Times(1).Return(true, nil)
// 				mockRepo.EXPECT().ProductAcquiredRights(ctx, "p1", []string{"A"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
// 					{
// 						SKU:          "s1",
// 						Metric:       "OPS",
// 						AcqLicenses:  5,
// 						TotalCost:    20,
// 						AvgUnitPrice: 4,
// 					},
// 					{
// 						SKU:          "s2",
// 						Metric:       "WS",
// 						AcqLicenses:  10,
// 						TotalCost:    50,
// 						AvgUnitPrice: 5,
// 					},
// 				}, nil)
// 				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
// 					{
// 						Name: "OPS",
// 						Type: repo.MetricOPSOracleProcessorStandard,
// 					},
// 					{
// 						Name: "WS",
// 						Type: repo.MetricOPSOracleProcessorStandard,
// 					},
// 				}, nil)

// 				cores := &repo.Attribute{
// 					ID:   "cores",
// 					Type: repo.DataTypeInt,
// 				}
// 				cpu := &repo.Attribute{
// 					ID:   "cpus",
// 					Type: repo.DataTypeInt,
// 				}
// 				corefactor := &repo.Attribute{
// 					ID:   "corefactor",
// 					Type: repo.DataTypeInt,
// 				}

// 				base := &repo.EquipmentType{
// 					ID:         "e2",
// 					ParentID:   "e3",
// 					Type:       "server",
// 					Attributes: []*repo.Attribute{cores, cpu, corefactor},
// 				}
// 				start := &repo.EquipmentType{
// 					ID:       "e1",
// 					Type:     "partition",
// 					ParentID: "e2",
// 				}
// 				agg := &repo.EquipmentType{
// 					ID:       "e3",
// 					Type:     "cluster",
// 					ParentID: "e4",
// 				}
// 				end := &repo.EquipmentType{
// 					ID:       "e4",
// 					Type:     "vcenter",
// 					ParentID: "e5",
// 				}
// 				endP := &repo.EquipmentType{
// 					ID:   "e5",
// 					Type: "datacenter",
// 				}

// 				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
// 				mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).Times(1).Return(nil, nil)
// 			},
// 			want: &v1.ListAcqRightsForApplicationsProductResponse{
// 				AcqRights: []*v1.ProductAcquiredRights{
// 					{
// 						SKU:            "s1",
// 						SwidTag:        "p1",
// 						Metric:         "OPS",
// 						NumAcqLicences: 5,
// 						TotalCost:      20,
// 						AvgUnitPrice:   4,
// 					},
// 					{
// 						SKU:            "s2",
// 						SwidTag:        "p1",
// 						Metric:         "WS",
// 						NumAcqLicences: 10,
// 						TotalCost:      50,
// 						AvgUnitPrice:   5,
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "FAILURE: Error in ListMetricOPS",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.ListAcqRightsForApplicationsProductRequest{
// 					AppId:  "a1",
// 					ProdId: "p1",
// 					Scope:  "A",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := mock.NewMockLicense(mockCtrl)
// 				rep = mockRepo
// 				mockRepo.EXPECT().ProductExistsForApplication(ctx, "p1", "a1", []string{"A"}).Times(1).Return(true, nil)
// 				mockRepo.EXPECT().ProductAcquiredRights(ctx, "p1", []string{"A"}).Times(1).Return("p1", []*repo.ProductAcquiredRight{
// 					{
// 						SKU:          "s1",
// 						Metric:       "OPS",
// 						AcqLicenses:  5,
// 						TotalCost:    20,
// 						AvgUnitPrice: 4,
// 					},
// 				}, nil)
// 				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
// 					{
// 						Name: "OPS",
// 						Type: repo.MetricOPSOracleProcessorStandard,
// 					},
// 				}, nil)

// 				cores := &repo.Attribute{
// 					ID:   "cores",
// 					Type: repo.DataTypeInt,
// 				}
// 				cpu := &repo.Attribute{
// 					ID:   "cpus",
// 					Type: repo.DataTypeInt,
// 				}
// 				corefactor := &repo.Attribute{
// 					ID:   "corefactor",
// 					Type: repo.DataTypeInt,
// 				}

// 				base := &repo.EquipmentType{
// 					ID:         "e2",
// 					ParentID:   "e3",
// 					Type:       "server",
// 					Attributes: []*repo.Attribute{cores, cpu, corefactor},
// 				}
// 				start := &repo.EquipmentType{
// 					ID:       "e1",
// 					Type:     "partition",
// 					ParentID: "e2",
// 				}
// 				agg := &repo.EquipmentType{
// 					ID:       "e3",
// 					Type:     "cluster",
// 					ParentID: "e4",
// 				}
// 				end := &repo.EquipmentType{
// 					ID:       "e4",
// 					Type:     "vcenter",
// 					ParentID: "e5",
// 				}
// 				endP := &repo.EquipmentType{
// 					ID:   "e5",
// 					Type: "datacenter",
// 				}

// 				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
// 				mockRepo.EXPECT().ProductApplicationEquipments(ctx, "p1", "a1", []string{"A"}).Times(1).Return(
// 					[]*repo.Equipment{
// 						{
// 							ID:      "ue1",
// 							EquipID: "ee1",
// 							Type:    "partition",
// 						},
// 					}, nil)
// 				mat := &repo.MetricOPSComputed{
// 					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
// 					BaseType:       base,
// 					AggregateLevel: agg,
// 					NumCoresAttr:   cores,
// 					NumCPUAttr:     cpu,
// 					CoreFactorAttr: corefactor,
// 				}

// 				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return(nil, errors.New("Internal Error"))
// 				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, "p1", "a1", mat, []string{"A"}).Times(1).Return(uint64(6), nil)

// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.setup()
// 			s := NewLicenseServiceServer(rep)
// 			got, err := s.ListAcqRightsForApplicationsProduct(tt.args.ctx, tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("licenseServiceServer.ListAcqRightsForApplicationsProduct() error = %v, wantErr %v , failed test case %s", err, tt.wantErr, tt.name)
// 			} else if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("licenseServiceServer.ListAcqRightsForApplicationsProduct() = %v, want %v", got, tt.want)
// 			} else {
// 				fmt.Println("Test case passed  : [", tt.name, "]")
// 			}
// 		})
// 	}
// }
