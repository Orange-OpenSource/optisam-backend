package prometheus

import (
	"github.com/pkg/errors"
	"contrib.go.opencensus.io/exporter/prometheus"
)

// NewExporter creates a new, configured Prometheus exporter.
func NewExporter(config Config) (*prometheus.Exporter, error) {
	exporter, err := prometheus.NewExporter(prometheus.Options{
		Namespace: config.Namespace,
	})

	return exporter, errors.Wrap(err, "failed to create prometheus exporter")
}
