package loader

import (
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

var (
	applicationProductSchema = map[string]string{
		"IdApplication": "application.id",
		"SWIDTag":       "application.product",
	}
)

func loadApplicationProducts(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Request, doneChan <-chan struct{}) {
	loadStaticTypes("applications", "IdApplication", ml, dg, masterDir, scopes, filenames, ch, doneChan, applicationProductsNquadsForRow)
}

func applicationProductsNquadsForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, []string, []string, string, bool) {
	//	nodeType := "product"
	nquads := make([]*api.NQuad, 0, len(row)+3)
	// appUID := uidForXid("app_" + row[xidIDX])
	uids := []string{}
	upserts := []string{}
	var created, updated string
	var equipUID string
	appUID, nqs, appUpsert := uidForXIDForType("app_"+row[xidIDX], "application", "application.id", row[xidIDX], dgraphTypeApplication)
	uids = append(uids, appUID)
	upserts = append(upserts, appUpsert)
	nquads = append(nquads, nqs...)
	for i := 0; i < len(row); i++ {
		if i == xidIDX {
			continue
		}
		predicate, ok := applicationProductSchema[cols[i]]
		// log.Println(predicate)
		if !ok {
			if cols[i] == "created" || cols[i] == "updated" {
				continue
			}
			// if we cannot find predicate in map set predicate to
			// csv coloumn name
			predicate = cols[i]
		}

		switch predicate {
		case "application.product":
			if row[i] == "" {
				continue
			}
			uid, nqs, upsert := uidForXIDForType(row[i], "product", "product.swidtag", row[i], dgraphTypeProduct)
			uids = append(uids, uid)
			upserts = append(upserts, upsert)
			nquads = append(nquads, nqs...)
			equipUID = uid
		case "updated":
			updated = row[i]
		case "created":
			created = row[i]
		}
	}

	if equipUID == "" {
		return nquads, uids, upserts, appUID, false
	}
	facets := updatedAndCreatedFacets(updated, created)
	// link both nodes
	nquads = append(nquads, &api.NQuad{
		Subject:   appUID,
		Predicate: "application.product",
		ObjectId:  equipUID,
		Facets:    facets,
	})

	return nquads, uids, upserts, appUID, false
}
