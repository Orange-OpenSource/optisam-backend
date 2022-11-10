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

// ReportRepository for Dgraph
type ReportRepository struct {
	dg *dgo.Dgraph
}

// NewReportRepository creates new Repository
func NewReportRepository(dg *dgo.Dgraph) *ReportRepository {
	return &ReportRepository{
		dg: dg,
	}
}

type JSONStringArr struct {
	Val []string
}

type Object struct {
	EquipmentTypes JSONStringArr
}

func (r *ReportRepository) EquipmentTypeParents(ctx context.Context, equipType string, scope string) ([]string, error) {
	depth, err := r.getRecursionDepth(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("equipmentTypeParents - cannot fetch recursion depth: %v", err)
	}

	q := `{
		Hierarchy(func: eq(metadata.equipment.type,` + equipType + `))@filter(eq(scopes,` + scope + `))@recurse(depth: ` + strconv.Itoa(depth) + `,loop:false)@normalize  {
			EquipmentTypes: metadata.equipment.type
			metadata.equipment.parent
		}
		}`

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("EquipmentTypeParents - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("equipmentTypeParents - cannot complete query transaction")
	}

	type data struct {
		Hierarchy []*Object
	}
	d := &data{}
	// fmt.Println(string(resp.GetJson()))
	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("EquipmentTypeParents - ", zap.String("reason", err.Error()), zap.String("query", q))
		fmt.Println(string(resp.GetJson()))
		return nil, fmt.Errorf("equipmentTypeParents - cannot unmarshal Json object")
	}
	if len(d.Hierarchy) == 0 {
		logger.Log.Error("EquipmentTypeParents - ", zap.String("reason", "unable to find equipment"), zap.String("param", equipType+","+scope))
		return nil, fmt.Errorf("equipmentTypeParents - unable to find equipment")
	}
	if len(d.Hierarchy[0].EquipmentTypes.Val) == 1 {
		return nil, repo.ErrNoData
	}
	return d.Hierarchy[0].EquipmentTypes.Val[1:], nil

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
func (r *ReportRepository) EquipmentTypeAttrs(ctx context.Context, eqtype string, scope string) ([]*repo.EquipmentAttributes, error) {
	q := `{
		EqTypeAttr(func: eq(metadata.equipment.type,` + eqtype + `))@filter(eq(scopes,"` + scope + `")) {
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
		return nil, fmt.Errorf("equipmentTypeAttrs - cannot complete query transaction")
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
		return nil, fmt.Errorf("equipmentTypeAttrs - cannot unmarshal Json object")
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

// ProductEquipments implements interface's ProductEquipments
func (r *ReportRepository) ProductEquipments(ctx context.Context, editor string, scope string, eqtype string) ([]*repo.ProductEquipment, error) {

	q := `{
		DirectEquipments(func: eq(product.editor,"` + editor + `")) @filter(eq(scopes,"` + scope + `")) @cascade{
			Swidtag: product.swidtag
	  		Equipments: product.equipment @filter(eq(equipment.type,"` + eqtype + `")){
	 	 		EquipmentID: equipment.id
	  	 		EquipmentType: equipment.type
			}
		}
	}`
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ProductEquipments - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("productEquipments - cannot complete query transaction")
	}

	type data struct {
		DirectEquipments []*repo.ProductEquipment
	}

	d := &data{}

	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("ProductEquipments - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("productEquipments - cannot unmarshal Json object")
	}

	if len(d.DirectEquipments) == 0 {
		return nil, repo.ErrNoData
	}
	return d.DirectEquipments, nil

}

type Object1 struct {
	EquipmentIDs   JSONStringArr
	EquipmentTypes JSONStringArr
}

// EquipmentParents implements interface's EquipmentParents
func (r *ReportRepository) EquipmentParents(ctx context.Context, equipID, equipType string, scope string) ([]*repo.Equipment, error) {
	depth, err := r.getRecursionDepth(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("equipmentParents - cannot fetch recursion depth: %v", err)
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
		return nil, fmt.Errorf("equipmentParents - cannot complete query transaction")
	}

	type data struct {
		EquipmentParents []*Object1
	}

	d := &data{}

	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("EquipmentParents - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("equipmentParents - cannot unmarshal Json object")
	}

	if len(d.EquipmentParents[0].EquipmentIDs.Val) == 1 {
		return nil, repo.ErrNoData
	}

	var res []*repo.Equipment

	for i := 0; i < len(d.EquipmentParents[0].EquipmentIDs.Val); i++ {
		if d.EquipmentParents[0].EquipmentIDs.Val[i] == equipID {
			continue
		}
		res = append(res, &repo.Equipment{
			EquipmentID:   d.EquipmentParents[0].EquipmentIDs.Val[i],
			EquipmentType: d.EquipmentParents[0].EquipmentTypes.Val[i],
		})
	}

	return res, nil

}

// EquipmentAttributes implements interface's EquipmentAttributes
func (r *ReportRepository) EquipmentAttributes(ctx context.Context, equipID, equipType string, attrs []*repo.EquipmentAttributes, scope string) (json.RawMessage, error) {

	q := `{
		ID as var(func: eq(equipment.id,"` + equipID + `"))@filter(eq(equipment.type,"` + equipType + `") AND eq(scopes,"` + scope + `")){}
		
		EquipmentAttributes(func: uid(ID)){
			` + getAttributes(attrs, equipType) + `
		   }
				
	}`

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("EquipmentAttributes - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("equipmentAttributes - cannot complete query transaction")
	}

	type data struct {
		EquipmentAttributes []json.RawMessage
	}

	d := &data{}

	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("EquipmentAttributes - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("equipmentAttributes - cannot unmarshal Json object")
	}

	return d.EquipmentAttributes[0], nil
}

func (r *ReportRepository) getRecursionDepth(ctx context.Context, scope string) (int, error) {
	q := `{
		var(func: has(metadata.equipment.type))@filter(eq(scopes,"` + scope + `")){
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

		if name.ParentIdentifier == true { // nolint: gocritic
			continue
		} else if name.AttributeIdentifier == true {
			attrString += name.AttributeName + ":" + "equipment.id \n"
		} else {
			attrString = attrString + name.AttributeName + ":" + "equipment." + equipType + "." + name.AttributeName + "\n"
		}

	}

	return attrString
}
