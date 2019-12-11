// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

// func Test_licenseServiceServer_GetProduct(t *testing.T) {

// 	//Obtaining a mock controller
// 	//Mock controller is responsible for tracking and asserting
// 	//the expectations of its associated mock objects
// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()
// 	mockLicense := mock.NewMockLicense(mockCtrl)
// 	ctx := context.Background()

// 	ServiceServer
// 	s := NewLicenseServiceServer(mockLicense)

// 	type args struct {
// 		ctx context.Context
// 		req *v1.ProductRequest
// 	}

// 	tests := []struct {
// 		name    string
// 		s       v1.LicenseServiceServer
// 		args    args
// 		mock    func()
// 		want    *v1.ProductResponse
// 		wantErr bool
// 	}{{
// 		name: "SUCCESS",
// 		s:    s,
// 		args: args{
// 			ctx: ctx,
// 			req:  &v1.ProductRequest{SwidTag:"ORAC249"},
// 		},
// 		mock: func() {
// 			p := &v1.ProductResponse{ProductInfo:&v1.ProductInfo{SwidTag:"ORAC249",Name:"Oracle DataBase",Editor:"Oracle"},
// 			ProductOptions:&v1.ProductOptions{NumOfOptions:1,Optioninfo:[]*OptionInfo{{SwidTag:"ORAC249",Name:"Oracle DataBase"}}},
// 			ProductRights:&v1.ProductRights{},
//   }
// 			Expect Do to be called once
// 			mockLicense.EXPECT().GetProductInformation(&ctx,req).Return(p, nil).Times(1)
// 		},
// 		want:     &v1.ProductResponse{ProductInfo:&v1.ProductInfo{SwidTag:"ORAC249",Name:"Oracle DataBase",Editor:"Oracle"},
// 		ProductOptions:&v1.ProductOptions{NumOfOptions:1,Optioninfo:[]*v1.OptionInfo{{SwidTag:"ORAC249",Name:"Oracle DataBase"}}},
// 		ProductRights:&v1.ProductRights{},
// },
// 		wantErr: false,
// 	},
// 		{
// 			name: "SUCCESS",
// 			s:    s,
// 			args: args{
// 				ctx: ctx,
// 				req:  &v1.ProductRequest{SwidTag:"WIND001"},
// 			},
// 			mock: func() {
// 				p :=  &v1.ProductResponse{ProductInfo:&v1.ProductInfo{SwidTag:"WIND001",Name:"Windows Server",Editor:"Windows"},
// 				ProductOptions:&v1.ProductOptions{NumOfOptions:1,Optioninfo:[]*v1.OptionInfo{{SwidTag:"WIND001",Name:"Windows Server"}}},
// 				ProductRights:&v1.ProductRights{},
// 	  }
// 				//Expect Do to be called once
// 				mockLicense.EXPECT().GetProductInformation(&ctx.req).Return(p, nil).Times(1)
// 			},
// 			want:     &v1.ProductResponse{ProductInfo:&v1.ProductInfo{SwidTag:"WIND001",Name:"Windows Server",Editor:"Windows"},
// 			ProductOptions:&v1.ProductOptions{NumOfOptions:1,Optioninfo:[]*v1.OptionInfo{{SwidTag:"WIND001",Name:"Windows Server"}}},
// 			ProductRights:&v1.ProductRights{},
//   },
// 			wantErr: false,
// 		},

// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.s.GetProduct(tt.args.ctx, tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("licenseServiceServer.GetProduct() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("licenseServiceServer.GetProduct() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_licenseServiceServer_ListApplications(t *testing.T) {

// 	//Obtaining a mock controller
// 	//Mock controller is responsible for tracking and asserting
// 	//the expectations of its associated mock objects
// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()
// 	mockLicense := mock.NewMockLicense(mockCtrl)
// 	ctx := context.Background()

// 	ServiceServer
// 	s := NewLicenseServiceServer(mockLicense)

// 	type args struct {
// 		ctx context.Context
// 		req *v1.ListApplicationsRequest
// 	}

//     tests := []struct {
// 		name    string
// 		s       v1.LicenseServiceServer
// 		args    args
// 		mock    func()
// 		want    *v1.ListApplicationsResponse
// 		wantErr bool
// 	}{{
// 		name: "SUCCESS",
// 		s:    s,
// 		args: args{
// 			ctx: ctx,
// 			req:  &v1.ListApplicationsRequest{PageNum:1,PageSize:3,SortOrder:"asc",SortBy:"name"},
// 		},
// 		mock: func() {
// 			p := &v1.ListApplicationsResponse{Applications: []*v1.Application{{ApplicationId: "92", Name: "Trence", ApplicationOwner: "Mylands"},
// 			{ApplicationId: "92", Name: "Trence", ApplicationOwner: "Mylands"},
// 		    {ApplicationId: "92", Name: "Trence", ApplicationOwner: "Mylands"}}}
// 			//Expect Do to be called once
// 			mockLicense.EXPECT().GetApplications(&ctx,req).Return(p, nil).Times(1)
// 		},
// 		want:    &v1.ListApplicationsResponse{Applications: []*v1.Application{{ApplicationId: "92", Name: "Trence", ApplicationOwner: "Mylands"},
// 		{ApplicationId: "92", Name: "Trence", ApplicationOwner: "Mylands"},
// 	    {ApplicationId: "92", Name: "Trence", ApplicationOwner: "Mylands"}}},
// 		wantErr: false,
// 	},
// 		{
// 			name: "SUCCESS",
// 			s:    s,
// 			args: args{
// 				ctx: ctx,
// 				req:  &v1.ListApplicationsRequest{PageNum:1,PageSize:5,SortOrder:"desc",SortBy:"name"},
// 			},
// 			mock: func() {
// 				p := &v1.ListApplicationsResponse{Applications: []*v1.Application{{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"},
// 				{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"},
// 				{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"},
// 				{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"},
// 				{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"}}}
// 				Expect Do to be called once
// 				mockLicense.EXPECT().GetApplications(&ctx.req).Return(p, nil).Times(1)
// 			},
// 			want:    &v1.ListApplicationsResponse{Applications: []*v1.Application{{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"},
// 			{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"},
// 			{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"},
// 			{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"},
// 			{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"}}},
// 			wantErr: false,
// 		},
// 		{
// 			name: "FAILURE",
// 			s:    s,
// 			args: args{
// 				ctx: ctx,
// 				req:  &v1.ListApplicationsRequest{PageNum:1,PageSize:3,SortOrder:"asc",SortBy:"name"},
// 			},
// 			mock: func() {
// 				p := &v1.ListApplicationsResponse{Applications: []*v1.Application{{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"},
// 				{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"},
// 				{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"}}}
// 				//Expect Do to be called once
// 				mockLicense.EXPECT().GetApplications(&ctx,req).Return(p, nil).Times(1)
// 			},
// 			want:    &v1.ListApplicationsResponse{Applications: []*v1.Application{{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"}}},
// 			wantErr: false,
// 		},
// 		{
// 			name: "FAILURE",
// 			s:    s,
// 			args: args{
// 				ctx: ctx,
// 				req:  &v1.ListApplicationsRequest{PageNum:1,PageSize:1,SortOrder:"desc",SortBy:"name"},
// 			},
// 			mock: func() {
// 				p := &v1.ListApplicationsResponse{Applications: []*v1.Application{}}
// 				Expect Do to be called once
// 				mockLicense.EXPECT().GetApplications(&ctx,req).Return(p, nil).Times(1)
// 			},
// 			want:    &v1.ListApplicationsResponse{Applications: []*v1.Application{{ApplicationId: "50", Name: "Monzaro", ApplicationOwner: "Noble"}}},
// 			wantErr: false,
// 		}}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.s.ListApplications(tt.args.ctx, tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("licenseServiceServer.ListApplications() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("licenseServiceServer.ListApplications() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_licenseServiceServer_ListProducts(t *testing.T) {

// 	Obtaining a mock controller
// 	Mock controller is responsible for tracking and asserting
// 	the expectations of its associated mock objects
// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()
// 	mockLicense := mock.NewMockLicense(mockCtrl)
// 	ctx := context.Background()

// 	ServiceServer
// 	s := NewLicenseServiceServer(mockLicense)

// 	type args struct {
// 		ctx context.Context
// 		req *v1.ListProductsRequest
// 	}
// 	tests := []struct {
// 		name    string
// 		s       v1.LicenseServiceServer
// 		args    args
// 		mock    func()
// 		want    *v1.ListProductsResponse
// 		wantErr bool
// 	}{{
// 		name: "SUCCESS",
// 		s:    s,
// 		args: args{
// 			ctx: ctx,
// 			req:  &v1.ListProductsRequest{PageNum:1,PageSize:3,SortOrder:"asc",SortBy:"name"},
// 		},
// 		mock: func() {
// 			p := &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"},
// 			{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"},
// 		    {SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"}}}
// 			Expect Do to be called once
// 			mockLicense.EXPECT().GetProducts(&ctx,req).Return(p, nil).Times(1)
// 		},
// 		want:    &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"},
// 		{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"},
// 	    {SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"}}},
// 		wantErr: false,
// 	},
// 		{
// 			name: "SUCCESS",
// 			s:    s,
// 			args: args{
// 				ctx: ctx,
// 				req:  &v1.ListProductsRequest{PageNum:1,PageSize:5,SortOrder:"desc",SortBy:"name"},
// 			},
// 			mock: func() {
// 				p := &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"},
// 				{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"},
// 				{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"},
// 				{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"},
// 				{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"}}}
// 				Expect Do to be called once
// 				mockLicense.EXPECT().GetProducts(&ctx.req).Return(p, nil).Times(1)
// 			},
// 			want:    &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"},
// 			{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"},
// 			{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"},
// 			{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"},
// 			{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"}}},
// 			wantErr: false,
// 		},
// 		{
// 			name: "FAILURE",
// 			s:    s,
// 			args: args{
// 				ctx: ctx,
// 				req:  &v1.ListProductsRequest{PageNum:1,PageSize:3,SortOrder:"asc",SortBy:"name"},
// 			},
// 			mock: func() {
// 				p := &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"},
// 				{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"},
// 				{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"}}}
// 				Expect Do to be called once
// 				mockLicense.EXPECT().GetProducts(&ctx,req).Return(p, nil).Times(1)
// 			},
// 			want:    &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat", Version: "10.2.0"}}},
// 			wantErr: false,
// 		},
// 		{
// 			name: "FAILURE",
// 			s:    s,
// 			args: args{
// 				ctx: ctx,
// 				req:  &v1.ListProductsRequest{PageNum:1,PageSize:1,SortOrder:"desc",SortBy:"name"},
// 			},
// 			mock: func() {
// 				p := &v1.ListProductsResponse{Products: []*v1.Product{}}
// 				Expect Do to be called once
// 				mockLicense.EXPECT().GetProducts(&ctx,req).Return(p, nil).Times(1)
// 			},
// 			want:    &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"}}},
// 			wantErr: false,
// 		}}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.s.ListProducts(tt.args.ctx, tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("licenseServiceServer.ListProducts() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("licenseServiceServer.ListProducts() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_licenseServiceServer_ListProducts(t *testing.T) {

// 	// Obtaining a mock controller
// 	// Mock controller is responsible for tracking and asserting
// 	// the expectations of its associated mock objects
// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()
// 	mockLicense := mock.NewMockLicense(mockCtrl)
// 	ctx := context.Background()

// 	// ServiceServer
// 	s := NewLicenseServiceServer(mockLicense)

// 	type args struct {
// 		ctx context.Context
// 		in1 *empty.Empty
// 	}
// 	tests := []struct {
// 		name    string
// 		s       v1.LicenseServiceServer
// 		args    args
// 		mock    func()
// 		want    *v1.ListProductsResponse
// 		wantErr bool
// 	}{{
// 		name: "SUCCESS",
// 		s:    s,
// 		args: args{
// 			ctx: ctx,
// 		},
// 		mock: func() {
// 			p := &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"}}}
// 			// Expect Do to be called once
// 			mockLicense.EXPECT().GetProducts(&ctx).Return(p, nil).Times(1)
// 		},
// 		want:    &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"}}},
// 		wantErr: false,
// 	},
// 		{
// 			name: "SUCCESS",
// 			s:    s,
// 			args: args{
// 				ctx: ctx,
// 			},
// 			mock: func() {
// 				p := &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"}}}
// 				// Expect Do to be called once
// 				mockLicense.EXPECT().GetProducts(&ctx).Return(p, nil).Times(1)
// 			},
// 			want:    &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "ORAC100", Name: "Oracle Database", Editor: "oracle"}}},
// 			wantErr: false,
// 		},
// 		{
// 			name: "FAILURE",
// 			s:    s,
// 			args: args{
// 				ctx: ctx,
// 			},
// 			mock: func() {
// 				p := &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"}}}
// 				// Expect Do to be called once
// 				mockLicense.EXPECT().GetProducts(&ctx).Return(p, nil).Times(1)
// 			},
// 			want:    &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat", Version: "10.2.0"}}},
// 			wantErr: false,
// 		},
// 		{
// 			name: "FAILURE",
// 			s:    s,
// 			args: args{
// 				ctx: ctx,
// 			},
// 			mock: func() {
// 				p := &v1.ListProductsResponse{Products: []*v1.Product{}}
// 				// Expect Do to be called once
// 				mockLicense.EXPECT().GetProducts(&ctx).Return(p, nil).Times(1)
// 			},
// 			want:    &v1.ListProductsResponse{Products: []*v1.Product{{SwidTag: "LINU103", Name: "Linux Red Hat", Editor: "redhat"}}},
// 			wantErr: false,
// 		}}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.mock()
// 			got, err := tt.s.ListProducts(tt.args.ctx, tt.args.in1)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("licenseServiceServer.ListProducts() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("licenseServiceServer.ListProducts() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
