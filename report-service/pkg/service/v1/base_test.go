package v1

import (
	"optisam-backend/common/optisam/logger"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := logger.Init(-1, ""); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
