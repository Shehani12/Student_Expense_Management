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

// decodeExpense reads an Expense from the JSON request body.
func decodeExpense(r *http.Request) (models.Expense, error) {
	var e models.Expense
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		return e, fmt.Errorf("%w: body must be valid JSON", service.ErrInvalidExpense)
	}
	return e, nil
}

func (h *ExpenseHandler) create(w http.ResponseWriter, r *http.Request) {
	e, err := decodeExpense(r)
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

func (h *ExpenseHandler) list(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.List())
}

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

func (h *ExpenseHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	e, err := decodeExpense(r)
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

func (h *ExpenseHandler) summary(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.Summary())
}
