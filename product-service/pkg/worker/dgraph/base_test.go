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
