// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package main

import (
	"fmt"
	"os"

	"optisam-backend/report-service/pkg/cmd"
)

//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --go_out=plugins=grpc:../../pkg/api/v1 report.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --grpc-gateway_out=logtostderr=true:../../pkg/api/v1 report.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --swagger_out=logtostderr=true:../../api/swagger/v1 report.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --validate_out=lang=go:../../pkg/api/v1 report.proto
//go:generate mockgen -destination=../../pkg/service/v1/mock/license_mock.go -package=mock optisam-backend/license-service/pkg/api/v1 LicenseServiceClient
func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
