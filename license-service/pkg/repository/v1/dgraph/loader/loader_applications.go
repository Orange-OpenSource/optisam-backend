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
	applicationSchema = map[string]string{
		"IdApplication": "application.id",
		"Name":          "application.name",
		"Version":       "application.version",
		"Owner":         "application.owner",
		"IdInstance":    "application.instance",
		"SWIDTag":       "application.product",
	}
)

func loadApplications(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Mutation, doneChan <-chan struct{}) {
	loadStaticTypes("applications", "IdApplication", ml, dg, masterDir, scopes, filenames, ch, doneChan, applicationsNquadsForRow)
}

func applicationsNquadsForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, string) {
	//	nodeType := "product"
	nquads := make([]*api.NQuad, 0, len(row)+3)
	//appUID := uidForXid("app_" + row[xidIDX])
	appUID, nqs := uidForXIDForType("app_"+row[xidIDX], "application", "application.id", row[xidIDX])
	nquads = append(nquads, nqs...)
	for i := 0; i < len(row); i++ {
		// if i == len(row) {
		// 	nquads = append(nquads, &api.NQuad{
		// 		Subject:     appUID,
		// 		Predicate:   "type",
		// 		ObjectValue: stringObjectValue("application"),
		// 	})
		// 	return nquads, appUID
		// }
		predicate, ok := applicationSchema[cols[i]]
		// log.Println(predicate)
		if !ok {
			// if we cannot find predicate in map set predicate to
			// csv coloumn name
			predicate = cols[i]
		}
		switch predicate {
		case "application.product":
			if row[i] == "" {
				continue
			}
			// make a new node of type product
			//	uid := uidForXid(row[i])
			uid, nqs := uidForXIDForType(row[i], "product", "product.swidtag", row[i])
			nquads = append(nquads, nqs...)
			//log.Println(row[xidIDX], row[i], prodUID, uid)
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
				Subject:   appUID,
				Predicate: predicate,
				ObjectId:  uid,
			})

		default:
			nquads = append(nquads, &api.NQuad{
				Subject:     appUID,
				Predicate:   predicate,
				ObjectValue: stringObjectValue(row[i]),
			})
		}
	}
	return nquads, appUID
}
