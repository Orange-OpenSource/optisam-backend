// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package loader

import (
	"errors"
	"fmt"
	"time"
)

const (
	updatedColumnName = "updated"
	createdColumnName = "created"
)

func isRowDirty(row []string, updatedIdx, createdIdx int) (time.Time, error) {
	switch {
	case updatedIdx < 0 && createdIdx < 0:
		//return time.Time{}, errors.New("updated and creted colums are missing")
		return time.Time{}, nil
	case len(row) <= updatedIdx && len(row) <= createdIdx:
		// both created and updated records are not present we must treat this record as dirty
		return time.Time{}, errors.New("updated and created  values are missing from row")
	case (len(row) <= updatedIdx) && (createdIdx > -1 && len(row) > createdIdx):
		t, err := time.Parse(time.RFC3339, row[createdIdx])
		if err != nil {
			return time.Time{}, fmt.Errorf("cannot parse created time : %s, err: %v", row[createdIdx], err)
			// we cannot parse the time we must log an error and proceed row as dirty
		}
		return t, nil
	case (updatedIdx > -1 && len(row) > updatedIdx) && len(row) <= createdIdx:
		t, err := time.Parse(time.RFC3339, row[updatedIdx])
		if err != nil {
			return time.Time{}, fmt.Errorf("cannot parse updated time : %s, err: %v", row[updatedIdx], err)
			// we cannot parse the time we must log an error and proceed row as dirty
		}
		return t, nil
	case len(row) > updatedIdx && len(row) > createdIdx:
		//fmt.Println("$$$$$$$$$$$$", row[updatedIdx], lastUpdate.String())
		timeRow := row[updatedIdx]
		if timeRow == "" {
			// if updated column is empty we should consider created
			timeRow = row[createdIdx]
		}
		// if both created and updated columsn are present then updated should be prefered
		t, err := time.Parse(time.RFC3339, timeRow)
		if err != nil {
			return t, fmt.Errorf("cannot parse updated time : %s, err: %v", timeRow, err)
			// we cannot parse the time we must log an error and proceed row as dirty
		}
		//fmt.Println("!!!!!!!!!!!!!!!!!!!!!!! "+lastUpdate.String(), t.String(), t.After(lastUpdate))

		return t, nil

	}
	// I have no idea how we end up here  but to be on safer side lets assume this row as dirty
	return time.Time{}, nil
}

//func checkCreated()
