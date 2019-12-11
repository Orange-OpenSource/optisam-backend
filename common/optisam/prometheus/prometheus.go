// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package prometheus

import (
	"github.com/pkg/errors"
	"go.opencensus.io/exporter/prometheus"
)

// NewExporter creates a new, configured Prometheus exporter.
func NewExporter(config Config) (*prometheus.Exporter, error) {
	exporter, err := prometheus.NewExporter(prometheus.Options{
		Namespace: config.Namespace,
	})

	return exporter, errors.Wrap(err, "failed to create prometheus exporter")
}
