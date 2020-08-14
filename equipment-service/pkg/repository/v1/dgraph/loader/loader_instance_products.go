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
	instanceProductsSchema = map[string]string{
		"IdInstance": "instance.id",
		"SWIDTag":    "instance.product",
	}
)

func loadInstanceProducts(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Request, doneChan <-chan struct{}) {
	loadStaticTypes("instances", "IdInstance", ml, dg, masterDir, scopes, filenames, ch, doneChan, instanceProductsForRow)
}

func instanceProductsForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, []string, []string, string, bool) {
	//	nodeType := "product"
	nquads := make([]*api.NQuad, 0, len(row)+3)
	var updated, created string
	var prodUID string
	uids, upserts := []string{}, []string{}
	//	instUID := uidForXid("inst_" + row[xidIDX])
	instUID, nqs, instUpsert := uidForXIDForType("inst_"+row[xidIDX], "instance", "instance.id", row[xidIDX], dgraphTypeInstance)
	nquads = append(nquads, nqs...)
	uids = append(uids, instUID)
	upserts = append(upserts, instUpsert)
	for i := 0; i < len(row); i++ {
		if i == xidIDX {
			continue
		}
		predicate, ok := instanceProductsSchema[cols[i]]
		if !ok {
			// if we cannot find predicate in map set predicate to
			// csv coloumn name
			predicate = cols[i]
		}
		//log.Println(predicate)
		switch predicate {
		case "instance.product":
			if row[i] == "" {
				continue
			}
			uid, nqs, upsert := uidForXIDForType(row[i], "product", "product.swidtag", row[i], dgraphTypeProduct)
			uids = append(uids, uid)
			upserts = append(upserts, upsert)
			nquads = append(nquads, nqs...)
			prodUID = uid
		case "updated":
			updated = row[i]
		case "created":
			created = row[i]
		default:
			nquads = append(nquads, &api.NQuad{
				Subject:     instUID,
				Predicate:   predicate,
				ObjectValue: stringObjectValue(row[i]),
			})
		}
	}

	if prodUID == "" {
		return nquads, uids, upserts, instUID, false
	}

	facets := updatedAndCreatedFacets(updated, created)

	nquads = append(nquads, &api.NQuad{
		Subject:   instUID,
		Predicate: "instance.product",
		ObjectId:  prodUID,
		Facets:    facets,
	})
	return nquads, uids, upserts, instUID, false
}
