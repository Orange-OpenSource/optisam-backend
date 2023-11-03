package v1

import (
	"os"
	"testing"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
)

func TestMain(m *testing.M) {
	if err := logger.Init(-1, ""); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
