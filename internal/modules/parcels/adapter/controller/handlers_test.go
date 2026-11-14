package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/modules/parcels/usecase"
)

type fakeUC struct{ usecase.Parcels }

func (fakeUC) Add(_ context.Context, tn string, c domain.Carrier, l string) (domain.Parcel, error) {
	return domain.NewParcel("1", tn, c, l, domain.Parcel{}.CreatedAt)
}

func TestAdd(t *testing.T) {
	for body, want := range map[string]int{
		`{"trackingNumber":"00340434123456","carrier":"dhl"}`:   http.StatusCreated,
		`{"trackingNumber":"00340434123456","carrier":"fedex"}`: http.StatusUnprocessableEntity,
	} {
		rec := httptest.NewRecorder()
		NewHandlers(fakeUC{}).Add(rec, httptest.NewRequest(http.MethodPost, "/api/parcels", strings.NewReader(body)))
		if rec.Code != want {
			t.Errorf("%s: %d", body, rec.Code)
		}
	}
}
