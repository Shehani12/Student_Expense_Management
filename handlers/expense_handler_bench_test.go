package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

// Benchmarks measure how long one request takes through the whole stack
// (router, handler, service and repository), without the network.
// Run with: go test ./handlers -bench . -benchmem

// newBenchServer returns a server that already holds n expenses.
func newBenchServer(b *testing.B, n int) *http.ServeMux {
	b.Helper()
	mux := newTestServer()
	categories := []string{"Food", "Transport", "Study"}
	for i := range n {
		body := fmt.Sprintf(`{"title":"Item %d","amount":%d.5,"category":"%s","date":"2026-10-%02d"}`,
			i, i%50+1, categories[i%3], i%28+1)
		if rec := do(mux, "POST", "/expenses", body); rec.Code != http.StatusCreated {
			b.Fatalf("setup failed: %d %s", rec.Code, rec.Body.String())
		}
	}
	return mux
}

// BenchmarkCreateExpense measures POST /expenses: decode JSON, validate, store.
func BenchmarkCreateExpense(b *testing.B) {
	mux := newTestServer()
	body := `{"title":"Lunch","amount":12.5,"category":"Food","date":"2026-10-14"}`
	for b.Loop() {
		do(mux, "POST", "/expenses", body)
	}
}

// BenchmarkGetExpense measures GET /expenses/{id} with 1000 stored expenses.
func BenchmarkGetExpense(b *testing.B) {
	mux := newBenchServer(b, 1000)
	for b.Loop() {
		do(mux, "GET", "/expenses/500", "")
	}
}

// BenchmarkSummary measures GET /summary, which reads all 1000 expenses.
func BenchmarkSummary(b *testing.B) {
	mux := newBenchServer(b, 1000)
	for b.Loop() {
		do(mux, "GET", "/summary", "")
	}
}

// BenchmarkGetExpenseParallel sends GET requests from many goroutines at once,
// like many students using the service together. RLock lets them read in parallel.
func BenchmarkGetExpenseParallel(b *testing.B) {
	mux := newBenchServer(b, 1000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			do(mux, "GET", "/expenses/500", "")
		}
	})
}
