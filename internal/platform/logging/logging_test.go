package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	New(&buf, slog.LevelInfo, "json").Info("hello", "k", "v")
	if !strings.Contains(buf.String(), `"msg":"hello"`) {
		t.Fatalf("expected JSON, got %q", buf.String())
	}
	buf.Reset()
	l := New(&buf, slog.LevelWarn, "text")
	l.Info("hidden")
	l.Warn("shown")
	if strings.Contains(buf.String(), "hidden") || !strings.Contains(buf.String(), "msg=shown") {
		t.Fatalf("unexpected text output %q", buf.String())
	}
}
