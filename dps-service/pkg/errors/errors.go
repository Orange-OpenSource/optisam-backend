package errors

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/status"
)

func init() {
	errMap["MissingFileName"] = &customError{2000, "MissingFile", "File name not found", http.StatusBadRequest}
	errMap["MissingKVForMap"] = &customError{2001, "MissingKVForMap", "headers for key-value in csv not mentioned", http.StatusBadRequest}
	errMap["InvalidCsvFile"] = &customError{2002, "InvalidCsvFile", "Csv file is invalid/corrupt", http.StatusBadRequest}
	errMap["HeadersMissing"] = &customError{2003, "HeadersMissing", "defined headers are not found in csv, wrong file ", http.StatusBadRequest}
	errMap["InvalidKeyValueHeaders"] = &customError{2004, "InvalidKeyValueHeaders", " header of key and value  can't be same", http.StatusBadRequest}
	errMap["InvalidDirPath"] = &customError{2005, "InvalidDirPath", "directory path is blank", http.StatusBadRequest}
	errMap["UnknownFileType"] = &customError{2006, "UnknownFileType", "Unknown file received, cann't processed ", http.StatusBadRequest}
	errMap["InvalidFileName"] = &customError{2007, "InvalidFileName", "File name is not as exxpected, required scope_filename.csv", http.StatusBadRequest}
	errMap["FileNotSupported"] = &customError{2008, "FileNotSupported", "This file is not supported", http.StatusBadRequest}
	errMap["TargetServiceNotSupported"] = &customError{2009, "TargetServiceNotSupported", "This target service is not supported", http.StatusServiceUnavailable}
	errMap["RetriesExceeded"] = &customError{2010, "RetriesExceeded", "retries cross max limit", http.StatusInternalServerError}
	errMap["ServiceUnavaliable"] = &customError{2011, "ServiceUnavaliable", "target service is refusing or down", http.StatusServiceUnavailable}
	errMap["ParsingError"] = &customError{2012, "ParsingError", "data parsing failed ", http.StatusBadRequest}
	errMap["InternalError"] = &customError{2013, "Internal Error", "database system is down", http.StatusInternalServerError}
	errMap["FileValidationFailureWithHeaders"] = &customError{2014, "FileValidationFailureWithHeaders", "Headers are missing in file", http.StatusBadRequest}
	errMap["FileValidationFailureWithScope"] = &customError{2015, "FileValidationFailureWithScope", "file validation is failed due to scope missing in name", http.StatusBadRequest}
	errMap["BadFile"] = &customError{2015, "BadFile", "Could not read the content of file/corrupt file/bad segment", http.StatusBadRequest}
	errMap["NoDataInFile"] = &customError{2016, "NoDataInFile", "no data in file", http.StatusBadRequest}
	errMap["MissingDataField "] = &customError{2017, "MissingDataField ", "field is missing in row", http.StatusBadRequest}
	errMap["DBError"] = &customError{2018, "DBError", "internal service DB error", http.StatusInternalServerError}
	errMap["NoContent"] = &customError{2019, "NoContent", "There is no data in scope", http.StatusNoContent}
	errMap["ScopeValidationError"] = &customError{2020, "ScopeValidationError", "Scope Validation Failed", http.StatusBadRequest}
	errMap["ClaimsNotFoundError"] = &customError{2021, "ClaimsNotFoundError", "cannot find claims in context", http.StatusBadRequest}
	errMap["NoErrorMapped"] = &customError{2022, "NoErrorMapped", "Error is not mapped in current scenario", http.StatusInternalServerError}

}

var (
	errMap = make(map[string]*customError)
)

type customError struct {
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

func (e *customError) Error() string { return e.Value }

func getError(value string) *customError {
	return errMap[value]
}

// CustomHTTPError is the GRPC Gateway custom handler
func CustomHTTPError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	if err != nil {
		const fallback = `{"code": 13, "message": "failed to marshal error message"}`
		cError := getError(status.Convert(err).Message())
		if cError == nil {
			s := status.Convert(err)
			pb := s.Proto()
			contentType := marshaler.ContentType()
			buf, _ := marshaler.Marshal(pb)
			w.Header().Set("Content-Type", contentType)
			st := runtime.HTTPStatusFromCode(s.Code())
			w.WriteHeader(st)
			w.Write(buf) // nolint: errcheck
		} else {
			w.Header().Set("Content-type", marshaler.ContentType())
			w.WriteHeader(cError.HTTPStatusCode)
			jErr := json.NewEncoder(w).Encode(customError{
				cError.ErrorCode,
				cError.Value,
				cError.Message,
				cError.HTTPStatusCode,
			})
			if jErr != nil {
				w.Write([]byte(fallback)) // nolint: errcheck
			}
		}
	}
}
