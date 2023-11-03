package v1

import (
	"context"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	dbmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/dbmock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
	queuemock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/queuemock"

	"github.com/golang/mock/gomock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/config"
	"go.uber.org/zap"
)

func TestGroupComplianceEditorCost(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.GroupComplianceEditorRequest
		output *v1.GroupComplianceEditorResponse
		mock   func(*v1.GroupComplianceEditorRequest)
		ctx    context.Context
		outErr bool
	}{

		{
			name:   "GroupComplianceEditorCost without context",
			input:  &v1.GroupComplianceEditorRequest{Scopes: []string{"OSN", "OFR"}, Editor: "Oracle"},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.GroupComplianceEditorRequest) {},
		},
		{
			name:  "GroupComplianceEditorCost with data",
			input: &v1.GroupComplianceEditorRequest{Scopes: []string{"OSN", "OFR"}, Editor: "Oracle"},
			ctx:   ctx,
			mock: func(data *v1.GroupComplianceEditorRequest) {
				dbObj.EXPECT().GetScopeCounterfietAmountEditor(ctx, db.GetScopeCounterfietAmountEditorParams{
					Column1: []string{"OSN", "OFR"},
					Editor:  "Oracle",
				}).Return([]db.GetScopeCounterfietAmountEditorRow{
					{
						Scope: "OSN",
						Cost:  1.0,
					},
				}, nil).Times(1)
				dbObj.EXPECT().GetScopeTotalAmountEditor(ctx, db.GetScopeTotalAmountEditorParams{
					Column1: []string{"OSN", "OFR"},
					Editor:  "Oracle",
				}).Return([]db.GetScopeTotalAmountEditorRow{
					{
						Scope: "OSN",
						Cost:  1.0,
					},
				}, nil).Times(1)
				dbObj.EXPECT().GetScopeUnderUsageCostEditor(ctx, db.GetScopeUnderUsageCostEditorParams{
					Column1: []string{"OSN", "OFR"},
					Editor:  "Oracle",
				}).Return([]db.GetScopeUnderUsageCostEditorRow{
					{
						Scope: "OSN",
						Cost:  1.0,
					},
				}, nil).Times(1)
			},
			output: &v1.GroupComplianceEditorResponse{
				Costs: &v1.ScopesEditorCosts{
					CounterFeiting: []*v1.ScopeCost{
						{
							Scope: "OSN",
							Cost:  1.0,
						},
					},
					UnderUsage: []*v1.ScopeCost{
						{
							Scope: "OSN",
							Cost:  1.0,
						},
					},
					Total: []*v1.ScopeCost{
						{
							Scope: "OSN",
							Cost:  1.0,
						},
					},
				},
				GroupCounterFeitingCost: 1,
				GroupUnderUsageCost:     1,
				GroupTotalCost:          1,
			},
			outErr: false,
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			_, err := s.GroupComplianceEditorCost(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
				// } else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				// 	t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}
