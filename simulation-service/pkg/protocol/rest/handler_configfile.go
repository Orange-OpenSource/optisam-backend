// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package rest

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/simulation-service/pkg/api/v1"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type handler struct {
	client v1.SimulationServiceClient
}

func (h *handler) CreateConfigHandler(ctx context.Context, res http.ResponseWriter, req *http.Request) {

	// Extract scopes from request
	scopesString := req.FormValue("scopes")

	if scopesString == "" {
		logger.Log.Error("Scopes were empty")
		// Ques : Is this error code right?
		// http.Error(res, "Can not find scopes", http.StatusBadRequest)
		//return
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
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authToken)

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
	_, err = h.client.CreateConfig(ctx, &v1.CreateConfigRequest{
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

func (h *handler) UpdateConfigHandler(ctx context.Context, res http.ResponseWriter, req *http.Request) {

	// Extract scopes from request
	scopesString := req.FormValue("scopes")

	if scopesString == "" {
		logger.Log.Error("Scopes were empty")
		// Ques : Is this error code right?
		// http.Error(res, "Can not find scopes", http.StatusBadRequest)
		//return
	}
	// // convert it into an array of scopes
	// scopes := strings.Split(scopesString, ",")

	vars := mux.Vars(req)
	configIDStr := vars["config_id"]

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
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authToken)

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
	_, err = h.client.UpdateConfig(ctx, &v1.UpdateConfigRequest{
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

func getAuthToken(req *http.Request) string {
	bearerToken := req.Header.Get("Authorization")
	authToken := strings.TrimPrefix(bearerToken, "Bearer")
	authToken = strings.TrimSpace(authToken)
	authToken = "bearer " + authToken

	return authToken
}

func getConfigValueObject(configFile *csv.Reader, columns []string) ([]*v1.ConfigValue, error) {
	var configObject []*v1.ConfigValue
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
		configObject = append(configObject, &v1.ConfigValue{
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

func getConfigData(multipartForm *multipart.Form, res http.ResponseWriter) ([]*v1.Data, error) {
	configData := []*v1.Data{}
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
			data := &v1.Data{
				Metadata: &v1.Metadata{
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
