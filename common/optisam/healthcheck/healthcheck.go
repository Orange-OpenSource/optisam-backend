// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package healthcheck

import (
	"net/http"

	health "github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/handlers"
)

// New returns a new health checker instance.
func New() *health.Health {
	healthChecker := health.New()
	return healthChecker
}

// Handler returns a new HTTP handler for a health checker.
func Handler(healthChecker health.IHealth) http.Handler {
	return handlers.NewJSONHandlerFunc(healthChecker, nil)
}
