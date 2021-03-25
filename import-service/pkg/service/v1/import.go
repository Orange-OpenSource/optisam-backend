// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/logger"
	rest_middleware "optisam-backend/common/optisam/middleware/rest"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"optisam-backend/import-service/pkg/config"
	v1Simulation "optisam-backend/simulation-service/pkg/api/v1"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type importServiceServer struct {
	// grpcServers map[string]*grpc.ClientConn
	dpsClient v1.DpsServiceClient
	simClient v1Simulation.SimulationServiceClient
	config    *config.Config
}

type uploadType string

const (
	metadataload uploadType = "metadata"
	dataload     uploadType = "data"
	rawdataload  uploadType = "globaldata"
)

var (
	globalFileExtensions []string = []string{"xlsx", "csv"}
)

// NewImportServiceServer creates import service
func NewImportServiceServer(grpcServers map[string]*grpc.ClientConn, config *config.Config) ImportServiceServer {
	return &importServiceServer{config: config, dpsClient: v1.NewDpsServiceClient(grpcServers["dps"]),
		simClient: v1Simulation.NewSimulationServiceClient(grpcServers["simulation"])}
}

func (i *importServiceServer) UploadDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	//origReq := req
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

	uploadedBy := userClaims.UserID
	// const _24K = (1 << 20) * 24
	if err := req.ParseMultipartForm(32 << 20); nil != err {
		logger.Log.Error("parse multi past form ", zap.Error(err))
		http.Error(res, "cannot store files", http.StatusInternalServerError)
		return
	}
	err := os.MkdirAll(i.config.Upload.UploadDir, os.ModePerm)
	if err != nil {
		logger.Log.Error("Cannot create Dir", zap.Error(err))
		http.Error(res, "cannot upload Error", http.StatusInternalServerError)
		return
	}
	var filenames []string
	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
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
		//ctx, _ := AnnotateContext(req.Context(), origReq)
		authStr := strings.Replace(req.Header.Get("Authorization"), "Bearer", "bearer", 1)
		md := metadata.Pairs("Authorization", authStr)
		ctx := metadata.NewOutgoingContext(req.Context(), md)
		// Notify call to DPS

		_, err := i.dpsClient.NotifyUpload(ctx, &v1.NotifyUploadRequest{
			Scope:      dataScope,
			Type:       "data",
			Files:      filenames,
			UploadedBy: uploadedBy,
		})
		if err != nil {
			logger.Log.Error("DPS call failed", zap.Error(err))
		}
		res.Write([]byte("Files Uploaded"))
	}
}

func (i *importServiceServer) UploadMetaDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	//origReq := req
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

	uploadedBy := userClaims.UserID
	// const _24K = (1 << 20) * 24
	if err := req.ParseMultipartForm(32 << 20); nil != err {
		logger.Log.Error("parse multi past form ", zap.Error(err))
		http.Error(res, "cannot store files", http.StatusInternalServerError)
		return
	}
	err := os.MkdirAll(i.config.Upload.UploadDir, os.ModePerm)
	if err != nil {
		logger.Log.Error("Cannot create Dir", zap.Error(err))
		http.Error(res, "cannot upload Error", http.StatusInternalServerError)
		return
	}
	var filenames []string
	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			logger.Log.Info("Import MetaData File Handler", zap.String("File", hdr.Filename), zap.String("uploadedBy", uploadedBy))
			if !helper.RegexContains(i.config.Upload.MetaDatafileAllowedRegex, hdr.Filename) {
				logger.Log.Error("Validation Error-File Not allowed", zap.Any("Regex", i.config.Upload.MetaDatafileAllowedRegex), zap.String("File", hdr.Filename))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles("", i.config.Upload.UploadDir, metadataload)
				return
			}
			// open uploaded
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
		//ctx, _ := AnnotateContext(req.Context(), origReq)
		authStr := strings.Replace(req.Header.Get("Authorization"), "Bearer", "bearer", 1)
		md := metadata.Pairs("Authorization", authStr)
		ctx := metadata.NewOutgoingContext(req.Context(), md)
		// Notify call to DPS
		_, err := i.dpsClient.NotifyUpload(ctx, &v1.NotifyUploadRequest{
			Scope:      metadataScope,
			Type:       "metadata",
			Files:      filenames,
			UploadedBy: uploadedBy,
		})
		if err != nil {
			logger.Log.Error("DPS call failed", zap.Error(err))
		}
		res.Write([]byte("Files Uploaded"))
	}
}

func (h *importServiceServer) CreateConfigHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {

	// Extract scopes from request
	scopesString := req.FormValue("scopes")

	if scopesString == "" {
		logger.Log.Error("Scopes were empty")
		// Ques : Is this error code right?
		http.Error(res, "Can not find scopes", http.StatusBadRequest)
		return
	}

	// // convert it into an array of scopes
	// scopes := strings.Split(scopesString, ",")

	//Extract config_name from request
	configName := req.FormValue("config_name")

	if configName == "" {
		logger.Log.Error("Config_name is required")
		http.Error(res, "Config name is required", http.StatusBadRequest)
		return
	}

	var IsLetter = regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString

	if !IsLetter(configName) || len(configName) > 50 {
		logger.Log.Error("ConfigName has not followed validation")
		http.Error(res, "Invalid Configuration name", http.StatusBadRequest)
		return
	}

	configName = strings.ToLower(configName)

	//Extract Equipment type from request
	equipType := req.FormValue("equipment_type")

	if equipType == "" {
		logger.Log.Error("EquipType is required")
		http.Error(res, "EquipType is required", http.StatusBadRequest)
		return
	}

	//TODO : To verify that how are we gonna save equip types and how to make call to compare if the equipment type is included.

	//get auth token and add it into context
	authToken := getAuthToken(req)
	ctx := metadata.AppendToOutgoingContext(req.Context(), "authorization", authToken)

	// If there is no file uploaded
	if len(req.MultipartForm.File) == 0 {
		http.Error(res, "No files found", http.StatusBadRequest)
		return
	}

	configData, err := getConfigData(req.MultipartForm, res)
	if err != nil {
		return
	}

	// calling create config
	_, err = h.simClient.CreateConfig(ctx, &v1Simulation.CreateConfigRequest{
		ConfigName:    configName,
		EquipmentType: equipType,
		Data:          configData,
	})

	if err != nil {
		logger.Log.Error("could not insert config data - CreateConfig()", zap.Error(err))
		http.Error(res, "Could not create configuration", http.StatusInternalServerError)
		return
	}

}

func (h *importServiceServer) UpdateConfigHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {

	// Extract scopes from request
	scopesString := req.FormValue("scopes")

	if scopesString == "" {
		logger.Log.Error("Scopes were empty")
		// Ques : Is this error code right?
		http.Error(res, "Can not find scopes", http.StatusBadRequest)
		return
	}
	// // convert it into an array of scopes
	// scopes := strings.Split(scopesString, ",")

	configIDStr := param.ByName("config_id")

	if configIDStr == "" {
		logger.Log.Error("Config_id is required")
		http.Error(res, "Config ID is required", http.StatusBadRequest)
		return
	}
	configID, err := strconv.Atoi(configIDStr)
	if err != nil {
		logger.Log.Error("Can not convert string to int")
		http.Error(res, "Internal error", http.StatusInternalServerError)
		return
	}

	//get auth token and add it into context
	authToken := getAuthToken(req)
	ctx := metadata.AppendToOutgoingContext(req.Context(), "authorization", authToken)

	//Extract deletedMetadataIDs from request
	deletedMetadataIDs := req.FormValue("deletedMetadataIDs")
	// If the request is empty
	if len(req.MultipartForm.File) == 0 && deletedMetadataIDs == "" {
		logger.Log.Error("Request is Empty!!")
		return
	}

	deletedMetadataIDsInt := make([]int32, 0)

	if deletedMetadataIDs != "" {
		deletedMetadataIDsArray := strings.Split(deletedMetadataIDs, ",")
		deletedMetadataIDsInt, err = convertStringArrayToInt(deletedMetadataIDsArray)
		if err != nil {
			logger.Log.Error("Can not convert string to int")
			http.Error(res, "Internal error", http.StatusInternalServerError)
			return
		}
		deletedMetadataIDsInt = removeRepeatedElem(deletedMetadataIDsInt)
	}

	configData, err := getConfigData(req.MultipartForm, res)
	if err != nil {
		return
	}

	// calling update config
	_, err = h.simClient.UpdateConfig(ctx, &v1Simulation.UpdateConfigRequest{
		ConfigId:           int32(configID),
		DeletedMetadataIds: deletedMetadataIDsInt,
		Data:               configData,
	})

	if err != nil {
		logger.Log.Error("could not update config - UpdateConfig()", zap.Error(err))
		http.Error(res, "Internal Error", http.StatusInternalServerError)
		return
	}

}

func removeFiles(scope string, dir string, datatype uploadType) {
	logger.Log.Info("Removing Files", zap.String("Scope", scope))
	var delFilesRegex string
	if datatype == "data" {
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
	var res []int32
	for _, id := range deletedMetadataIDs {
		intID, err := strconv.Atoi(id)
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
	for key, _ := range hmap {
		res = append(res, key)

	}

	fmt.Println(res)
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
				return nil, errors.New("Error")
			}
			configFile, err := hdr.Open()
			if err != nil {
				logger.Log.Error("Can not open file - Open() ", zap.Error(err))
				http.Error(res, "can not open file", http.StatusInternalServerError)
				return nil, errors.New("Error")
			}
			defer configFile.Close()

			//parse the file
			configCsv := csv.NewReader(configFile)
			configCsv.Comma = ';'

			columns, err := configCsv.Read()
			if err == io.EOF {
				logger.Log.Error("config file is empty ", zap.Error(err))
				http.Error(res, "config file is empty", http.StatusNotFound)
				return nil, errors.New("Error")
			}
			if err != nil {
				logger.Log.Error("can not read config file - Read() ", zap.Error(err))
				http.Error(res, "can not read config file", http.StatusUnprocessableEntity)
				return nil, errors.New("Error")
			}
			if columns[0] != attrName {
				http.Error(res, "can not read config file", http.StatusUnprocessableEntity)
				return nil, errors.New("Error")
			}

			// Get config values object
			configValues, err := getConfigValueObject(configCsv, columns)
			if err != nil {
				logger.Log.Error("Error in reading config file ", zap.Error(err))
				http.Error(res, "can not read config file", http.StatusUnprocessableEntity)
				return nil, errors.New("Error")
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

func getAuthToken(req *http.Request) string {
	bearerToken := req.Header.Get("Authorization")
	authToken := strings.TrimPrefix(bearerToken, "Bearer")
	authToken = strings.TrimSpace(authToken)
	authToken = "bearer " + authToken

	return authToken
}
func (i *importServiceServer) UploadGlobalDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	//origReq := req
	scope := req.FormValue("scope")
	if scope == "" {
		http.Error(res, "ScopeIsMissing", http.StatusBadRequest)
		return
	}
	isDeleteOldInventory := req.FormValue("isDeleteOldInventory") == "true"

	userClaims, ok := rest_middleware.RetrieveClaims(req.Context())
	if !ok {
		http.Error(res, "ClaimsNotFound", http.StatusBadRequest)
		return
	}
	if userClaims.Role == claims.RoleUser {
		http.Error(res, "RoleValidationFailed", http.StatusForbidden)
		return
	}
	
	if !helper.Contains(userClaims.Socpes, scope) {
		http.Error(res, "ScopeValidationFailed", http.StatusForbidden)
		return
	}

	uploadedBy := userClaims.UserID
	if err := req.ParseMultipartForm(32 << 20); nil != err {
		logger.Log.Debug("parsing multipartFrom Error :", zap.Error(err))
		http.Error(res, "FormParsingError", http.StatusInternalServerError)
		return
	}
	globalFileDir := fmt.Sprintf("%s/%s", i.config.Upload.RawDataUploadDir, scope)
	err := os.MkdirAll(globalFileDir, os.ModePerm)
	if err != nil {
		logger.Log.Debug("Cannot create Dir, Error :", zap.Error(err))
		http.Error(res, "DirCreationError", http.StatusInternalServerError)
		return
	}
	var filenames []string
	var hdrs []*multipart.FileHeader
	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			ext := getglobalFileExtension(hdr.Filename)
			if !helper.Contains(globalFileExtensions, ext) {
				http.Error(res, "FileExtensionValidationFailure", http.StatusInternalServerError)
				return
			}
			hdrs = append(hdrs, hdr)
			filenames = append(filenames, hdr.Filename)
		}
	}
	authStr := strings.Replace(req.Header.Get("Authorization"), "Bearer", "bearer", 1)
	md := metadata.Pairs("Authorization", authStr)
	ctx := metadata.NewOutgoingContext(req.Context(), md)
	_, err = i.dpsClient.NotifyUpload(ctx, &v1.NotifyUploadRequest{
		Scope:                scope,
		Type:                 "globaldata",
		Files:                filenames,
		UploadedBy:           uploadedBy,
		IsDeleteOldInventory: isDeleteOldInventory,
	})
	if err != nil {
		logger.Log.Debug("DPS globaldata failed", zap.Error(err))
		http.Error(res, "InternalServerError", http.StatusInternalServerError)
		return
	}

	for _, hdr := range hdrs {
		infile, err := hdr.Open()
		if err != nil {
			logger.Log.Debug("cannot open file hdr", zap.Error(err), zap.String("file", hdr.Filename))
			http.Error(res, "FileFormHeaderError", http.StatusInternalServerError)
			removeFiles("", globalFileDir, rawdataload)
			return
		}
		// open destination
		var outfile *os.File
		fn := filepath.Join(globalFileDir, hdr.Filename)
		if outfile, err = os.Create(fn); nil != err {
			logger.Log.Debug("cannot create file", zap.Error(err), zap.String("file", hdr.Filename))
			http.Error(res, "FileCreationError", http.StatusInternalServerError)
			removeFiles("", globalFileDir, rawdataload)
			return
		}
		if _, err = io.Copy(outfile, infile); nil != err {
			logger.Log.Debug("cannot copy content of files", zap.Error(err), zap.String("file", hdr.Filename))
			if err := os.Remove(fn); err != nil {
				logger.Log.Debug("cannot remove", zap.Error(err), zap.String("file", hdr.Filename))
				http.Error(res, "FileRemovingError", http.StatusInternalServerError)
				removeFiles("", globalFileDir, rawdataload)
				return
			}
			http.Error(res, "ContentCopyFailure", http.StatusInternalServerError)
			outfile.Close()
			return
		}
		outfile.Close()
		res.Write([]byte(fmt.Sprintf("%s file uploaded\n", hdr.Filename)))
	}

}

func getglobalFileExtension(fileName string) string {
	if fileName == "" {
		return ""
	}
	temp := strings.Split(fileName, ".")
	if len(temp) < 2 {
		return ""
	}
	return temp[len(temp)-1]
}
