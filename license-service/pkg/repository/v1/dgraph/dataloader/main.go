package main

import (
	"os"

	"go.uber.org/zap"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/dgraph/dataloader/cmd"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logger.Log.Error("command failed", zap.Error(err))
		os.Exit(1)
	}
}
