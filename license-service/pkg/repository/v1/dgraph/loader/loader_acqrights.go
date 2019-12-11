// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package loader

import (
	"log"
	"strings"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
)

var (
	acqRightsSchema = map[string]*schemaNode{
		"Entity":                            &schemaNode{predName: "acqRights.entity", dataConv: stringConverter{}},
		"SKU":                               &schemaNode{predName: "acqRights.SKU", dataConv: stringConverter{}},
		"SWIDTag":                           &schemaNode{predName: "acqRights.swidtag", dataConv: stringConverter{}},
		"Product name":                      &schemaNode{predName: "acqRights.productName", dataConv: stringConverter{}},
		"Editor":                            &schemaNode{predName: "acqRights.editor", dataConv: stringConverter{}},
		"Metric":                            &schemaNode{predName: "acqRights.metric", dataConv: stringConverter{}},
		"Acquired licenses number":          &schemaNode{predName: "acqRights.numOfAcqLicences", dataConv: intConverter{}},
		"Licenses under maintenance number": &schemaNode{predName: "acqRights.numOfLicencesUnderMaintenance", dataConv: intConverter{}},
		"AVG Unit Price":                    &schemaNode{predName: "acqRights.averageUnitPrice", dataConv: floatConverter{}},
		"AVG Maintenant Unit Price":         &schemaNode{predName: "acqRights.averageMaintenantUnitPrice", dataConv: floatConverter{}},
		"Total purchase cost":               &schemaNode{predName: "acqRights.totalPurchaseCost", dataConv: floatConverter{}},
		"Total maintenance cost":            &schemaNode{predName: "acqRights.totalMaintenanceCost", dataConv: floatConverter{}},
		"Total cost":                        &schemaNode{predName: "acqRights.totalCost", dataConv: floatConverter{}},
		"updated":                           &schemaNode{predName: "updated", dataConv: stringConverter{}},
		"created":                           &schemaNode{predName: "created", dataConv: stringConverter{}},
	}
)

func loadAcquiredRights(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Mutation, doneChan <-chan struct{}) {
	loadStaticTypes("acquired rights", "SKU", ml, dg, masterDir, scopes, filenames, ch, doneChan, acquiredNquadsForRow)
}

func acquiredNquadsForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, string) {
	//	nodeType := "product"
	nquads := make([]*api.NQuad, 0, len(row)+3)
	//	acqRightUID := uidForXid(row[xidIDX])
	acqRightUID, nqs := uidForXIDForType(row[xidIDX], "acqRights", "acqRights.SKU", row[xidIDX])
	nquads = append(nquads, nqs...)
	for i := 0; i < len(row); i++ {
		// if i == len(row) {
		// 	nquads = append(nquads, &api.NQuad{
		// 		Subject:     acqRightUID,
		// 		Predicate:   "type",
		// 		ObjectValue: stringObjectValue("acqRights"),
		// 	})
		// 	return nquads, acqRightUID
		// }
		schNode, ok := acqRightsSchema[cols[i]]
		if !ok {
			log.Printf("not found: %s, map: %#v", cols[i], acqRightsSchema)
			// if we cannot find predicate in map set predicate to
			// csv coloumn name
			nquads = append(nquads, &api.NQuad{
				Subject:     acqRightUID,
				Predicate:   strings.Replace(strings.TrimSpace(cols[i]), " ", "_", -1),
				ObjectValue: stringObjectValue(row[i]),
			})
			continue
		}
		predicate := schNode.predName
		switch predicate {
		case "acqRights.swidtag":
			if row[i] == "" {
				continue
			}
			acqRightUID, nqs := uidForXIDForType(row[xidIDX], "acqRights", "acqRights.SKU", row[xidIDX])
			nquads = append(nquads, nqs...)
			nquads = append(nquads, &api.NQuad{
				Subject:     acqRightUID,
				Predicate:   predicate,
				ObjectValue: stringObjectValue(row[i]),
			})
			// make a new node of type product
			//uid := uidForXid(row[i])
			uid, nqs := uidForXIDForType(row[i], "product", "product.swidtag", row[i])
			nquads = append(nquads, nqs...)
			//log.Println(row[xidIDX], row[i], prodUID, uid)
			nquads = append(nquads, &api.NQuad{
				Subject:     uid,
				Predicate:   "type",
				ObjectValue: stringObjectValue("product"),
			})
			// assign XID to node
			nquads = append(nquads, &api.NQuad{
				Subject:     uid,
				Predicate:   "product.swidtag",
				ObjectValue: stringObjectValue(row[i]),
			})
			// link both nodes
			nquads = append(nquads, &api.NQuad{
				Subject:   uid,
				Predicate: "product.acqRights",
				ObjectId:  acqRightUID,
			})
		default:
			//log.Println(predicate)
			val, err := schNode.dataConv.convert(row[i])
			if err != nil {
				log.Printf("acquiredNquadsForRow - failed to convert data for SKU: %s, data: %s, error: %v", row[xidIDX], row[i], err)
				nquads = append(nquads, &api.NQuad{
					Subject:     acqRightUID,
					Predicate:   predicate + ".failure",
					ObjectValue: defaultObjectValue(row[i]),
				})
				continue
			}
			nquads = append(nquads, &api.NQuad{
				Subject:     acqRightUID,
				Predicate:   predicate,
				ObjectValue: val,
			})
		}
	}
	return nquads, acqRightUID
}
