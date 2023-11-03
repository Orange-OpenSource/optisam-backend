package dgraph

import (
	"log"
	"os"
	"testing"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/dgraph"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

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
