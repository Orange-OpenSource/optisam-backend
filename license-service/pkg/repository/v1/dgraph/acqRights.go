package dgraph

import (
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"sort"
	"strings"

	"go.uber.org/zap"
)

type predAcqRights string

// String implements stringer
func (p predAcqRights) String() string {
	return string(p)
}

const (
	predAcqRightsSKU         predAcqRights = "acqRights.SKU"
	predAcqRightsSwidTag     predAcqRights = "acqRights.swidtag"
	predAcqRightsProductName predAcqRights = "acqRights.productName"
	predAcqRightsEditor      predAcqRights = "acqRights.editor"
	predAcqRightsMetric      predAcqRights = "acqRights.metric"
)

func scopeFilters(scopes []string) []string {
	return []string{
		fmt.Sprintf("eq(scopes,[%s])", strings.Join(scopes, ",")),
	}
}

func agregateFilters(filters ...[]string) string {
	var aggFilters []string
	for _, filter := range filters {
		aggFilters = append(aggFilters, filter...)
	}
	return "@filter( " + strings.Join(aggFilters, " AND ") + " )"
}

func acquiredRightsFilter(filter *v1.AggregateFilter) []string {
	if filter == nil || len(filter.Filters) == 0 {
		return nil
	}
	sort.Sort(filter)
	filters := make([]string, 0, len(filter.Filters))
	for _, f := range filter.Filters {
		pred, err := acquiredRightsPredForFilteringKey(v1.AcquiredRightsSearchKey(f.Key()))
		if err != nil {
			logger.Log.Error("dgraph - acquiredRightsFilter - ", zap.String("reason", err.Error()))
			continue
		}
		switch pred {
		case predAcqRightsSKU, predAcqRightsSwidTag, predAcqRightsProductName, predAcqRightsEditor, predAcqRightsMetric:
			filters = append(filters, stringFilter(pred.String(), f))
		}
	}
	return filters
}

func acquiredRightsPredForFilteringKey(key v1.AcquiredRightsSearchKey) (predAcqRights, error) {
	switch key {
	case v1.AcquiredRightsSearchKeySKU:
		return predAcqRightsSKU, nil
	case v1.AcquiredRightsSearchKeySwidTag:
		return predAcqRightsSwidTag, nil
	case v1.AcquiredRightsSearchKeyProductName:
		return predAcqRightsProductName, nil
	case v1.AcquiredRightsSearchKeyEditor:
		return predAcqRightsEditor, nil
	case v1.AcquiredRightsSearchKeyMetric:
		return predAcqRightsMetric, nil
	default:
		return "", fmt.Errorf("acquiredRightsPredForFilteringKey - unknown acquired key")
	}
}

func typeFilters(typeName, typeValue string) []string {
	return []string{fmt.Sprintf("eq("+typeName+",%s)", typeValue)}
}
