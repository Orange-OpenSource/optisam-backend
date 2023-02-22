package v1

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ImportServiceServer interface {
	UploadDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	UploadMetaDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	CreateConfigHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	UpdateConfigHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	UploadGlobalDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	DownloadFile(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	UploadFiles(res http.ResponseWriter, req *http.Request, param httprouter.Params)
	UploadCatalogData(res http.ResponseWriter, req *http.Request, param httprouter.Params)
}
