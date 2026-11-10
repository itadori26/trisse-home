package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/system/usecase"
)

type fakeMonitor struct{ usecase.Monitor }

func (fakeMonitor) Current(context.Context) (usecase.SnapshotView, error) {
	return usecase.SnapshotView{CPUPercent: 12.345, Memory: usecase.UsageView{Total: 100, Used: 25, Percent: 25}}, nil
}

func TestSnapshotRoundsPercentages(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeMonitor{}).Snapshot(rec, httptest.NewRequest(http.MethodGet, "/api/system", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"cpuPercent":12.3`) || !strings.Contains(rec.Body.String(), `"disks":[]`) {
		t.Fatalf("unexpected %d %s", rec.Code, rec.Body)
	}
}
