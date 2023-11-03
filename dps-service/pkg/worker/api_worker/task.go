package apiworker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	application "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/application-service/pkg/api/v1"

	product "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/product-service/pkg/api/v1"

	equipment "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/equipment-service/pkg/api/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/worker/constants"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/worker/models"

	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	dataToRPCMappings = make(map[string]map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
)

func init() {
	dataToRPCMappings[constants.APPLICATIONS] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.ApplicationsInstances] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.InstancesProducts] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.ApplicationEquipments] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.PRODUCTS] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.ApplicationsProducts] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.ProductsEquipments] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.ProductsAcquiredRights] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.METADATA] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)
	dataToRPCMappings[constants.EQUIPMENTS] = make(map[string]func(context.Context, models.Envlope, grpc.ClientConnInterface) error)

	dataToRPCMappings[constants.METADATA][constants.UPSERT] = sendUpsertMetaDataReq
	dataToRPCMappings[constants.EQUIPMENTS][constants.UPSERT] = sendUpsertEqDataReq
	dataToRPCMappings[constants.EQUIPMENTS][constants.DROP] = sendDropEquipmentDataReq
	dataToRPCMappings[constants.APPLICATIONS][constants.UPSERT] = sendUpsertApplicationReq
	dataToRPCMappings[constants.APPLICATIONS][constants.DELETE] = sendDeleteApplicationReq
	dataToRPCMappings[constants.APPLICATIONS][constants.DROP] = sendDropApplicationDataReq
	dataToRPCMappings[constants.ApplicationsInstances][constants.UPSERT] = sendUpsertInstanceReq
	dataToRPCMappings[constants.ApplicationsInstances][constants.DELETE] = sendDeleteInstanceReq
	dataToRPCMappings[constants.InstancesProducts][constants.UPSERT] = sendUpsertInstanceReq
	dataToRPCMappings[constants.ApplicationEquipments][constants.UPSERT] = sendUpsertApplicationEquipReq
	dataToRPCMappings[constants.PRODUCTS][constants.UPSERT] = sendUpsertProductReq
	dataToRPCMappings[constants.PRODUCTS][constants.DROP] = sendDropProductDataReq
	dataToRPCMappings[constants.ApplicationsProducts][constants.UPSERT] = sendUpsertProductReq
	dataToRPCMappings[constants.ProductsEquipments][constants.UPSERT] = sendUpsertProductReq
	dataToRPCMappings[constants.ProductsAcquiredRights][constants.UPSERT] = sendUpsertAcqRightsReq
}

var rpcTimeOut time.Duration

// setRpcTimeOut sets rpctimeout default  3000 milisecond ()
func setRpcTimeOut(t time.Duration) {
	rpcTimeOut = 3000
	if t > 0 {
		rpcTimeOut = t
	}
}

// GetDataCountInPayload ...
func GetDataCountInPayload(data []byte, fileType string) int32 {
	var count int32
	switch fileType {
	case constants.ApplicationsProducts:
		var temp product.UpsertProductRequest
		err := json.Unmarshal(data, &temp)
		if err != nil {
			log.Println("Failed to unmarshal for success/failed count calculation , err:", err)
			return count
		}
		count = int32(len(temp.Applications.ApplicationId))

	case constants.ProductsEquipments:
		var temp product.UpsertProductRequest
		err := json.Unmarshal(data, &temp)
		if err != nil {
			log.Println("Failed to unmarshal for success/failed count calculation , err:", err)
			return count
		}

		count = int32(len(temp.Equipments.Equipmentusers))
	case constants.InstancesProducts:
		var temp application.UpsertInstanceRequest
		err := json.Unmarshal(data, &temp)
		if err != nil {
			log.Println("Failed to unmarshal for success/failed count calculation , err:", err)
			return count
		}
		count = int32(len(temp.Products.ProductId))

	case constants.ApplicationEquipments:
		var temp application.UpsertApplicationEquipRequest
		err := json.Unmarshal(data, &temp)
		if err != nil {
			log.Println("Failed to unmarshal for success/failed count calculation , err:", err)
			return count
		}
		count = int32(len(temp.Equipments.EquipmentId))

	default:
		count = 1
	}

	return count
}

func sendUpsertEqDataReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	appData := equipment.UpsertEquipmentRequest{}

	err = jsonpb.Unmarshal(bytes.NewReader(data.Data), &appData)
	if err != nil {
		log.Println("Failed to marshal data ")
		return
	}
	resp, err := equipment.NewEquipmentServiceClient(cc).UpsertEquipment(ctx, &appData)
	if err != nil {
		log.Println("FAILED sendUpsertEqDataReq err :", err, " for data ", appData)
		return err
	}
	if resp.Success == false {
		log.Println("FAILEDTOSEND")
	}
	return
}
func sendUpsertMetaDataReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	appData := equipment.UpsertMetadataRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		log.Println("Failed to marshal data ")
		return
	}
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
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	appData := product.UpsertAcqRightsRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to marshal data ", "err", err.Error())
		return
	}
	appData.Ppid = fmt.Sprintf("%v", data.GlobalFileID)
	resp, err := product.NewProductServiceClient(cc).UpsertAcqRights(ctx, &appData)
	if err != nil {
		logger.Log.Sugar().Errorf("FAILED sendUpsertAcqRightsReq err :", "err", err, "for data ", appData)
		return err
	}
	if !resp.Success {
		logger.Log.Sugar().Errorf("FAILEDTOSEND")
	}
	return
}

func sendUpsertProductReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	appData := product.UpsertProductRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to marshal data ", "err", err.Error())
		return
	}
	appData.Ppid = fmt.Sprintf("%v", data.GlobalFileID)
	resp, err := product.NewProductServiceClient(cc).UpsertProduct(ctx, &appData)
	if err != nil {
		logger.Log.Sugar().Errorf("FAILED sendUpsertProductReq err :", "err", err, " for data ", appData)
		return err
	}
	if !resp.Success {
		logger.Log.Sugar().Errorf("FAILEDTOSEND")
	}
	return
}

func sendUpsertApplicationReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	appData := application.UpsertApplicationRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		log.Println("Failed to marshal data ")
		return status.Error(codes.Internal, "ParsingError")
	}
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

func sendUpsertApplicationEquipReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	appData := application.UpsertApplicationEquipRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		log.Println("Failed to marshal data ")
		return status.Error(codes.Internal, "ParsingError")
	}
	resp, err := application.NewApplicationServiceClient(cc).UpsertApplicationEquip(ctx, &appData)
	if err != nil {
		log.Println("FAILED sendUpsertApplicationEquipReq err :", err, " for data ", appData)
		return err
	}

	if resp.Success == false {
		log.Println("FAILEDTOSEND")
	}
	return
}

func sendDeleteApplicationReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	appData := application.DeleteApplicationRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		return
	}
	//log.Println("DEBUG sending data  to service %s  is [%+v]", data.TargetService, appData)
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
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
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

func sendDeleteApplicationsReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	appData := application.DropApplicationDataRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		return
	}
	// log.Printf("DEBUG sending data  to service %s  is [%+v]", data.TargetService, appData)
	resp, err := application.NewApplicationServiceClient(cc).DropApplicationData(ctx, &appData)
	if err != nil {
		log.Println("FAILED sendDeleteApplicationsReq err :", err, " for data ", appData)
		return err
	}
	if resp.Success == false {
		log.Println("FAILEDTOSEND")
	}
	return
}

func sendUpsertInstanceReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	appData := application.UpsertInstanceRequest{}
	err = json.Unmarshal(data.Data, &appData)
	if err != nil {
		return
	}
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

func sendDropApplicationDataReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	appReq := &application.DropApplicationDataRequest{}
	err = json.Unmarshal(data.Data, &appReq)
	if err != nil {
		return
	}
	resp, err := application.NewApplicationServiceClient(cc).DropApplicationData(ctx, appReq)
	if err != nil {
		log.Println("FAILED sendDropApplicationDataReq err :", err, " for data ", appReq)
		return err
	}
	if !resp.Success {
		log.Println("FAILEDTOSEND")
	}
	return
}

func sendDropEquipmentDataReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	equipReq := &equipment.DropEquipmentDataRequest{}
	err = json.Unmarshal(data.Data, &equipReq)
	if err != nil {
		return
	}
	resp, err := equipment.NewEquipmentServiceClient(cc).DropEquipmentData(ctx, equipReq)
	if err != nil {
		log.Println("FAILED sendDropEquipmentDataReq err :", err, " for data ", equipReq)
		return err
	}
	if !resp.Success {
		log.Println("FAILEDTOSEND")
	}
	return
}

func sendDropProductDataReq(ctx context.Context, data models.Envlope, cc grpc.ClientConnInterface) (err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*rpcTimeOut))
	defer cancel()
	prodReq := &product.DropProductDataRequest{}
	err = json.Unmarshal(data.Data, &prodReq)
	if err != nil {
		logger.Log.Sugar().Errorf("unmarshalling error", "err", err.Error())
		return
	}
	prodReq.Ppid = fmt.Sprintf("%v", data.GlobalFileID)
	resp, err := product.NewProductServiceClient(cc).DropProductData(ctx, prodReq)
	if err != nil {
		logger.Log.Sugar().Errorf("FAILED sendDropProductDataReq err :", "err", err.Error(), " for data ", prodReq)
		return err
	}
	if !resp.Success {
		logger.Log.Sugar().Errorf("FAILEDTOSEND sendDropProductDataReq")
	}
	return
}
