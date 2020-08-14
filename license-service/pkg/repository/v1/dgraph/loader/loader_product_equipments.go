// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package loader

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

var (
	productEquipmentSchema = map[string]string{
		"SWIDTag":     "product.swidtag",
		"IdEquipment": "product.equipment",
		"NbUsers":     "users.count",
	}
)

func loadProductEquipments(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Request, doneChan <-chan struct{}) {
	loadStaticTypes("products", "SWIDTag", ml, dg, masterDir, scopes, filenames, ch, doneChan, productEquipmentsNquadsForRow)
}

func productEquipmentsNquadsForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, []string, []string, string, bool) {
	//	nodeType := "product"
	var updated, created string
	nquads := make([]*api.NQuad, 0, len(row)+3)
	//prodUID := uidForXid(row[xidIDX])
	swidTag := row[xidIDX]
	var equipID string
	var equipUID string
	var nbOfUsers string
	uids, upserts := []string{}, []string{}
	prodUID, nqs, prodUpsert := uidForXIDForType(row[xidIDX], "product", "product.swidtag", row[xidIDX], dgraphTypeProduct)
	uids = append(uids, prodUID)
	upserts = append(upserts, prodUpsert)
	nquads = append(nquads, nqs...)
	for i := 0; i < len(row); i++ {
		if i == xidIDX {
			continue
		}
		predicate, ok := productEquipmentSchema[cols[i]]
		if !ok {
			// if we cannot find predicate in map set predicate to
			// csv coloumn name
			predicate = cols[i]
		}
		switch predicate {
		case "product.equipment":
			if row[i] == "" {
				continue
			}
			uid, nqs, upsert := uidForXIDForType(row[i], "equipment", "equipment.id", row[i], dgraphTypeEquipment)
			uids = append(uids, uid)
			upserts = append(upserts, upsert)
			nquads = append(nquads, nqs...)
			equipID = row[i]
			equipUID = uid
			nquads = append(nquads, nqs...)

		case "users.count":
			if row[i] == "" {
				continue
			}
			nbOfUsers = row[i]
		case "updated":
			updated = row[i]
		case "created":
			created = row[i]
		default:
			nquads = append(nquads, &api.NQuad{
				Subject:     prodUID,
				Predicate:   predicate,
				ObjectValue: stringObjectValue(row[i]),
			})
		}
	}

	facets := updatedAndCreatedFacets(updated, created)

	if equipUID != "" {
		nquads = append(nquads, &api.NQuad{
			Subject:   prodUID,
			Predicate: "product.equipment",
			ObjectId:  equipUID,
			Facets:    facets,
		})
	}

	if nbOfUsers == "" {
		return nquads, uids, upserts, prodUID, false
	}

	usersID := "user_" + swidTag + "_" + equipID
	usersUID, nqs, userUpsert := uidForXIDForType(usersID, "instance_users", "users.id", usersID, dgraphTypeUser)
	uids = append(uids, usersUID)
	upserts = append(upserts, userUpsert)
	nquads = append(nquads, nqs...)

	nquads = append(nquads, &api.NQuad{
		Subject:   prodUID,
		Predicate: "product.users",
		ObjectId:  usersUID,
		Facets:    facets,
	})

	nquads = append(nquads, &api.NQuad{
		Subject:   equipUID,
		Predicate: "equipment.users",
		ObjectId:  usersUID,
		Facets:    facets,
	})

	cnv := intConverter{}
	val, err := cnv.convert(nbOfUsers)
	if err != nil {
		log.Printf("failed to convert data for NbOfUsers: error: %v", err)
		nquads = append(nquads, &api.NQuad{
			Subject:     usersUID,
			Predicate:   "users.count" + ".failure",
			ObjectValue: defaultObjectValue(nbOfUsers),
		})
		return nquads, uids, upserts, usersUID, false
	}

	nquads = append(nquads, &api.NQuad{
		Subject:     usersUID,
		Predicate:   "users.count",
		ObjectValue: val,
	})

	return nquads, uids, upserts, prodUID, false
}

func updatedAndCreatedFacets(updated, created string) []*api.Facet {
	facets := make([]*api.Facet, 0, 2)

	if updated != "" {
		if genRDF {
			facets = append(facets, &api.Facet{
				Key:     "updated",
				Value:   []byte(updated),
				ValType: api.Facet_DATETIME,
			})
		} else {
			ut, err := time.Parse(time.RFC3339, updated)
			if err != nil {
				log.Printf("updatedAndCreatedFacets - Parse updated, err:%v", err)
				facets = append(facets, &api.Facet{
					Key:     "updated_str",
					Value:   []byte(updated),
					ValType: api.Facet_STRING,
				})
			}
			utBytes, err := ut.MarshalBinary()
			if err != nil {
				log.Printf("updatedAndCreatedFacets - MarshalBinary updated, err:%v", err)
				facets = append(facets, &api.Facet{
					Key:     "updated_str",
					Value:   []byte(updated),
					ValType: api.Facet_STRING,
				})
			} else {
				facets = append(facets, &api.Facet{
					Key:     "updated",
					Value:   utBytes,
					ValType: api.Facet_DATETIME,
				})
			}
		}
	}
	if created != "" {
		if genRDF {
			facets = append(facets, &api.Facet{
				Key:     "created",
				Value:   []byte(created),
				ValType: api.Facet_DATETIME,
			})
		} else {
			ut, err := time.Parse(time.RFC3339, created)
			if err != nil {
				log.Printf("updatedAndCreatedFacets - Parse created, err:%v", err)
				facets = append(facets, &api.Facet{
					Key:     "created_str",
					Value:   []byte(updated),
					ValType: api.Facet_STRING,
				})
			}
			utBytes, err := ut.MarshalBinary()
			if err != nil {
				log.Printf("updatedAndCreatedFacets - MarshalBinary created, err:%v", err)
				facets = append(facets, &api.Facet{
					Key:     "created_str",
					Value:   []byte(updated),
					ValType: api.Facet_STRING,
				})
			} else {
				facets = append(facets, &api.Facet{
					Key:     "created",
					Value:   utBytes,
					ValType: api.Facet_DATETIME,
				})
			}
		}
	}
	return facets
}

func facetsRDF(facets []*api.Facet) string {
	keyVals := make([]string, 0, len(facets))

	for _, facet := range facets {
		switch facet.ValType {
		case api.Facet_STRING:
			keyVals = append(keyVals, fmt.Sprintf("%s=\"%s\"", facet.Key, facet.Value))
		case api.Facet_DATETIME:
			keyVals = append(keyVals, fmt.Sprintf("%s=%s", facet.Key, facet.Value))
		default:
			keyVals = append(keyVals, fmt.Sprintf("%s=\"%s\"", facet.Key, facet.Value))
		}
	}

	if len(keyVals) == 0 {
		return ""
	}
	return " ( " + strings.Join(keyVals, ",") + " ) "
}
