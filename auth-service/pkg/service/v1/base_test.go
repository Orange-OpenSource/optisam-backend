package v1

import (
	"os"
	"testing"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
)

func TestMain(m *testing.M) {
	logger.Init(-1, "")
	os.Exit(m.Run())
}
