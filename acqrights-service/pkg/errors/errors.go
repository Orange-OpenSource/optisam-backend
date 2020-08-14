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
	"optisam-backend/common/optisam/logger"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

func init() {
	//Business Error
	ErrMap["ClaimsNotFoundError"] = CustomError{4001, "ClaimsNotFoundError", "cannot find claims in context", http.StatusBadRequest}
	ErrMap["DgraphQueryExecutionError"] = CustomError{4002, "DgraphQueryExecutionError", "cannot complete query transaction", http.StatusBadRequest}
	ErrMap["DgraphDataUnmarshalError"] = CustomError{4003, "DgraphDataUnmarshalError", "cannot unmarshal Json object", http.StatusBadRequest}
	ErrMap["AcqRightsCountIsZeroError"] = CustomError{4004, "AcqRightsCountIsZeroError", "length of total count cannot be zero", http.StatusBadRequest}

	//Validation Error
	ErrMap["DataValidationError"] = CustomError{4100, "DataValidationError", "Data Validation Failed", http.StatusBadRequest}
	ErrMap["ScopeValidationError"] = CustomError{4100, "ScopeValidationError", "Scope Validation Failed", http.StatusBadRequest}

	//Database Error
	ErrMap["DBError"] = CustomError{4200, "DB", "Database Operation Failed", http.StatusInternalServerError}

	//Authentication Error
	ErrMap["NoTokenError"] = CustomError{1001, "NoTokenError", "No Token in Authorization Header", http.StatusUnauthorized}
	ErrMap["ParseTokenError"] = CustomError{1002, "ParseTokenError", "Token cannot be parsed", http.StatusUnauthorized}
	ErrMap["InvalidTokenError"] = CustomError{1003, "InvalidTokenError", "Token Signature Verification Failed", http.StatusUnauthorized}
	ErrMap["InvalidClaimsError"] = CustomError{1004, "InvalidClaimsError", "Claims Attached are invalid", http.StatusUnauthorized}
	ErrMap["InvalidAPIKeyError"] = CustomError{1005, "InvalidAPIKeyError", "No Key in X-API-KEY Header", http.StatusUnauthorized}
	ErrMap["NoAuthNError"] = CustomError{1006, "NoAuthNError", "No Authentication Scheme in Request", http.StatusUnauthorized}
}

var (
	ErrMap = make(map[string]CustomError)
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

func getError(value string) CustomError {
	return ErrMap[value]
}

func CustomHTTPError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	if err != nil {
		const fallback = `{"error": "failed to marshal error message"}`
		customError := getError(status.Convert(err).Message())
		w.Header().Set("Content-type", marshaler.ContentType())
		w.WriteHeader(customError.HTTPStatusCode)
		jErr := json.NewEncoder(w).Encode(CustomError{
			customError.ErrorCode,
			customError.Value,
			customError.Message,
			customError.HTTPStatusCode,
		})
		if jErr != nil {
			_, err := w.Write([]byte(fallback))
			logger.Log.Error("CustomHTTPError", zap.Error(err))
		}
	}
}
