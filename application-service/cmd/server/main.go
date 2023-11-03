package main

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/application-service/pkg/cmd"
)

//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../thirdparty/  --go_out=paths=source_relative:../../pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../pkg/api/v1 application.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../pkg/api/v1 application.proto
//go:generate protoc --proto_path=../../api/proto/v1  --proto_path=../../thirdparty/ --openapiv2_out=logtostderr=true,json_names_for_fields=false:../../api/swagger/v1 application.proto
//go:generate protoc --proto_path=../../api/proto/v1  --proto_path=../../thirdparty/  --validate_out=lang=go,paths=source_relative:../../pkg/api/v1 application.proto
//go:generate mockgen -destination=../../pkg/api/v1/mock/mock.go -source=../../pkg/api/v1/application_grpc.pb.go ApplicationServiceClient

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
