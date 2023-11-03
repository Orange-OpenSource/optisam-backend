package main

import (
	"fmt"
	"os"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/cmd"
)

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
