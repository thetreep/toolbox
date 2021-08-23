package tracing

import (
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/cockroachdb/errors"
	"go.opencensus.io/trace"
)

func GCloudTracing(projectID string) error {
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: projectID,
	})
	if err != nil {
		return errors.Wrap(err, "gcp tracing exporter")
	}

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	return nil
}
