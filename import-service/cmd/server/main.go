package main

import (
	"fmt"
	"os"

	"optisam-backend/import-service/pkg/cmd"
)

func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
