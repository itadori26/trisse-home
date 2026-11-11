// Command budgets measures the footprint of a build and compares it with
// the budgets in budgets.json (ADR 0011, lean by default). It writes a
// Markdown table (to stdout and, in CI, the job summary), stores the
// measurements as JSON for the next comparison and exits with status 1 when
// a budget is exceeded.
//
//	go run ./scripts/budgets -dist internal/platform/webui/dist -binary bin/trisse-home \
//	    -image-bytes 26000000 -memory "18.5MiB / 7.6GiB" -baseline main.json -out measurements.json
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Metric names, also the keys in budgets.json and the measurement files.
const (
	InitialJS  = "initialJsGzipKB"
	InitialCSS = "initialCssGzipKB"
	LazyChunk  = "largestLazyChunkGzipKB"
	Binary     = "binaryMB"
	Image      = "imageMB"
	Memory     = "idleMemoryMB"
)

var order = []string{InitialJS, InitialCSS, LazyChunk, Binary, Image, Memory}

var labels = map[string]string{
	InitialJS:  "Initial JS (gzip)",
	InitialCSS: "Initial CSS (gzip)",
	LazyChunk:  "Largest lazy JS chunk (gzip)",
	Binary:     "Go binary",
	Image:      "Container image",
	Memory:     "Idle memory (RSS after 60 s)",
}

func unit(name string) string {
	if strings.HasSuffix(name, "KB") {
		return "KB"
	}
	return "MB"
}

// Measurements maps metric names to values (KB or MB, see the name).
type Measurements map[string]float64

func gzipSize(data []byte) int {
	var buf bytes.Buffer
	w, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression) // valid level: cannot fail
	_, _ = w.Write(data)
	_ = w.Close()
	return buf.Len()
}

var assetRef = regexp.MustCompile(`(?:src|href)="/?(assets/[^"]+)"`)

// measureDist reads index.html: every JS/CSS asset it references is loaded
// for the first paint; every other JS file is a lazily loaded chunk.
func measureDist(dir string, m Measurements) error {
	index, err := os.ReadFile(filepath.Join(dir, "index.html")) //nolint:gosec // path from the command line
	if err != nil {
		return err
	}
	initial := map[string]bool{}
	for _, match := range assetRef.FindAllSubmatch(index, -1) {
		initial[string(match[1])] = true
	}
	files, err := filepath.Glob(filepath.Join(dir, "assets", "*"))
	if err != nil {
		return err
	}
	var js, css, lazy int
	for _, f := range files {
		data, err := os.ReadFile(f) //nolint:gosec // files of the build output
		if err != nil {
			return err
		}
		rel := "assets/" + filepath.Base(f)
		size := gzipSize(data)
		switch {
		case initial[rel] && strings.HasSuffix(f, ".js"):
			js += size
		case initial[rel] && strings.HasSuffix(f, ".css"):
			css += size
		case strings.HasSuffix(f, ".js"):
			lazy = max(lazy, size)
		}
	}
	m[InitialJS], m[InitialCSS], m[LazyChunk] = kb(js), kb(css), kb(lazy)
	return nil
}

func kb(b int) float64     { return round(float64(b) / 1000) }
func mb(b float64) float64 { return round(b / 1e6) }
func round(f float64) float64 {
	v, _ := strconv.ParseFloat(strconv.FormatFloat(f, 'f', 1, 64), 64)
	return v
}

var memPattern = regexp.MustCompile(`^\s*([\d.]+)\s*([KMG]i?B|B)`)

// parseMemory reads the first value of `docker stats --format {{.MemUsage}}`.
func parseMemory(s string) (float64, error) {
	match := memPattern.FindStringSubmatch(s)
	if match == nil {
		return 0, fmt.Errorf("cannot parse memory %q", s)
	}
	v, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, err
	}
	factor := map[string]float64{"B": 1, "KB": 1e3, "MB": 1e6, "GB": 1e9, "KiB": 1 << 10, "MiB": 1 << 20, "GiB": 1 << 30}[match[2]]
	return mb(v * factor), nil
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path) //nolint:gosec // path from the command line
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// report renders the table and returns the names of exceeded budgets.
func report(w io.Writer, got, budgets, baseline Measurements) []string {
	var exceeded []string
	fmt.Fprintln(w, "### Resource budgets (ADR 0011)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "| Metric | Value | Budget | Δ main | |")
	fmt.Fprintln(w, "| --- | ---: | ---: | ---: | --- |")
	for _, name := range order {
		v, ok := got[name]
		if !ok {
			continue
		}
		budget, hasBudget := budgets[name]
		status, budgetText := "✅", "–"
		if hasBudget {
			budgetText = fmt.Sprintf("%.1f %s", budget, unit(name))
			if v > budget {
				status = "❌"
				exceeded = append(exceeded, name)
			}
		}
		delta := "–"
		if b, ok := baseline[name]; ok {
			delta = fmt.Sprintf("%+.1f %s", v-b, unit(name))
		}
		fmt.Fprintf(w, "| %s | %.1f %s | %s | %s | %s |\n", labels[name], v, unit(name), budgetText, delta, status)
	}
	sort.Strings(exceeded)
	return exceeded
}

func run() error {
	budgetsPath := flag.String("budgets", "budgets.json", "budget file")
	dist := flag.String("dist", "", "frontend build output (index.html + assets)")
	binary := flag.String("binary", "", "Go binary")
	imageBytes := flag.Float64("image-bytes", 0, "container image size in bytes (docker image inspect)")
	memory := flag.String("memory", "", "docker stats MemUsage, e.g. \"18.5MiB / 7.6GiB\"")
	baselinePath := flag.String("baseline", "", "measurements of main for the delta column (optional)")
	out := flag.String("out", "", "write the measurements as JSON")
	flag.Parse()

	var budgets struct {
		Budgets Measurements `json:"budgets"`
	}
	if err := readJSON(*budgetsPath, &budgets); err != nil {
		return fmt.Errorf("read budgets: %w", err)
	}
	got := Measurements{}
	if *dist != "" {
		if err := measureDist(*dist, got); err != nil {
			return fmt.Errorf("measure frontend: %w", err)
		}
	}
	if *binary != "" {
		st, err := os.Stat(*binary)
		if err != nil {
			return err
		}
		got[Binary] = mb(float64(st.Size()))
	}
	if *imageBytes > 0 {
		got[Image] = mb(*imageBytes)
	}
	if *memory != "" {
		v, err := parseMemory(*memory)
		if err != nil {
			return err
		}
		got[Memory] = v
	}
	baseline := Measurements{}
	if *baselinePath != "" {
		if err := readJSON(*baselinePath, &baseline); err != nil {
			fmt.Fprintln(os.Stderr, "budgets: no baseline:", err)
		}
	}
	var table bytes.Buffer
	exceeded := report(&table, got, budgets.Budgets, baseline)
	_, _ = os.Stdout.Write(table.Bytes())
	if summary := os.Getenv("GITHUB_STEP_SUMMARY"); summary != "" {
		f, err := os.OpenFile(summary, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600) //nolint:gosec // path set by GitHub Actions
		if err == nil {
			_, _ = f.Write(table.Bytes())
			_ = f.Close()
		}
	}
	if *out != "" {
		data, _ := json.MarshalIndent(got, "", "  ") // map of floats: cannot fail
		if err := os.WriteFile(*out, data, 0o600); err != nil {
			return err
		}
	}
	if len(exceeded) > 0 {
		return fmt.Errorf("budget exceeded: %s (see budgets.json and docs/adr/0011-lean-by-default.md)", strings.Join(exceeded, ", "))
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "budgets:", err)
		os.Exit(1)
	}
}
