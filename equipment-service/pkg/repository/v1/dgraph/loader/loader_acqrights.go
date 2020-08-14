// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package loader

import (
	"log"
	"strings"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

var (
	acqRightsSchema = map[string]*schemaNode{
		"Entity":                            {predName: "acqRights.entity", dataConv: stringConverter{}},
		"SKU":                               {predName: "acqRights.SKU", dataConv: stringConverter{}},
		"SWIDTag":                           {predName: "acqRights.swidtag", dataConv: stringConverter{}},
		"Product name":                      {predName: "acqRights.productName", dataConv: stringConverter{}},
		"Editor":                            {predName: "acqRights.editor", dataConv: stringConverter{}},
		"Metric":                            {predName: "acqRights.metric", dataConv: stringConverter{}},
		"Acquired licenses number":          {predName: "acqRights.numOfAcqLicences", dataConv: intConverter{}},
		"Licenses under maintenance number": {predName: "acqRights.numOfLicencesUnderMaintenance", dataConv: intConverter{}},
		"AVG Unit Price":                    {predName: "acqRights.averageUnitPrice", dataConv: floatConverter{}},
		"AVG Maintenant Unit Price":         {predName: "acqRights.averageMaintenantUnitPrice", dataConv: floatConverter{}},
		"Total purchase cost":               {predName: "acqRights.totalPurchaseCost", dataConv: floatConverter{}},
		"Total maintenance cost":            {predName: "acqRights.totalMaintenanceCost", dataConv: floatConverter{}},
		"Total cost":                        {predName: "acqRights.totalCost", dataConv: floatConverter{}},
		"updated":                           {predName: "updated", dataConv: stringConverter{}},
		"created":                           {predName: "created", dataConv: stringConverter{}},
	}
)

func loadAcquiredRights(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Request, doneChan <-chan struct{}) {
	loadStaticTypes("acquired rights", "SKU", ml, dg, masterDir, scopes, filenames, ch, doneChan, acquiredNquadsForRow)
}

func acquiredNquadsForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, []string, []string, string, bool) {
	//	nodeType := "product"
	nquads := make([]*api.NQuad, 0, len(row)+3)
	uids := []string{}
	upserts := []string{}
	//	acqRightUID := uidForXid(row[xidIDX])
	acqRightUID, nqs, acqupsert := uidForXIDForType(row[xidIDX], "acqRights", "acqRights.SKU", row[xidIDX], dgraphTypeAcquiredRights)
	uids = append(uids, acqRightUID)
	upserts = append(upserts, acqupsert)
	nquads = append(nquads, nqs...)
	for i := 0; i < len(row); i++ {
		if i == xidIDX {
			continue
		}
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
			uid, nqs, upsert := uidForXIDForType(row[i], "product", "product.swidtag", row[i], dgraphTypeProduct)
			uids = append(uids, uid)
			upserts = append(upserts, upsert)
			nquads = append(nquads, nqs...)
			//log.Println(row[xidIDX], row[i], prodUID, uid)
			// link both nodes
			nquads = append(nquads, &api.NQuad{
				Subject:     acqRightUID,
				Predicate:   predicate,
				ObjectValue: stringObjectValue(row[i]),
			})
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
	return nquads, uids, upserts, acqRightUID, true
}
