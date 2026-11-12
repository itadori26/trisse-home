package kernel

import "context"

// MetricType is the kind of a metric sample (#109).
type MetricType string

// Metric types.
const (
	MetricGauge   MetricType = "gauge"
	MetricCounter MetricType = "counter"
)

// Label is a metric label. Modules must only use labels with a small,
// bounded set of values (e.g. a service name), never IDs of requests.
type Label struct{ Name, Value string }

// MetricSample is a format-neutral measurement; the platform renders it
// (Prometheus text format). Samples with the same Name form one family
// and must share Help and Type.
type MetricSample struct {
	Name   string // with the trisse_ prefix, snake_case
	Help   string
	Type   MetricType
	Labels []Label
	Value  float64
}

// MetricSource is implemented by modules that contribute metrics.
type MetricSource interface {
	Metrics(ctx context.Context) []MetricSample
}

// Gauge builds a gauge sample.
func Gauge(name, help string, value float64, labels ...Label) MetricSample {
	return MetricSample{Name: name, Help: help, Type: MetricGauge, Labels: labels, Value: value}
}

// Bool converts a flag to 0 or 1.
func Bool(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
