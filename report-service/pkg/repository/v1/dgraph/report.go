// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"optisam-backend/common/optisam/logger"
	repo "optisam-backend/report-service/pkg/repository/v1"
	"strconv"

	"github.com/dgraph-io/dgo/v2"
	"go.uber.org/zap"
)

//ReportRepository for Dgraph
type ReportRepository struct {
	dg *dgo.Dgraph
}

//NewReportRepository creates new Repository
func NewReportRepository(dg *dgo.Dgraph) *ReportRepository {
	return &ReportRepository{
		dg: dg,
	}
}

type equip struct {
	EquipmentType string
	Parent        *equip
}

type JSONStringArr struct {
	Val []string
}

type Object struct {
	EquipmentTypes JSONStringArr
}

func (r *ReportRepository) EquipmentTypeParents(ctx context.Context, equipType string) ([]string, error) {
	depth, err := r.getRecursionDepth(ctx)
	if err != nil {
		return nil, fmt.Errorf("EquipmentTypeParents - cannot fetch recursion depth: %v", err)
	}

	q := `{
		Heirarchy(func: eq(metadata.equipment.type,` + equipType + `)) @recurse(depth: ` + strconv.Itoa(depth) + `,loop:false)@normalize  {
			EquipmentTypes: metadata.equipment.type
			metadata.equipment.parent
		}
		}`

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("EquipmentTypeParents - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("EquipmentTypeParents - cannot complete query transaction")
	}

	type data struct {
		Heirarchy []*Object
	}
	d := &data{}
	fmt.Println(string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("EquipmentTypeParents - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("EquipmentTypeParents - cannot unmarshal Json object")
	}

	if len(d.Heirarchy[0].EquipmentTypes.Val) == 1 {
		return nil, repo.ErrNoData
	}

	return d.Heirarchy[0].EquipmentTypes.Val[1:], nil

}

// UnmarshalJSON Implements Unmarshal JSON
func (o *JSONStringArr) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("no bytes to unmarshal")
	}
	// See if we can guess based on the first character
	switch b[0] {
	case '[':
		return o.unmarshalMany(b)
	default:
		return o.unmarshalSingle(b)
	}

}

func (o *JSONStringArr) unmarshalSingle(b []byte) error {
	var t string
	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}
	o.Val = []string{t}
	return nil
}

func (o *JSONStringArr) unmarshalMany(b []byte) error {
	var t []string
	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}
	o.Val = t
	return nil
}

// EquipmentTypeAttrs implements interface's EquipmentTypeAttrs
func (r *ReportRepository) EquipmentTypeAttrs(ctx context.Context, eqtype string) ([]*repo.EquipmentAttributes, error) {
	q := `{
		EqTypeAttr(func: eq(metadata.equipment.type,` + eqtype + `)) {
		  Attributes: metadata.equipment.attribute{
		  AttributeName: attribute.name
		  AttributeIdentifier: attribute.identifier
		  ParentIdentifier: attribute.parentIdentifier
		}
		}
		}`
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("EquipmentTypeAttrs - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("EquipmentTypeAttrs - cannot complete query transaction")
	}

	type Attribute struct {
		AttributeName       string
		AttributeIdentifier bool
		ParentIdentifier    bool
	}

	type object struct {
		Attributes []*Attribute
	}

	type data struct {
		EqTypeAttr []*object
	}

	d := &data{}

	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("EquipmentTypeAttrs - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("EquipmentTypeAttrs - cannot unmarshal Json object")
	}

	var res []*repo.EquipmentAttributes

	if len(d.EqTypeAttr[0].Attributes) != 0 {
		for i := range d.EqTypeAttr[0].Attributes {
			res = append(res, &repo.EquipmentAttributes{
				AttributeName:       d.EqTypeAttr[0].Attributes[i].AttributeName,
				ParentIdentifier:    d.EqTypeAttr[0].Attributes[i].ParentIdentifier,
				AttributeIdentifier: d.EqTypeAttr[0].Attributes[i].AttributeIdentifier,
			})
		}
	} else {
		return nil, repo.ErrNoData
	}

	return res, nil

}

//ProductEquipments implements interface's ProductEquipments
func (r *ReportRepository) ProductEquipments(ctx context.Context, swidTag string, scope string, eqtype string) ([]*repo.ProductEquipment, error) {

	q := `{
		DirectEquipments(func: eq(product.swidtag,"` + swidTag + `"))@filter(eq(scopes,"` + scope + `")){
	  	Equipments: product.equipment @filter(eq(equipment.type,"` + eqtype + `")){
	 	 EquipmentID: equipment.id
	  	 EquipmentType: equipment.type
	  
	}
	}
			
			}`
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ProductEquipments - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("ProductEquipments - cannot complete query transaction")
	}

	type object struct {
		Equipments []*repo.ProductEquipment
	}

	type data struct {
		DirectEquipments []*object
	}

	d := &data{}

	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("ProductEquipments - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("ProductEquipments - cannot unmarshal Json object")
	}

	if len(d.DirectEquipments) == 0 {
		return nil, repo.ErrNoData
	}

	return d.DirectEquipments[0].Equipments, nil

}

type equipmentType struct {
	EquipmentID   string
	EquipmentType string
	Parent        []*equipmentType
}

type Object1 struct {
	EquipmentIDs   JSONStringArr
	EquipmentTypes JSONStringArr
}

//EquipmentParents implements interface's EquipmentParents
func (r *ReportRepository) EquipmentParents(ctx context.Context, equipID, equipType string, scope string) ([]*repo.ProductEquipment, error) {
	depth, err := r.getRecursionDepth(ctx)
	if err != nil {
		return nil, fmt.Errorf("EquipmentParents - cannot fetch recursion depth: %v", err)
	}

	q := `{
		ID as var(func: eq(equipment.id,"` + equipID + `"))@filter(eq(equipment.type,"` + equipType + `") AND eq(scopes,"` + scope + `")){}
	  
	 	EquipmentParents(func: uid(ID))@recurse(depth:` + strconv.Itoa(depth) + `, loop: false)@normalize{
			EquipmentIDs: equipment.id
			EquipmentTypes: equipment.type
			equipment.parent
		}
				
	}`
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("EquipmentParents - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("EquipmentParents - cannot complete query transaction")
	}

	type data struct {
		EquipmentParents []*Object1
	}

	d := &data{}

	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("EquipmentParents - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("EquipmentParents - cannot unmarshal Json object")
	}

	if len(d.EquipmentParents[0].EquipmentIDs.Val) == 1 {
		return nil, repo.ErrNoData
	}

	var res []*repo.ProductEquipment

	for i := 0; i < len(d.EquipmentParents[0].EquipmentIDs.Val); i++ {
		if d.EquipmentParents[0].EquipmentIDs.Val[i] == equipID {
			continue
		}
		res = append(res, &repo.ProductEquipment{
			EquipmentID:   d.EquipmentParents[0].EquipmentIDs.Val[i],
			EquipmentType: d.EquipmentParents[0].EquipmentTypes.Val[i],
		})
	}

	return res, nil

}

//EquipmentAttributes implements interface's EquipmentAttributes
func (r *ReportRepository) EquipmentAttributes(ctx context.Context, equipID, equipType string, attrs []*repo.EquipmentAttributes) (json.RawMessage, error) {

	q := `{
		ID as var(func: eq(equipment.id,"` + equipID + `"))@filter(eq(equipment.type,"` + equipType + `")){}
		
		EquipmentAttributes(func: uid(ID)){
			` + getAttributes(attrs, equipType) + `
		   }
				
	}`

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("EquipmentAttributes - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("EquipmentAttributes - cannot complete query transaction")
	}

	type data struct {
		EquipmentAttributes []json.RawMessage
	}

	d := &data{}

	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("EquipmentAttributes - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("EquipmentAttributes - cannot unmarshal Json object")
	}

	return json.RawMessage(d.EquipmentAttributes[0]), nil
}

func (r *ReportRepository) getRecursionDepth(ctx context.Context) (int, error) {
	q := `{
		var(func: has(metadata.equipment.type)){
		  t as count(metadata.equipment.type)
	  }
		
	  recursiondepth() {
		depth: sum(val(t))
	  }
	  
	  }`
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("getRecursionDepth - ", zap.String("reason", err.Error()), zap.String("query", q))
		return -1, fmt.Errorf("getRecursionDepth - cannot complete query transaction")
	}

	type dep struct {
		Depth int
	}

	type data struct {
		Recursiondepth []*dep
	}

	d := &data{}
	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("getRecursionDepth - ", zap.String("reason", err.Error()), zap.String("query", q))
		return -1, fmt.Errorf("getRecursionDepth - cannot unmarshal Json object")
	}

	return d.Recursiondepth[0].Depth, nil

}

func getAttributes(attrs []*repo.EquipmentAttributes, equipType string) string {
	attrString := ""

	for _, name := range attrs {

		if name.ParentIdentifier == true {
			continue
		} else if name.AttributeIdentifier == true {
			attrString += name.AttributeName + ":" + "equipment.id \n"
		} else {
			attrString = attrString + name.AttributeName + ":" + "equipment." + equipType + "." + name.AttributeName + "\n"
		}

	}

	return attrString
}
