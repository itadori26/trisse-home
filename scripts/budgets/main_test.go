package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMeasureDist(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "assets"), 0o750)
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("index.html", `<script type="module" src="/assets/index-a.js"></script><link rel="stylesheet" href="/assets/index-b.css">`)
	noise := func(n int) string { // incompressible content of n bytes
		b := make([]byte, n/2)
		_, _ = rand.Read(b)
		return hex.EncodeToString(b)
	}
	write("assets/index-a.js", noise(8000))
	write("assets/index-b.css", noise(2000))
	write("assets/lazy-c.js", noise(4000))
	m := Measurements{}
	if err := measureDist(dir, m); err != nil {
		t.Fatal(err)
	}
	if m[InitialJS] <= 0 || m[InitialCSS] <= 0 || m[LazyChunk] <= 0 || m[LazyChunk] >= m[InitialJS] {
		t.Fatalf("%v", m)
	}
}

func TestParseMemory(t *testing.T) {
	for in, want := range map[string]float64{"18.5MiB / 7.6GiB": 19.4, "900KiB / 1GiB": 0.9, "1.2GiB / 8GiB": 1288.5} {
		if got, err := parseMemory(in); err != nil || got != want {
			t.Errorf("%s: %v %v, want %v", in, got, err, want)
		}
	}
	if _, err := parseMemory("n/a"); err == nil {
		t.Fatal("garbage accepted")
	}
}

func TestReport(t *testing.T) {
	var buf bytes.Buffer
	exceeded := report(&buf, Measurements{InitialJS: 150, Binary: 20}, Measurements{InitialJS: 140, Binary: 25}, Measurements{InitialJS: 130})
	if len(exceeded) != 1 || exceeded[0] != InitialJS {
		t.Fatalf("exceeded: %v", exceeded)
	}
	if !strings.Contains(buf.String(), "+20.0 KB") || !strings.Contains(buf.String(), "❌") {
		t.Fatal(buf.String())
	}
}

func TestReportListsThePiImage(t *testing.T) {
	var buf bytes.Buffer
	exceeded := report(&buf, Measurements{PiImage: 560}, Measurements{PiImage: 613}, Measurements{})
	if len(exceeded) != 0 || !strings.Contains(buf.String(), "Raspberry Pi image (.img.xz) | 560.0 MB | 613.0 MB") {
		t.Fatalf("exceeded %v, table:\n%s", exceeded, buf.String())
	}
}
