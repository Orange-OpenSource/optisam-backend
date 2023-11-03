package main

import (
	"fmt"
	"os"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/cmd"
)

//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../thirdparty/ --go_out=paths=source_relative:../../pkg/api/v1 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:../../pkg/api/v1 notification.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../thirdparty/ --grpc-gateway_out=paths=source_relative:../../pkg/api/v1 notification.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../thirdparty/ --openapiv2_out=logtostderr=true,json_names_for_fields=false:../../api/swagger/v1  notification.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../thirdparty/ --validate_out=lang=go,paths=source_relative:../../pkg/api/v1 notification.proto
//go:generate mockgen -destination=../../pkg/api/v1/mock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/api/v1 NotificationServiceClient

func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
