package main

import (
	"fmt"
	"os"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/cmd"
)

//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../thirdparty/  --go_out=paths=source_relative:../../pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../pkg/api/v1 report.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../pkg/api/v1 report.proto
//go:generate protoc --proto_path=../../api/proto/v1  --proto_path=../../thirdparty/ --openapiv2_out=logtostderr=true,json_names_for_fields=false:../../api/swagger/v1 report.proto
//go:generate protoc --proto_path=../../api/proto/v1  --proto_path=../../thirdparty/  --validate_out=lang=go,paths=source_relative:../../pkg/api/v1 report.proto
//go:generate mockgen -destination=../../pkg/api/v1/mock/mock.go -source=../../pkg/api/v1/report_grpc.pb.go ReportServiceClient

//go:generate protoc --proto_path=../../thirdparty/license-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/license-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/license-service/pkg/api/v1 license.proto
//go:generate protoc --proto_path=../../thirdparty/license-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/license-service/pkg/api/v1 license.proto
//go:generate protoc --proto_path=../../thirdparty/license-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/license-service/pkg/api/v1 license.proto
//go:generate mockgen -destination=../../thirdparty/license-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/license-service/pkg/api/v1/license_grpc.pb.go LicenseServiceClient

//go:generate protoc --proto_path=../../thirdparty/product-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/product-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/product-service/pkg/api/v1 product.proto
//go:generate protoc --proto_path=../../thirdparty/product-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/product-service/pkg/api/v1 product.proto
//go:generate protoc --proto_path=../../thirdparty/product-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/product-service/pkg/api/v1 product.proto
//go:generate mockgen -destination=../../thirdparty/product-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/product-service/pkg/api/v1/product_grpc.pb.go ProductServiceClient

func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
