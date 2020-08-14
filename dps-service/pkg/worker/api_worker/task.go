// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package apiworker

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	acq "optisam-backend/acqrights-service/pkg/api/v1"
	application "optisam-backend/application-service/pkg/api/v1"
	"optisam-backend/dps-service/pkg/worker/constants"
	"optisam-backend/dps-service/pkg/worker/models"
	equipment "optisam-backend/equipment-service/pkg/api/v1"
	product "optisam-backend/product-service/pkg/api/v1"

	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
)

var (
	dataToRPCMappings = make(map[string]map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
)

func init() {
	dataToRPCMappings[constants.APPLICATIONS] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.APPLICATIONS_INSTANCES] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.INSTANCES_PRODUCTS] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.INSTANCES_EQUIPMENTS] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.PRODUCTS] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.APPLICATIONS_PRODUCTS] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.PRODUCTS_EQUIPMENTS] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.PRODUCTS_ACQUIREDRIGHTS] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.METADATA] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.EQUIPMENTS] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)

	dataToRPCMappings[constants.METADATA][constants.UPSERT] = sendUpsertMetaDataReq
	dataToRPCMappings[constants.EQUIPMENTS][constants.UPSERT] = sendUpsertEqDataReq
	dataToRPCMappings[constants.APPLICATIONS][constants.UPSERT] = sendUpsertApplicationReq
	dataToRPCMappings[constants.APPLICATIONS][constants.DELETE] = sendDeleteApplicationReq
	dataToRPCMappings[constants.APPLICATIONS_INSTANCES][constants.UPSERT] = sendUpsertInstanceReq
	dataToRPCMappings[constants.APPLICATIONS_INSTANCES][constants.DELETE] = sendDeleteInstanceReq
	dataToRPCMappings[constants.INSTANCES_PRODUCTS][constants.UPSERT] = sendUpsertInstanceReq
	dataToRPCMappings[constants.INSTANCES_EQUIPMENTS][constants.UPSERT] = sendUpsertInstanceReq
	dataToRPCMappings[constants.PRODUCTS][constants.UPSERT] = sendUpsertProductReq
	dataToRPCMappings[constants.APPLICATIONS_PRODUCTS][constants.UPSERT] = sendUpsertProductReq
	dataToRPCMappings[constants.PRODUCTS_EQUIPMENTS][constants.UPSERT] = sendUpsertProductReq
	dataToRPCMappings[constants.PRODUCTS_ACQUIREDRIGHTS][constants.UPSERT] = sendUpsertAcqRightsReq
}

func getDataCountInPayload(data []byte, fileType string) int32 {
	var count int32
	switch fileType {
	case constants.APPLICATIONS_PRODUCTS:
		var temp product.UpsertProductRequest
		err := json.Unmarshal(data, &temp)
		if err != nil {
			log.Println("Failed to unmarshal for success/failed count calculation , err:", err)
			return count
		}
		count = int32(len(temp.Applications.ApplicationId))

	case constants.PRODUCTS_EQUIPMENTS:
		var temp product.UpsertProductRequest
		err := json.Unmarshal(data, &temp)
		if err != nil {
			log.Println("Failed to unmarshal for success/failed count calculation , err:", err)
			return count
		}

		count = int32(len(temp.Equipments.Equipmentusers))
	case constants.INSTANCES_PRODUCTS:
		var temp application.UpsertInstanceRequest
		err := json.Unmarshal(data, &temp)
		if err != nil {
			log.Println("Failed to unmarshal for success/failed count calculation , err:", err)
			return count
		}
		count = int32(len(temp.Products.ProductId))

	case constants.INSTANCES_EQUIPMENTS:
		var temp application.UpsertInstanceRequest
		err := json.Unmarshal(data, &temp)
		if err != nil {
			log.Println("Failed to unmarshal for success/failed count calculation , err:", err)
			return count
		}
		count = int32(len(temp.Equipments.EquipmentId))

	}

	return count
}

func sendUpsertEqDataReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	appData := equipment.UpsertEquipmentRequest{}

	err = jsonpb.Unmarshal(bytes.NewReader(data.Data), &appData)
	if err != nil {
		log.Println("Failed to marshal data ")
		return
	}
	// log.Printf("DEBUG sending data  to service %s  is [%+v]", data.TargetService, appData)
	resp, err := equipment.NewEquipmentServiceClient(cc).UpsertEquipment(ctx, &appData)
	if err != nil {
		log.Println("FAILED sendUpsertMetaDataReq err :", err, " for data ", appData)
		return err
	}
	if resp.Success == false {
		log.Println("FAILEDTOSEND")
	}
	return
}
func sendUpsertMetaDataReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	appData := equipment.UpsertMetadataRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		log.Println("Failed to marshal data ")
		return
	}
	// log.Printf("DEBUG sending data  to service %s  is [%+v]", data.TargetService, appData)
	resp, err := equipment.NewEquipmentServiceClient(cc).UpsertMetadata(ctx, &appData)
	if err != nil {
		log.Println("FAILED sendUpsertMetaDataReq err :", err, " for data ", appData)
		return err
	}
	if resp.Success == false {
		log.Println("FAILEDTOSEND")
	}
	return
}

func sendUpsertAcqRightsReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	appData := acq.UpsertAcqRightsRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		log.Println("Failed to marshal data ")
		return
	}
	// log.Printf("DEBUG sending data  to service %s  is [%+v]", data.TargetService, appData)
	resp, err := acq.NewAcqRightsServiceClient(cc).UpsertAcqRights(ctx, &appData)
	if err != nil {
		log.Println("FAILED sendUpsertAcqRightsReq err :", err, " for data ", appData)
		return err
	}
	if resp.Success == false {
		log.Println("FAILEDTOSEND")
	}
	return
}

func sendUpsertProductReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	appData := product.UpsertProductRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		log.Println("Failed to marshal data ")
		return
	}
	// log.Printf("DEBUG sending data  to service %s  is [%+v] ,  and %+v", data.TargetService, appData, cc)
	resp, err := product.NewProductServiceClient(cc).UpsertProduct(ctx, &appData)
	if err != nil {
		log.Println("FAILED sendUpsertProductReq err :", err, " for data ", appData)
		return err
	}
	if resp.Success == false {
		log.Println("FAILEDTOSEND")
	}
	return
}

func sendUpsertApplicationReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	appData := application.UpsertApplicationRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		log.Println("Failed to marshal data ")
		return
	}
	// log.Printf("DEBUG sending data  to service %s  is [%+v]", data.TargetService, appData)
	resp, err := application.NewApplicationServiceClient(cc).UpsertApplication(ctx, &appData)
	if err != nil {
		log.Println("FAILED sendUpsertApplicationReq err :", err, " for data ", appData)
		return err
	}
	if resp.Success == false {
		log.Println("FAILEDTOSEND")
	}
	return
}

func sendDeleteApplicationReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	appData := application.DeleteApplicationRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		return
	}
	// log.Printf("DEBUG sending data  to service %s  is [%+v]", data.TargetService, appData)
	resp, err := application.NewApplicationServiceClient(cc).DeleteApplication(ctx, &appData)
	if err != nil {
		log.Println("FAILED sendDeleteApplicationReq err :", err, " for data ", appData)
		return err
	}
	if resp.Success == false {
		log.Println("FAILEDTOSEND")
	}
	return
}

func sendDeleteInstanceReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	appData := application.DeleteInstanceRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		return
	}
	// log.Printf("DEBUG sending data  to service %s  is [%+v]", data.TargetService, appData)
	resp, err := application.NewApplicationServiceClient(cc).DeleteInstance(ctx, &appData)
	if err != nil {
		log.Println("FAILED sendDeleteInstanceReq err :", err, " for data ", appData)
		return err
	}
	if resp.Success == false {
		log.Println("FAILEDTOSEND")
	}
	return
}

func sendUpsertInstanceReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	appData := application.UpsertInstanceRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		return
	}
	// log.Printf("DEBUG sending data  to service %s  is [%+v]", data.TargetService, appData)
	resp, err := application.NewApplicationServiceClient(cc).UpsertInstance(ctx, &appData)
	if err != nil {
		log.Println("FAILED sendUpsertInstanceReq :", err, " for data ", appData)
		return err
	}
	if resp.Success == false {
		log.Println("FAILEDTOSEND")
	}
	return
}
