// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
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
