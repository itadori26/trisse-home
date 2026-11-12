package metrics

import (
	"runtime"
	"runtime/metrics"
	"time"

	"trisse-home/internal/shared/kernel"
)

var started = time.Now()

// runtimeMetrics maps runtime/metrics names to exported families.
var runtimeMetrics = []struct {
	key, name, help string
	typ             kernel.MetricType
}{
	{"/sched/goroutines:goroutines", "go_goroutines", "Number of goroutines that currently exist.", kernel.MetricGauge},
	{"/memory/classes/heap/objects:bytes", "go_memstats_heap_alloc_bytes", "Bytes of allocated heap objects.", kernel.MetricGauge},
	{"/memory/classes/total:bytes", "go_memstats_sys_bytes", "Bytes of memory obtained from the OS.", kernel.MetricGauge},
	{"/gc/cycles/total:gc-cycles", "go_gc_cycles_total", "Number of completed GC cycles.", kernel.MetricCounter},
}

// Runtime returns Go runtime and process metrics.
func Runtime(version string) []kernel.MetricSample {
	samples := make([]metrics.Sample, len(runtimeMetrics))
	for i, m := range runtimeMetrics {
		samples[i].Name = m.key
	}
	metrics.Read(samples)
	out := make([]kernel.MetricSample, 0, len(samples)+3)
	for i, s := range samples {
		var v float64
		switch s.Value.Kind() {
		case metrics.KindUint64:
			v = float64(s.Value.Uint64())
		case metrics.KindFloat64:
			v = s.Value.Float64()
		default:
			continue // not supported by this Go version
		}
		m := runtimeMetrics[i]
		out = append(out, kernel.MetricSample{Name: m.name, Help: m.help, Type: m.typ, Value: v})
	}
	out = append(out,
		kernel.Gauge("process_start_time_seconds", "Start time of the process since the Unix epoch in seconds.", float64(started.Unix())),
		kernel.Gauge("trisse_build_info", "Version of the running build (value is always 1).", 1,
			kernel.Label{Name: "version", Value: version}, kernel.Label{Name: "goversion", Value: runtime.Version()}),
	)
	return out
}
