package v1

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	v1Acc "optisam-backend/account-service/pkg/api/v1"
	v1Catalog "optisam-backend/catalog-service/pkg/api/v1"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	rest_middleware "optisam-backend/common/optisam/middleware/rest"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"optisam-backend/import-service/pkg/config"
	v1Simulation "optisam-backend/simulation-service/pkg/api/v1"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/julienschmidt/httprouter"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type importServiceServer struct {
	// grpcServers map[string]*grpc.ClientConn
	dpsClient     v1.DpsServiceClient
	simClient     v1Simulation.SimulationServiceClient
	accClient     v1Acc.AccountServiceClient
	catalogClient v1Catalog.ProductCatalogClient
	config        *config.Config
}

type uploadType string
type downloadType string

const (
	metadataload uploadType   = "metadata"
	dataload     uploadType   = "data"
	rawdataload  uploadType   = "globaldata"
	analysis     uploadType   = "analysis"
	source       uploadType   = "source"
	corefactor   uploadType   = "corefactor"
	errorFile    downloadType = "error"
	GENERIC      string       = "GENERIC"
	GEN          string       = "GEN"
	XLSX         string       = ".xlsx"
	CSV          string       = ".csv"
)

// NewImportServiceServer creates import service
func NewImportServiceServer(grpcServers map[string]*grpc.ClientConn, config *config.Config) ImportServiceServer {
	return &importServiceServer{config: config, dpsClient: v1.NewDpsServiceClient(grpcServers["dps"]),
		simClient:     v1Simulation.NewSimulationServiceClient(grpcServers["simulation"]),
		accClient:     v1Acc.NewAccountServiceClient(grpcServers["account"]),
		catalogClient: v1Catalog.NewProductCatalogClient(grpcServers["catalog"]),
	}
}

// UploadFiles will be used for upload global,metadata,data files and analysis file (optimization future)
func (i *importServiceServer) UploadFiles(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	var err error
	dataScope := req.FormValue("scope")
	if dataScope == "" {
		logger.Log.Error("ScopeNotFound")
		http.Error(res, "ScopeNotFound", http.StatusBadRequest)
		return
	}
	uploadCategory := req.FormValue("uploadType")
	if dataScope == "" {
		logger.Log.Error("uploadType")
		http.Error(res, "uploadType", http.StatusBadRequest)
		return
	}
	userClaims, ok := rest_middleware.RetrieveClaims(req.Context())
	if !ok {
		logger.Log.Error("ClaimsNotFound")
		http.Error(res, "ClaimsNotFound", http.StatusBadRequest)
		return
	}

	if !helper.Contains(userClaims.Socpes, dataScope) {
		http.Error(res, "ScopeValidationFailed", http.StatusUnauthorized)
		return
	}

	if err = req.ParseMultipartForm(32 << 20); nil != err {
		logger.Log.Error("ParsingFailure", zap.Error(err))
		http.Error(res, "ParsingFailure", http.StatusInternalServerError)
		return
	}
	var status int
	var resp interface{}
	switch uploadCategory {
	case string(analysis):
		dstDir := fmt.Sprintf("%s/%s/analysis", i.config.Upload.RawDataUploadDir, dataScope)
		resp, status, err = uploadFileForAnalysis(i, req, dataScope, dstDir)
	case string(corefactor):
		resp, status, err = saveCoreFactorReference(i, req)
		if err != nil {
			logger.Log.Error("Failed to upload file ", zap.Error(err))
			http.Error(res, err.Error(), status)
			return
		}
	default:
		err = errors.New("unknownUploadCategoryReceived")
		status = http.StatusBadRequest
	}
	if err != nil {
		logger.Log.Error("Failed to upload file ", zap.Error(err))
		http.Error(res, err.Error(), status)
		return
	}
	out, jrr := json.Marshal(resp)
	if jrr != nil {
		logger.Log.Error("Failed to marshal the response", zap.Error(jrr))
		http.Error(res, "ResponseParsingFailure", http.StatusInternalServerError)
	}
	res.Write(out) //nolint

}

func saveCoreFactorReference(i *importServiceServer, req *http.Request) (interface{}, int, error) {
	file, fileInfo, err := req.FormFile("file")
	if fileInfo.Size > i.config.MaxFileSize*1024*1024 {
		logger.Log.Error("File uploaded is larger than allowed", zap.Error(err))
		return nil, http.StatusBadRequest, errors.New("maximum file allowded is :" + strconv.FormatInt(i.config.MaxFileSize, 10) + "Mbs")
	}
	if err != nil {
		logger.Log.Error("Failed to read reference file", zap.Error(err))
		return nil, http.StatusBadRequest, err
	}
	defer file.Close()
	f, err := excelize.OpenReader(file)
	if err != nil {
		logger.Log.Error("Failed to parse reference file", zap.Error(err))
		return nil, http.StatusBadRequest, err
	}
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		logger.Log.Error("Reference file doesn't have any sheet", zap.Error(err))
		return nil, http.StatusBadRequest, err
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		logger.Log.Error("Failed to read the sheet", zap.Error(err), zap.String("sheet", sheets[0]))
		return nil, http.StatusInternalServerError, err
	} else if len(rows) < 2 {
		logger.Log.Error("inapropiate sheet, no reference value found", zap.Error(err), zap.String("sheet", sheets[0]))
		return nil, http.StatusInternalServerError, err
	}
	rows = rows[1:]
	dataToSend := make(map[string]map[string]string)
	for _, v := range rows {
		logger.Log.Debug("reference row", zap.Any("row", v))
		if len(v) == 3 {
			mf := v[0]
			ml := v[1]
			if mf == "" {
				mf = "default"
			}
			if ml == "" {
				ml = "default"
			}
			if dataToSend[mf] == nil {
				dataToSend[mf] = make(map[string]string)
			}
			dataToSend[mf][ml] = v[2]
		}
	}
	byteData, err := json.Marshal(dataToSend)
	if err != nil {
		logger.Log.Error("Marshaling failure", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}
	logger.Log.Debug("sending data to dps ", zap.Any("referecnce data", dataToSend))

	resp, err := i.dpsClient.StoreCoreFactorReference(req.Context(), &v1.StoreReferenceDataRequest{
		ReferenceData: byteData,
		Filename:      fileInfo.Filename,
	})
	if err != nil {
		logger.Log.Error(" unable to store core factor reference", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}
	return resp, http.StatusOK, nil
}

func uploadFileForAnalysis(i *importServiceServer, req *http.Request, scope, dstDir string) (interface{}, int, error) {
	var resp interface{}
	err := os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		logger.Log.Error("AnalysisDirectoryCreationFailure", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}
	fileName := ""
	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			if hdr.Size > i.config.MaxFileSize*1024*1024 {
				logger.Log.Error("File uploaded is larger than allowed", zap.Error(err))
				return nil, http.StatusBadRequest, errors.New("maximum file allowded is :" + strconv.FormatInt(i.config.MaxFileSize, 10) + "Mbs")
			}
			if !strings.Contains(hdr.Filename, XLSX) {
				err = errors.New("InvalidFileExtension") //nolint
				return nil, http.StatusBadRequest, err
			}

			infile, err := hdr.Open() //nolint
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			var outfile *os.File
			fileName = fmt.Sprintf("%d", time.Now().Nanosecond()) + "_" + hdr.Filename
			fn := filepath.Join(dstDir, fileName)
			outfile, err = os.Create(fn)
			if err != nil {
				return nil, http.StatusInternalServerError, err
			}

			if _, err = io.Copy(outfile, infile); err != nil {
				if err2 := outfile.Close(); err2 != nil {
					logger.Log.Error("FileCloseFailure", zap.Error(err2))
					return nil, http.StatusInternalServerError, err2
				}
				logger.Log.Error("ContentCopyFailure", zap.Error(err))
				if err1 := os.Remove(fn); err1 != nil {
					err = err1
				}
				return nil, http.StatusInternalServerError, err
			}
			if err = outfile.Close(); err != nil {
				logger.Log.Error("FileCloseFailure", zap.Error(err))
				return nil, http.StatusInternalServerError, err
			}
			infile.Close()
		}
		ctx1, cancel := context.WithDeadline(req.Context(), time.Now().Add(time.Second*600))
		defer cancel()
		resp, err = i.dpsClient.DataAnalysis(ctx1, &v1.DataAnalysisRequest{
			Scope: scope,
			File:  fileName,
		})

		if err != nil {
			logger.Log.Error("AnalysisFailure", zap.Error(err))
			return nil, http.StatusInternalServerError, err
		}
	}

	return resp, http.StatusOK, nil
}

func (i *importServiceServer) UploadDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	// origReq := req
	dataScope := req.FormValue("scope")
	if dataScope == "" {
		logger.Log.Error("No Scope for Data")
		return
	}
	userClaims, ok := rest_middleware.RetrieveClaims(req.Context())
	if !ok {
		http.Error(res, "cannot store files", http.StatusInternalServerError)
		return
	}
	if userClaims.Role == claims.RoleUser {
		http.Error(res, "RoleValidationFailed", http.StatusForbidden)
		return
	}

	if !helper.Contains(userClaims.Socpes, dataScope) {
		http.Error(res, "ScopeValidationFailed", http.StatusForbidden)
		return
	}
	scopeinfo, err := i.accClient.GetScope(req.Context(), &v1Acc.GetScopeRequest{Scope: dataScope})
	if err != nil {
		logger.Log.Error("service/v1 - UploadDataHandler - account/GetScope - fetching scope info", zap.String("reason", err.Error()))
		http.Error(res, "Unable to get scope info", http.StatusInternalServerError)
		return
	}
	if scopeinfo.ScopeType == v1Acc.ScopeType_GENERIC.String() {
		http.Error(res, "Can not upload data for generic scope", http.StatusForbidden)
		return
	}
	uploadedBy := userClaims.UserID
	// const _24K = (1 << 20) * 24
	if parseerr := req.ParseMultipartForm(32 << 20); parseerr != nil {
		logger.Log.Error("parse multi past form ", zap.Error(parseerr))
		http.Error(res, "cannot store files", http.StatusInternalServerError)
		return
	}
	err1 := os.MkdirAll(i.config.Upload.UploadDir, os.ModePerm)
	if err1 != nil {
		logger.Log.Error("Cannot create Dir", zap.Error(err1))
		http.Error(res, "cannot upload Error", http.StatusInternalServerError)
		return
	}
	var filenames []string
	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			if hdr.Size > i.config.MaxFileSize*1024*1024 {
				logger.Log.Error("File uploaded is larger than allowed", zap.Error(err))
				http.Error(res, "maximum file allowded is :"+strconv.FormatInt(i.config.MaxFileSize, 10)+"Mbs", http.StatusBadRequest)
				return
			}
			logger.Log.Info("Import File Handler", zap.String("File", hdr.Filename), zap.String("uploadedBy", uploadedBy))
			if !helper.RegexContains(i.config.Upload.DataFileAllowedRegex, hdr.Filename) {
				logger.Log.Error("Validation Error-File Not allowed", zap.String("File", hdr.Filename))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles(dataScope, i.config.Upload.UploadDir, dataload)
				return
			}
			// open uploaded
			infile, err := hdr.Open()
			if err != nil {
				logger.Log.Error("cannot open file directory", zap.Error(err))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles(dataScope, i.config.Upload.UploadDir, dataload)
				return
			}
			// open destination
			var outfile *os.File
			fn := filepath.Join(i.config.Upload.UploadDir, dataScope+"_"+hdr.Filename)

			if outfile, err = os.Create(fn); nil != err {
				logger.Log.Error("cannot create file", zap.Error(err))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles(dataScope, i.config.Upload.UploadDir, dataload)
				return
			}
			if _, err = io.Copy(outfile, infile); nil != err {
				logger.Log.Error("cannot copy content of files", zap.Error(err))
				// if all contents are not copied remove the files
				if err := os.Remove(fn); err != nil {
					logger.Log.Error("cannot remove", zap.Error(err))
					http.Error(res, "cannot upload Error", http.StatusInternalServerError)
					removeFiles(dataScope, i.config.Upload.UploadDir, dataload)
					return
				}
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				outfile.Close()
				return
			}
			outfile.Close()
			filenames = append(filenames, fmt.Sprintf("%s_%s", dataScope, hdr.Filename))
		}
		// ctx, _ := AnnotateContext(req.Context(), origReq)
		// authStr := strings.Replace(req.Header.Get("Authorization"), "Bearer", "bearer", 1)
		// md := metadata.Pairs("Authorization", authStr)
		// ctx := metadata.NewOutgoingContext(req.Context(), md)
		// Notify call to DPS

		_, err := i.dpsClient.NotifyUpload(req.Context(), &v1.NotifyUploadRequest{
			Scope:      dataScope,
			Type:       "data",
			Files:      filenames,
			UploadedBy: uploadedBy,
		})
		if err != nil {
			logger.Log.Error("DPS call failed", zap.Error(err))
			errMsg := "InternalServerError"
			errDesc := strings.Split(err.Error(), "=")
			if strings.TrimSpace(errDesc[len(errDesc)-1]) == "Injection is already running" || strings.TrimSpace(errDesc[len(errDesc)-1]) == "Deletion is already running" {
				errMsg = errDesc[len(errDesc)-1]
			}
			http.Error(res, errMsg, http.StatusInternalServerError)
			return
		}
		res.Write([]byte("Files Uploaded")) // nolint: errcheck
	}
}

func (i *importServiceServer) UploadMetaDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	// origReq := req
	metadataScope := req.FormValue("scope")
	if metadataScope == "" {
		logger.Log.Error("No Scope for metaData")
		return
	}
	userClaims, ok := rest_middleware.RetrieveClaims(req.Context())
	if !ok {
		http.Error(res, "cannot store files", http.StatusInternalServerError)
		return
	}
	if userClaims.Role == claims.RoleUser {
		http.Error(res, "RoleValidationFailed", http.StatusForbidden)
		return
	}

	if !helper.Contains(userClaims.Socpes, metadataScope) {
		http.Error(res, "ScopeValidationFailed", http.StatusForbidden)
		return
	}
	scopeinfo, err := i.accClient.GetScope(req.Context(), &v1Acc.GetScopeRequest{Scope: metadataScope})
	if err != nil {
		logger.Log.Error("service/v1 - UploadDataHandler - account/GetScope - fetching scope info", zap.String("reason", err.Error()))
		http.Error(res, "Unable to get scope info", http.StatusInternalServerError)
		return
	}
	if scopeinfo.ScopeType == v1Acc.ScopeType_GENERIC.String() {
		http.Error(res, "Can not upload data for generic scope", http.StatusForbidden)
		return
	}
	uploadedBy := userClaims.UserID
	// const _24K = (1 << 20) * 24
	if parseerr := req.ParseMultipartForm(32 << 20); parseerr != nil {
		logger.Log.Error("parse multi past form ", zap.Error(parseerr))
		http.Error(res, "cannot store files", http.StatusInternalServerError)
		return
	}
	err1 := os.MkdirAll(i.config.Upload.UploadDir, os.ModePerm)
	if err1 != nil {
		logger.Log.Error("Cannot create Dir", zap.Error(err1))
		http.Error(res, "cannot upload Error", http.StatusInternalServerError)
		return
	}
	var filenames []string
	// for _, _ = range req.MultipartForm.File {

	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			if hdr.Size > i.config.MaxFileSize*1024*1024 {
				logger.Log.Error("File uploaded is larger than allowed", zap.Error(err))
				http.Error(res, "maximum file allowded is :"+strconv.FormatInt(i.config.MaxFileSize, 10)+"Mbs", http.StatusBadRequest)
				return
			}
			logger.Log.Info("Import MetaData File Handler", zap.String("File", hdr.Filename), zap.String("uploadedBy", uploadedBy))
			if !helper.RegexContains(i.config.Upload.MetaDatafileAllowedRegex, hdr.Filename) {
				logger.Log.Error("Validation Error-File Not allowed", zap.Any("Regex", i.config.Upload.MetaDatafileAllowedRegex), zap.String("File", hdr.Filename))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles("", i.config.Upload.UploadDir, metadataload)
				return
			}
			// 	// open uploaded
			infile, err := hdr.Open()
			if err != nil {
				logger.Log.Error("cannot open file directory", zap.Error(err))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles("", i.config.Upload.UploadDir, metadataload)
				return
			}
			// open destination
			var outfile *os.File
			fn := filepath.Join(i.config.Upload.UploadDir, metadataScope+"_"+hdr.Filename)
			if outfile, err = os.Create(fn); nil != err {
				logger.Log.Error("cannot create file", zap.Error(err))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles("", i.config.Upload.UploadDir, metadataload)
				return
			}
			if _, err = io.Copy(outfile, infile); nil != err {
				logger.Log.Error("cannot copy content of files", zap.Error(err))
				// if all contents are not copied remove the files
				if err := os.Remove(fn); err != nil {
					logger.Log.Error("cannot remove", zap.Error(err))
					http.Error(res, "cannot upload Error", http.StatusInternalServerError)
					removeFiles("", i.config.Upload.UploadDir, metadataload)
					return
				}
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				outfile.Close()
				return
			}
			outfile.Close()
			filenames = append(filenames, fmt.Sprintf("%s_%s", metadataScope, hdr.Filename))
		}
		// ctx, _ := AnnotateContext(req.Context(), origReq)
		// authStr := strings.Replace(req.Header.Get("Authorization"), "Bearer", "bearer", 1)
		// md := metadata.Pairs("Authorization", authStr)
		// ctx := metadata.NewOutgoingContext(req.Context(), md)
		// Notify call to DPS
		_, err := i.dpsClient.NotifyUpload(req.Context(), &v1.NotifyUploadRequest{
			Scope:      metadataScope,
			Type:       "metadata",
			Files:      filenames,
			UploadedBy: uploadedBy,
		})
		if err != nil {
			logger.Log.Error("DPS call failed", zap.Error(err))
		}
		res.Write([]byte("Files Uploaded")) // nolint: errcheck
	}
}

func (i *importServiceServer) CreateConfigHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	userClaims, ok := rest_middleware.RetrieveClaims(req.Context())
	if !ok {
		http.Error(res, "import/CreateConfigHandler - cannot retrieve claims", http.StatusInternalServerError)
		return
	}
	if userClaims.Role == claims.RoleUser {
		http.Error(res, "import/CreateConfigHandler - RoleValidationFailed", http.StatusForbidden)
		return
	}
	// Extract scopes from request
	scope := req.FormValue("scope")
	if scope == "" {
		logger.Log.Error("import/CreateConfigHandler - scope was empty")
		http.Error(res, "import/CreateConfigHandler - Can not find scope", http.StatusBadRequest)
		return
	}
	if !helper.Contains(userClaims.Socpes, scope) {
		http.Error(res, "import/CreateConfigHandler - Admin does not have access to scope", http.StatusUnauthorized)
		return
	}
	// Extract config_name from request
	configName := req.FormValue("config_name")
	if configName == "" {
		logger.Log.Error("import/CreateConfigHandler - Config_name is required")
		http.Error(res, "import/CreateConfigHandler - Config name is required", http.StatusBadRequest)
		return
	}

	var IsLetter = regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString

	if !IsLetter(configName) || len(configName) > 50 {
		logger.Log.Error("import/CreateConfigHandler - ConfigName has not followed validation")
		http.Error(res, "import/CreateConfigHandler - Invalid Configuration name", http.StatusBadRequest)
		return
	}

	configName = strings.ToLower(configName)

	// Extract Equipment type from request
	equipType := req.FormValue("equipment_type")

	if equipType == "" {
		logger.Log.Error("import/CreateConfigHandler - EquipType is required")
		http.Error(res, "import/CreateConfigHandler - EquipType is required", http.StatusBadRequest)
		return
	}

	// TODO : To verify that how are we gonna save equip types and how to make call to compare if the equipment type is included.

	// get auth token and add it into context
	// authToken := getAuthToken(req)
	// ctx := metadata.AppendToOutgoingContext(req.Context(), "authorization", authToken)

	// If there is no file uploaded
	if len(req.MultipartForm.File) == 0 {
		http.Error(res, "import/CreateConfigHandler - No files found", http.StatusBadRequest)
		return
	}

	configData, err := getConfigData(req.MultipartForm, res)
	if err != nil {
		return
	}

	// calling create config
	_, err = i.simClient.CreateConfig(req.Context(), &v1Simulation.CreateConfigRequest{
		ConfigName:    configName,
		EquipmentType: equipType,
		Data:          configData,
		Scope:         scope,
	})

	if err != nil {
		logger.Log.Error("import/CreateConfigHandler - simulation/CreateConfig - could not insert config data - CreateConfig()", zap.Error(err))
		http.Error(res, "import/CreateConfigHandler - simulation/CreateConfig - Could not create configuration", http.StatusInternalServerError)
		return
	}

}

func (i *importServiceServer) UpdateConfigHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	userClaims, ok := rest_middleware.RetrieveClaims(req.Context())
	if !ok {
		http.Error(res, "import/UpdateConfigHandler - cannot retrieve claims", http.StatusInternalServerError)
		return
	}
	if userClaims.Role == claims.RoleUser {
		http.Error(res, "import/UpdateConfigHandler - RoleValidationFailed", http.StatusForbidden)
		return
	}
	// Extract scopes from request
	scope := req.FormValue("scope")
	if scope == "" {
		logger.Log.Error("import/UpdateConfigHandler - scope was empty")
		http.Error(res, "import/UpdateConfigHandler - Can not find scope", http.StatusBadRequest)
		return
	}
	if !helper.Contains(userClaims.Socpes, scope) {
		http.Error(res, "import/UpdateConfigHandler - Admin does not have access to scope", http.StatusUnauthorized)
		return
	}

	configIDStr := param.ByName("config_id")

	if configIDStr == "" {
		logger.Log.Error("import/UpdateConfigHandler - Config_id is required")
		http.Error(res, "import/UpdateConfigHandler - Config ID is required", http.StatusBadRequest)
		return
	}
	configID, err := strconv.Atoi(configIDStr) // nolint: gosec
	if err != nil {
		logger.Log.Error("import/UpdateConfigHandler - Can not convert string to int")
		http.Error(res, "import/UpdateConfigHandler - Internal error", http.StatusInternalServerError)
		return
	}

	// //get auth token and add it into context
	// authToken := getAuthToken(req)
	// ctx := metadata.AppendToOutgoingContext(req.Context(), "authorization", authToken)

	// Extract deletedMetadataIDs from request
	deletedMetadataIDs := req.FormValue("deletedMetadataIDs")
	// If the request is empty
	if len(req.MultipartForm.File) == 0 && deletedMetadataIDs == "" {
		logger.Log.Error("import/UpdateConfigHandler - Request is Empty!!")
		return
	}

	deletedMetadataIDsInt := make([]int32, 0)

	if deletedMetadataIDs != "" {
		deletedMetadataIDsArray := strings.Split(deletedMetadataIDs, ",")
		deletedMetadataIDsInt, err = convertStringArrayToInt(deletedMetadataIDsArray)
		if err != nil {
			logger.Log.Error("import/UpdateConfigHandler - Can not convert string to int")
			http.Error(res, "import/UpdateConfigHandler - Internal error", http.StatusInternalServerError)
			return
		}
		deletedMetadataIDsInt = removeRepeatedElem(deletedMetadataIDsInt)
	}

	configData, err := getConfigData(req.MultipartForm, res)
	if err != nil {
		return
	}

	// calling update config
	_, err = i.simClient.UpdateConfig(req.Context(), &v1Simulation.UpdateConfigRequest{
		ConfigId:           int32(configID),
		DeletedMetadataIds: deletedMetadataIDsInt,
		Data:               configData,
		Scope:              scope,
	})

	if err != nil {
		logger.Log.Error("import/UpdateConfigHandler - simulation/UpdateConfig - could not update config - UpdateConfig()", zap.Error(err))
		http.Error(res, "import/UpdateConfigHandler - simulation/UpdateConfig - Internal Error", http.StatusInternalServerError)
		return
	}

}

func removeFiles(scope string, dir string, datatype uploadType) {
	logger.Log.Info("Removing Files", zap.String("Scope", scope))
	var delFilesRegex string
	if datatype == "data" { // nolint: gocritic
		delFilesRegex = scope + "_*"
	} else if datatype == "globaldata" {
		delFilesRegex = "*"
	} else {
		delFilesRegex = "metadata_*"
	}
	files, err := filepath.Glob(path.Join(dir, delFilesRegex))
	if err != nil {
		logger.Log.Error("Failed to list files", zap.Error(err))
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			logger.Log.Error("Failed to list files", zap.Error(err))
		}
	}
}

func getConfigValueObject(configFile *csv.Reader, columns []string) ([]*v1Simulation.ConfigValue, error) {
	var configObject []*v1Simulation.ConfigValue
	for {
		var values = make(map[string]string, len(columns))
		record, err := configFile.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		key := record[0]
		for i := range record {
			values[columns[i]] = record[i]
		}
		byteValues, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}
		configObject = append(configObject, &v1Simulation.ConfigValue{
			Key:   key,
			Value: byteValues,
		})
	}
	return configObject, nil
}

func convertStringArrayToInt(deletedMetadataIDs []string) ([]int32, error) {
	res := make([]int32, len(deletedMetadataIDs))
	for _, id := range deletedMetadataIDs {
		intID, err := strconv.Atoi(id) // nolint: gosec
		if err != nil {
			return nil, err
		}
		res = append(res, int32(intID))
	}

	return res, nil
}

func removeRepeatedElem(array []int32) []int32 {
	var hmap = make(map[int32]int32, len(array))

	for i := 0; i < len(array); i++ {
		_, ok := hmap[array[i]]
		if ok == true {
			hmap[array[i]] = 1
		} else {
			hmap[array[i]] = 0
		}
	}

	var res = make([]int32, 0, len(array))
	for key := range hmap {
		res = append(res, key)

	}

	// fmt.Println(res)
	return res

}

func getConfigData(multipartForm *multipart.Form, res http.ResponseWriter) ([]*v1Simulation.Data, error) {
	configData := []*v1Simulation.Data{}
	var attrMap = make(map[string]int, len(multipartForm.File))
	// Loop through files
	for attrName, fHeaders := range multipartForm.File {
		for _, hdr := range fHeaders {

			// Extract fileName from header
			fileName := hdr.Filename

			// Handled the case of when more than one file is uploaded for single attribute
			_, ok := attrMap[attrName]

			if ok == false {
				attrMap[attrName] = 1
			} else {
				http.Error(res, "Only one file per attribute is allowed", http.StatusBadRequest)
				return nil, errors.New("error")
			}
			configFile, err := hdr.Open()
			if err != nil {
				logger.Log.Error("Can not open file - Open() ", zap.Error(err))
				http.Error(res, "can not open file", http.StatusInternalServerError)
				return nil, errors.New("error")
			}
			defer configFile.Close()

			// parse the file
			configCsv := csv.NewReader(configFile)
			configCsv.Comma = ';'

			columns, err := configCsv.Read()
			if err == io.EOF {
				logger.Log.Error("config file is empty ", zap.Error(err))
				http.Error(res, "config file is empty", http.StatusNotFound)
				return nil, errors.New("error")
			}
			if err != nil {
				logger.Log.Error("can not read config file - Read() ", zap.Error(err))
				http.Error(res, "can not read config file", http.StatusUnprocessableEntity)
				return nil, errors.New("error")
			}
			if columns[0] != attrName {
				http.Error(res, "can not read config file", http.StatusUnprocessableEntity)
				return nil, errors.New("error")
			}

			// Get config values object
			configValues, err := getConfigValueObject(configCsv, columns)
			if err != nil {
				logger.Log.Error("Error in reading config file ", zap.Error(err))
				http.Error(res, "can not read config file", http.StatusUnprocessableEntity)
				return nil, errors.New("error")
			}

			// Making request array
			data := &v1Simulation.Data{
				Metadata: &v1Simulation.Metadata{
					AttributeName:  attrName,
					ConfigFilename: fileName,
				},
				Values: configValues,
			}

			configData = append(configData, data)

		}
	}

	return configData, nil
}

func (i *importServiceServer) UploadGlobalDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) { //nolint
	var globalFileDir, genericFile, analysisID string
	var filenames []string
	var hdrs []*multipart.FileHeader
	stype := v1.NotifyUploadRequest_GENERIC

	scope := req.FormValue("scope")
	if scope == "" {
		http.Error(res, "ScopeIsMissing", http.StatusBadRequest)
		return
	}

	userClaims, ok := rest_middleware.RetrieveClaims(req.Context())
	if !ok {
		http.Error(res, "ClaimsNotFound", http.StatusBadRequest)
		return
	}

	if !helper.Contains(userClaims.Socpes, scope) {
		http.Error(res, "ScopeValidationFailed", http.StatusForbidden)
		return
	}

	scopeInfo, err := i.accClient.GetScope(req.Context(), &v1Acc.GetScopeRequest{Scope: scope})
	if err != nil {
		logger.Log.Error("service/v1 - UploadGlobalDataHandler - account/GetScope - fetching scope info", zap.String("reason", err.Error()))
		http.Error(res, "InternalError", http.StatusInternalServerError)
		return
	}

	uploadedBy := userClaims.UserID
	if err = req.ParseMultipartForm(32 << 20); nil != err {
		logger.Log.Debug("parsing multipartFrom Error :", zap.Error(err))
		http.Error(res, "FormParsingError", http.StatusInternalServerError)
		return
	}

	if scopeInfo.ScopeType == GENERIC {
		genericFile = req.FormValue("file")
		if genericFile == "" {
			logger.Log.Debug("FileNameMissing")
			http.Error(res, "FileNameMissing", http.StatusBadRequest)
			return
		}
		if !strings.Contains(genericFile, XLSX) {
			logger.Log.Debug("InvalidFileReceived")
			http.Error(res, "InvalidFileReceived", http.StatusBadRequest)
			return
		}
	}

	if scopeInfo.ScopeType != GENERIC {
		stype = v1.NotifyUploadRequest_SPECIFIC
		globalFileDir = fmt.Sprintf("%s/%s", i.config.Upload.RawDataUploadDir, scope)
		err = os.MkdirAll(globalFileDir, os.ModePerm)
		if err != nil {
			logger.Log.Debug("Cannot create Dir, Error :", zap.Error(err))
			http.Error(res, "DirCreationError", http.StatusInternalServerError)
			return
		}
		for _, fheaders := range req.MultipartForm.File {
			for _, hdr := range fheaders {
				if hdr.Size > i.config.MaxFileSize*1024*1024 {
					logger.Log.Error("File uploaded is larger than allowed", zap.Error(err))
					http.Error(res, "maximum file allowded is :"+strconv.FormatInt(i.config.MaxFileSize, 10)+"Mbs", http.StatusBadRequest)
					return
				}
				ext := getglobalFileExtension(hdr.Filename)
				if stype == v1.NotifyUploadRequest_GENERIC {
					if !strings.Contains(ext, XLSX) {
						http.Error(res, "GenerifcFileExtensionValidationFailure", http.StatusBadRequest)
						return
					}
				} else {
					if !strings.Contains(ext, CSV) {
						http.Error(res, "SpecificFileExtensionValidationFailure", http.StatusBadRequest)
						return
					}
				}
				hdrs = append(hdrs, hdr)
				filenames = append(filenames, hdr.Filename)
			}
		}
	} else {
		temp := strings.Split(genericFile, "_")
		if len(temp) < 3 {
			logger.Log.Debug("UnknownFileReceived", zap.String("expectation", "good_time_file.xlsx"))
			http.Error(res, "UnknownFileReceived", http.StatusBadRequest)
			return
		}
		analysisID = temp[1]
		temp = temp[2:]
		filenames = append(filenames, strings.Join(temp, "_"))
		logger.Log.Debug("parsing from generic file", zap.String("targetFile", filenames[0]), zap.String("analysis_id", analysisID))
	}

	dpsResp, err := i.dpsClient.NotifyUpload(req.Context(), &v1.NotifyUploadRequest{
		Scope:      scope,
		Type:       "globaldata",
		Files:      filenames,
		UploadedBy: uploadedBy,
		ScopeType:  stype,
		AnalysisId: analysisID,
	})
	if err != nil {
		logger.Log.Debug("DPS globaldata failed", zap.Error(err))
		errMsg := "InternalServerError"
		errDesc := strings.Split(err.Error(), "=")
		if strings.TrimSpace(errDesc[len(errDesc)-1]) == "Injection is already running" || strings.TrimSpace(errDesc[len(errDesc)-1]) == "Deletion is already running" {
			errMsg = errDesc[len(errDesc)-1]
		}
		http.Error(res, errMsg, http.StatusInternalServerError)
		return
	}

	if scopeInfo.ScopeType != GENERIC {
		var fileName string
		for _, hdr := range hdrs {
			if scopeInfo.ScopeType == GENERIC {
				fileName = filenames[0]
			} else {
				fileName = hdr.Filename
			}
			infile, err := hdr.Open()
			if err != nil {
				logger.Log.Debug("cannot open file hdr", zap.Error(err), zap.String("file", fileName))
				http.Error(res, "FileFormHeaderError", http.StatusInternalServerError)
				removeFiles("", globalFileDir, rawdataload)
				return
			}
			// open destination
			var outfile *os.File
			fn := filepath.Join(globalFileDir, fmt.Sprintf("%d_%s", dpsResp.FileUploadId[fileName], fileName))
			if outfile, err = os.Create(fn); nil != err {
				logger.Log.Debug("cannot create file", zap.Error(err), zap.String("file", fileName))
				http.Error(res, "FileCreationError", http.StatusInternalServerError)
				removeFiles("", globalFileDir, rawdataload)
				return
			}
			if _, err = io.Copy(outfile, infile); nil != err {
				logger.Log.Debug("cannot copy content of files", zap.Error(err), zap.String("file", fileName))
				if err := os.Remove(fn); err != nil {
					logger.Log.Debug("cannot remove", zap.Error(err), zap.String("file", fileName))
					http.Error(res, "FileRemovingError", http.StatusInternalServerError)
					removeFiles("", globalFileDir, rawdataload)
					return
				}
				http.Error(res, "ContentCopyFailure", http.StatusInternalServerError)
				outfile.Close()
				return
			}
			outfile.Close()
			res.Write([]byte(fmt.Sprintf("%s file uploaded\n", fileName))) // nolint: errcheck
		}
	} else {

		dst := fmt.Sprintf("%s/GEN/%s_%d_%s", i.config.Upload.RawDataUploadDir, scope, dpsResp.FileUploadId[filenames[0]], filenames[0])
		src := fmt.Sprintf("%s/%s/analysis/%s", i.config.Upload.RawDataUploadDir, scope, genericFile)
		logger.Log.Error("storing global file for nifi", zap.String("dst", dst), zap.String("src", src))
		if err := os.Rename(src, dst); err != nil {
			logger.Log.Error("Failed to move generic file from analysis to nifi src dir", zap.Error(err))
			http.Error(res, "ContentCopyFailure", http.StatusInternalServerError)
		}
	}

}

func getglobalFileExtension(fileName string) string {
	if fileName == "" {
		return ""
	}
	temp := strings.SplitAfter(fileName, ".")
	if len(temp) < 2 {
		return ""
	}
	return fmt.Sprintf(".%s", temp[len(temp)-1])
}

func (i *importServiceServer) DownloadFile(res http.ResponseWriter, req *http.Request, param httprouter.Params) { // nolint
	uploadID := ""
	fileName := ""
	scopeType := ""
	userClaims, ok := rest_middleware.RetrieveClaims(req.Context())
	if !ok {
		logger.Log.Error("ClaimsNotFound")
		http.Error(res, "ClaimsNotFound", http.StatusInternalServerError)
		return
	}
	scope := req.FormValue("scope")
	if scope == "" {
		logger.Log.Error("ScopeIsMissing")
		http.Error(res, "ScopeIsMissing", http.StatusBadRequest)
		return
	}
	if !helper.Contains(userClaims.Socpes, scope) {
		logger.Log.Error("ScopeValidationFailed")
		http.Error(res, "ScopeValidationFailed", http.StatusUnauthorized)
		return
	}
	if userClaims.Role == claims.RoleUser {
		logger.Log.Error("RoleValidationFailed")
		http.Error(res, "RoleValidationFailed", http.StatusForbidden)
		return
	}

	downloadType := req.FormValue("downloadType")
	if downloadType == "" {
		logger.Log.Error("downloadTypeIsMissing")
		http.Error(res, "downloadTypeIsMissing", http.StatusBadRequest)
		return
	}
	var isOlderGeneric bool
	if downloadType == string(errorFile) || downloadType == string(source) {
		uploadID = req.FormValue("uploadId")
		if uploadID == "" {
			logger.Log.Error("UploadIdIsMissing")
			http.Error(res, "UploadIdIsMissing", http.StatusBadRequest)
			return
		}
		id, err := strconv.ParseInt(uploadID, 10, 64)
		if err != nil {
			logger.Log.Error("BadUploadIdReceived")
			http.Error(res, "BadUploadIdReceived", http.StatusBadRequest)
			return
		}
		resp, err := i.dpsClient.GetAnalysisFileInfo(req.Context(), &v1.GetAnalysisFileInfoRequest{
			Scope:    scope,
			UploadId: int32(id),
			FileType: downloadType,
		})
		if err != nil {
			logger.Log.Error("Failed to get fileInfo", zap.Error(err), zap.String("uploadID", uploadID), zap.String("downloadType", downloadType))
			http.Error(res, "Failed to get fileInfo", http.StatusInternalServerError)
			return
		}
		fileName = resp.FileName
		scopeType = resp.ScopeType
		isOlderGeneric = resp.IsOlderGeneric
	} else if downloadType == string(analysis) {
		fileName = req.FormValue("fileName")
		if scope == "" {
			logger.Log.Error("FileNameIsMissing")
			http.Error(res, "FileNameIsMissing", http.StatusBadRequest)
			return
		}
	}
	fileLocation := ""
	switch string(downloadType) { //nolint
	case string(errorFile):
		fileLocation = path.Join(i.config.Upload.RawDataUploadDir, scope, "errors", fileName)
	case string(analysis):
		fileLocation = path.Join(i.config.Upload.RawDataUploadDir, scope, "analysis", fileName)
	case string(source):
		if scopeType == GENERIC {
			if isOlderGeneric { // older generic files
				fileRegex := fileName + "*"
				fileLocation = path.Join(i.config.Upload.RawDataUploadDir, "GEN", "archive", fileRegex)
			} else {
				fileLocation = path.Join(i.config.Upload.RawDataUploadDir, scope, "analysis", fileName)
			}
		} else {
			fileRegex := fileName + "*"
			fileLocation = path.Join(i.config.Upload.RawDataUploadDir, scope, "archive", fileRegex)
		}
	default:
		http.Error(res, "InvalidDownloadTypeReceived", http.StatusBadRequest)
		return
	}
	logger.Log.Debug("looking for file ", zap.String("filelocation", fileLocation))
	file, err := filepath.Glob(fileLocation)
	if err != nil || file == nil {
		logger.Log.Error("Download - File does not exist", zap.Error(err), zap.String("file", fileLocation))
		http.Error(res, "File does not exist", http.StatusNotFound)
		return
	}
	if scopeType != GENERIC {
		fileLocation = file[0]
	}
	fileData, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		logger.Log.Error("Download - error in reading file", zap.Error(err), zap.String("file", fileLocation))
		http.Error(res, "error in reading file", http.StatusInternalServerError)
		return
	}
	http.ServeContent(res, req, fileName, time.Now().UTC(), bytes.NewReader(fileData))
	return
}

func (i *importServiceServer) UploadCatalogData(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	var err error

	userClaims, ok := rest_middleware.RetrieveClaims(req.Context())
	if !ok {
		logger.Log.Error("ClaimsNotFound")
		http.Error(res, "ClaimsNotFound", http.StatusBadRequest)
		return
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		logger.Log.Error("Role Validation Error")
		http.Error(res, "Role Validation Error", http.StatusBadRequest)
		return
	}

	if err = req.ParseMultipartForm(32 << 20); nil != err {
		logger.Log.Error("ParsingFailure", zap.Error(err))
		http.Error(res, "ParsingFailure", http.StatusInternalServerError)
		return
	}
	var status int
	var resp interface{}

	resp, status, err = saveCatalogProducts(i, req)
	if err != nil {
		logger.Log.Error("Failed to upload file ", zap.Error(err))
		http.Error(res, err.Error(), status)
		return
	}
	if err != nil {
		logger.Log.Error("Failed to upload file ", zap.Error(err))
		http.Error(res, err.Error(), status)
		return
	}
	out, jrr := json.Marshal(resp)
	if jrr != nil {
		logger.Log.Error("Failed to marshal the response", zap.Error(jrr))
		http.Error(res, "ResponseParsingFailure", http.StatusInternalServerError)
	}
	res.Write(out) //nolint
}

func saveCatalogProducts(i *importServiceServer, req *http.Request) (*v1Catalog.UploadResponse, int, error) {
	file, fileInfo, err := req.FormFile("file")
	defer file.Close()
	if fileInfo.Size > i.config.MaxFileSize*1024*1024 {
		logger.Log.Error("File uploaded is larger than allowed", zap.Error(err))
		return nil, http.StatusBadRequest, errors.New("maximum file allowded is :" + strconv.FormatInt(i.config.MaxFileSize, 10) + "Mbs")
	}
	if err != nil {
		logger.Log.Error("Failed to read reference file", zap.Error(err))
		return nil, http.StatusBadRequest, err
	}
	f, err := excelize.OpenReader(file)
	if err != nil {
		logger.Log.Error("Failed to parse reference file", zap.Error(err))
		return nil, http.StatusBadRequest, err
	}
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		logger.Log.Error("Reference file doesn't have any sheet", zap.Error(err))
		return nil, http.StatusBadRequest, err
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		logger.Log.Error("Failed to read the sheet", zap.Error(err), zap.String("sheet", sheets[0]))
		return nil, http.StatusInternalServerError, err
	} else if len(rows) < 2 {
		logger.Log.Error("inapropiate sheet, no reference value found", zap.Error(err), zap.String("sheet", sheets[0]))
		return nil, http.StatusInternalServerError, err
	}
	//save headers index and validate all present

	headers := rows[0]
	headersindexarray := make([]int, len(headers))
	var headerscount int
	for index := 0; index < len(headers); index++ {
		switch strings.ToLower(headers[index]) {
		case "editor":
			headersindexarray[0] = index
			headerscount = headerscount + 1
		case "name":
			headersindexarray[1] = index
			headerscount = headerscount + 1
		case "licensing":
			headersindexarray[2] = index
			headerscount = headerscount + 1
		case "general information":
			headersindexarray[3] = index
			headerscount = headerscount + 1
		case "version":
			headersindexarray[4] = index
			headerscount = headerscount + 1
		case "eol":
			headersindexarray[5] = index
			headerscount = headerscount + 1
		case "eos":
			headersindexarray[6] = index
			headerscount = headerscount + 1
		}
	}
	if headerscount < 7 {
		err = errors.New("missing headers")
		logger.Log.Error("unable to import catalog products from sheet", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	rows = rows[1:]
	dataToSend := v1Catalog.UploadRecords{}
	for _, v := range rows {
		if len(v) == 0 {
			continue
		}
		gn := (greaternumber(headersindexarray[0], headersindexarray[1]) + 1)
		if len(v) >= gn {
			if v[headersindexarray[0]] == "" || v[headersindexarray[1]] == "" {
				logger.Log.Info("Wrong Number of arguments")
				continue
			}
		} else {
			logger.Log.Info("Wrong Number of arguments")
			continue

		}
		var eoltime, eostime time.Time
		if len(v) > headersindexarray[5] {
			if strings.Contains(v[headersindexarray[5]], "/") {
				v[headersindexarray[5]] = strings.ReplaceAll(v[headersindexarray[5]], "/", "-")
			}
			eoltime, _ = time.Parse("02-01-2006", v[headersindexarray[5]])
		} else {
			eoltime, _ = time.Parse("02-01-2006", "")
		}
		eoltimeObject, err := ptypes.TimestampProto(eoltime)
		if err != nil {
			logger.Log.Error("unable to import process record", zap.Error(err))
			continue
		}
		if len(v) > headersindexarray[6] {
			if strings.Contains(v[headersindexarray[6]], "/") {
				v[headersindexarray[6]] = strings.ReplaceAll(v[headersindexarray[6]], "/", "-")
			}
			eostime, _ = time.Parse("02-01-2006", v[headersindexarray[6]])
		} else {
			eoltime, _ = time.Parse("02-01-2006", "")
		}
		eostimeObject, err := ptypes.TimestampProto(eostime)
		if err != nil {
			logger.Log.Error("unable to import process record", zap.Error(err))
			continue
		}
		var version string
		if len(v) > headersindexarray[4] {
			version = v[headersindexarray[4]]
		}
		var generalInfo, licensing string
		if len(v) > headersindexarray[3] {
			generalInfo = v[headersindexarray[3]]
		}
		if len(v) > headersindexarray[2] {
			licensing = v[headersindexarray[2]]
		}
		row := v1Catalog.Upload{
			Editor:             v[headersindexarray[0]],
			Name:               v[headersindexarray[1]],
			Licensing:          licensing,
			GenearlInformation: generalInfo,
			Version:            version,
			EndOfLife:          eoltimeObject,
			EndOfSupport:       eostimeObject,
		}
		dataToSend.Data = append(dataToSend.Data, &row)
	}
	logger.Log.Info("v1/service - Calling Catalog Bulk Import" + fmt.Sprint(time.Now()))
	dataToSend.FileName = fileInfo.Filename
	resp, err := i.catalogClient.BulkFileUpload(req.Context(), &dataToSend)
	if err != nil {
		logger.Log.Error(" unable to import catalog products from sheet", zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}
	logger.Log.Info("v1/service - Calling Catalog Bulk Import Finished " + fmt.Sprint(time.Now()))

	return resp, http.StatusOK, nil
}

func greaternumber(num1 int, num2 int) int {
	if num1 > num2 {
		return num1
	}
	return num2
}
