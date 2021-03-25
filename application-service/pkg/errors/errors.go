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
	"google.golang.org/grpc/status"
)

func init() {
	//Business Error
	errMap["ClaimsNotFoundError"] = &customError{4001, "ClaimsNotFoundError", "cannot find claims in context", http.StatusBadRequest}
	errMap["DgraphQueryExecutionError"] = &customError{4002, "DgraphQueryExecutionError", "cannot complete query transaction", http.StatusBadRequest}
	errMap["DgraphDataUnmarshalError"] = &customError{4003, "DgraphDataUnmarshalError", "cannot unmarshal Json object", http.StatusBadRequest}
	errMap["AcqRightsCountIsZeroError"] = &customError{4004, "AcqRightsCountIsZeroError", "length of total count cannot be zero", http.StatusBadRequest}

	//Validation Error
	errMap["DataValidationError"] = &customError{4100, "DataValidationError", "Data Validation Failed", http.StatusBadRequest}
	errMap["ScopeValidationError"] = &customError{4100, "ScopeValidationError", "Scope Validation Failed", http.StatusBadRequest}

	//Database Error
	errMap["DBError"] = &customError{4200, "DB", "Database Operation Failed", http.StatusInternalServerError}

	//Authentication Error
	errMap["NoTokenError"] = &customError{1001, "NoTokenError", "No Token in Authorization Header", http.StatusUnauthorized}
	errMap["ParseTokenError"] = &customError{1002, "ParseTokenError", "Token cannot be parsed", http.StatusUnauthorized}
	errMap["InvalidTokenError"] = &customError{1003, "InvalidTokenError", "Token Signature Verification Failed", http.StatusUnauthorized}
	errMap["InvalidClaimsError"] = &customError{1004, "InvalidClaimsError", "Claims Attached are invalid", http.StatusUnauthorized}
	errMap["InvalidAPIKeyError"] = &customError{1005, "InvalidAPIKeyError", "No Key in X-API-KEY Header", http.StatusUnauthorized}
	errMap["NoAuthNError"] = &customError{1006, "NoAuthNError", "No Authentication Scheme in Request", http.StatusUnauthorized}
	errMap["NoErrorMapped"] = &customError{1007, "NoErrorMapped", "Error is not mapped in current scenario", http.StatusInternalServerError}
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
	// err, ok := errMap[value]
	// if !ok {
	// 	return errMap["NoErrorMapped"]
	// }
	return errMap[value]
}

// CustomHTTPError is the GRPC Gateway custom handler
func CustomHTTPError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	if err != nil {
		const fallback = `{"code": 13, "message": "failed to marshal error message"}`
		cError := getError(grpc.ErrorDesc(err))
		if cError == nil {
			s := status.Convert(err)
			pb := s.Proto()
			contentType := marshaler.ContentType()
			buf, _ := marshaler.Marshal(pb)
			w.Header().Set("Content-Type", contentType)
			st := runtime.HTTPStatusFromCode(s.Code())
			w.WriteHeader(st)
			w.Write(buf)
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
				w.Write([]byte(fallback))
			}
		}
	}
}
