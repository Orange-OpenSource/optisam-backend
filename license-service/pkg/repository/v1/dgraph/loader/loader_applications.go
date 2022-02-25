package loader

import (
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

var (
	applicationSchema = map[string]string{
		"IdApplication": "application.id",
		"Name":          "application.name",
		"Version":       "application.version",
		"Owner":         "application.owner",
	}
)

func loadApplications(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Request, doneChan <-chan struct{}) {
	loadStaticTypes("applications", "IdApplication", ml, dg, masterDir, scopes, filenames, ch, doneChan, applicationsNquadsForRow)
}

func applicationsNquadsForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, []string, []string, string, bool) {
	//	nodeType := "product"
	nquads := make([]*api.NQuad, 0, len(row)+3)
	// appUID := uidForXid("app_" + row[xidIDX])
	uids := []string{}
	upserts := []string{}
	appUID, nqs, upsert := uidForXIDForType("app_"+row[xidIDX], "application", "application.id", row[xidIDX], dgraphTypeApplication)
	uids = append(uids, appUID)
	upserts = append(upserts, upsert)
	nquads = append(nquads, nqs...)
	for i := 0; i < len(row); i++ {
		if i == xidIDX {
			continue
		}
		//
		predicate, ok := applicationSchema[cols[i]]
		// log.Println(predicate)
		if !ok {
			// if we cannot find predicate in map set predicate to
			// csv coloumn name
			predicate = cols[i]
		}
		nquads = append(nquads, &api.NQuad{
			Subject:     appUID,
			Predicate:   predicate,
			ObjectValue: stringObjectValue(row[i]),
		})

	}
	return nquads, uids, upserts, appUID, true
}
