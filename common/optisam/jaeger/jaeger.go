package jaeger

import (
	"contrib.go.opencensus.io/exporter/jaeger"
	"github.com/pkg/errors"
)

// NewExporter creates a new, configured Jaeger exporter.
func NewExporter(config Config) (*jaeger.Exporter, error) {
	exporter, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: config.CollectorEndpoint,
		AgentEndpoint:     config.AgentEndpoint,
		Username:          config.Username,
		Password:          config.Password,
		Process: jaeger.Process{
			ServiceName: config.ServiceName,
		},
	})

	return exporter, errors.Wrap(err, "failed to create jaeger exporter")
}
