package loader

import (
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

var (
	instanceEquipmentSchema = map[string]string{
		"IdInstance":  "instance.id",
		"IdEquipment": "instance.equipment",
	}
)

func loadInstanceEquipments(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Request, doneChan <-chan struct{}) {
	loadStaticTypes("instances", "IdInstance", ml, dg, masterDir, scopes, filenames, ch, doneChan, instanceEquipmentsForRow)
}

func instanceEquipmentsForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, []string, []string, string, bool) {
	//	nodeType := "product"
	nquads := make([]*api.NQuad, 0, len(row)+3)
	var created, updated string
	var equipUID string
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
		predicate, ok := instanceEquipmentSchema[cols[i]]
		if !ok {
			// if we cannot find predicate in map set predicate to
			// csv coloumn name
			predicate = cols[i]
		}
		// log.Println(predicate)
		switch predicate {
		case "instance.equipment":
			if row[i] == "" {
				continue
			}
			//	uid := uidForXid(row[i])
			uid, nqs, upsert := uidForXIDForType(row[i], "equipment", "equipment.id", row[i], dgraphTypeEquipment)
			uids = append(uids, uid)
			upserts = append(upserts, upsert)
			nquads = append(nquads, nqs...)
			equipUID = uid

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

	if equipUID == "" {
		return nquads, uids, upserts, instUID, false
	}

	facets := updatedAndCreatedFacets(updated, created)

	nquads = append(nquads, &api.NQuad{
		Subject:   instUID,
		Predicate: "instance.equipment",
		ObjectId:  equipUID,
		Facets:    facets,
	})
	return nquads, uids, upserts, instUID, false
}
