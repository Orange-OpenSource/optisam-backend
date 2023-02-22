package errors

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/status"
)

func init() {
	// Business Error
	errMap["CloseSoureError"] = &customError{4001, "CloseSoureError", "closesource must have value", http.StatusBadRequest}
	errMap["OpenSourceError"] = &customError{4002, "OpenSourceError", "opensouce must have value", http.StatusBadRequest}
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
