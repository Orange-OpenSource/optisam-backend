package main

import (
	"fmt"
	"os"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/cmd"
)

//go:generate protoc --proto_path=../../thirdparty/account-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/account-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/account-service/pkg/api/v1 account.proto
//go:generate protoc --proto_path=../../thirdparty/account-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/account-service/pkg/api/v1 account.proto
//go:generate protoc --proto_path=../../thirdparty/account-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/account-service/pkg/api/v1 account.proto
//go:generate mockgen -destination=../../thirdparty/account-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/account-service/pkg/api/v1/account_grpc.pb.go AccountServiceClient

//go:generate protoc --proto_path=../../thirdparty/notification-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/notification-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/notification-service/pkg/api/v1 notification.proto
//go:generate protoc --proto_path=../../thirdparty/notification-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/notification-service/pkg/api/v1 notification.proto
//go:generate protoc --proto_path=../../thirdparty/notification-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/notification-service/pkg/api/v1 notification.proto
//go:generate mockgen -destination=../../thirdparty/notification-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/notification-service/pkg/api/v1/notification_grpc.pb.go NotificationServiceClient

//go:generate protoc --proto_path=../../thirdparty/catalog-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/catalog-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/catalog-service/pkg/api/v1 catalog.proto
//go:generate protoc --proto_path=../../thirdparty/catalog-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/catalog-service/pkg/api/v1 catalog.proto
//go:generate protoc --proto_path=../../thirdparty/catalog-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/catalog-service/pkg/api/v1 catalog.proto
//go:generate mockgen -destination=../../thirdparty/catalog-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/catalog-service/pkg/api/v1/catalog_grpc.pb.go CatalogServiceClient

//go:generate protoc --proto_path=../../thirdparty/dps-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/dps-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/dps-service/pkg/api/v1 dps.proto
//go:generate protoc --proto_path=../../thirdparty/dps-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/dps-service/pkg/api/v1 dps.proto
//go:generate protoc --proto_path=../../thirdparty/dps-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/dps-service/pkg/api/v1 dps.proto
//go:generate mockgen -destination=../../thirdparty/dps-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/dps-service/pkg/api/v1/dps_grpc.pb.go DpsServiceClient

//go:generate protoc --proto_path=../../thirdparty/product-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/product-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/product-service/pkg/api/v1 product.proto
//go:generate protoc --proto_path=../../thirdparty/product-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/product-service/pkg/api/v1 product.proto
//go:generate protoc --proto_path=../../thirdparty/product-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/product-service/pkg/api/v1 product.proto
//go:generate mockgen -destination=../../thirdparty/product-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/product-service/pkg/api/v1/product_grpc.pb.go ProductServiceClient

//go:generate protoc --proto_path=../../thirdparty/simulation-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/simulation-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/simulation-service/pkg/api/v1 simulation.proto
//go:generate protoc --proto_path=../../thirdparty/simulation-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/simulation-service/pkg/api/v1 simulation.proto
//go:generate protoc --proto_path=../../thirdparty/simulation-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/simulation-service/pkg/api/v1 simulation.proto
//go:generate mockgen -destination=../../thirdparty/simulation-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/simulation-service/pkg/api/v1/simulation_grpc.pb.go SimulationServiceClient
func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
