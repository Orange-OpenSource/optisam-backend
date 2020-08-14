// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package main

import (
	"context"
	"fmt"
	"log"
	middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/equipment-service/pkg/api/v1"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/grpc"
)

func main() {
	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithChainUnaryInterceptor(middleware.AddAuthNClientInterceptor("12345678"))}
	conn, err := grpc.Dial("localhost:12090", opts...)
	if err != nil {
		log.Fatalf("connection failed")
	}
	defer conn.Close()

	client := v1.NewEquipmentServiceClient(conn)
	resp, err := client.UpsertEquipment(context.Background(), &v1.UpsertEquipmentRequest{
		EqType: "server",
		Scope:  "France",
		EqData: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"server_code":             {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
				"server_hostname":         {Kind: &structpb.Value_StringValue{StringValue: "SERV1"}},
				"server_processorsNumber": {Kind: &structpb.Value_NumberValue{NumberValue: 1}},
				"server_coresNumber":      {Kind: &structpb.Value_NumberValue{NumberValue: 1}},
				"parent_hostname":         {Kind: &structpb.Value_StringValue{StringValue: "CL1"}},
				"corefactor_oracle":       {Kind: &structpb.Value_NumberValue{NumberValue: 1}},
				"sag":                     {Kind: &structpb.Value_NumberValue{NumberValue: 1}},
				"pvu":                     {Kind: &structpb.Value_NumberValue{NumberValue: 1}},
				"created":                 {Kind: &structpb.Value_StringValue{StringValue: "2019-08-27T09:58:56.0260078ZA"}},
				"updated":                 {Kind: &structpb.Value_StringValue{StringValue: "2019-08-27T09:58:56.0260078Z"}},
			}},
	})
	fmt.Printf("resp %v", resp.GetSuccess())
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
