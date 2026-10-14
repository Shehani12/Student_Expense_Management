package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"expense-api/models"
	"expense-api/service"
)

// ExpenseHandler turns HTTP requests into service calls.
type ExpenseHandler struct {
	svc *service.ExpenseService
}

// NewExpenseHandler creates a handler using the given service.
func NewExpenseHandler(svc *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{svc: svc}
}

// RegisterRoutes connects each URL pattern to its handler function.
// Since Go 1.22 a pattern holds both the HTTP method and the path, and
// {id} marks a path variable that r.PathValue("id") reads.
func (h *ExpenseHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /expenses", h.create)
	mux.HandleFunc("GET /expenses", h.list)
	mux.HandleFunc("GET /expenses/{id}", h.get)
	mux.HandleFunc("PUT /expenses/{id}", h.update)
	mux.HandleFunc("DELETE /expenses/{id}", h.delete)
	mux.HandleFunc("GET /summary", h.summary)
}

// writeJSON sends a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError converts an error into a status code and a JSON message.
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, service.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, service.ErrInvalidExpense):
		status = http.StatusBadRequest
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// parseID reads the {id} part of the URL as a number.
func parseID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return 0, fmt.Errorf("%w: id must be a number", service.ErrInvalidExpense)
	}
	return id, nil
}

// maxBodyBytes limits how large a request body can be (1 MB).
const maxBodyBytes = 1 << 20

// decodeExpense reads an Expense from the JSON request body. Unknown fields
// (for example a misspelled "amout") are rejected instead of being ignored.
func decodeExpense(w http.ResponseWriter, r *http.Request) (models.Expense, error) {
	var e models.Expense
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&e); err != nil {
		return e, fmt.Errorf("%w: body must be valid JSON", service.ErrInvalidExpense)
	}
	return e, nil
}

// create handles POST /expenses: it adds an expense and returns it with 201 Created.
// Bad aspect (repetitive error handling): every call that can fail needs its own
// "if err != nil" block. writeError shortens each block, but cannot remove them.
func (h *ExpenseHandler) create(w http.ResponseWriter, r *http.Request) {
	e, err := decodeExpense(w, r)
	if err != nil {
		writeError(w, err)
		return
	}
	created, err := h.svc.Add(e)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// list handles GET /expenses: it returns every expense, sorted by ID.
func (h *ExpenseHandler) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.List())
}

// get handles GET /expenses/{id}: it returns one expense, or 404 if it does not exist.
func (h *ExpenseHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	e, err := h.svc.Get(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// update handles PUT /expenses/{id}: it replaces an expense and returns the new version.
func (h *ExpenseHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	e, err := decodeExpense(w, r)
	if err != nil {
		writeError(w, err)
		return
	}
	updated, err := h.svc.Update(id, e)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// delete handles DELETE /expenses/{id}: it removes an expense and returns 204 No Content.
func (h *ExpenseHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.svc.Delete(id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// summary handles GET /summary: it returns the count, the total and the total per category.
func (h *ExpenseHandler) summary(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.Summary())
}
