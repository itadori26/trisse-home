package controller

import (
	"net/http"
	"time"

	"trisse-home/internal/modules/todos/adapter/presenter"
	"trisse-home/internal/modules/todos/usecase"
	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
)

// UseCases bundles the to-do input ports.
type UseCases struct {
	List           usecase.ListTodos
	Create         usecase.CreateTodo
	Update         usecase.UpdateTodo
	Delete         usecase.DeleteTodo
	ClearCompleted usecase.ClearCompleted
}

// Handlers serves the to-do endpoints.
type Handlers struct{ uc UseCases }

// NewHandlers returns to-do handlers.
func NewHandlers(uc UseCases) *Handlers { return &Handlers{uc: uc} }

func (d TodoInputDTO) toInput() (usecase.TodoInput, error) {
	in := usecase.TodoInput{Title: d.Title}
	if d.Due != "" {
		due, err := time.Parse(time.DateOnly, d.Due)
		if err != nil {
			return in, kernel.NewValidationError("due", "The due date must look like 2026-12-24.")
		}
		in.Due = due
	}
	return in, nil
}

// List handles GET /api/todos?done=true.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.List.Execute(r.Context(), r.URL.Query().Get("done") == "true")
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Todos(res))
}

// Create handles POST /api/todos.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto TodoInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	in, err := dto.toInput()
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Create.Execute(r.Context(), in)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Todo(res))
}

// Update handles PUT /api/todos/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto TodoInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	in, err := dto.toInput()
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Update.Execute(r.Context(), r.PathValue("id"), in, dto.Done)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Todo(res))
}

// Delete handles DELETE /api/todos/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete.Execute(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ClearCompleted handles POST /api/todos/clear-completed.
func (h *Handlers) ClearCompleted(w http.ResponseWriter, r *http.Request) {
	n, err := h.uc.ClearCompleted.Execute(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, ClearedDTO{Removed: n})
}
