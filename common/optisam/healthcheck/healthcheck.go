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
