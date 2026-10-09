package service

import (
	"errors"
	"testing"

	"expense-api/models"
	"expense-api/repository"
)

func newTestService() *ExpenseService {
	return NewExpenseService(repository.NewInMemoryExpenseRepository())
}

func TestAddRejectsEmptyTitle(t *testing.T) {
	svc := newTestService()
	_, err := svc.Add(models.Expense{Title: "", Amount: 5, Category: "Food"})
	if !errors.Is(err, ErrInvalidExpense) {
		t.Errorf("expected ErrInvalidExpense, got %v", err)
	}
}

func TestGetMissingExpense(t *testing.T) {
	svc := newTestService()
	if _, err := svc.Get(99); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSummaryTotals(t *testing.T) {
	svc := newTestService()
	svc.Add(models.Expense{Title: "Lunch", Amount: 10, Category: "Food"})
	svc.Add(models.Expense{Title: "Snack", Amount: 5, Category: "Food"})
	svc.Add(models.Expense{Title: "Bus", Amount: 3, Category: "Transport"})

	sum := svc.Summary()
	if sum.TotalAmount != 18 || sum.ByCategory["Food"] != 15 {
		t.Errorf("unexpected summary: %+v", sum)
	}
}
