// Package metrics renders metrics in the Prometheus text exposition format
// (version 0.0.4) without a client library (ADR 0010), and provides Go
// runtime and HTTP server metrics (#109).
package metrics

import (
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// ContentType of the exposition format.
const ContentType = "text/plain; version=0.0.4; charset=utf-8"

// Write renders samples grouped by family name (sorted for stable output);
// HELP and TYPE are taken from the first sample of a family.
func Write(w io.Writer, samples []kernel.MetricSample) error {
	sorted := make([]kernel.MetricSample, len(samples))
	copy(sorted, samples)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	var bw strings.Builder
	last := ""
	for _, s := range sorted {
		if !validName(s.Name) {
			continue
		}
		if s.Name != last {
			last = s.Name
			typ := s.Type
			if typ == "" {
				typ = kernel.MetricGauge
			}
			bw.WriteString("# HELP " + s.Name + " " + escapeHelp(s.Help) + "\n")
			bw.WriteString("# TYPE " + s.Name + " " + string(typ) + "\n")
		}
		bw.WriteString(s.Name)
		if len(s.Labels) > 0 {
			bw.WriteByte('{')
			for i, l := range s.Labels {
				if i > 0 {
					bw.WriteByte(',')
				}
				bw.WriteString(l.Name + `="` + escapeLabel(l.Value) + `"`)
			}
			bw.WriteByte('}')
		}
		bw.WriteByte(' ')
		bw.WriteString(formatFloat(s.Value))
		bw.WriteByte('\n')
	}
	_, err := io.WriteString(w, bw.String())
	return err
}

func validName(n string) bool {
	if n == "" {
		return false
	}
	for i, r := range n {
		ok := r == '_' || r == ':' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (i > 0 && r >= '0' && r <= '9')
		if !ok {
			return false
		}
	}
	return true
}

var (
	helpEscaper  = strings.NewReplacer(`\`, `\\`, "\n", `\n`)
	labelEscaper = strings.NewReplacer(`\`, `\\`, "\n", `\n`, `"`, `\"`)
)

func escapeHelp(s string) string  { return helpEscaper.Replace(s) }
func escapeLabel(s string) string { return labelEscaper.Replace(s) }

func formatFloat(v float64) string {
	switch {
	case math.IsInf(v, 1):
		return "+Inf"
	case math.IsInf(v, -1):
		return "-Inf"
	case math.IsNaN(v):
		return "NaN"
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}
