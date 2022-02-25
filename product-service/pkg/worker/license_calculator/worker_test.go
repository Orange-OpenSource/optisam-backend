package licensecalculator

/*
func TestLicenseCalWorker_DoWork(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	cTime := time.Now()
	parser := cron.NewParser(cron.Descriptor)
	temp, _ := parser.Parse("@every 12h")
	nTime := temp.Next(cTime)
	var mockCtrl *gomock.Controller
	var licenseClient l_v1.LicenseServiceClient
	var rep repo.Product
	type args struct {
		ctx context.Context
		j   *job.Job
	}
	tests := []struct {
		name    string
		w       *LicenseCalWorker
		args    args
		setup   func()
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "lcalw"},
					Status: job.JobStatusPENDING,
					Data:   json.RawMessage(`{"updatedBy":"cron"}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				rep = mockRepo
				mockLicenseClient := mocklicense.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mockLicenseClient
				mockRepo.EXPECT().ListAcqrightsProducts(ctx).Times(1).Return([]db.ListAcqrightsProductsRow{
					{
						Swidtag: "P1",
						Scope:   "Scope1",
					},
					{
						Swidtag: "P2",
						Scope:   "Scope2",
					},
				}, nil)
				gomock.InOrder(
					mockLicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &l_v1.ListAcquiredRightsForProductRequest{
						SwidTag: "P1",
						Scope:   "Scope1",
					}).Times(1).Return(&l_v1.ListAcquiredRightsForProductResponse{
						AcqRights: []*l_v1.ProductAcquiredRights{
							{
								SKU:            "Acq1",
								SwidTag:        "P1",
								NumCptLicences: 20,
								AvgUnitPrice:   10,
							},
							{
								SKU:            "Acq2,Acq3",
								SwidTag:        "P1",
								NumCptLicences: 10,
								AvgUnitPrice:   10,
							},
						},
					}, nil),
					mockRepo.EXPECT().AddComputedLicenses(ctx, db.AddComputedLicensesParams{
						Sku:              "Acq1",
						Computedlicenses: 20,
						Computedcost:     decimal.NewFromFloat(10 * float64(20)),
						Scope:            "Scope1",
					}).Times(1).Return(nil),
					mockRepo.EXPECT().AddComputedLicenses(ctx, db.AddComputedLicensesParams{
						Sku:              "Acq2",
						Computedlicenses: 10,
						Computedcost:     decimal.NewFromFloat(10 * float64(10)),
						Scope:            "Scope1",
					}).Times(1).Return(nil),
					mockRepo.EXPECT().AddComputedLicenses(ctx, db.AddComputedLicensesParams{
						Sku:              "Acq3",
						Computedlicenses: 10,
						Computedcost:     decimal.NewFromFloat(10 * float64(10)),
						Scope:            "Scope1",
					}).Times(1).Return(nil),
					mockRepo.EXPECT().UpsertDashboardUpdates(ctx, db.UpsertDashboardUpdatesParams{
						UpdatedAt:    cTime,
						NextUpdateAt: sql.NullTime{Time: nTime, Valid: true},
						UpdatedBy:    "cron",
						Scope:        "Scope1",
					}).Times(1).Return(nil),

					mockLicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &l_v1.ListAcquiredRightsForProductRequest{
						SwidTag: "P2",
						Scope:   "Scope2",
					}).Times(1).Return(&l_v1.ListAcquiredRightsForProductResponse{
						AcqRights: []*l_v1.ProductAcquiredRights{
							{
								SKU:            "Acq4",
								SwidTag:        "P2",
								NumCptLicences: 10,
								AvgUnitPrice:   5,
							},
						},
					}, nil),
					mockRepo.EXPECT().AddComputedLicenses(ctx, db.AddComputedLicensesParams{
						Sku:              "Acq4",
						Computedlicenses: 10,
						Computedcost:     decimal.NewFromFloat(5 * float64(10)),
						Scope:            "Scope2",
					}).Times(1).Return(nil),

					mockRepo.EXPECT().UpsertDashboardUpdates(ctx, db.UpsertDashboardUpdatesParams{
						UpdatedAt:    cTime,
						NextUpdateAt: sql.NullTime{Time: nTime, Valid: true},
						UpdatedBy:    "cron",
						Scope:        "Scope2",
					}).Times(1).Return(nil),
				)
			},
		},
		{name: "SUCCESS - no products",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "lcalw"},
					Status: job.JobStatusPENDING,
					Data:   json.RawMessage(`{"updatedBy":"cron"}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				rep = mockRepo
				mockLicenseClient := mocklicense.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mockLicenseClient
				mockRepo.EXPECT().ListAcqrightsProducts(ctx).Times(1).Return(nil, sql.ErrNoRows)
			},
		},
		{name: "SUCCESS - ListAcqRightsForProduct - no response",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "lcalw"},
					Status: job.JobStatusPENDING,
					Data:   json.RawMessage(`{"updatedBy":"cron"}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				rep = mockRepo
				mockLicenseClient := mocklicense.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mockLicenseClient
				mockRepo.EXPECT().ListAcqrightsProducts(ctx).Times(1).Return([]db.ListAcqrightsProductsRow{
					{
						Swidtag: "P1",
						Scope:   "Scope1",
					},
					{
						Swidtag: "P2",
						Scope:   "Scope2",
					},
				}, nil)
				gomock.InOrder(
					mockLicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &l_v1.ListAcquiredRightsForProductRequest{
						SwidTag: "P1",
						Scope:   "Scope1",
					}).Times(1).Return(nil, nil),
					mockLicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &l_v1.ListAcquiredRightsForProductRequest{
						SwidTag: "P2",
						Scope:   "Scope2",
					}).Times(1).Return(&l_v1.ListAcquiredRightsForProductResponse{
						AcqRights: []*l_v1.ProductAcquiredRights{
							{
								SKU:            "Acq3",
								SwidTag:        "P2",
								NumCptLicences: 10,
								AvgUnitPrice:   5,
							},
						},
					}, nil),
					mockRepo.EXPECT().AddComputedLicenses(ctx, db.AddComputedLicensesParams{
						Sku:              "Acq3",
						Computedlicenses: 10,
						Computedcost:     decimal.NewFromFloat(5 * float64(10)),
						Scope:            "Scope2",
					}).Times(1).Return(nil),
					mockRepo.EXPECT().UpsertDashboardUpdates(ctx, db.UpsertDashboardUpdatesParams{
						UpdatedAt:    cTime,
						NextUpdateAt: sql.NullTime{Time: nTime, Valid: true},
						UpdatedBy:    "cron",
						Scope:        "Scope2",
					}).Times(1).Return(nil),
				)
			},
		},
		{name: "FAILURE - ListAcqrightsProducts - DBError",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "lcalw"},
					Status: job.JobStatusPENDING,
					Data:   json.RawMessage(`{"updatedBy":"cron"}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				rep = mockRepo
				mockLicenseClient := mocklicense.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mockLicenseClient
				mockRepo.EXPECT().ListAcqrightsProducts(ctx).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "SUCCESS - ListAcqRightsForProduct - can not fetch acqrights for product",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "lcalw"},
					Status: job.JobStatusPENDING,
					Data:   json.RawMessage(`{"updatedBy":"cron"}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				rep = mockRepo
				mockLicenseClient := mocklicense.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mockLicenseClient
				mockRepo.EXPECT().ListAcqrightsProducts(ctx).Times(1).Return([]db.ListAcqrightsProductsRow{
					{
						Swidtag: "P1",
						Scope:   "Scope1",
					},
					{
						Swidtag: "P2",
						Scope:   "Scope2",
					},
				}, nil)
				gomock.InOrder(
					mockLicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &l_v1.ListAcquiredRightsForProductRequest{
						SwidTag: "P1",
						Scope:   "Scope1",
					}).Times(1).Return(nil, errors.New("Internal")),
					mockLicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &l_v1.ListAcquiredRightsForProductRequest{
						SwidTag: "P2",
						Scope:   "Scope2",
					}).Times(1).Return(nil, errors.New("Internal")),
				)
			},
		},
		{name: "FAILURE - AddComputedLicenses - DBError",
			args: args{
				ctx: ctx,
				j: &job.Job{
					Type:   sql.NullString{String: "lcalw"},
					Status: job.JobStatusPENDING,
					Data:   json.RawMessage(`{"updatedBy":"cron"}`),
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				rep = mockRepo
				mockLicenseClient := mocklicense.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mockLicenseClient
				mockRepo.EXPECT().ListAcqrightsProducts(ctx).Times(1).Return([]db.ListAcqrightsProductsRow{
					{
						Swidtag: "P1",
						Scope:   "Scope1",
					},
				}, nil)
				gomock.InOrder(
					mockLicenseClient.EXPECT().ListAcqRightsForProduct(ctx, &l_v1.ListAcquiredRightsForProductRequest{
						SwidTag: "P1",
						Scope:   "Scope1",
					}).Times(1).Return(&l_v1.ListAcquiredRightsForProductResponse{
						AcqRights: []*l_v1.ProductAcquiredRights{
							{
								SKU:            "Acq1",
								SwidTag:        "P1",
								NumCptLicences: 20,
								AvgUnitPrice:   10,
							},
							{
								SKU:            "Acq2",
								SwidTag:        "P1",
								NumCptLicences: 10,
								AvgUnitPrice:   10,
							},
						},
					}, nil),
					mockRepo.EXPECT().AddComputedLicenses(ctx, db.AddComputedLicensesParams{
						Sku:              "Acq1",
						Computedlicenses: 20,
						Computedcost:     decimal.NewFromFloat(10 * float64(20)),
						Scope:            "Scope1",
					}).Times(1).Return(errors.New("Internal")),
				)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			w := &LicenseCalWorker{
				id:            "lcalw",
				productRepo:   rep,
				licenseClient: licenseClient,
				cronTime:      "@every 12h",
			}
			if err := w.DoWork(tt.args.ctx, tt.args.j); (err != nil) != tt.wantErr {
				t.Errorf("LicenseCalWorker.DoWork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
*/
