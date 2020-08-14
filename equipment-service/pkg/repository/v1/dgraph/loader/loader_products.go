// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package loader

import (
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

var (
	productSchema = map[string]string{
		"Name":        "product.name",
		"Version":     "product.version",
		"Category":    "product.category",
		"Editor":      "product.editor",
		"SWIDTag":     "product.swidtag",
		"IsOptionOf":  "product.child",
		"IdEquipment": "product.equipment",
	}
)

func loadProducts(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Request, doneChan <-chan struct{}) {
	loadStaticTypes("products", "SWIDTag", ml, dg, masterDir, scopes, filenames, ch, doneChan, productsNquadsForRow)
}

func productsNquadsForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, []string, []string, string, bool) {
	//	nodeType := "product"
	nquads := make([]*api.NQuad, 0, len(row)+3)
	uids, upserts := []string{}, []string{}
	//prodUID := uidForXid(row[xidIDX])
	prodUID, nqs, prodUpsert := uidForXIDForType(row[xidIDX], "product", "product.swidtag", row[xidIDX], dgraphTypeProduct)
	uids = append(uids, prodUID)
	upserts = append(upserts, prodUpsert)
	nquads = append(nquads, nqs...)
	for i := 0; i < len(row); i++ {
		if i == xidIDX {
			continue
		}
		//
		// if i == len(row) {
		// 	nquads = append(nquads, &api.NQuad{
		// 		Subject:     prodUID,
		// 		Predicate:   "type_name",
		// 		ObjectValue: stringObjectValue("product"),
		// 	})
		// 	return nquads, prodUID
		// }
		predicate, ok := productSchema[cols[i]]
		if !ok {
			// if we cannot find predicate in map set predicate to
			// csv coloumn name
			predicate = cols[i]
		}
		switch predicate {
		case "product.child":
			if row[i] == "" {
				continue
			}
			// make a new node of type product
			//uid := uidForXid(row[i])
			//log.Println(row[xidIDX], row[i], prodUID, uid)
			uid, nqs, upsert := uidForXIDForType(row[i], "product", "product.swidtag", row[i], dgraphTypeProduct)
			uids = append(uids, uid)
			upserts = append(upserts, upsert)
			nquads = append(nquads, nqs...)
			// nquads = append(nquads, &api.NQuad{
			// 	Subject:     uid,
			// 	Predicate:   "type_name",
			// 	ObjectValue: stringObjectValue("product"),
			// })
			// // assign XID to node
			// nquads = append(nquads, &api.NQuad{
			// 	Subject:     uid,
			// 	Predicate:   "product.swidtag",
			// 	ObjectValue: stringObjectValue(row[i]),
			// })
			// link bot nodes
			nquads = append(nquads, &api.NQuad{
				Subject:   uid,
				Predicate: predicate,
				ObjectId:  prodUID,
			})
		case "product.editor":
			if row[i] == "" {
				continue
			}
			uid, nqs, upsert := uidForXIDForType("editor_"+row[i], "editor", "editor.name", row[i], dgraphTypeEditor)
			uids = append(uids, uid)
			upserts = append(upserts, upsert)
			nquads = append(nquads, nqs...)
			nquads = append(nquads, &api.NQuad{
				Subject:   uid,
				Predicate: "editor.product",
				ObjectId:  prodUID,
			})
			nquads = append(nquads, &api.NQuad{
				Subject:     prodUID,
				Predicate:   "product.editor",
				ObjectValue: stringObjectValue(row[i]),
			})
			// make a new node of type product

		default:
			nquads = append(nquads, &api.NQuad{
				Subject:     prodUID,
				Predicate:   predicate,
				ObjectValue: stringObjectValue(row[i]),
			})
		}
	}
	return nquads, uids, upserts, prodUID, true
}
