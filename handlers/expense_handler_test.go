package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"expense-api/repository"
	"expense-api/service"
)

// newTestServer builds the full stack with an empty in-memory repository.
func newTestServer() *http.ServeMux {
	svc := service.NewExpenseService(repository.NewInMemoryExpenseRepository())
	mux := http.NewServeMux()
	NewExpenseHandler(svc).RegisterRoutes(mux)
	return mux
}

// do sends one request to the server and returns the recorded response.
func do(mux *http.ServeMux, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestStatusCodes(t *testing.T) {
	mux := newTestServer()
	valid := `{"title":"Lunch","amount":5,"category":"Food","date":"2026-10-12"}`

	tests := []struct {
		name         string
		method, path string
		body         string
		want         int
	}{
		{"create valid", "POST", "/expenses", valid, http.StatusCreated},
		{"create invalid JSON", "POST", "/expenses", `{`, http.StatusBadRequest},
		{"create unknown field", "POST", "/expenses", `{"title":"Lunch","amout":5,"category":"Food"}`, http.StatusBadRequest},
		{"create bad date", "POST", "/expenses", `{"title":"Lunch","amount":5,"category":"Food","date":"tomorrow"}`, http.StatusBadRequest},
		{"list", "GET", "/expenses", "", http.StatusOK},
		{"get existing", "GET", "/expenses/1", "", http.StatusOK},
		{"get missing", "GET", "/expenses/99", "", http.StatusNotFound},
		{"get non-numeric id", "GET", "/expenses/abc", "", http.StatusBadRequest},
		{"update existing", "PUT", "/expenses/1", valid, http.StatusOK},
		{"update missing", "PUT", "/expenses/99", valid, http.StatusNotFound},
		{"summary", "GET", "/summary", "", http.StatusOK},
		{"delete existing", "DELETE", "/expenses/1", "", http.StatusNoContent},
		{"delete again", "DELETE", "/expenses/1", "", http.StatusNotFound},
	}
	// The cases run in order and share one server, so later cases see earlier changes.
	for _, tt := range tests {
		rec := do(mux, tt.method, tt.path, tt.body)
		if rec.Code != tt.want {
			t.Errorf("%s: %s %s returned %d, want %d (body: %s)",
				tt.name, tt.method, tt.path, rec.Code, tt.want, rec.Body.String())
		}
	}
}

func TestLogRequestsKeepsStatus(t *testing.T) {
	handler := LogRequests(newTestServer())
	req := httptest.NewRequest("GET", "/expenses/99", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("middleware changed the status: got %d, want 404", rec.Code)
	}
}
