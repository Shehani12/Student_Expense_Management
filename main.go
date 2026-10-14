package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"expense-api/handlers"
	"expense-api/repository"
	"expense-api/service"
)

func main() {
	// Build the layers from the bottom up and connect them.
	repo := repository.NewInMemoryExpenseRepository()
	svc := service.NewExpenseService(repo)
	handler := handlers.NewExpenseHandler(svc)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Timeouts stop slow or stuck clients from holding connections open forever.
	// LogRequests wraps the router so every request is printed to the terminal.
	server := &http.Server{
		Addr:              ":8080",
		Handler:           handlers.LogRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start the server in its own goroutine so main can wait for events.
	// The buffered channel carries any start-up error back to main.
	serverErr := make(chan error, 1)
	go func() {
		log.Println("Expense API listening on http://localhost:8080")
		serverErr <- server.ListenAndServe()
	}()

	// Ctrl+C is an operating-system event. signal.NotifyContext turns it into
	// a context that is cancelled when the signal arrives.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// select waits for whichever event happens first.
	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	case <-ctx.Done():
		log.Println("Shutting down: finishing requests already in progress...")
		// Give running requests up to 5 seconds to finish.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
		log.Println("Server stopped")
	}
}
