// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package loader

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/equipment-service/pkg/repository/v1"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

type equipLoader struct {
	eqTypes []*v1.EquipmentType
}

// LoadEquipments ...
func loadEquipments(ml *MasterLoader, ch chan<- *api.Request, masterDir string, scopes []string, files []string, eqTypes []*v1.EquipmentType, doneChan <-chan struct{}) {
	log.Println(files, len(eqTypes))
	for _, scope := range scopes {
		scopeLoader := ml.GetLoader(filepath.Base(scope))
		for _, file := range files {
			for _, et := range eqTypes {
				if et.SourceName == filepath.Base(file) {
					fileLoader := scopeLoader.GetLoader(masterDir, filepath.Base(file))
					func(fl *FileLoader, f, s string, eqType *v1.EquipmentType) {
						load := func(version string) (time.Time, error) {
							return loadEquipmentFile(fl, ch, masterDir, s, version, f, eqType, doneChan)
						}
						fl.SetLoaderFunc(load)

					}(fileLoader, file, scope, et)
				}
			}
		}
	}
}

func loadEquipmentFile(l Loader, ch chan<- *api.Request, masterDir, scope, version string, filename string, eqType *v1.EquipmentType, doneChan <-chan struct{}) (retUpdatedOn time.Time, retErr error) {
	log.Printf("start equip loading %s: %s \n", eqType.Type, filename)
	defer log.Printf("end equip loading %s: %s \n", eqType.Type, filename)
	updatedOn := l.UpdatedOn()
	filename = filepath.Join(masterDir, scope, version, filename)
	f, err := readFile(filename)
	if err != nil {
		logger.Log.Error("error opening file", zap.String("filename:", filename), zap.String("reason", err.Error()))
		return updatedOn, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = ';'
	columns, err := r.Read()
	if err == io.EOF {
		return updatedOn, err
	} else if err != nil {
		logger.Log.Error("error reading header ", zap.String("filename:", filename), zap.String("reason", err.Error()))
		return updatedOn, err
	}

	log.Println(columns)

	// find primary key index
	attr, err := eqType.PrimaryKeyAttribute()
	if err != nil {
		logger.Log.Error("cannot find primary key attribute ", zap.String("filename:", filename), zap.String("reason", err.Error()))
		return updatedOn, err
	}

	index := findColoumnIdx(attr.MappedTo, columns)

	if index < 0 {
		logger.Log.Error("cannot find xid", zap.String("filename:", filename), zap.String("XID_col", attr.MappedTo))
		return updatedOn, err
	}

	attrMap := make(map[int]*v1.Attribute)
	for _, attr := range eqType.Attributes {
		index := findColoumnIdx(attr.MappedTo, columns)
		if index < 0 {
			logger.Log.Error("cannot find index ", zap.String("mapped _to", attr.MappedTo), zap.String("filename:", filename), zap.String("XID_col", attr.MappedTo))
			continue
		}
		attrMap[index] = attr
	}

	for idx := range columns {
		if idx == index {
			continue
		}
		_, ok := attrMap[idx]
		if !ok {
			logger.Log.Info("no mapping is found for csv coloumn", zap.String("csv_coloumn", columns[index]), zap.String("filename:", filename), zap.String("col", attr.MappedTo))
		}
	}

	mu := &api.Mutation{
		//CommitNow: true,
		Set: make([]*api.NQuad, 0, 8000),
	}

	fc := floatConverter{}
	ic := intConverter{}
	updatedIdx := findColoumnIdx(updatedColumnName, columns)
	createdIdx := findColoumnIdx(createdColumnName, columns)
	maxUpdated := time.Time{}
	defer func() {
		if maxUpdated.After(updatedOn) {
			updatedOn = maxUpdated
			retUpdatedOn = maxUpdated
		}
	}()
	rowNum := 0
	upserts := make(map[string]string)
	for {
		rowNum++
		row, err := r.Read()
		if err != nil && err != io.EOF {
			logger.Log.Error("error reading file", zap.String("filename:", filename), zap.String("reason", err.Error()))
			return updatedOn, err
		}
		if len(row) < index+1 {
			if err == io.EOF {
				break
			}
			logger.Log.Error("pk index is not in row ", zap.String("filename:", filename), zap.Strings("row", row))
			continue
		}
		row[index] = strings.TrimSpace(row[index])
		if row[index] == "" {
			logger.Log.Error("primary key is empty skipping row", zap.Int("row", rowNum), zap.String("filename:", filename), zap.String("xidClm", columns[index]))
			continue
		}

		t, err := isRowDirty(row, updatedIdx, createdIdx)
		if err != nil {
			// We got an error while checking the state we must consider this as dirty state
			logger.Log.Error("error while checking state", zap.Error(err), zap.String("file", filename))
		}

		if err == nil && l.CurrentState() != LoaderStateCreated && !t.After(updatedOn) {
			continue
		}

		if t.After(maxUpdated) {
			maxUpdated = t
		}

		//	uid := uidForXid(row[index])
		uid, nqs, upsert := uidForXIDForType(row[index], "equipment", "equipment.id", row[index], dgraphTypeEquipment, dgraphType("Equipment"+eqType.Type))
		upserts[uid] = upsert
		mu.Set = append(mu.Set, nqs...)
		mu.Set = append(mu.Set, scopeNquad(scope, uid)...)
		mu.Set = append(mu.Set,
			// &api.NQuad{
			// 	Subject:     uid,
			// 	Predicate:   "equipment.id",
			// 	ObjectValue: stringObjectValue(row[index]),
			// },
			&api.NQuad{
				Subject:     uid,
				Predicate:   "equipment.type",
				ObjectValue: stringObjectValue(eqType.Type),
			},
			// &api.NQuad{
			// 	Subject:     uid,
			// 	Predicate:   "type_name",
			// 	ObjectValue: stringObjectValue("equipment"),
			// },
		)

		for idx := range row {
			row[idx] = strings.TrimSpace(row[idx])
			if row[idx] == "" {
				continue
			}

			attr, ok := attrMap[idx]
			if !ok {
				//log.Println(columns[idx])
				continue // this is not mapped
			}

			if attr.IsIdentifier {
				continue // we already handled this case
			}

			if attr.IsParentIdentifier {
				parentUID, nqs, parentUpsert := uidForXIDForType(row[idx], "equipment", "equipment.id", row[idx], dgraphTypeEquipment, dgraphType("Equipment"+eqType.ParentType))
				upserts[parentUID] = parentUpsert
				mu.Set = append(mu.Set, nqs...)
				mu.Set = append(mu.Set,
					// &api.NQuad{
					// 	Subject:     parentUID,
					// 	Predicate:   "equipment.id",
					// 	ObjectValue: stringObjectValue(row[idx]),
					// },
					// &api.NQuad{
					// 	Subject:     parentUID,
					// 	Predicate:   "type_name",
					// 	ObjectValue: stringObjectValue("equipment"),
					// },
					&api.NQuad{
						Subject:   uid,
						Predicate: "equipment.parent",
						ObjectId:  parentUID,
					},
				)
				continue
			}
			switch attr.Type {
			case v1.DataTypeString:
				mu.Set = append(mu.Set,
					&api.NQuad{
						Subject:     uid,
						Predicate:   "equipment." + eqType.Type + "." + attr.Name,
						ObjectValue: stringObjectValue(row[idx]),
					},
				)
			case v1.DataTypeFloat:
				val, err := fc.convert(row[idx])
				if err != nil {
					logger.Log.Error("error converting data ", zap.String("filename:", filename), zap.String("header", columns[idx]), zap.String("col_val", row[idx]), zap.String("reason", err.Error()))
					mu.Set = append(mu.Set,
						&api.NQuad{
							Subject:     uid,
							Predicate:   "equipment." + eqType.Type + "." + attr.Name + "." + "failure",
							ObjectValue: val,
						},
					)
				}
				mu.Set = append(mu.Set,
					&api.NQuad{
						Subject:     uid,
						Predicate:   "equipment." + eqType.Type + "." + attr.Name,
						ObjectValue: val,
					},
				)
			case v1.DataTypeInt:
				val, err := ic.convert(row[idx])
				if err != nil {
					logger.Log.Error("error converting data ", zap.String("filename:", filename), zap.String("header", columns[idx]), zap.String("col_val", row[idx]), zap.String("reason", err.Error()))
					mu.Set = append(mu.Set,
						&api.NQuad{
							Subject:     uid,
							Predicate:   "equipment." + eqType.Type + "." + attr.Name + "." + "failure",
							ObjectValue: val,
						},
					)
				}
				mu.Set = append(mu.Set,
					&api.NQuad{
						Subject:     uid,
						Predicate:   "equipment." + eqType.Type + "." + attr.Name,
						ObjectValue: val,
					},
				)
			default:
				logger.Log.Error("unknown data type", zap.String("filename:", filename), zap.String("header", columns[idx]), zap.String("col_val", row[idx]), zap.String("dataType", attr.Type.String()))
			}
		}

		if err == io.EOF {
			// If err is equal to end of file error we must break the loop as all the data is ended.
			break
		}

		if len(mu.Set) < batchSize {
			continue
		}

		select {
		case <-doneChan:
			return updatedOn, errors.New("loader stopped all records not processed")
		case ch <- &api.Request{
			Query:     upsertQueries(upserts),
			Mutations: []*api.Mutation{mu},
		}:
		}
		upserts = make(map[string]string)
		mu = &api.Mutation{
			//CommitNow: true,
		}

	}
	if len(mu.Set) == 0 {
		return updatedOn, nil
	}
	select {
	case <-doneChan:
		if len(mu.Set) != 0 {
			return updatedOn, errors.New("file processing is not complete after eof")
		}
		return updatedOn, nil
	case ch <- &api.Request{
		Query:     upsertQueries(upserts),
		Mutations: []*api.Mutation{mu},
	}:
	}
	return updatedOn, nil
}

func findColoumnIdx(column string, columns []string) int {
	for idx := range columns {
		if columns[idx] == column {
			return idx
		}
	}
	return -1
}

func upsertQueries(upserts map[string]string) string {
	querySlice := make([]string, 0, len(upserts)+2)
	querySlice = append(querySlice, "query {")
	for _, v := range upserts {
		querySlice = append(querySlice, v)
	}
	querySlice = append(querySlice, "}")
	return strings.Join(querySlice, "\n")
}
