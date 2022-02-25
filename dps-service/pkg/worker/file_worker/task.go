package fileworker

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	// acq "optisam-backend/acqrights-service/pkg/api/v1"
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

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// getFileTypeFromFileName return FileType in uppercase
// It can be
func getFileTypeFromFileName(fileName, scope string) (fileType string, err error) {
	fileName = getFileName(fileName)
	fileName = strings.ToUpper(fileName)
	sep := fmt.Sprintf("%s_", strings.ToUpper(scope))
	if !strings.Contains(fileName, sep) {
		err = status.Error(codes.Internal, "InvalidFileName")
		return
	}
	fileType = strings.Split(strings.Split(fileName, sep)[1], constants.FileExtension)[0]
	return
}

// nolint: nakedret
func fileProcessing(jobData gendb.UploadedDataFile) (data models.FileData, err error) {
	data.DuplicateRecords = make([]interface{}, 0)
	var fileType string
	var expectedHeaders []string
	if jobData.FileName == "" {
		err = status.Error(codes.Internal, "MissingFileName")
		return
	}
	if strings.Contains(strings.ToUpper(jobData.FileName), constants.METADATA) {
		data, err = csvFileToSchemaData(jobData.FileName)
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
			return data, status.Error(codes.Internal, constants.BadFile)
		}
		// For equipment, dynamic processing is required
		if strings.Contains(fileType, "EQUIPMENT_") {
			data, err = getEquipment(fileType, jobData.FileName)
			if err != nil {
				data.FileFailureReason = err.Error()
				return data, status.Error(codes.Internal, data.FileFailureReason)
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
					data.FileFailureReason = constants.BadFile
				}
				return
			}
		}
	}
	data.Scope = jobData.Scope
	data.FileName = jobData.FileName
	data.UploadID = jobData.UploadID
	data.GlobalID = jobData.Gid
	return
}

// Headers are updated, no No space is allowed in headers and these are case insensitive
// nolint: nakedret
func getHeadersForFileType(fileType string) (headers []string, err error) {
	headers = []string{}
	switch fileType {
	case constants.PRODUCTS:
		headers = []string{"swidtag", "version", "category", "editor", "isoptionof", "name", "flag"}

	case constants.APPLICATIONS:
		headers = []string{"application_id", "version", "owner", "name", "domain", "flag"}

	case constants.ApplicationsInstances:
		headers = []string{"application_id", "instance_id", "environment", "flag"}

	case constants.ApplicationsProducts:
		headers = []string{"application_id", "swidtag", "flag"}

	case constants.ProductsEquipments:
		headers = []string{"equipment_id", "swidtag", "nbusers", "flag"}

	case constants.InstancesProducts:
		headers = []string{"instance_id", "swidtag", "flag"}

	case constants.InstancesEquipments:
		headers = []string{"instance_id", "equipment_id", "flag"}

	case constants.ProductsAcquiredRights:
		headers = []string{"product_version", "sku", "swidtag", "product_name", "editor", "metric", "acquired_licenses", "total_license_cost", "total_maintenance_cost", "unit_price", "maintenance_unit_price", "total_cost", "flag", "maintenance_start", "maintenance_end", "maintenance_licenses"}

	default:
		err = status.Error(codes.Internal, "FileNotSupported")
	}
	return
}

func csvFileToSchemaData(fileName string) (data models.FileData, err error) {
	file := fmt.Sprintf("%s/%s", config.GetConfig().FilesLocation, fileName)
	csvFile, err := os.Open(file)
	if err != nil {
		logger.Log.Error("Failed to open file", zap.Error(err), zap.Any("File", file))
		return
	}
	defer csvFile.Close()
	scanner := bufio.NewScanner(csvFile)
	success := scanner.Scan()
	if success == false {
		err = scanner.Err()
		if err == nil {
			data.FileFailureReason = "EmptyFile"
		} else {
			data.FileFailureReason = constants.BadFile
		}
		err = errors.New(data.FileFailureReason)
		return
	}
	row := scanner.Text()
	// schemaType := strings.Split(strings.Split(fileName, constants.SCOPE_DELIMETER)[2], constants.FILE_EXTENSION)[0]

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
			data.IsOptionOf = list[headers.IndexesOfHeaders[constants.ISOPTIONOF]]
			data.Category = list[headers.IndexesOfHeaders[constants.CATEGORY]]
			data.SwidTag = list[headers.IndexesOfHeaders[constants.SWIDTAG]]
			data.Action = constants.ActionType[list[headers.IndexesOfHeaders[constants.FLAG]]]
			oldData, ok := resp.Products[data.SwidTag]
			if ok {
				resp.DuplicateRecords = append(resp.DuplicateRecords, oldData)
			}
			resp.Products[data.SwidTag] = data
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	if s.Err() != nil {
		err = errors.New("badfile")
	}
	return

}

func getApplications(s *bufio.Scanner, headers models.HeadersInfo) (resp models.FileData, err error) {
	resp.Applications = make(map[string]models.ApplicationInfo)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.APPID]]) > 0 {
			data := models.ApplicationInfo{}
			data.ID = list[headers.IndexesOfHeaders[constants.APPID]]
			data.Name = list[headers.IndexesOfHeaders[constants.NAME]]
			data.Owner = list[headers.IndexesOfHeaders[constants.OWNER]]
			data.Version = list[headers.IndexesOfHeaders[constants.VERSION]]
			data.Domain = list[headers.IndexesOfHeaders[constants.DOMAIN]]
			data.Action = constants.ActionType[list[headers.IndexesOfHeaders[constants.FLAG]]]
			oldData, ok := resp.Applications[data.ID]
			if ok {
				resp.DuplicateRecords = append(resp.DuplicateRecords, oldData)
			}
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

func getApplicationsAndProducts(s *bufio.Scanner, headers models.HeadersInfo) (models.FileData, error) {
	records := make(map[string]bool)
	resp := models.FileData{}
	resp.AppProducts = make(map[string]map[string][]string)
	resp.AppProducts[constants.UPSERT] = make(map[string][]string)
	resp.AppProducts[constants.DELETE] = make(map[string][]string)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.APPID]]) > 0 && len(list[headers.IndexesOfHeaders[constants.SWIDTAG]]) > 0 {
			prodID := list[headers.IndexesOfHeaders[constants.SWIDTAG]]
			appID := list[headers.IndexesOfHeaders[constants.APPID]]
			action := constants.ActionType[list[headers.IndexesOfHeaders[constants.FLAG]]]
			_, ok := records[row]
			if ok {
				resp.DuplicateRecords = append(resp.DuplicateRecords, models.ProdApplink{
					ProdID: prodID,
					AppID:  appID,
					Action: action,
				})
			} else {
				records[row] = true
				resp.AppProducts[action][prodID] = append(resp.AppProducts[action][prodID], appID)
			}

			//)
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	return resp, s.Err()
}

func getInstancesOfProducts(s *bufio.Scanner, headers models.HeadersInfo) (models.FileData, error) {
	records := make(map[string]bool)
	resp := models.FileData{}
	resp.ProdInstances = make(map[string]map[string][]string)
	resp.ProdInstances[constants.UPSERT] = make(map[string][]string)
	resp.ProdInstances[constants.DELETE] = make(map[string][]string)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.INSTID]]) > 0 && len(list[headers.IndexesOfHeaders[constants.SWIDTAG]]) > 0 {
			instanceID := list[headers.IndexesOfHeaders[constants.INSTID]]
			prodID := list[headers.IndexesOfHeaders[constants.SWIDTAG]]
			action := constants.ActionType[list[headers.IndexesOfHeaders[constants.FLAG]]]
			_, ok := records[row]
			if ok {
				resp.DuplicateRecords = append(resp.DuplicateRecords, models.ProdInstancelink{
					ProdID:     prodID,
					InstanceID: instanceID,
					Action:     action,
				})
			} else {
				records[row] = true
				resp.ProdInstances[action][instanceID] = append(resp.ProdInstances[action][instanceID], prodID)
			}

		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	return resp, s.Err()
}

func getInstanceOfApplications(s *bufio.Scanner, headers models.HeadersInfo) (models.FileData, error) {
	records := make(map[string]bool)
	resp := models.FileData{}
	resp.AppInstances = make(map[string][]models.AppInstance)

	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.INSTID]]) > 0 && len(list[headers.IndexesOfHeaders[constants.APPID]]) > 0 {
			data := models.AppInstance{}
			data.ID = list[headers.IndexesOfHeaders[constants.INSTID]]
			appID := list[headers.IndexesOfHeaders[constants.APPID]]
			data.Env = list[headers.IndexesOfHeaders[constants.ENVIRONMENT]]
			data.Action = constants.ActionType[list[headers.IndexesOfHeaders[constants.FLAG]]]
			_, ok := records[row]
			if ok {
				resp.DuplicateRecords = append(resp.DuplicateRecords, models.AppInstanceLink{
					AppID:      appID,
					InstanceID: data.ID,
					Env:        data.Env,
					Action:     data.Action,
				})
			} else {
				records[row] = true
				resp.AppInstances[appID] = append(resp.AppInstances[appID], data)
			}

		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	return resp, s.Err()
}

func getEquipmentsOfProducts(s *bufio.Scanner, headers models.HeadersInfo) (models.FileData, error) {
	records := make(map[string]bool)
	resp := models.FileData{}
	resp.ProdEquipments = make(map[string]map[string][]models.ProdEquipemtInfo)
	resp.ProdEquipments[constants.UPSERT] = make(map[string][]models.ProdEquipemtInfo)
	resp.ProdEquipments[constants.DELETE] = make(map[string][]models.ProdEquipemtInfo)

	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.SWIDTAG]]) > 0 && len(list[headers.IndexesOfHeaders[constants.EQUIPID]]) > 0 {
			temp := models.ProdEquipemtInfo{}
			prodID := list[headers.IndexesOfHeaders[constants.SWIDTAG]]
			temp.EquipID = list[headers.IndexesOfHeaders[constants.EQUIPID]]
			temp.NbUsers = list[headers.IndexesOfHeaders[constants.NBUSERS]]
			action := constants.ActionType[list[headers.IndexesOfHeaders[constants.FLAG]]]

			_, ok := records[row]
			if ok {
				resp.DuplicateRecords = append(resp.DuplicateRecords, models.ProductEquipmentLink{
					ProdID:  prodID,
					EquipID: temp.EquipID,
					NbUser:  temp.NbUsers,
					Action:  action,
				})
			} else {
				records[row] = true
				resp.ProdEquipments[action][prodID] = append(resp.ProdEquipments[action][prodID], temp)
			}
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	return resp, s.Err()
}

func getEquipmentsOnInstances(s *bufio.Scanner, headers models.HeadersInfo) (models.FileData, error) {
	records := make(map[string]bool)
	resp := models.FileData{}
	resp.EquipInstances = make(map[string]map[string][]string)
	resp.EquipInstances[constants.UPSERT] = make(map[string][]string)
	resp.EquipInstances[constants.DELETE] = make(map[string][]string)

	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.EQUIPID]]) > 0 && len(list[headers.IndexesOfHeaders[constants.INSTID]]) > 0 {
			instanceID := list[headers.IndexesOfHeaders[constants.INSTID]]
			equipID := list[headers.IndexesOfHeaders[constants.EQUIPID]]
			action := constants.ActionType[list[headers.IndexesOfHeaders[constants.FLAG]]]

			_, ok := records[row]
			if ok {
				resp.DuplicateRecords = append(resp.DuplicateRecords, models.EquipmentInstanceLink{
					InstanceID: instanceID,
					EquipID:    equipID,
					Action:     action,
				})
			} else {
				records[row] = true
				resp.EquipInstances[action][instanceID] = append(resp.EquipInstances[action][instanceID], equipID)
			}
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	return resp, s.Err()
}

// nolint: nakedret
func getAcqRightsOfProducts(s *bufio.Scanner, headers models.HeadersInfo) (models.FileData, error) {
	resp := models.FileData{}
	resp.AcqRights = make(map[string]models.AcqRightsInfo)
	var err error
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		if len(list) >= headers.MaxIndexVal+1 && len(list[headers.IndexesOfHeaders[constants.SKU]]) > 0 {
			temp := models.AcqRightsInfo{}
			temp.Version = list[headers.IndexesOfHeaders[constants.PRODUCTVERSION]]
			temp.SwidTag = list[headers.IndexesOfHeaders[constants.SWIDTAG]]
			temp.Sku = list[headers.IndexesOfHeaders[constants.SKU]]
			temp.ProductName = list[headers.IndexesOfHeaders[constants.PRODUCTNAME]]
			temp.Editor = list[headers.IndexesOfHeaders[constants.EDITOR]]
			temp.Metric = list[headers.IndexesOfHeaders[constants.METRIC]]
			temp.NumOfAcqLic, _ = strconv.Atoi(list[headers.IndexesOfHeaders[constants.ACQLICNO]])
			temp.NumOfMaintenanceLic, _ = strconv.Atoi(list[headers.IndexesOfHeaders[constants.LICUNDERMAINTENANCENO]])
			temp.AvgPrice, _ = strconv.ParseFloat(list[headers.IndexesOfHeaders[constants.AVGUNITPRICE]], 64)
			temp.AvgMaintenantPrice, _ = strconv.ParseFloat(list[headers.IndexesOfHeaders[constants.AVGMAINENANCEUNITPRICE]], 64)
			temp.TotalPurchasedCost, _ = strconv.ParseFloat(list[headers.IndexesOfHeaders[constants.TOTALPURCHASECOST]], 64)
			temp.TotalMaintenanceCost, _ = strconv.ParseFloat(list[headers.IndexesOfHeaders[constants.TOTALMAINENANCECOST]], 64)
			temp.TotalCost, _ = strconv.ParseFloat(list[headers.IndexesOfHeaders[constants.TOTALCOST]], 64)
			temp.Action = constants.ActionType[list[headers.IndexesOfHeaders[constants.FLAG]]]
			temp.StartOfMaintenance = list[headers.IndexesOfHeaders[constants.StartOfMaintenance]]
			temp.EndOfMaintenance = list[headers.IndexesOfHeaders[constants.EndOfMaintenance]]
			oldData, ok := resp.AcqRights[temp.Sku]
			if ok {
				resp.DuplicateRecords = append(resp.DuplicateRecords, oldData)
			}
			resp.AcqRights[temp.Sku] = temp
		} else {
			resp.InvalidCount++
			resp.InvalidDataRowNum = append(resp.InvalidDataRowNum, int(resp.TotalCount)+1)
		}
		resp.TotalCount++
	}
	err = s.Err()
	return resp, err
}

// nolint: nakedret
func csvToFileData(fileType, fileName string, expectedHeaders []string) (models.FileData, error) {
	var headers models.HeadersInfo
	resp := models.FileData{}
	var err error
	file := fmt.Sprintf("%s/%s", config.GetConfig().FilesLocation, fileName)
	logger.Log.Info("Looking for file   >>>>>>>>>>>>>>>>>>>>>>>>>>>>> : ", zap.Any("file", file))
	csvFile, err := os.Open(file)
	if err != nil {
		logger.Log.Error("Failed to open file", zap.Error(err), zap.Any("File", file))
		resp.FileFailureReason = constants.BadFile
		return resp, status.Error(codes.Internal, "BadFile")
	}
	defer csvFile.Close()
	scanner := bufio.NewScanner(csvFile)
	success := scanner.Scan()
	if success == false {
		err = scanner.Err()
		if err == nil {
			resp.FileFailureReason = "EmptyFile"
		} else {
			resp.FileFailureReason = constants.BadFile
		}
		err = errors.New(resp.FileFailureReason)
		return resp, err
	}

	headers, err = getIndexOfHeaders(scanner.Text(), expectedHeaders)
	if err != nil {
		resp.FileFailureReason = err.Error()
		return resp, err
	}
	switch fileType {
	case constants.PRODUCTS:
		resp, err = getProducts(scanner, headers)

	case constants.APPLICATIONS:
		resp, err = getApplications(scanner, headers)

	case constants.ProductsEquipments:
		resp, err = getEquipmentsOfProducts(scanner, headers)

	case constants.ProductsAcquiredRights:
		resp, err = getAcqRightsOfProducts(scanner, headers)

	case constants.InstancesProducts:
		resp, err = getInstancesOfProducts(scanner, headers)

	case constants.InstancesEquipments:
		resp, err = getEquipmentsOnInstances(scanner, headers)

	case constants.ApplicationsInstances:
		resp, err = getInstanceOfApplications(scanner, headers)

	case constants.ApplicationsProducts:
		resp, err = getApplicationsAndProducts(scanner, headers)

	default:
		err = status.Error(codes.Internal, "FileNotSupported")
		return resp, err
	}

	if resp.TotalCount == 0 {
		err = status.Error(codes.Internal, "NoDataInFile")
	}
	if err != nil {
		resp.FileFailureReason = err.Error()
	}
	resp.FileType = fileType
	resp.TargetServices = constants.SERVICES[fileType]
	return resp, err
}

func getFileName(fileName string) string {
	temp := strings.Split(fileName, constants.NifiFileDelimeter)
	if len(temp) == 3 {
		fileName = temp[2]
	}
	return fileName
}

func createAPITypeJobs(data models.FileData) (jobs []job.Job, err error) {
	for _, targetService := range data.TargetServices {
		switch targetService {
		case constants.AppService:
			jobs = createAppServiceJobs(data, targetService)

		case constants.ProdService:
			jobs = createProdServiceJobs(data, targetService)

		case constants.EquipService:
			jobs = createEquipServiceJobs(data, targetService)

		default:
			err = status.Error(codes.Internal, "TargetServiceNotSupported")
		}
	}
	return
}

// nolint: nakedret
func createEquipServiceJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
	// For  Metadata Processing
	if len(data.Schema) > 0 {
		fileAsSource := strings.Split(data.FileName, fmt.Sprintf("%s_", strings.ToUpper(data.Scope)))[1]
		envlope := getEnvlope(targetService, data.FileType, data.FileName, data.TransfromedFileName, data.UploadID, data.GlobalID)
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
		envlope := getEnvlope(targetService, "EQUIPMENTS", data.FileName, data.TransfromedFileName, data.UploadID, data.GlobalID)
		for k, v := range data.Equipments {
			for _, rec := range v {
				// Marshal Map
				b, _ := json.Marshal(rec)
				// fmt.Printf("json %s", string(b))
				// structpb := &structpb.Struct{}
				// unmarshaler := jsonpb.Unmarshaler{}
				// unmarshal bytes to structpb
				// err = unmarshaler.Unmarshal(bytes.NewReader(b), structpb)
				// if err != nil {
				// 	logger.Log.Error("Failed To Unmarshal to structpb", zap.Error(err))
				// }
				eqData := models.EquipmentRequest{Scope: data.Scope, EqType: strings.ToLower(k), EqData: b}
				envlope.TargetAction = constants.UPSERT
				// marshal to specific job
				envlope.Data, err = json.Marshal(eqData)
				if err != nil {
					log.Println("Failed to marshal jobdata, err:", err)
					return
				}
				// marshal to generic envelope
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

// nolint: nakedret
func createProdAcqRightsJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for _, val := range data.AcqRights {
		envlope := getEnvlope(targetService, data.FileType, data.FileName, data.TransfromedFileName, data.UploadID, data.GlobalID)
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
			AvgUnitPrice:            val.AvgPrice,
			AvgMaintenanceUnitPrice: val.AvgMaintenantPrice,
			TotalPurchaseCost:       val.TotalPurchasedCost,
			TotalMaintenanceCost:    val.TotalMaintenanceCost,
			TotalCost:               val.TotalCost,
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

	case constants.ApplicationsProducts:
		jobs = createAppProductsJobs(data, targetService)

	case constants.ProductsEquipments:
		jobs = createProdEquipJobs(data, targetService)

	case constants.ProductsAcquiredRights:
		jobs = createProdAcqRightsJobs(data, targetService)
	}
	return
}

func createAppServiceJobs(data models.FileData, targetService string) (jobs []job.Job) {
	switch data.FileType {
	case constants.APPLICATIONS:
		jobs = createApplicationJobs(data, targetService)

	case constants.ApplicationsInstances:
		jobs = createAppInstanceJobs(data, targetService)

	case constants.InstancesProducts:
		jobs = createInstanceProdJobs(data, targetService)

	case constants.InstancesEquipments:
		jobs = createInstanceEquipJobs(data, targetService)
	}
	return
}

func createProdEquipJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for action, prodAndEquip := range data.ProdEquipments {
		for prodID, equips := range prodAndEquip {
			envlope := getEnvlope(targetService, data.FileType, data.FileName, data.TransfromedFileName, data.UploadID, data.GlobalID)
			jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
			appData := product.UpsertProductRequest{
				SwidTag: prodID,
				Scope:   data.Scope,
				Equipments: &product.UpsertProductRequestEquipment{
					Operation:      constants.APIAction[action],
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
		nb, _ := strconv.Atoi(val.NbUsers) // nolint: gosec
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
			envlope := getEnvlope(targetService, data.FileType, data.FileName, data.TransfromedFileName, data.UploadID, data.GlobalID)
			jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
			appData := product.UpsertProductRequest{
				SwidTag: prodID,
				Scope:   data.Scope,
				Applications: &product.UpsertProductRequestApplication{
					Operation:     constants.APIAction[action],
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
		envlope := getEnvlope(targetService, data.FileType, data.FileName, data.TransfromedFileName, data.UploadID, data.GlobalID)
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
			envlope := getEnvlope(targetService, data.FileType, data.FileName, data.TransfromedFileName, data.UploadID, data.GlobalID)
			jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
			appData := application.UpsertInstanceRequest{
				InstanceId: instanceID,
				Scope:      data.Scope,
				Equipments: &application.UpsertInstanceRequestEquipment{
					Operation:   constants.APIAction[action],
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
			envlope := getEnvlope(targetService, data.FileType, data.FileName, data.TransfromedFileName, data.UploadID, data.GlobalID)
			jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
			appData := application.UpsertInstanceRequest{
				InstanceId: instanceID,
				Scope:      data.Scope,
				Products: &application.UpsertInstanceRequestProduct{
					Operation: constants.APIAction[action],
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

// nolint: nakedret
func createAppInstanceJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for appID, list := range data.AppInstances {
		for _, val := range list {
			envlope := getEnvlope(targetService, data.FileType, data.FileName, data.TransfromedFileName, data.UploadID, data.GlobalID)
			jobObj := job.Job{Status: job.JobStatusFAILED, Type: constants.APITYPE}
			var appData interface{}
			if val.Action == constants.UPSERT {
				appData = application.UpsertInstanceRequest{
					ApplicationId: appID,
					InstanceId:    val.ID,
					InstanceName:  val.Env,
					Scope:         data.Scope,
				}
				envlope.TargetAction = constants.UPSERT
			} else {
				appData = application.DeleteInstanceRequest{
					ApplicationId: appID,
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

// nolint: nakedret
func createApplicationJobs(data models.FileData, targetService string) (jobs []job.Job) {
	var err error
	for _, val := range data.Applications {
		envlope := getEnvlope(targetService, data.FileType, data.FileName, data.TransfromedFileName, data.UploadID, data.GlobalID)
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

func archiveFile(fileName string, uploadID int32) error {
	newfile := fmt.Sprintf("%s/%d_%s", config.GetConfig().ArchiveLocation, uploadID, fileName)
	newfile = strings.Replace(newfile, fmt.Sprintf("%s#", constants.PROCESSING), "", 1)
	oldFile := fmt.Sprintf("%s/%s", config.GetConfig().FilesLocation, fileName)
	logger.Log.Error(" Archieving ", zap.Any("oldfile", oldFile), zap.Any("newfile", newfile))
	return os.Rename(oldFile, newfile)
}

func getEnvlope(service, fileType, fileName, transformedFile string, id, gid int32) models.Envlope {
	return models.Envlope{
		TargetService:       service,
		TargetRPC:           fileType,
		UploadID:            id,
		FileName:            fileName,
		GlobalFileID:        gid,
		TransfromedFileName: transformedFile,
	}
}

func getEquipment(fileType, fileName string) (models.FileData, error) {
	file := fmt.Sprintf("%s/%s", config.GetConfig().FilesLocation, fileName)
	eqType := strings.Split(fileType, "_")[1]
	logger.Log.Info("Looking for equipment file   >>>>>>>>>>>>>>>>> : ", zap.Any("file", file))
	// time.Sleep(5)
	data, duplicate, err := getDynamicEquipmentFromCsv(file)
	if err != nil {
		logger.Log.Error("Error reading equipment file", zap.Error(err), zap.Any("file", file))
		return models.FileData{}, err
	}

	resp := models.FileData{}
	resp.Equipments = make(map[string][]map[string]interface{})
	resp.Equipments[eqType] = data
	resp.TotalCount = int32(len(data)) + int32(len(duplicate))
	resp.FileType = fileType
	resp.DuplicateRecords = duplicate
	resp.TargetServices = constants.SERVICES[fileType]
	return resp, nil
}

// nolint: nakedret
func getDynamicEquipmentFromCsv(file string) (resp []map[string]interface{}, duplicate []interface{}, err error) {
	csvFile, err := os.Open(file)
	if err != nil {
		logger.Log.Error("Failed to open file", zap.Error(err), zap.Any("File", file))
		return
	}
	defer csvFile.Close()
	s := bufio.NewScanner(csvFile)
	success := s.Scan()
	if !success {
		err = s.Err()
		if err == nil {
			err = errors.New("emptyfile")
		} else {
			err = errors.New("badfile")
		}
		return
	}

	headers := make(map[int]string)
	for key, val := range strings.Split(s.Text(), constants.DELIMETER) {
		headers[key] = val
	}
	hlen := len(headers)
	records := make(map[string]bool)
	for s.Scan() {
		row := s.Text()
		list := strings.Split(row, constants.DELIMETER)
		// TODO should we allow this
		if len(list) >= hlen {
			temp := make(map[string]interface{})
			for index, val := range list {
				// var out interface{}
				// var pErr error
				// out, pErr = strconv.ParseInt(val, 10, 64)
				// if pErr != nil {
				// 	out, pErr = strconv.ParseFloat(val, 64)
				// 	if pErr != nil {
				// 		out, pErr = strconv.ParseBool(val)
				// 		if pErr != nil {
				// 			// the value is string
				// 			out = val
				// 		}
				// 	}
				// }
				temp[headers[index]] = val
			}
			ok := records[row]
			if ok {
				duplicate = append(duplicate, temp)
			} else {
				records[row] = true
				resp = append(resp, temp)
			}
		}
	}
	err = s.Err()
	if len(resp) == 0 || err != nil {
		err = errors.New("badfile")
		return
	}
	logger.Log.Info("Equipment File processed ", zap.Any("file", file))
	return
}
