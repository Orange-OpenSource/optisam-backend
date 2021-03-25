// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package fileworker

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	//acq "optisam-backend/acqrights-service/pkg/api/v1"
	application "optisam-backend/application-service/pkg/api/v1"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	"optisam-backend/dps-service/pkg/config"
	gendb "optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	"optisam-backend/dps-service/pkg/worker/constants"
	"optisam-backend/dps-service/pkg/worker/models"
	equipment "optisam-backend/equipment-service/pkg/api/v1"
	product "optisam-backend/product-service/pkg/api/v1"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//getFileTypeFromFileName return FileType in uppercase
//It can be
func getFileTypeFromFileName(fileName, scope string) (fileType string, err error) {
	fileName = strings.ToUpper(fileName)
	sep := fmt.Sprintf("%s_", strings.ToUpper(scope))
	if !strings.Contains(fileName, sep) {
		err = status.Error(codes.Internal, "InvalidFileName")
		return
	}
	fileType = strings.Split(strings.Split(fileName, sep)[1], constants.FILE_EXTENSION)[0]
	return
}

func fileProcessing(jobData gendb.UploadedDataFile) (data models.FileData, err error) {
	var fileType string
	var expectedHeaders []string
	if jobData.FileName == "" {
		err = status.Error(codes.Internal, "MissingFileName")
		return
	}
	if strings.Contains(strings.ToUpper(jobData.FileName), constants.METADATA) {
		data, err = csvFileToSchemaData(jobData.FileName, jobData.Scope)
		if err != nil {
			data.FileFailureReason = err.Error()
			return data, status.Error(codes.Internal, data.FileFailureReason)
		}
		data.FileType = constants.METADATA
		data.TargetServices = constants.SERVICES[data.FileType]
	} else {
		fileType, err = getFileTypeFromFileName(jobData.FileName, jobData.Scope)
		if err != nil {
			data.FileFailureReason = err.Error()
			return data, status.Error(codes.Internal, data.FileFailureReason)
		}
		//For equipment, dynamic processing is required
		if strings.Contains(fileType, "EQUIPMENT_") {
			data, err = getEquipment(fileType, jobData.FileName)
			if err != nil {
				data.FileFailureReason = "InvalidFileName"
				return data, status.Error(codes.Internal, "InvalidFileName")
			}
			data.TargetServices = constants.SERVICES[constants.EQUIPMENTS]
		} else {
			expectedHeaders, err = getHeadersForFileType(fileType)
			if err != nil {
				data.FileFailureReason = err.Error()
				return data, status.Error(codes.Internal, "FileNotSupported")
			}
			data, err = csvToFileData(fileType, jobData.FileName, expectedHeaders)
			if err != nil {
				log.Println("Failed to read data from  file ", jobData.FileName, " with err ", err)
				if data.FileFailureReason == "" {
					data.FileFailureReason = "BadFile"
				}
				return
			}
		}
	}
	data.Scope = jobData.Scope
	data.FileName = jobData.FileName
	data.UploadID = jobData.UploadID
	return
}

//Headers are updated, no No space is allowed in headers and these are case insensitive
func getHeadersForFileType(fileType string) (headers []string, err error) {
	headers = []string{}
	switch fileType {
	case constants.PRODUCTS:
		headers = []string{"swidtag", "version", "category", "editor", "isoptionof", "name", "flag"}

	case constants.APPLICATIONS:
		headers = []string{"application_id", "version", "owner", "name", "domain", "flag"}

	case constants.APPLICATIONS_INSTANCES:
		headers = []string{"application_id", "instance_id", "environment", "flag"}

	case constants.APPLICATIONS_PRODUCTS:
		headers = []string{"application_id", "swidtag", "flag"}

	case constants.PRODUCTS_EQUIPMENTS:
		headers = []string{"equipment_id", "swidtag", "nbusers", "flag"}

	case constants.INSTANCES_PRODUCTS:
		headers = []string{"instance_id", "swidtag", "flag"}

	case constants.INSTANCES_EQUIPMENTS:
		headers = []string{"instance_id", "equipment_id", "flag"}

	case constants.PRODUCTS_ACQUIREDRIGHTS:
		headers = []string{"product_version", "entity", "sku", "swidtag", "product_name", "editor", "metric", "acquired_licenses", "total_license_cost", "total_maintenance_cost", "unit_price", "maintenance_unit_price", "total_cost", "flag", "maintenance_start", "maintenance_end", "maintenance_licenses"}

	default:
		err = status.Error(codes.Internal, "FileNotSupported")
	}
	return
}

func csvFileToSchemaData(fileName, scope string) (data models.FileData, err error) {
	file := fmt.Sprintf("%s/%s", config.GetConfig().FilesLocation, fileName)
	csvFile, err := os.Open(file)
	if err != nil {
		return
	}
	defer csvFile.Close()
	scanner := bufio.NewScanner(csvFile)
	if !scanner.Scan() {
		err = scanner.Err()
		return
	}
	row := scanner.Text()
	//schemaType := strings.Split(strings.Split(fileName, constants.SCOPE_DELIMETER)[2], constants.FILE_EXTENSION)[0]

	for _, val := range strings.Split(row, constants.DELIMETER) {
		data.Schema = append(data.Schema, val)
	}
	data.TotalCount++
	return
}

func getIndexOfHeaders(firstRow string, expectedHeaders []string) (headers models.HeadersInfo, err error) {
	headers.IndexesOfHeaders = make(map[string]int)
	headers.MaxIndexVal = 0
	for _, val := range expectedHeaders {
		headers.IndexesOfHeaders[strings.ToLower(val)] = -1
	}
	firstRow = strings.ToLower(firstRow)
	actualHeaders := strings.Split(firstRow, constants.DELIMETER)

	if len(headers.IndexesOfHeaders) > len(actualHeaders) {
		err = status.Error(codes.Internal, "HeadersMissing")
		return
	}

	for i, data := range actualHeaders {
		headers.IndexesOfHeaders[data] = i
	}
	for key, val := range headers.IndexesOfHeaders {
		if val == -1 {
			log.Println(" mandatory header field [ ", key, "] is missing ")
			err = status.Error(codes.Internal, "HeadersMissing")
			return
		}
		if val > headers.MaxIndexVal {
			headers.MaxIndexVal = val
		}
	}
	return
}

func getProducts(s *bufio.Scanner, headers models.HeadersInfo) (resp models.FileData, err error) {
	resp.Products = make(map[string]models.ProductInfo)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.SWIDTAG]]) > 0 {
			data := models.ProductInfo{}
			data.Name = list[headers.IndexesOfHeaders[constants.NAME]]
			data.Version = list[headers.IndexesOfHeaders[constants.VERSION]]
			data.Editor = list[headers.IndexesOfHeaders[constants.EDITOR]]
			data.IsOptionOf = list[headers.IndexesOfHeaders[constants.IS_OPTION_OF]]
			data.Category = list[headers.IndexesOfHeaders[constants.CATEGORY]]
			data.SwidTag = list[headers.IndexesOfHeaders[constants.SWIDTAG]]
			data.Action = constants.ACTION_TYPE[list[headers.IndexesOfHeaders[constants.FLAG]]]
			resp.Products[data.SwidTag] = data
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	if s.Err() != nil {
		err = errors.New("BadFile")
	}
	return

}

func getApplications(s *bufio.Scanner, headers models.HeadersInfo) (resp models.FileData, err error) {
	resp.Applications = make(map[string]models.ApplicationInfo)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.APP_ID]]) > 0 {
			data := models.ApplicationInfo{}
			data.ID = list[headers.IndexesOfHeaders[constants.APP_ID]]
			data.Name = list[headers.IndexesOfHeaders[constants.NAME]]
			data.Owner = list[headers.IndexesOfHeaders[constants.OWNER]]
			data.Version = list[headers.IndexesOfHeaders[constants.VERSION]]
			data.Domain = list[headers.IndexesOfHeaders[constants.DOMAIN]]
			data.Action = constants.ACTION_TYPE[list[headers.IndexesOfHeaders[constants.FLAG]]]
			resp.Applications[data.ID] = data
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++

	}
	err = s.Err()
	return
}

func getApplicationsAndProducts(s *bufio.Scanner, headers models.HeadersInfo) (resp models.FileData, err error) {
	resp.AppProducts = make(map[string]map[string][]string)
	resp.AppProducts[constants.UPSERT] = make(map[string][]string)
	resp.AppProducts[constants.DELETE] = make(map[string][]string)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.APP_ID]]) > 0 && len(list[headers.IndexesOfHeaders[constants.SWIDTAG]]) > 0 {
			prodID := list[headers.IndexesOfHeaders[constants.SWIDTAG]]
			appID := list[headers.IndexesOfHeaders[constants.APP_ID]]
			action := constants.ACTION_TYPE[list[headers.IndexesOfHeaders[constants.FLAG]]]
			resp.AppProducts[action][prodID] = append(resp.AppProducts[action][prodID], appID)
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	err = s.Err()
	return
}

func getInstancesOfProducts(s *bufio.Scanner, headers models.HeadersInfo) (resp models.FileData, err error) {
	resp.ProdInstances = make(map[string]map[string][]string)
	resp.ProdInstances[constants.UPSERT] = make(map[string][]string)
	resp.ProdInstances[constants.DELETE] = make(map[string][]string)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.INST_ID]]) > 0 && len(list[headers.IndexesOfHeaders[constants.SWIDTAG]]) > 0 {
			instanceID := list[headers.IndexesOfHeaders[constants.INST_ID]]
			prodId := list[headers.IndexesOfHeaders[constants.SWIDTAG]]
			action := constants.ACTION_TYPE[list[headers.IndexesOfHeaders[constants.FLAG]]]
			resp.ProdInstances[action][instanceID] = append(resp.ProdInstances[action][instanceID], prodId)
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	err = s.Err()
	return
}

func getInstanceOfApplications(s *bufio.Scanner, headers models.HeadersInfo) (resp models.FileData, err error) {
	resp.AppInstances = make(map[string][]models.AppInstance)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.INST_ID]]) > 0 && len(list[headers.IndexesOfHeaders[constants.APP_ID]]) > 0 {
			data := models.AppInstance{}
			data.ID = list[headers.IndexesOfHeaders[constants.INST_ID]]
			appID := list[headers.IndexesOfHeaders[constants.APP_ID]]
			data.Env = list[headers.IndexesOfHeaders[constants.ENVIRONMENT]]
			data.Action = constants.ACTION_TYPE[list[headers.IndexesOfHeaders[constants.FLAG]]]
			resp.AppInstances[appID] = append(resp.AppInstances[appID], data)
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	err = s.Err()
	return
}

func getEquipmentsOfProducts(s *bufio.Scanner, headers models.HeadersInfo) (resp models.FileData, err error) {
	resp.ProdEquipments = make(map[string]map[string][]models.ProdEquipemtInfo)
	resp.ProdEquipments[constants.UPSERT] = make(map[string][]models.ProdEquipemtInfo)
	resp.ProdEquipments[constants.DELETE] = make(map[string][]models.ProdEquipemtInfo)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.SWIDTAG]]) > 0 && len(list[headers.IndexesOfHeaders[constants.EQUIP_ID]]) > 0 {
			temp := models.ProdEquipemtInfo{}
			prodID := list[headers.IndexesOfHeaders[constants.SWIDTAG]]
			temp.EquipID = list[headers.IndexesOfHeaders[constants.EQUIP_ID]]
			temp.NbUsers = list[headers.IndexesOfHeaders[constants.NBUSERS]]
			action := constants.ACTION_TYPE[list[headers.IndexesOfHeaders[constants.FLAG]]]
			resp.ProdEquipments[action][prodID] = append(resp.ProdEquipments[action][prodID], temp)
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	err = s.Err()
	return
}

func getEquipmentsOnInstances(s *bufio.Scanner, headers models.HeadersInfo) (resp models.FileData, err error) {
	resp.EquipInstances = make(map[string]map[string][]string)
	resp.EquipInstances[constants.UPSERT] = make(map[string][]string)
	resp.EquipInstances[constants.DELETE] = make(map[string][]string)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.EQUIP_ID]]) > 0 && len(list[headers.IndexesOfHeaders[constants.INST_ID]]) > 0 {
			instanceID := list[headers.IndexesOfHeaders[constants.INST_ID]]
			equipID := list[headers.IndexesOfHeaders[constants.EQUIP_ID]]
			action := constants.ACTION_TYPE[list[headers.IndexesOfHeaders[constants.FLAG]]]
			resp.EquipInstances[action][instanceID] = append(resp.EquipInstances[action][instanceID], equipID)
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	err = s.Err()
	return
}

func getAcqRightsOfProducts(s *bufio.Scanner, headers models.HeadersInfo) (resp models.FileData, err error) {
	resp.AcqRights = make(map[string]models.AcqRightsInfo)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.SKU]]) > 0 {
			temp := models.AcqRightsInfo{}
			temp.Version = list[headers.IndexesOfHeaders[constants.PRODUCT_VERSION]]
			temp.SwidTag = list[headers.IndexesOfHeaders[constants.SWIDTAG]]
			temp.Sku = list[headers.IndexesOfHeaders[constants.SKU]]
			temp.Entity = list[headers.IndexesOfHeaders[constants.ENTITY]]
			temp.ProductName = list[headers.IndexesOfHeaders[constants.PRODUCT_NAME]]
			temp.Editor = list[headers.IndexesOfHeaders[constants.EDITOR]]
			temp.Metric = list[headers.IndexesOfHeaders[constants.METRIC]]
			temp.NumOfAcqLic, _ = strconv.Atoi(list[headers.IndexesOfHeaders[constants.ACQ_LIC_NO]])
			temp.NumOfMaintenanceLic, _ = strconv.Atoi(list[headers.IndexesOfHeaders[constants.LIC_UNDER_MAINTENANCE_NO]])
			temp.AvgPrice, _ = strconv.ParseFloat(list[headers.IndexesOfHeaders[constants.AVG_UNIT_PRICE]], 64)
			temp.AvgMaintenantPrice, err = strconv.ParseFloat(list[headers.IndexesOfHeaders[constants.AVG_MAINENANCE_UNIT_PRICE]], 64)
			temp.TotalPurchasedCost, _ = strconv.ParseFloat(list[headers.IndexesOfHeaders[constants.TOTAL_PURCHASE_COST]], 64)
			temp.TotalMaintenanceCost, _ = strconv.ParseFloat(list[headers.IndexesOfHeaders[constants.TOTAL_MAINENANCE_COST]], 64)
			temp.TotalCost, _ = strconv.ParseFloat(list[headers.IndexesOfHeaders[constants.TOTAL_COST]], 64)
			temp.Action = constants.ACTION_TYPE[list[headers.IndexesOfHeaders[constants.FLAG]]]
			temp.StartOfMaintenance = list[headers.IndexesOfHeaders[constants.START_OF_MAINTENANCE]]
			temp.EndOfMaintenance = list[headers.IndexesOfHeaders[constants.END_OF_MAINTENANCE]]
			resp.AcqRights[temp.Sku] = temp
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	err = s.Err()
	return
}

func csvToFileData(fileType, fileName string, expectedHeaders []string) (resp models.FileData, err error) {
	var headers models.HeadersInfo
	file := fmt.Sprintf("%s/%s", config.GetConfig().FilesLocation, fileName)
	log.Println("Looking for file   >>>>>>>>>>>>>>>>>>>>>>>>>>>>> : ", file)
	csvFile, err := os.Open(file)
	if err != nil {
		resp.FileFailureReason = "BadFile"
		return resp, status.Error(codes.Internal, "BadFile")
	}
	defer csvFile.Close()
	scanner := bufio.NewScanner(csvFile)
	if !scanner.Scan() {
		err = scanner.Err()
		resp.FileFailureReason = "BadFile"
		return
	}
	headers, err = getIndexOfHeaders(scanner.Text(), expectedHeaders)
	if err != nil {
		resp.FileFailureReason = err.Error()
		return
	}
	switch fileType {
	case constants.PRODUCTS:
		resp, err = getProducts(scanner, headers)

	case constants.APPLICATIONS:
		resp, err = getApplications(scanner, headers)

	case constants.PRODUCTS_EQUIPMENTS:
		resp, err = getEquipmentsOfProducts(scanner, headers)

	case constants.PRODUCTS_ACQUIREDRIGHTS:
		resp, err = getAcqRightsOfProducts(scanner, headers)

	case constants.INSTANCES_PRODUCTS:
		resp, err = getInstancesOfProducts(scanner, headers)

	case constants.INSTANCES_EQUIPMENTS:
		resp, err = getEquipmentsOnInstances(scanner, headers)

	case constants.APPLICATIONS_INSTANCES:
		resp, err = getInstanceOfApplications(scanner, headers)

	case constants.APPLICATIONS_PRODUCTS:
		resp, err = getApplicationsAndProducts(scanner, headers)

	default:
		err = status.Error(codes.Internal, "FileNotSupported")
		return
	}

	if resp.TotalCount == 0 {
		err = status.Error(codes.Internal, "NoDataInFile")
	}
	if err != nil {
		resp.FileFailureReason = err.Error()
	}
	log.Println("File ", fileName, " has total records ", resp.TotalCount, " invalid", resp.InvalidCount, " row", resp.InvalidDataRowNum)
	resp.FileType = fileType
	resp.FileName = fileName
	resp.TargetServices = constants.SERVICES[fileType]
	return
}

func createAPITypeJobs(data models.FileData) (jobs []job.Job, err error) {
	for _, targetService := range data.TargetServices {
		switch targetService {
		case constants.APP_SERVICE:
			jobs = createAppServiceJobs(data, targetService)

		case constants.PROD_SERVICE:
			jobs = createProdServiceJobs(data, targetService)

		case constants.EQUIP_SERVICE:
			jobs = createEquipServiceJobs(data, targetService)

		default:
			err = status.Error(codes.Internal, "TargetServiceNotSupported")
		}
	}
	return
}

func createEquipServiceJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
	// For  Metadata Processing
	if len(data.Schema) > 0 {
		fileAsSource := strings.Split(data.FileName, fmt.Sprintf("%s_", strings.ToUpper(data.Scope)))[1]
		envlope := getEnvlope(targetService, data.FileType, data.FileName, data.UploadID)
		appData := equipment.UpsertMetadataRequest{
			MetadataType:       "equipment",
			MetadataSource:     fileAsSource,
			MetadataAttributes: data.Schema,
			Scope:              data.Scope,
		}

		envlope.TargetAction = constants.UPSERT
		envlope.Data, err = json.Marshal(appData)
		if err != nil {
			log.Println("Failed to marshal jobdata, err:", err)
			return
		}
		jobObj.Data, err = json.Marshal(envlope)
		if err != nil {
			log.Println("Failed to marshal envlope, err:", err)
			return
		}
		jobObj.Status = job.JobStatusPENDING
		jobs = append(jobs, jobObj)
	} else {
		envlope := getEnvlope(targetService, "EQUIPMENTS", data.FileName, data.UploadID)
		for k, v := range data.Equipments {
			for _, rec := range v {
				//Marshal Map
				b, _ := json.Marshal(rec)
				//fmt.Printf("json %s", string(b))
				// structpb := &structpb.Struct{}
				// unmarshaler := jsonpb.Unmarshaler{}
				// //unmarshal bytes to structpb
				// err = unmarshaler.Unmarshal(bytes.NewReader(b), structpb)
				// if err != nil {
				// 	logger.Log.Error("Failed To Unmarshal to structpb", zap.Error(err))
				// }
				eqData := models.EquipmentRequest{Scope: data.Scope, EqType: strings.ToLower(k), EqData: b}
				envlope.TargetAction = constants.UPSERT
				//marshal to specific job
				envlope.Data, err = json.Marshal(eqData)
				if err != nil {
					log.Println("Failed to marshal jobdata, err:", err)
					return
				}
				//marshal to generic envelope
				jobObj.Data, err = json.Marshal(envlope)
				if err != nil {
					log.Println("Failed to marshal envlope, err:", err)
					return
				}
				jobObj.Status = job.JobStatusPENDING
				jobs = append(jobs, jobObj)
			}
		}
	}

	return
}

func createProdAcqRightsJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for _, val := range data.AcqRights {
		envlope := getEnvlope(targetService, data.FileType, data.FileName, data.UploadID)
		jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
		appData := product.UpsertAcqRightsRequest{
			Version:                 val.Version,
			Sku:                     val.Sku,
			Swidtag:                 val.SwidTag,
			ProductName:             val.ProductName,
			ProductEditor:           val.Editor,
			MetricType:              val.Metric,
			NumLicensesAcquired:     int32(val.NumOfAcqLic),
			NumLicencesMaintainance: int32(val.NumOfMaintenanceLic),
			AvgUnitPrice:            float64(val.AvgPrice),
			AvgMaintenanceUnitPrice: float64(val.AvgMaintenantPrice),
			TotalPurchaseCost:       float64(val.TotalPurchasedCost),
			TotalMaintenanceCost:    float64(val.TotalMaintenanceCost),
			TotalCost:               float64(val.TotalCost),
			Entity:                  val.Entity,
			Scope:                   data.Scope,
			StartOfMaintenance:      val.StartOfMaintenance,
			EndOfMaintenance:        val.EndOfMaintenance,
		}
		envlope.TargetAction = constants.UPSERT
		envlope.Data, err = json.Marshal(appData)
		if err != nil {
			log.Println("Failed to marshal jobdata, err:", err)
			return
		}
		jobObj.Data, err = json.Marshal(envlope)
		if err != nil {
			log.Println("Failed to marshal envlope, err:", err)
			return
		}
		jobObj.Status = job.JobStatusPENDING
		jobs = append(jobs, jobObj)
	}
	return
}

func createProdServiceJobs(data models.FileData, targetService string) (jobs []job.Job) {
	switch data.FileType {
	case constants.PRODUCTS:
		jobs = createProductJobs(data, targetService)

	case constants.APPLICATIONS_PRODUCTS:
		jobs = createAppProductsJobs(data, targetService)

	case constants.PRODUCTS_EQUIPMENTS:
		jobs = createProdEquipJobs(data, targetService)

	case constants.PRODUCTS_ACQUIREDRIGHTS:
		jobs = createProdAcqRightsJobs(data, targetService)
	}
	return
}

func createAppServiceJobs(data models.FileData, targetService string) (jobs []job.Job) {
	switch data.FileType {
	case constants.APPLICATIONS:
		jobs = createApplicationJobs(data, targetService)

	case constants.APPLICATIONS_INSTANCES:
		jobs = createAppInstanceJobs(data, targetService)

	case constants.INSTANCES_PRODUCTS:
		jobs = createInstanceProdJobs(data, targetService)

	case constants.INSTANCES_EQUIPMENTS:
		jobs = createInstanceEquipJobs(data, targetService)
	}
	return
}

func createProdEquipJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for action, prodAndEquip := range data.ProdEquipments {
		for prodID, equips := range prodAndEquip {
			envlope := getEnvlope(targetService, data.FileType, data.FileName, data.UploadID)
			jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
			appData := product.UpsertProductRequest{
				SwidTag: prodID,
				Scope:   data.Scope,
				Equipments: &product.UpsertProductRequestEquipment{
					Operation:      constants.API_ACTION[action],
					Equipmentusers: convertProdEquipments(equips),
				},
			}
			envlope.TargetAction = constants.UPSERT
			envlope.Data, err = json.Marshal(appData)
			if err != nil {
				log.Println("Failed to marshal jobdata, err:", err)
				return
			}
			jobObj.Data, err = json.Marshal(envlope)
			if err != nil {
				log.Println("Failed to marshal envlope, err:", err)
				return
			}
			jobObj.Status = job.JobStatusPENDING
			jobs = append(jobs, jobObj)
		}
	}
	return
}
func convertProdEquipments(data []models.ProdEquipemtInfo) (res []*product.UpsertProductRequestEquipmentEquipmentuser) {
	for _, val := range data {
		nb, _ := strconv.Atoi(val.NbUsers)
		temp := product.UpsertProductRequestEquipmentEquipmentuser{
			EquipmentId: val.EquipID,
			NumUser:     int32(nb),
		}
		res = append(res, &temp)
	}
	return
}

func createAppProductsJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for action, prodAndApps := range data.AppProducts {
		for prodID, applications := range prodAndApps {
			envlope := getEnvlope(targetService, data.FileType, data.FileName, data.UploadID)
			jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
			appData := product.UpsertProductRequest{
				SwidTag: prodID,
				Scope:   data.Scope,
				Applications: &product.UpsertProductRequestApplication{
					Operation:     constants.API_ACTION[action],
					ApplicationId: applications,
				},
			}
			envlope.TargetAction = constants.UPSERT
			envlope.Data, err = json.Marshal(appData)
			if err != nil {
				log.Println("Failed to marshal jobdata, err:", err)
				return
			}
			jobObj.Data, err = json.Marshal(envlope)
			if err != nil {
				log.Println("Failed to marshal envlope, err:", err)
				return
			}
			jobObj.Status = job.JobStatusPENDING
			jobs = append(jobs, jobObj)
		}
	}
	return
}

func createProductJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for _, val := range data.Products {
		envlope := getEnvlope(targetService, data.FileType, data.FileName, data.UploadID)
		jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
		appData := product.UpsertProductRequest{
			SwidTag:  val.SwidTag,
			Name:     val.Name,
			Version:  val.Version,
			Editor:   val.Editor,
			Category: val.Category,
			OptionOf: val.IsOptionOf,
			Scope:    data.Scope,
		}
		envlope.TargetAction = constants.UPSERT
		envlope.Data, err = json.Marshal(appData)
		if err != nil {
			log.Println("Failed to marshal jobdata, err:", err)
			return
		}
		jobObj.Data, err = json.Marshal(envlope)
		if err != nil {
			log.Println("Failed to marshal envlope, err:", err)
			return
		}
		jobObj.Status = job.JobStatusPENDING
		jobs = append(jobs, jobObj)
	}
	return
}

func createInstanceEquipJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for action, instanceAndEquipments := range data.EquipInstances {
		for instanceID, equipments := range instanceAndEquipments {
			envlope := getEnvlope(targetService, data.FileType, data.FileName, data.UploadID)
			jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
			appData := application.UpsertInstanceRequest{
				InstanceId: instanceID,
				Scope:      data.Scope,
				Equipments: &application.UpsertInstanceRequestEquipment{
					Operation:   constants.API_ACTION[action],
					EquipmentId: equipments,
				},
			}
			envlope.TargetAction = constants.UPSERT
			envlope.Data, err = json.Marshal(appData)
			if err != nil {
				log.Println("Failed to marshal jobdata, err:", err)
				return
			}
			jobObj.Data, err = json.Marshal(envlope)
			if err != nil {
				log.Println("Failed to marshal envlope, err:", err)
				return
			}
			jobObj.Status = job.JobStatusPENDING
			jobs = append(jobs, jobObj)
		}
	}
	return
}

func createInstanceProdJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for action, instanceAndProducts := range data.ProdInstances {
		for instanceID, products := range instanceAndProducts {
			envlope := getEnvlope(targetService, data.FileType, data.FileName, data.UploadID)
			jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
			appData := application.UpsertInstanceRequest{
				InstanceId: instanceID,
				Scope:      data.Scope,
				Products: &application.UpsertInstanceRequestProduct{
					Operation: constants.API_ACTION[action],
					ProductId: products,
				},
			}
			envlope.TargetAction = constants.UPSERT
			envlope.Data, err = json.Marshal(appData)
			if err != nil {
				log.Println("Failed to marshal jobdata, err:", err)
				return
			}
			jobObj.Data, err = json.Marshal(envlope)
			if err != nil {
				log.Println("Failed to marshal envlope, err:", err)
				return
			}
			jobObj.Status = job.JobStatusPENDING
			jobs = append(jobs, jobObj)
		}
	}
	return
}

func createAppInstanceJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for appId, list := range data.AppInstances {
		for _, val := range list {
			envlope := getEnvlope(targetService, data.FileType, data.FileName, data.UploadID)
			jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
			var appData interface{}
			if val.Action == constants.UPSERT {
				appData = application.UpsertInstanceRequest{
					ApplicationId: appId,
					InstanceId:    val.ID,
					InstanceName:  val.Env,
					Scope:         data.Scope,
				}
				envlope.TargetAction = constants.UPSERT
			} else {
				appData = application.DeleteInstanceRequest{
					ApplicationId: appId,
					InstanceId:    val.ID,
				}
				envlope.TargetAction = constants.DELETE
			}
			envlope.Data, err = json.Marshal(appData)
			if err != nil {
				log.Println("Failed to marshal jobdata, err:", err)
				return
			}
			jobObj.Data, err = json.Marshal(envlope)
			if err != nil {
				log.Println("Failed to marshal envlope, err:", err)
				return
			}
			jobObj.Status = job.JobStatusPENDING
			jobs = append(jobs, jobObj)
		}
	}
	return
}

func createApplicationJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for _, val := range data.Applications {
		envlope := getEnvlope(targetService, data.FileType, data.FileName, data.UploadID)
		jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
		var appData interface{}
		if val.Action == constants.UPSERT {
			appData = application.UpsertApplicationRequest{
				ApplicationId: val.ID,
				Name:          val.Name,
				Version:       val.Version,
				Owner:         val.Owner,
				Scope:         data.Scope,
				Domain:        val.Domain,
			}
			envlope.TargetAction = constants.UPSERT
		} else {
			appData = application.DeleteApplicationRequest{
				ApplicationId: val.ID,
			}
			envlope.TargetAction = constants.DELETE
		}
		envlope.Data, err = json.Marshal(appData)
		if err != nil {
			log.Println("Failed to marshal jobdata, err:", err)
			return
		}
		jobObj.Data, err = json.Marshal(envlope)
		if err != nil {
			log.Println("Failed to marshal envlope, err:", err)
			return
		}
		jobObj.Status = job.JobStatusPENDING
		jobs = append(jobs, jobObj)
	}
	return
}

func archiveFile(fileName string, uploadId int32) error {
	newfile := fmt.Sprintf("%s/%d_%s", config.GetConfig().ArchiveLocation, uploadId, fileName)
	oldFile := fmt.Sprintf("%s/%s", config.GetConfig().FilesLocation, fileName)
	log.Println(" Archieving filr from ", oldFile, " to ", newfile)
	return os.Rename(oldFile, newfile)
}

func getEnvlope(service, fileType, fileName string, id int32) models.Envlope {
	return models.Envlope{
		TargetService: service,
		TargetRPC:     fileType,
		UploadID:      id,
		FileName:      fileName,
	}
}

func getEquipment(fileType, fileName string) (models.FileData, error) {
	file := fmt.Sprintf("%s/%s", config.GetConfig().FilesLocation, fileName)
	eqType := strings.Split(fileType, "_")[1]
	log.Println("Looking for file   >>>>>>>>>>>>>>>>> : ", file, fileType)
	time.Sleep(5)
	data, err := getDynamicEquipmentFromCsv(file)
	if err != nil {
		logger.Log.Error("Error reading equipment csv", zap.Error(err))
		return models.FileData{}, err
	}

	resp := models.FileData{}
	resp.Equipments = make(map[string][]map[string]interface{})
	resp.Equipments[eqType] = data
	resp.TotalCount = int32(len(data))
	resp.FileType = fileType
	resp.FileName = fileName
	resp.TargetServices = constants.SERVICES[fileType]
	return resp, nil
}

func getDynamicEquipmentFromCsv(file string) (resp []map[string]interface{}, err error) {

	csvFile, err := os.Open(file)
	if err != nil {
		logger.Log.Error("The file is not found", zap.Error(err))
		return nil, err
	}
	defer csvFile.Close()
	s := bufio.NewScanner(csvFile)
	if !s.Scan() {
		err = s.Err()
		return
	}

	headers := make(map[int]string)
	for key, val := range strings.Split(s.Text(), constants.DELIMETER) {
		headers[key] = val
	}
	hlen := len(headers)
	for s.Scan() {
		list := strings.Split(s.Text(), constants.DELIMETER)
		// TODO should we allow this
		if len(list) >= hlen {
			temp := make(map[string]interface{})
			for index, val := range list {
				var out interface{}
				var pErr error
				out, pErr = strconv.ParseInt(val, 10, 64)
				if pErr != nil {
					out, pErr = strconv.ParseFloat(val, 64)
					if pErr != nil {
						out, pErr = strconv.ParseBool(val)
						if pErr != nil {
							//the value is string
							out = val
						}
					}
				}
				temp[headers[index]] = out
			}
			resp = append(resp, temp)
		}
	}
	err = s.Err()
	log.Println("<<<<<<<<<<<>>>>>>>>>>>> Equipment File Processed in DPS service ")

	return
}
