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

type ImportServiceServer interface {
	UploadDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	UploadMetaDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
}
