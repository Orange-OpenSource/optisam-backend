package main

import (
	"fmt"
	"optisam-backend/import-service/pkg/cmd"
	"os"
)

//go:generate mockgen -destination=../../pkg/service/v1/mock/dps_mock.go -package=mock optisam-backend/dps-service/pkg/api/v1 DpsServiceClient
func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
