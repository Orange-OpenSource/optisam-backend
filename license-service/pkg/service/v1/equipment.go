// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"fmt"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func equipmentTypeExistsByID(ID string, eqTypes []*repo.EquipmentType) (*repo.EquipmentType, error) {
	for _, eqt := range eqTypes {
		if eqt.ID == ID {
			return eqt, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "equipment not exists")
}

func equipmentTypeExistsByType(eqType string, eqTypes []*repo.EquipmentType) int {
	for i := 0; i < len(eqTypes); i++ {
		if eqTypes[i].Type == eqType {
			return i
		}
	}
	return -1
}

func attributeUsed(name string, attr []*repo.Attribute) bool {
	for _, attrMap := range attr {
		if name == attrMap.MappedTo {
			return true
		}
	}
	return false
}

func validateEquipUpdation(mappedTo []string, equip *repo.EquipmentType, parentID string, newAttr []*v1.Attribute) error {
	countParentKey := 0
	for _, attr := range newAttr {
		if attr.PrimaryKey {
			return status.Error(codes.InvalidArgument, "primary key not required")
		}
		if attr.ParentIdentifier {
			countParentKey++
			if attr.DataType != v1.DataTypes_STRING {
				return status.Error(codes.InvalidArgument, "only string data type is allowed for parent identifier")
			}
		}
	}

	if equip.ParentID != "" && countParentKey > 0 {
		return status.Error(codes.InvalidArgument, "no parent identifier required when parent is already present ")
	}

	if parentID == "" && countParentKey > 0 {
		return status.Error(codes.InvalidArgument, "parent is not selected for equipment type ")
	}

	if countParentKey > 1 {
		return status.Errorf(codes.InvalidArgument, "multiple parent keys are found")
	}
	return validateNewAttributes(mappedTo, equip.Attributes, newAttr)
}

func validateNewAttributes(mappedTo []string, oldAttr []*repo.Attribute, newAttr []*v1.Attribute) error {
	names := make(map[string]struct{})
	mappings := make(map[string]string)

	for _, attr := range oldAttr {
		name := strings.ToUpper(attr.Name)
		names[name] = struct{}{}
		mappings[attr.MappedTo] = name
	}
	// vaidations on attributes
	for _, attr := range newAttr {
		// check if name if unique or not
		name := strings.ToUpper(attr.Name)
		_, ok := names[name]
		if ok {
			// we arlready have this name for some other attribute
			return status.Errorf(codes.InvalidArgument, "attribute name: %v, is already given to some other attribte", attr.Name)
		}

		// atttribute name does not exist before
		// make an entry
		names[name] = struct{}{}
		// check if mapping of equipment exists
		mappingFound := false
		for _, mapping := range mappedTo {
			if mapping == attr.MappedTo {
				mappingFound = true
				break
			}
		}

		if !mappingFound {
			return status.Errorf(codes.InvalidArgument, "mapping:%v does not exist", attr.MappedTo)
		}

		attrName, ok := mappings[attr.MappedTo]
		if ok {
			// mapping is already assigned to some other attributes for some other attribute
			return status.Errorf(codes.InvalidArgument, "attribute mapping: %v, is already given to attribte: %v", attr.MappedTo, attrName)
		}

		// atttribute mapping does not exist before
		// make an entry
		mappings[attr.MappedTo] = attr.Name

		if attr.Searchable {
			if !attr.Displayed {
				return status.Error(codes.InvalidArgument, "searchable attribute should always be displayable")
			}
		}
	}
	return nil
}

func servAttrToRepoAttr(attr *v1.Attribute) *repo.Attribute {
	repoAttr := &repo.Attribute{
		ID:                 attr.ID,
		Name:               attr.Name,
		Type:               repo.DataType(attr.DataType),
		IsIdentifier:       attr.PrimaryKey,
		IsSearchable:       attr.Searchable,
		IsDisplayed:        attr.Displayed,
		IsParentIdentifier: attr.ParentIdentifier,
		MappedTo:           attr.MappedTo,
		IsSimulated:        attr.Simulated,
	}

	switch attr.DataType {
	case v1.DataTypes_INT:
		repoAttr.IntVal = int(attr.GetIntVal())
		repoAttr.IntValOld = int(attr.GetIntValOld())
	case v1.DataTypes_FLOAT:
		repoAttr.FloatVal = attr.GetFloatVal()
		repoAttr.FloatValOld = attr.GetFloatValOld()
	case v1.DataTypes_STRING:
		repoAttr.StringVal = attr.GetStringVal()
		repoAttr.StringValOld = attr.GetStringValOld()
	}

	fmt.Println("Repo Attributes: ", repoAttr)

	return repoAttr

}
