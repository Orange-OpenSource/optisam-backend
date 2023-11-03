package main

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/cmd"
)

//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../thirdparty/  --go_out=paths=source_relative:../../pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../pkg/api/v1 product.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../pkg/api/v1 product.proto
//go:generate protoc --proto_path=../../api/proto/v1  --proto_path=../../thirdparty/ --openapiv2_out=logtostderr=true,json_names_for_fields=false:../../api/swagger/v1 product.proto
//go:generate protoc --proto_path=../../api/proto/v1  --proto_path=../../thirdparty/  --validate_out=lang=go,paths=source_relative:../../pkg/api/v1 product.proto
//go:generate mockgen -destination=../../pkg/api/v1/mock/mock.go -source=../../pkg/api/v1/product_grpc.pb.go ProductServiceClient

//go:generate protoc --proto_path=../../thirdparty/application-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/application-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/application-service/pkg/api/v1 application.proto
//go:generate protoc --proto_path=../../thirdparty/application-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/application-service/pkg/api/v1 application.proto
//go:generate protoc --proto_path=../../thirdparty/application-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/application-service/pkg/api/v1 application.proto
//go:generate mockgen -destination=../../thirdparty/application-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/application-service/pkg/api/v1/application_grpc.pb.go ApplicationServiceClient

//go:generate protoc --proto_path=../../thirdparty/equipment-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/equipment-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/equipment-service/pkg/api/v1 equipment.proto
//go:generate protoc --proto_path=../../thirdparty/equipment-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/equipment-service/pkg/api/v1 equipment.proto
//go:generate protoc --proto_path=../../thirdparty/equipment-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/equipment-service/pkg/api/v1 equipment.proto
//go:generate mockgen -destination=../../thirdparty/equipment-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/equipment-service/pkg/api/v1/equipment_grpc.pb.go EquipmentServiceClient

//go:generate protoc --proto_path=../../thirdparty/account-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/account-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/account-service/pkg/api/v1 account.proto
//go:generate protoc --proto_path=../../thirdparty/account-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/account-service/pkg/api/v1 account.proto
//go:generate protoc --proto_path=../../thirdparty/account-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/account-service/pkg/api/v1 account.proto
//go:generate mockgen -destination=../../thirdparty/account-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/account-service/pkg/api/v1/account_grpc.pb.go AccountServiceClient

//go:generate protoc --proto_path=../../thirdparty/metric-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/metric-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/metric-service/pkg/api/v1 metric.proto
//go:generate protoc --proto_path=../../thirdparty/metric-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/metric-service/pkg/api/v1 metric.proto
//go:generate protoc --proto_path=../../thirdparty/metric-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/metric-service/pkg/api/v1 metric.proto
//go:generate mockgen -destination=../../thirdparty/metric-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/metric-service/pkg/api/v1/metric_grpc.pb.go MetricServiceClient

//go:generate protoc --proto_path=../../thirdparty/license-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/license-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/license-service/pkg/api/v1 license.proto
//go:generate protoc --proto_path=../../thirdparty/license-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/license-service/pkg/api/v1 license.proto
//go:generate protoc --proto_path=../../thirdparty/license-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/license-service/pkg/api/v1 license.proto
//go:generate mockgen -destination=../../thirdparty/license-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/license-service/pkg/api/v1/license_grpc.pb.go LicenseServiceClient

//go:generate protoc --proto_path=../../thirdparty/notification-service/proto --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../thirdparty/notification-service/pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../thirdparty/notification-service/pkg/api/v1 notification.proto
//go:generate protoc --proto_path=../../thirdparty/notification-service/proto  --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../thirdparty/notification-service/pkg/api/v1 notification.proto
//go:generate protoc --proto_path=../../thirdparty/notification-service/proto --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../thirdparty/notification-service/pkg/api/v1 notification.proto
//go:generate mockgen -destination=../../thirdparty/notification-service/pkg/api/v1/mock/mock.go -source=../../thirdparty/notification-service/pkg/api/v1/notification_grpc.pb.go NotificationServiceClient

func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
