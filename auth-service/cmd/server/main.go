package main

import (
	"fmt"
	"os"

	"optisam-backend/auth-service/pkg/cmd"
)

//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --go_out=plugins=grpc:../../pkg/api/v1 auth.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --grpc-gateway_out=logtostderr=true:../../pkg/api/v1 auth.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --swagger_out=logtostderr=true:../../api/swagger/v1 auth.proto
//go:generate protoc --proto_path=../../api/proto/v1 --proto_path=../../../common/third_party --validate_out=lang=go:../../pkg/api/v1 auth.proto
func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
