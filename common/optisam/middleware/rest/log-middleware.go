// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package rest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	custom_logger "optisam-backend/common/optisam/logger"
	"strings"
	"time"

	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

type LoggerKey struct{}
type LoggerUserDetails struct {
	UserID string
	Role   string
}

func (l *LoggerUserDetails) String() string {
	return fmt.Sprintf("%v %v )", l.UserID, l.Role)
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) Write(p []byte) (int, error) {
	log.Printf("resp %v", string(p))
	return lrw.ResponseWriter.Write(p)
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// AddLogger logs request/response pair
func AddLogger(logger *zap.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uDetails LoggerUserDetails
		ctx := context.WithValue(r.Context(), LoggerKey{}, &uDetails)

		// We do not want to be spammed by Kubernetes health check.
		// Do not log Kubernetes health check.
		// You can change this behavior as you wish.
		if r.Header.Get("X-Liveness-Probe") == "Healthz" {
			h.ServeHTTP(w, r)
			return
		}
		buf, bodyErr := ioutil.ReadAll(r.Body)
		if bodyErr != nil {
			log.Print("bodyErr ", bodyErr.Error())
			http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
			return
		}

		rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
		r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
		// id := GetReqID(ctx)

		// Prepare fields to log
		var scheme string
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
		proto := r.Proto
		method := r.Method
		remoteAddr := r.RemoteAddr
		userAgent := r.UserAgent()
		uri := strings.Join([]string{scheme, "://", r.Host, r.RequestURI}, "")

		t1 := time.Now()
		lrw := newLoggingResponseWriter(w)

		h.ServeHTTP(lrw, r.WithContext(ctx))
		resp := &bytes.Buffer{}
		if _, err := io.Copy(w, resp); err != nil {
			log.Printf("Failed to send out response: %v", err)
		}
		sc := trace.FromContext(ctx).SpanContext()
		userDetails, ok := ctx.Value(LoggerKey{}).(*LoggerUserDetails)
		if !ok {
			//TODO: Do we need to handle this
		}
		if userDetails == nil {
			userDetails = &LoggerUserDetails{UserID: "", Role: ""}
		}

		if custom_logger.GlobalLevel == zap.DebugLevel {
			reqBody, _ := ioutil.ReadAll(rdr1)
			logger.Info("request completed",
				// zap.String("request-id", id),
				zap.String("trace-id", sc.TraceID.String()),
				zap.String("span-id", sc.SpanID.String()),
				zap.Bool("is-sampled", sc.IsSampled()),
				zap.String("user-id", userDetails.UserID),
				zap.String("user-role", userDetails.Role),
				zap.String("http-scheme", scheme),
				zap.String("http-proto", proto),
				zap.String("http-method", method),
				zap.Any("http-req-body", string(reqBody)),
				zap.Any("http-res-body", resp),
				zap.Any("http-status", lrw.statusCode),
				zap.String("remote-addr", remoteAddr),
				zap.String("user-agent", userAgent),
				zap.String("uri", uri),
				zap.Float64("elapsed-ms", float64(time.Since(t1).Nanoseconds())/1000000.0),
			)
		} else {
			logger.Info("request completed",
				// zap.String("request-id", id),
				zap.String("trace-id", sc.TraceID.String()),
				zap.String("span-id", sc.SpanID.String()),
				zap.Bool("is-sampled", sc.IsSampled()),
				zap.String("user-id", userDetails.UserID),
				zap.String("user-role", userDetails.Role),
				zap.String("http-scheme", scheme),
				zap.String("http-proto", proto),
				zap.String("http-method", method),
				zap.Any("http-status", lrw.statusCode),
				zap.String("remote-addr", remoteAddr),
				zap.String("user-agent", userAgent),
				zap.String("uri", uri),
				zap.Float64("elapsed-ms", float64(time.Since(t1).Nanoseconds())/1000000.0),
			)
		}
	})
}
