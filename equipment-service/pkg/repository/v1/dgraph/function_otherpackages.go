// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"fmt"
	v1 "optisam-backend/equipment-service/pkg/repository/v1"
	"regexp"
	"strings"
)

type dgraphSortOrder string

type totalRecords struct {
	TotalCount int32
}

type prodPred string

func (p prodPred) String() string {
	return string(p)
}

const (
	offset   string = "$offset"
	pagesize string = "$pagesize"
)

const (
	prodPredName      prodPred = "product.name"
	prodPredSwidTag   prodPred = "product.swidtag"
	prodPredVersion   prodPred = "product.version"
	prodPredEditor    prodPred = "product.editor"
	prodPredNumOfApp  prodPred = "val(numOfApplications)"
	prodPredNumOfEqp  prodPred = "val(numOfEquipments)"
	prodPredTotalCost prodPred = "val(totalCost)"
)

// String implements string interface
func (so dgraphSortOrder) String() string {
	return string(so)
}

const (
	sortASC  dgraphSortOrder = "orderasc"
	sortDESC dgraphSortOrder = "orderdesc"
)

func sortOrderForDgraph(key v1.SortOrder) (dgraphSortOrder, error) {
	switch key {
	case 0:
		return sortASC, nil
	case 1:
		return sortDESC, nil
	default:
		return "", fmt.Errorf("sortOrderForDgraph - cannot find dgraph predicate for key: %d", key)
	}
}

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

func stringFilter(pred string, q v1.Queryable) string {
	vals := q.Values()
	if len(vals) == 0 {
		return stringFilterSingle(q.Type(), pred, q.Value())
	}
	filters := make([]string, 0, len(vals))
	for _, val := range vals {
		filters = append(filters, stringFilterSingle(q.Type(), pred, val))
	}
	return " ( " + strings.Join(filters, "OR") + " ) "
}

func stringFilterSingle(typ v1.Filtertype, pred string, val interface{}) string {
	strVal, ok := val.(string)
	if ok {
		return stringFilterValString(typ, pred, strVal)
	}
	switch typ {
	case v1.EqFilter:
		return fmt.Sprintf(" (eq(%v,\"%v\")) ", pred, val)
	case v1.RegexFilter:
		return fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, val)
	default:
		// By default, regex filter is used.
		return fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, val)
	}
}

func stringFilterValString(typ v1.Filtertype, pred string, val string) string {
	switch typ {
	case v1.EqFilter:
		return fmt.Sprintf(" (eq(%v,\"%v\")) ", pred, val)
	case v1.RegexFilter:
		val = regexp.QuoteMeta(val)
		return fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, val)
	default:
		val = regexp.QuoteMeta(val)
		// By default, regex filter is used.
		return fmt.Sprintf(" (regexp(%v,/^%v/i)) ", pred, val)
	}
}
