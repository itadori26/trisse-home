package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/diskhealth/domain"
	"trisse-home/internal/modules/diskhealth/usecase"
)

type fakeUC struct {
	usecase.DiskHealth
	disks []domain.Disk
	err   error
}

func (f fakeUC) Disks(context.Context) ([]domain.Disk, error) { return f.disks, f.err }

func TestListWithoutExport(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeUC{err: domain.ErrNoExport}).List(rec, httptest.NewRequest(http.MethodGet, "/api/diskhealth", nil))
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "smart-export.sh") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestList(t *testing.T) {
	temp := 34
	d := domain.Evaluate(domain.Report{Device: "/dev/sda", Protocol: domain.ProtocolATA, At: time.Now(), Temperature: &temp,
		Attributes: []domain.Attribute{{ID: 197, Raw: 4}}}, time.Now(), domain.DefaultStaleAfter)
	rec := httptest.NewRecorder()
	NewHandlers(fakeUC{disks: []domain.Disk{d}}).List(rec, httptest.NewRequest(http.MethodGet, "/api/diskhealth", nil))
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, `"status":"warning"`) || !strings.Contains(body, `"pending":4`) || !strings.Contains(body, `"reallocated":null`) {
		t.Fatalf("%d %s", rec.Code, body)
	}
}
