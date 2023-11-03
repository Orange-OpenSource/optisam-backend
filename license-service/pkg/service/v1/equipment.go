package v1

import (
	"fmt"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func equipmentTypeExistsByID(id string, eqTypes []*repo.EquipmentType) (*repo.EquipmentType, error) {
	for _, eqt := range eqTypes {
		if eqt.ID == id {
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

func getChildEquipmentsByParentType(parentType string, eqTypes []*repo.EquipmentType) []string {
	var equipIndex []string
	for i := 0; i < len(eqTypes); i++ {
		if eqTypes[i].ParentType == parentType {
			equipIndex = append(equipIndex, eqTypes[i].Type)
		}
	}
	return equipIndex
}
