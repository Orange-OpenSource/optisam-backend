package v1

import (
	"context"
	"database/sql"
	"fmt"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	dbmock "optisam-backend/catalog-service/pkg/repository/v1/mock"
	"optisam-backend/catalog-service/pkg/repository/v1/postgres/db"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ctx = grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"s1", "s2", "s3"},
	})
)

func TestInsertProduct(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProductCatalog(mockCtrl)
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	testSet := []struct {
		name   string
		input  *v1.Product
		output *v1.Product
		mock   func(*v1.Product)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "UpsertProductWithCorrectData",
			input: &v1.Product{
				Name:               "product1",
				EditorID:           "f84e136c-28f8-4f77-980e-3532c81bc4ea",
				Metrics:            []string{"abc,cd"},
				GenearlInformation: "ginfo",
				ContracttTips:      "ct",
				LocationType:       "Both",
				OpenSource: &v1.OpenSource{IsOpenSource: true, OpenLicences: "string",
					OpenSourceType: "NONE"},
				CloseSource: &v1.CloseSource{IsCloseSource: false},
				Version: []*v1.Version{{Name: "v1", Recommendation: "string",
					EndOfLife: timestamppb.Now(), EndOfSupport: timestamppb.Now()}},
				Recommendation: "recomm",
				UsefulLinks:    []string{"useful", "links"},
				SupportVendors: []string{"support", "vendors"},
				CreatedOn:      timestamppb.Now(),
				UpdatedOn:      timestamppb.Now(),
				SwidtagProduct: "swidtag",
			},
			output: &v1.Product{
				Id:                 uuid.New().String(),
				Name:               "product1",
				EditorID:           "f84e136c-28f8-4f77-980e-3532c81bc4ea",
				Metrics:            []string{"abc,cd"},
				GenearlInformation: "ginfo",
				ContracttTips:      "ct",
				LocationType:       "Both",
				OpenSource: &v1.OpenSource{IsOpenSource: true, OpenLicences: "string",
					OpenSourceType: "None"},
				CloseSource: &v1.CloseSource{IsCloseSource: false},
				Version: []*v1.Version{{Id: uuid.New().String(), Name: "v1", Recommendation: "string",
					EndOfLife: timestamppb.Now(), EndOfSupport: timestamppb.Now()}},
				Recommendation: "recomm",
				UsefulLinks:    []string{"useful", "links"},
				SupportVendors: []string{"support", "vendors"},
				CreatedOn:      timestamppb.Now(),
				UpdatedOn:      timestamppb.Now(),
				SwidtagProduct: "swidtag",
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.Product) {
				// userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				// if !ok {
				//  t.Errorf("cannot find claims in context")
				// }
				dbObj.EXPECT().InsertProductCatalog(ctx, input).Return(nil).Times(1)
				dbObj.EXPECT().InsertProductTx(ctx, input).Return(input, nil).Times(1)
				dbObj.EXPECT().GetEditorCatalog(ctx, input.EditorID).Return(db.EditorCatalog{
					ID:                 "f84e136c-28f8-4f77-980e-3532c81bc4ea",
					Name:               "WantEditor",
					GeneralInformation: sql.NullString{String: "wantstring", Valid: true},
					PartnerManagers:    nil,
					Audits:             nil,
					Vendors:            nil,
					CreatedOn:          time.Now(),
					UpdatedOn:          time.Now()}, nil).Times(1)
				dbObj.EXPECT().InsertVersionCatalog(ctx, input.Version).Return(nil).Times(1)
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductCatalogServer(dbObj, &workerqueue.Queue{}, nil)
			fmt.Println("a ...any,", test.ctx, "poiuytgrfed", test.input)
			got, err := s.InsertProduct(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *&got.Name, *(&test.output.Name)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)
			} else {
				fmt.Println(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}
