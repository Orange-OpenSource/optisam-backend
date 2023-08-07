package rest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/middleware/grpc"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// func trace() string {
// 	var line int
// 	pc, file, line, ok := runtime.Caller(1)
// 	if !ok {
// 		return "?", string(0), "?"
// 	}

// 	fn := runtime.FuncForPC(pc)
// 	if fn == nil {
// 		return file, line, "?"
// 	}
// 	return file, string(line), fn.Name()
// }

func getFilterDetails(fil *Editorfilters, handler *handler, ctx context.Context, filterType string, errorChan chan error, wg *sync.WaitGroup) {
	logger.Log.Info(fmt.Sprintf("getFilterDetails for %v", filterType), zap.Any("start time", time.Now()))
	var (
		rows         *sql.Rows
		filterDetail FilterDetail
		err          error
	)
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			errorChan <- fmt.Errorf("Panic Occured: %v for %v", r, filterType)
		}
	}()
	switch filterType {
	case "countryCode":
		rows, err = handler.Db.QueryContext(ctx, countryCode)
	case "groupContract":
		rows, err = handler.Db.QueryContext(ctx, groupContract)
	case "editorScopes":
		rows, err = handler.Db.QueryContext(ctx, editorScopes)
	case "auditYears":
		rows, err = handler.Db.QueryContext(ctx, auditYears)
	}
	if err != nil {
		logger.Log.Error("rest - getFilterDetails ", zap.String("Reason: ", err.Error()))
		errorChan <- err
	}
	defer rows.Close()
	scopesmaps := make(map[string]string)
	if filterType == "editorScopes" {
		scopes, err := handler.pCRepo.GetAllScope(ctx)
		if err != nil {
			logger.Log.Error("couldn't fetch scopes from redis", zap.Any("error", err))
		}
		if len(scopes) > 0 {
			for _, v := range scopes {
				scopesmaps[v.ScopeCode] = v.ScopeName
			}
		} else {
			scopesmaps = createScopeMap(handler)
		}
	}

	for rows.Next() {
		var filter Filter
		if err = rows.Scan(
			&filter.Name,
			&filter.Count,
			&filterDetail.TotalCount,
		); err != nil {
			logger.Log.Error("rest - getFilterDetails ", zap.String("Reason: ", err.Error()))
		}
		if filterType == "editorScopes" && scopesmaps != nil {
			filter.Name = scopesmaps[filter.Name]
		}
		filterDetail.Filter = append(filterDetail.Filter, filter)
	}
	sort.Slice(filterDetail.Filter, func(i, j int) bool {
		return strings.ToLower(filterDetail.Filter[i].Name) < strings.ToLower(filterDetail.Filter[j].Name)
	})
	switch filterType {
	case "countryCode":
		fil.CountryCode = filterDetail
	case "groupContract":
		fil.GroupContract = filterDetail
	case "editorScopes":
		fil.Entities = filterDetail
	case "auditYears":
		sort.Slice(filterDetail.Filter, func(i, j int) bool {
			return strings.ToLower(filterDetail.Filter[i].Name) > strings.ToLower(filterDetail.Filter[j].Name)
		})
		fil.Year = filterDetail
	}
	logger.Log.Info(fmt.Sprintf("getFilterDetails for %v", filterType), zap.Any("end time", time.Now()))
}

func (handler *handler) GetEditorFilters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	logger.Log.Info("Get Filters", zap.Any("start time", time.Now()))
	var fil Editorfilters
	var wg sync.WaitGroup
	errorChan := make(chan error, 4)
	wg.Add(4)
	go getFilterDetails(&fil, handler, r.Context(), "countryCode", errorChan, &wg)
	go getFilterDetails(&fil, handler, r.Context(), "groupContract", errorChan, &wg)
	go getFilterDetails(&fil, handler, r.Context(), "editorScopes", errorChan, &wg)
	go getFilterDetails(&fil, handler, r.Context(), "auditYears", errorChan, &wg)
	wg.Wait()
	var errorchan error
	if len(errorChan) > 0 {
		for err := range errorChan {
			errorchan = err
			break
		}
	}
	if errorchan != nil {
		logger.Log.Error("rest - getEditorFilters ", zap.String("Reason: ", errorchan.Error()))
		sendErrorResponse(http.StatusInternalServerError, errorchan.Error(), w)
		return
	}
	json.NewEncoder(w).Encode(fil)
	logger.Log.Info("GetFilters", zap.Any("end", time.Now()))
}

func createScopeMap(handler *handler) map[string]string {
	scopesmap := make(map[string]string)
	cronCtx, err := CreateSharedContext(handler.AuthAPI, handler.Application)
	if err != nil {
		logger.Log.Error("couldnt fetch token, will try next time when API will execute", zap.Any("error", err))
		return nil
	}
	if cronCtx != nil {
		logger.Log.Info("createScopeMap", zap.Any("before AddClaimsInContext", time.Now()))
		cronAPIKeyCtx, err := grpc.AddClaimsInContext(*cronCtx, handler.VerifyKey, handler.APIKey)
		logger.Log.Info("createScopeMap", zap.Any("after AddClaimsInContext", time.Now()))
		if err != nil {
			logger.Log.Error("Cron AddClaims Failed", zap.Error(err))
			return nil
		}
		logger.Log.Info("createScopeMap", zap.Any("before list scopes call", time.Now()))
		listScopesNames, err := handler.account.ListScopes(cronAPIKeyCtx, &accv1.ListScopesRequest{})
		logger.Log.Info("createScopeMap", zap.Any("after list scopes call", time.Now()))
		if err != nil {
			logger.Log.Error("couldnt fetch scope name, will try next time when API will execute", zap.Any("error", err))
			return nil
		}
		if listScopesNames != nil {
			for _, v := range listScopesNames.Scopes {
				scopesmap[v.ScopeCode] = v.ScopeName
			}
		}
	}
	return scopesmap
}
