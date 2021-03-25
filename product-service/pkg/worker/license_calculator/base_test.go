// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package licensecalculator

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
