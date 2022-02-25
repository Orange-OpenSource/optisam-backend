package main

import (
	"os"

	"go.uber.org/zap"

	"optisam-backend/common/optisam/logger"
	"optisam-backend/license-service/pkg/repository/v1/dgraph/dataloader/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logger.Log.Error("command failed", zap.Error(err))
		os.Exit(1)
	}
}
