// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package rest

import (
	"net/http"

	"github.com/rs/cors"
)

// AddCORS add Cross-origin resource sharing header (Access-Control-Allow-Origin) to http responses
func AddCORS(allowedOrigins []string, handler http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut, http.MethodPatch, http.MethodOptions},
		AllowedHeaders: []string{"*"},
		Debug:          false,
	})
	// Insert the middleware
	return c.Handler(handler)
}
