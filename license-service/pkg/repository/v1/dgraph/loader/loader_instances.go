// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package loader

import (
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
)

var (
	instanceSchema = map[string]string{
		"IdInstance":  "instance.id",
		"Environment": "instance.environment",
		"SWIDTag":     "instance.product",
		"IdEquipment": "instance.equipment",
	}
)

func loadInstances(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Mutation, doneChan <-chan struct{}) {
	loadStaticTypes("instances", "IdInstance", ml, dg, masterDir, scopes, filenames, ch, doneChan, instancesForRow)
}

func instancesForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, string) {
	//	nodeType := "product"
	nquads := make([]*api.NQuad, 0, len(row)+3)
	//	instUID := uidForXid("inst_" + row[xidIDX])
	instUID, nqs := uidForXIDForType("inst_"+row[xidIDX], "product", "product.swidtag", row[xidIDX])
	nquads = append(nquads, nqs...)
	for i := 0; i < len(row); i++ {
		// if i == len(row) {
		// 	nquads = append(nquads, &api.NQuad{
		// 		Subject:     instUID,
		// 		Predicate:   "type",
		// 		ObjectValue: stringObjectValue("instance"),
		// 	})
		// 	return nquads, instUID
		// }

		predicate, ok := instanceSchema[cols[i]]
		if !ok {
			val, ok := applicationSchema[cols[i]]
			if ok && val == "application.id" {
				predicate = val
			} else {
				// if we cannot find predicate in map set predicate to
				// csv coloumn name
				predicate = cols[i]
			}
		}
		//log.Println(predicate)
		switch predicate {
		case "instance.product":
			if row[i] == "" {
				continue
			}
			// make a new node of type product
			//	uid := uidForXid(row[i])
			//log.Println(row[xidIDX], row[i], prodUID, uid)
			uid, nqs := uidForXIDForType(row[i], "product", "product.swidtag", row[i])
			nquads = append(nquads, nqs...)
			// nquads = append(nquads, &api.NQuad{
			// 	Subject:     uid,
			// 	Predicate:   "type",
			// 	ObjectValue: stringObjectValue("product"),
			// })
			// // assign XID to node
			// nquads = append(nquads, &api.NQuad{
			// 	Subject:     uid,
			// 	Predicate:   "product.swidtag",
			// 	ObjectValue: stringObjectValue(row[i]),
			// })
			// link both nodes
			nquads = append(nquads, &api.NQuad{
				Subject:   instUID,
				Predicate: predicate,
				ObjectId:  uid,
			})
		case "application.id":
			if row[i] == "" {
				continue
			}
			// make a new node of type product
			//uid := uidForXid("app_" + row[i])
			//log.Println(row[xidIDX], row[i], prodUID, uid)
			uid, nqs := uidForXIDForType("app_"+row[i], "application", "application.id", row[i])
			nquads = append(nquads, nqs...)
			nquads = append(nquads, &api.NQuad{
				Subject:     uid,
				Predicate:   "type",
				ObjectValue: stringObjectValue("application"),
			})
			// assign XID to node
			nquads = append(nquads, &api.NQuad{
				Subject:     uid,
				Predicate:   "application.id",
				ObjectValue: stringObjectValue(row[i]),
			})
			// link bot nodes
			nquads = append(nquads, &api.NQuad{
				Subject:   uid,
				Predicate: "application.instance",
				ObjectId:  instUID,
			})
			nquads = append(nquads, &api.NQuad{
				Subject:     instUID,
				Predicate:   "instance.id",
				ObjectValue: stringObjectValue(row[xidIDX]),
			})
		case "instance.equipment":
			//	uid := uidForXid(row[i])
			uid, nqs := uidForXIDForType(row[i], "equipment", "equipment.id", row[i])
			nquads = append(nquads, nqs...)
			// nquads = append(nquads, &api.NQuad{
			// 	Subject:     uid,
			// 	Predicate:   "type",
			// 	ObjectValue: stringObjectValue("equipment"),
			// })
			// nquads = append(nquads, &api.NQuad{
			// 	Subject:     uid,
			// 	Predicate:   "equipment.id",
			// 	ObjectValue: stringObjectValue(row[i]),
			// })
			nquads = append(nquads, &api.NQuad{
				Subject:   instUID,
				Predicate: "instance.equipment",
				ObjectId:  uid,
			})
		default:
			nquads = append(nquads, &api.NQuad{
				Subject:     instUID,
				Predicate:   predicate,
				ObjectValue: stringObjectValue(row[i]),
			})
		}
	}
	return nquads, instUID
}
