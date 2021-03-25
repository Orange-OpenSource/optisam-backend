// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//for go mock
//mockgen.exe -source=../../../optisam-backend/dps-service/pkg/api/v1/dps.pb.go  -destination=./mock/dps_mock.go -package=mock
type ImportServiceServer interface {
	UploadDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	UploadMetaDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	CreateConfigHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	UpdateConfigHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	UploadGlobalDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
}
