// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"fmt"
	v1 "optisam-backend/license-service/pkg/repository/v1"
)

type insPred string

const (
	insPredName      insPred = "instance.id" //"instance.name"
	insPredEnv       insPred = "instance.environment"
	insPredID        insPred = "instance.id"
	insPredNumOfProd insPred = "val(numOfProducts)"
	insPredNumOfEqp  insPred = "val(numOfEquipments)"
)

func keyToPredForInstance(key int32) (insPred, error) {
	switch key {
	case 0:
		return insPredName, nil
	case 1:
		return insPredEnv, nil
	case 2:
		return insPredNumOfProd, nil
	case 3:
		return insPredNumOfEqp, nil
	default:
		return "", fmt.Errorf("keyToPredForProduct - cannot find dgraph predicate for key: %d", key)
	}
}

func keyToPredForGetInstancesForApplicationsProduct(key int32) (insPred, error) {
	switch key {
	case 0:
		return insPredName, nil
	case 1:
		return insPredEnv, nil
	case 2:
		return insPredNumOfProd, nil
	case 3:
		return insPredNumOfEqp, nil
	default:
		return "", fmt.Errorf("keyToPredForProduct - cannot find dgraph predicate for key: %d", key)
	}
}

// TODO: as this is in dgraph we need to change it ot sortOrder
type dgraphSortOrder string

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
