// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package errors

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

func init() {
	ErrMap["MissingFileName"] = &CustomError{2000, "MissingFile", "File name not found", http.StatusBadRequest}
	ErrMap["MissingKVForMap"] = &CustomError{2001, "MissingKVForMap", "headers for key-value in csv not mentioned", http.StatusBadRequest}
	ErrMap["InvalidCsvFile"] = &CustomError{2002, "InvalidCsvFile", "Csv file is invalid/corrupt", http.StatusBadRequest}
	ErrMap["HeadersMissing"] = &CustomError{2003, "HeadersMissing", "defined headers are not found in csv, wrong file ", http.StatusBadRequest}
	ErrMap["InvalidKeyValueHeaders"] = &CustomError{2004, "InvalidKeyValueHeaders", " header of key and value  can't be same", http.StatusBadRequest}
	ErrMap["InvalidDirPath"] = &CustomError{2005, "InvalidDirPath", "directory path is blank", http.StatusBadRequest}
	ErrMap["UnknownFileType"] = &CustomError{2006, "UnknownFileType", "Unknown file received, cann't processed ", http.StatusBadRequest}
	ErrMap["InvalidFileName"] = &CustomError{2007, "InvalidFileName", "File name is not as exxpected, required scope_filename.csv", http.StatusBadRequest}
	ErrMap["FileNotSupported"] = &CustomError{2008, "FileNotSupported", "This file is not supported", http.StatusBadRequest}
	ErrMap["TargetServiceNotSupported"] = &CustomError{2009, "TargetServiceNotSupported", "This target service is not supported", http.StatusBadRequest}
}

var (
	ErrMap = make(map[string]*CustomError)
)

type CustomError struct {
	// Code is the error code that this descriptor describes.
	ErrorCode int `json:"errorcode"`

	// Value provides a unique, string key, often captilized with
	// underscores, to identify the error code. This value is used as the
	// keyed value when serializing api errors.
	Value string `json:"value"`

	// Message is a short, human readable decription of the error condition
	// included in API responses.
	Message string `json:"message"`

	// HTTPStatusCode provides the http status code that is associated with
	// this error condition.
	HTTPStatusCode int `json:"-"`
}

func (e *CustomError) Error() string { return e.Value }

func GetError(value string) *CustomError {
	return ErrMap[value]
}

func CustomHTTPError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	if err != nil {
		const fallback = `{"error": "failed to marshal error message"}`
		customError := *GetError(grpc.ErrorDesc(err))
		w.Header().Set("Content-type", marshaler.ContentType())
		w.WriteHeader(customError.HTTPStatusCode)
		jErr := json.NewEncoder(w).Encode(CustomError{
			customError.ErrorCode,
			customError.Value,
			customError.Message,
			customError.HTTPStatusCode,
		})
		if jErr != nil {
			w.Write([]byte(fallback))
		}
	}
}
