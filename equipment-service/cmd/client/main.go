package main

import (
	"context"
	"log"
	"optisam-backend/common/optisam/logger"
	middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/equipment-service/pkg/api/v1"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	err := logger.Init(-1, "")
	if err != nil {
		log.Fatalf("logger failed")
	}
	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithChainUnaryInterceptor(middleware.AddAuthNClientInterceptor("12345678"))}
	conn, err := grpc.Dial("localhost:14090", opts...)
	if err != nil {
		log.Fatalf("connection failed")
	}
	defer conn.Close()

	client := v1.NewEquipmentServiceClient(conn)
	_, err = client.UpsertEquipment(context.Background(), &v1.UpsertEquipmentRequest{
		EqType: "server",
		Scope:  "OFR",
		EqData: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"server_id":               {Kind: &structpb.Value_StringValue{StringValue: "123"}},
				"server_hostname":         {Kind: &structpb.Value_StringValue{StringValue: "SERV1"}},
				"server_processorsNumber": {Kind: &structpb.Value_StringValue{StringValue: "1"}},
				"server_coresNumber":      {Kind: &structpb.Value_StringValue{StringValue: "1"}},
				"parent_hostname":         {Kind: &structpb.Value_StringValue{StringValue: "CL1"}},
				"core_factor":             {Kind: &structpb.Value_StringValue{StringValue: "0.5"}},
				"core_per_processor":      {Kind: &structpb.Value_StringValue{StringValue: "1"}},
				"sag":                     {Kind: &structpb.Value_StringValue{StringValue: "1"}},
				"pvu":                     {Kind: &structpb.Value_StringValue{StringValue: "1"}},
				"created":                 {Kind: &structpb.Value_StringValue{StringValue: "2019-08-27T09:58:56.0260078ZA"}},
				"updated":                 {Kind: &structpb.Value_StringValue{StringValue: "2019-08-27T09:58:56.0260078Z"}},
			}},
	})
	if err != nil {
		logger.Log.Error("Upsert Equipment Failed", zap.Error(err))
	}
	// resp, err = client.UpsertEquipment(context.Background(), &v1.UpsertEquipmentRequest{
	// 	EqType: "partition",
	// 	Scope:  "France",
	// 	EqData: &structpb.Struct{
	// 		Fields: map[string]*structpb.Value{
	// 			"partition_code":     {Kind: &structpb.Value_StringValue{StringValue: "PART1"}},
	// 			"partition_hostname": {Kind: &structpb.Value_StringValue{StringValue: "PA_001"}},
	// 			"parent_id":          {Kind: &structpb.Value_StringValue{StringValue: "SERV1"}},
	// 			"created":            {Kind: &structpb.Value_StringValue{StringValue: "2019-08-27T09:58:56.0260078ZA"}},
	// 			"updated":            {Kind: &structpb.Value_StringValue{StringValue: "2019-08-27T09:58:56.0260078Z"}},
	// 		}},
	// })
	// fmt.Printf("resp %v", resp.GetSuccess())

}
