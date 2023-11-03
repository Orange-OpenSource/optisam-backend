package helper

import (
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
)

func CustomErrorHandle(logType, msg string, logParams map[string]interface{}) {

	logger := logger.Log.Sugar()

	if logType == "Errorw" {
		logger.Errorw(msg, logParams)
	}
	if logType == "Infow" {
		logger.Infow(msg, logParams)
	}
	if logType == "Debugw" {
		logger.Debugw(msg, logParams)
	}
	if logType == "Debug" {
		logger.Debug(msg, logParams)
	}
}
