package main

import (
	"log"
	"net/http"

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

	log.Println("Expense API listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
