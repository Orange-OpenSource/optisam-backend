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
