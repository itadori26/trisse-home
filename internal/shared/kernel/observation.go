package kernel

import "context"

// Observation is the latest state of something a module monitors, e.g. a
// service, a container, a disk or a backup. Modules report observations
// without knowing who consumes them (the notifications module, or nobody).
type Observation struct {
	// Kind groups observations for notification rules: service, container,
	// disk, tls or backup.
	Kind string
	// Key identifies the observed object uniquely within its kind.
	Key string
	// Label is a human-readable name.
	Label string
	// Healthy is the producer's verdict. For threshold kinds (disk, tls) the
	// consumer may judge Value against its own threshold instead.
	Healthy bool
	// Value carries a measurement such as a disk usage in percent or the days
	// until a certificate expires.
	Value float64
	// Detail is an optional explanation, e.g. an error message.
	Detail string
}

// ObservationSink receives observations (output port shared by modules).
type ObservationSink interface {
	Observe(ctx context.Context, o Observation)
}

// DiscardObservations is an ObservationSink that ignores everything.
type DiscardObservations struct{}

// Observe implements ObservationSink.
func (DiscardObservations) Observe(context.Context, Observation) {}
