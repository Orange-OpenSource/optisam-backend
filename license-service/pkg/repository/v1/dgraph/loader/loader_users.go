package loader

var (
	usersSchema = map[string]string{
		"NbUsers":     "users.count",
		"IdEquipment": "equipment.users",
		"SWIDTag":     "product.users",
	}
)

// // func loadUsers(ml *MasterLoader, dg *dgo.Dgraph, masterDir string, scopes, filenames []string, ch chan<- *api.Mutation, doneChan <-chan struct{}) {
// // 	loadStaticTypes("users", "NbUsers", ml, dg, masterDir, scopes, filenames, ch, doneChan, usersForRow)
// // }

// // func usersForRow(cols []string, scope string, row []string, xidIDX int) ([]*api.NQuad, string, bool) {
// // 	//	nodeType := "product"
// // 	nquads := make([]*api.NQuad, 0, len(row)+3)

// // 	if len(row) < 3 {
// // 		return nil, "", false
// // 	}
// // 	prodUID := ""
// // 	equipUID := ""
// // 	swidTag := ""
// // 	equipID := ""
// // 	nbOfUsers := ""
// // 	for i := 0; i < len(row); i++ {

// // 		predicate, ok := usersSchema[cols[i]]
// // 		if !ok {
// // 			if cols[i] == "created" || cols[i] == "updated" {
// // 				continue
// // 			}
// // 			// if we cannot find predicate in map set predicate to
// // 			// csv coloumn name
// // 			predicate = cols[i]
// // 		}
// // 		//log.Println(predicate)
// // 		switch predicate {
// // 		case "product.users":
// // 			// make a new node of type product
// // 			swidTag = row[i]
// // 			uid, nqs := uidForXIDForType(swidTag, "product", "product.swidtag", swidTag, dgraphTypeProduct)
// // 			nquads = append(nquads, nqs...)
// // 			prodUID = uid
// // 			// nquads = append(nquads, &api.NQuad{
// // 			// 	Subject:     prodUID,
// // 			// 	Predicate:   "type_name",
// // 			// 	ObjectValue: stringObjectValue("product"),
// // 			// })
// // 			// // assign XID to node
// // 			// nquads = append(nquads, &api.NQuad{
// // 			// 	Subject:     prodUID,
// // 			// 	Predicate:   "product.swidtag",
// // 			// 	ObjectValue: stringObjectValue(row[i]),
// // 			// })
// //             // log.Println(row[xidIDX], row[i], prodUID, uid)

// // 		case "equipment.users":
// // 			// equipUID = uidForXid(row[i])
// // 			equipID = row[i]
// // 			uid, nqs := uidForXIDForType(equipID, "equipment", "equipment.id", equipID, dgraphTypeEquipment)
// // 			nquads = append(nquads, nqs...)
// // 			equipUID = uid
// // 			// nquads = append(nquads, &api.NQuad{
// // 			// 	Subject:     equipUID,
// // 			// 	Predicate:   "type_name",
// // 			// 	ObjectValue: stringObjectValue("equipment"),
// // 			// })
// // 			// nquads = append(nquads, &api.NQuad{
// // 			// 	Subject:     equipUID,
// // 			// 	Predicate:   "equipment.id",
// // 			// 	ObjectValue: stringObjectValue(row[i]),
// // 			// })
// // 		case "users.count":
// // 			nbOfUsers = row[i]
// // 		}
// // 	}
// // 	usersID := "user_" + swidTag + equipID
// // 	//usersUID := uidForXid(usersID)
// // 	usersUID, nqs := uidForXIDForType(usersID, "instance_users", "users.id", usersID, dgraphTypeUser)
// // 	nquads = append(nquads, nqs...)
// // 	// nquads = append(nquads, &api.NQuad{
// // 	// 	Subject:     usersUID,
// // 	// 	Predicate:   "type_name",
// // 	// 	ObjectValue: stringObjectValue("instance_users"),
// // 	// })
// // 	// nquads = append(nquads, &api.NQuad{
// // 	// 	Subject:     usersUID,
// // 	// 	Predicate:   "users.id",
// // 	// 	ObjectValue: stringObjectValue(usersID),
// // 	// })

// // 	nquads = append(nquads, &api.NQuad{
// // 		Subject:   prodUID,
// // 		Predicate: "product.users",
// // 		ObjectId:  usersUID,
// // 	})

// // 	nquads = append(nquads, &api.NQuad{
// // 		Subject:   equipUID,
// // 		Predicate: "equipment.users",
// // 		ObjectId:  usersUID,
// // 	})

// // 	cnv := intConverter{}
// // 	val, err := cnv.convert(nbOfUsers)
// // 	if err != nil {
// // 		log.Printf("failed to convert data for NbOfUsers: error: %v", err)
// // 		nquads = append(nquads, &api.NQuad{
// // 			Subject:     usersUID,
// // 			Predicate:   "users.count" + ".failure",
// // 			ObjectValue: defaultObjectValue(nbOfUsers),
// // 		})
// // 		return nquads, usersUID, false
// // 	}

// // 	nquads = append(nquads, &api.NQuad{
// // 		Subject:     usersUID,
// // 		Predicate:   "users.count",
// // 		ObjectValue: val,
// // 	})

// // 	return nquads, usersUID, false
// }
