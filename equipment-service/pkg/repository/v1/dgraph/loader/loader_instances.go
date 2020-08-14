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
	instanceSchema = map[string]string{
		"IdInstance":  "instance.id",
		"Environment": "instance.environment",
	}
)

func loadInstances(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Request, doneChan <-chan struct{}) {
	loadStaticTypes("instances", "IdInstance", ml, dg, masterDir, scopes, filenames, ch, doneChan, instancesForRow)
}

func instancesForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, []string, []string, string, bool) {
	//	nodeType := "product"
	nquads := make([]*api.NQuad, 0, len(row)+3)
	uids, upserts := []string{}, []string{}
	//	instUID := uidForXid("inst_" + row[xidIDX])
	instUID, nqs, instUpsert := uidForXIDForType("inst_"+row[xidIDX], "instance", "instance.id", row[xidIDX], dgraphTypeInstance)
	uids = append(uids, instUID)
	upserts = append(upserts, instUpsert)
	nquads = append(nquads, nqs...)
	for i := 0; i < len(row); i++ {
		if i == xidIDX {
			continue
		}
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
		case "application.id":
			if row[i] == "" {
				continue
			}
			// make a new node of type product
			//uid := uidForXid("app_" + row[i])
			//log.Println(row[xidIDX], row[i], prodUID, uid)
			uid, nqs, upsert := uidForXIDForType("app_"+row[i], "application", "application.id", row[i], dgraphTypeApplication)
			uids = append(uids, uid)
			upserts = append(upserts, upsert)
			nquads = append(nquads, nqs...)
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
		default:
			nquads = append(nquads, &api.NQuad{
				Subject:     instUID,
				Predicate:   predicate,
				ObjectValue: stringObjectValue(row[i]),
			})
		}
	}
	return nquads, uids, upserts, instUID, true
}
