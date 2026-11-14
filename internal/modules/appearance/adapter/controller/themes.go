package controller

import (
	"io"
	"net/http"

	"trisse-home/internal/modules/appearance/adapter/gateway"
	"trisse-home/internal/modules/appearance/adapter/presenter"
	"trisse-home/internal/modules/appearance/usecase"
	"trisse-home/internal/shared/httpapi"
)

// ThemeHandlers serves /api/appearance/themes (#111).
type ThemeHandlers struct{ uc usecase.Themes }

// NewThemeHandlers returns the gallery handlers.
func NewThemeHandlers(uc usecase.Themes) *ThemeHandlers { return &ThemeHandlers{uc: uc} }

// Gallery handles GET /api/appearance/themes.
func (h *ThemeHandlers) Gallery(w http.ResponseWriter, r *http.Request) {
	g, err := h.uc.Gallery(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Gallery(g))
}

// Apply handles POST /api/appearance/themes/{id}/apply.
func (h *ThemeHandlers) Apply(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Apply(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Appearance(res))
}

// Restore handles POST /api/appearance/themes/restore.
func (h *ThemeHandlers) Restore(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Restore(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Appearance(res))
}

// Import handles POST /api/appearance/themes with the theme file as body.
func (h *ThemeHandlers) Import(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, gateway.MaxThemeBytes))
	if err != nil {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "too_large", "The theme file must be at most 16 KB.")
		return
	}
	t, err := h.uc.Import(r.Context(), data)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Theme(t))
}

// Export handles GET /api/appearance/themes/{id}: the shareable JSON file.
func (h *ThemeHandlers) Export(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	data, err := h.uc.Export(r.Context(), id)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// The id is validated (a-z, 0-9, -), so it is safe in the header.
	w.Header().Set("Content-Disposition", `attachment; filename="trisse-theme-`+id+`.json"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data) //nolint:gosec // JSON built by EncodeTheme from validated tokens, sent as an attachment with nosniff

}

// Delete handles DELETE /api/appearance/themes/{id} (imported themes only).
func (h *ThemeHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
