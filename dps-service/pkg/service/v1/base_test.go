package v1

import (
	"optisam-backend/common/optisam/logger"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logger.Init(-1, "")
	os.Exit(m.Run())
}
