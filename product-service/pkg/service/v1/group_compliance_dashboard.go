package v1

import (
	"context"
	"sync"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ProductServiceServer) GroupComplianceEditorCost(ctx context.Context, req *v1.GroupComplianceEditorRequest) (*v1.GroupComplianceEditorResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("rest - GroupComplianceEditorCost ", zap.String("Reason: ", "ClaimsNotFoundError"))
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		logger.Log.Error("rest - GroupComplianceEditorCost ", zap.String("Reason: ", "user doesnot have access to Group Compliance EditorCost"))
		return nil, status.Error(codes.PermissionDenied, "user doesnot have access to Group Compliance EditorCost")
	}

	scopes := req.GetScopes()
	// string{"OSN", "OFR", "UIT", "OLN", "OCM", "CLR", "SPC", "VER", "PER", "ANJ", "MAN", "API", "OJO", "MTS", "AAK", "AKA", "MON", "CHA", "REG", "LIC", "DPK", "ACR", "MTX", "PCT", "POI", "BUG", "ACQ", "PRS", "OSK"}
	editor := req.GetEditor()
	apiresp := v1.GroupComplianceEditorResponse{}
	costs := v1.ScopesEditorCosts{}
	var (
		wg         sync.WaitGroup
		uc, cc, tc float64
	)
	errorChan := make(chan error, 1)
	wg.Add(3)
	go counterfietCosts(s, scopes, editor, &costs, ctx, errorChan, &wg, &cc)
	go underUsageCosts(s, scopes, editor, &costs, ctx, errorChan, &wg, &uc)
	go totalCosts(s, scopes, editor, &costs, ctx, errorChan, &wg, &tc)
	wg.Wait()
	var errorchan error
	if len(errorChan) > 0 {
		for err := range errorChan {
			errorchan = err
			logger.Log.Error("rest - GroupComplianceEditorCost ", zap.String("Reason: ", err.Error()))
			break
		}
	}
	if errorchan != nil {
		logger.Log.Error("rest - getEditorFilters ", zap.String("Reason: ", errorchan.Error()))
		return &apiresp, nil
	}
	apiresp.Costs = &costs
	apiresp.GroupCounterFeitingCost = cc
	apiresp.GroupTotalCost = tc
	apiresp.GroupUnderUsageCost = uc
	return &apiresp, nil
}

func counterfietCosts(s *ProductServiceServer, scopes []string, editor string, costs *v1.ScopesEditorCosts, ctx context.Context, errorChan chan error, wg *sync.WaitGroup, groupTotal *float64) {
	defer wg.Done()
	dbresp, err := s.ProductRepo.GetScopeCounterfietAmountEditor(ctx, db.GetScopeCounterfietAmountEditorParams{
		Column1: scopes,
		Editor:  editor,
	})
	if err != nil {
		logger.Log.Error("rest - GroupComplianceEditorCost - counterfietCosts", zap.String("Reason: ", err.Error()))
		errorChan <- err
		return
	}
	var ccosts []*v1.ScopeCost
	scpMap := make(map[string]bool)
	for _, i := range dbresp {
		obj := v1.ScopeCost{
			Scope: i.Scope,
			Cost:  i.Cost,
		}
		*groupTotal = *groupTotal + i.Cost
		ccosts = append(ccosts, &obj)
		scpMap[i.Scope] = true
	}
	//populate remaing to 0
	for _, scpname := range scopes {
		if !scpMap[scpname] {
			obj := v1.ScopeCost{
				Scope: scpname,
				Cost:  0,
			}
			ccosts = append(ccosts, &obj)
		}
	}

	costs.CounterFeiting = ccosts
}

func underUsageCosts(s *ProductServiceServer, scopes []string, editor string, costs *v1.ScopesEditorCosts, ctx context.Context, errorChan chan error, wg *sync.WaitGroup, groupTotal *float64) {
	defer wg.Done()
	dbresp, err := s.ProductRepo.GetScopeUnderUsageCostEditor(ctx, db.GetScopeUnderUsageCostEditorParams{
		Column1: scopes,
		Editor:  editor,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GroupComplianceEditorCost - underUsageCosts", zap.Error(err))
		errorChan <- err
		return
	}
	var ccosts []*v1.ScopeCost
	scpMap := make(map[string]bool)
	for _, i := range dbresp {
		obj := v1.ScopeCost{
			Scope: i.Scope,
			Cost:  i.Cost,
		}
		*groupTotal = *groupTotal + i.Cost
		ccosts = append(ccosts, &obj)
		scpMap[i.Scope] = true
	}
	for _, scpname := range scopes {
		if !scpMap[scpname] {
			obj := v1.ScopeCost{
				Scope: scpname,
				Cost:  0,
			}
			ccosts = append(ccosts, &obj)
		}
	}
	costs.UnderUsage = ccosts
}

func totalCosts(s *ProductServiceServer, scopes []string, editor string, costs *v1.ScopesEditorCosts, ctx context.Context, errorChan chan error, wg *sync.WaitGroup, groupTotal *float64) {
	defer wg.Done()
	// res.Costs.CounterFeiting
	dbresp, err := s.ProductRepo.GetScopeTotalAmountEditor(ctx, db.GetScopeTotalAmountEditorParams{
		Column1: scopes,
		Editor:  editor,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GroupComplianceEditorCost - totalCosts", zap.Error(err))
		errorChan <- err
		return
	}
	var ccosts []*v1.ScopeCost
	scpMap := make(map[string]bool)
	for _, i := range dbresp {
		obj := v1.ScopeCost{
			Scope: i.Scope,
			Cost:  i.Cost,
		}
		*groupTotal = *groupTotal + i.Cost

		ccosts = append(ccosts, &obj)
		scpMap[i.Scope] = true
	}
	for _, scpname := range scopes {
		if !scpMap[scpname] {
			obj := v1.ScopeCost{
				Scope: scpname,
				Cost:  0,
			}
			ccosts = append(ccosts, &obj)
		}
	}
	costs.Total = ccosts
}
