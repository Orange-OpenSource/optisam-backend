package rest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"optisam-backend/common/optisam/logger"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

func getProductFilterDetails(fil *Productsfilters, handler *handler, ctx context.Context, filterType string, errorChan chan error, wg *sync.WaitGroup) {
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
	case "delpymentType":
		rows, err = handler.Db.QueryContext(ctx, locations)
	case "recommendationType":
		rows, err = handler.Db.QueryContext(ctx, recommendationType)
	case "licencingType":
		rows, err = handler.Db.QueryContext(ctx, licensings)
	case "entityType":
		rows, err = handler.Db.QueryContext(ctx, scopes)
	case "vendorType":
		rows, err = handler.Db.QueryContext(ctx, vendors)
	}
	if err != nil {
		logger.Log.Error("rest - getProductFilterDetails ", zap.String("Reason: ", err.Error()))
		errorChan <- err
	}
	defer rows.Close()

	scopesmaps := make(map[string]string)
	if filterType == "entityType" {
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
			logger.Log.Error("rest - getProductFilterDetails ", zap.String("Reason: ", err.Error()))
		}
		if filterType == "entityType" && scopesmaps != nil {
			filter.Name = scopesmaps[filter.Name]
		}
		filterDetail.Filter = append(filterDetail.Filter, filter)
	}
	sort.Slice(filterDetail.Filter, func(i, j int) bool {
		return strings.ToLower(filterDetail.Filter[i].Name) < strings.ToLower(filterDetail.Filter[j].Name)
	})
	switch filterType {
	case "delpymentType":
		fil.DeploymentType = filterDetail
	case "recommendationType":
		fil.Recommendation = filterDetail
	case "licencingType":
		fil.Licensing = filterDetail
	case "entityType":
		fil.Entities = filterDetail
	case "vendorType":
		fil.Vendors = filterDetail
	}
}

func (handler *handler) GetProductFilters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	logger.Log.Info("Get Product Filters", zap.Any("start time", time.Now()))
	var fil Productsfilters
	var wg sync.WaitGroup
	errorChan := make(chan error, 5)
	wg.Add(5)
	go getProductFilterDetails(&fil, handler, r.Context(), "delpymentType", errorChan, &wg)
	go getProductFilterDetails(&fil, handler, r.Context(), "recommendationType", errorChan, &wg)
	go getProductFilterDetails(&fil, handler, r.Context(), "licencingType", errorChan, &wg)
	go getProductFilterDetails(&fil, handler, r.Context(), "entityType", errorChan, &wg)
	go getProductFilterDetails(&fil, handler, r.Context(), "vendorType", errorChan, &wg)
	wg.Wait()
	var errorchan error
	if len(errorChan) > 0 {
		for err := range errorChan {
			errorchan = err
			break
		}
	}
	if errorchan != nil {
		logger.Log.Error("rest - getProductFilters ", zap.String("Reason: ", errorchan.Error()))
		sendErrorResponse(http.StatusInternalServerError, errorchan.Error(), w)
		return
	}
	json.NewEncoder(w).Encode(fil)
	logger.Log.Info("GetProducts", zap.Any("end", time.Now()))
}
