// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"log"
	"optisam-backend/common/optisam/dgraph"
	"optisam-backend/common/optisam/logger"
	"os"
	"testing"

	dgo "github.com/dgraph-io/dgo/v2"
)

var dg *dgo.Dgraph

func TestMain(m *testing.M) {
	logger.Init(-1, "")
	d, err := dgraph.NewDgraphConnection(&dgraph.Config{
		Hosts: []string{":9080"},
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	dg = d
	os.Exit(m.Run())

}
