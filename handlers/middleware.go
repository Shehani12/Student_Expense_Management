package handlers

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder wraps http.ResponseWriter so the status code can be read
// after the handler has run. Embedding the interface promotes all of its
// methods, so only WriteHeader needs to be written here (composition, not
// inheritance).
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader remembers the status code, then passes it on.
func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// LogRequests is middleware: a function that takes a handler and returns a
// new handler. The returned function is a closure that can still use next.
// It logs the method, path, status code and time taken for every request.
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s -> %d (%v)", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}
