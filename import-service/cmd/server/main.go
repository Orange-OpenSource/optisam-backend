// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package main

import (
	"fmt"
	"os"

	"optisam-backend/import-service/pkg/cmd"
)

//go:generate mockgen -destination=../../pkg/service/v1/mock/dps_mock.go -package=mock optisam-backend/dps-service/pkg/api/v1 DpsServiceClient
func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
