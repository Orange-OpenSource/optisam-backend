package main

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"optisam-backend/license-service/pkg/cmd"
)

//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --go_out=paths=source_relative:../../pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../pkg/api/v1 license.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --grpc-gateway_out=paths=source_relative:../../pkg/api/v1 license.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --openapiv2_out=logtostderr=true,json_names_for_fields=false:../../api/swagger/v1 license.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --validate_out=lang=go,paths=source_relative:../../pkg/api/v1 license.proto
//go:generate mockgen -destination=../../pkg/api/v1/mock/mock.go -package=mock optisam-backend/license-service/pkg/api/v1 LicenseServiceClient

func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
