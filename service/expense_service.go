package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"expense-api/models"
	"expense-api/repository"
)

// Custom errors. Go returns errors as values instead of throwing exceptions.
var (
	ErrNotFound       = errors.New("expense not found")
	ErrInvalidExpense = errors.New("invalid expense")
)

// Summary holds the totals returned by the /summary endpoint.
type Summary struct {
	Count       int                `json:"count"`
	TotalAmount float64            `json:"total_amount"`
	ByCategory  map[string]float64 `json:"by_category"`
}

// ExpenseService contains the business rules. It depends on the
// repository interface, not on a specific storage type.
type ExpenseService struct {
	repo repository.ExpenseRepository
}

// NewExpenseService creates a service using the given repository.
func NewExpenseService(repo repository.ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

// clean trims spaces from the text fields, so " Food " and "Food" are
// treated as the same category.
func clean(e models.Expense) models.Expense {
	e.Title = strings.TrimSpace(e.Title)
	e.Category = strings.TrimSpace(e.Category)
	e.Date = strings.TrimSpace(e.Date)
	e.Note = strings.TrimSpace(e.Note)
	return e
}

// validate checks the rules for an expense. Each error is wrapped with %w,
// so callers can test it with errors.Is(err, ErrInvalidExpense).
func validate(e models.Expense) error {
	if e.Title == "" {
		return fmt.Errorf("%w: title must not be empty", ErrInvalidExpense)
	}
	// Bad aspect (zero values): if "amount" is missing from the JSON, Go sets it
	// to 0, so a missing amount and a real 0 look the same. This check catches both.
	if e.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidExpense)
	}
	if e.Category == "" {
		return fmt.Errorf("%w: category must not be empty", ErrInvalidExpense)
	}
	// The date is optional, but if given it must be a real YYYY-MM-DD date.
	if e.Date != "" {
		if _, err := time.Parse(time.DateOnly, e.Date); err != nil {
			return fmt.Errorf("%w: date must be in YYYY-MM-DD format", ErrInvalidExpense)
		}
	}
	return nil
}

// roundCents rounds an amount to 2 decimal places.
// Bad aspect (no decimal type): float64 cannot store most decimals exactly, so
// 0.1 + 0.2 gives 0.30000000000000004. Totals are rounded before being returned.
func roundCents(amount float64) float64 {
	return math.Round(amount*100) / 100
}

// Add cleans, validates and stores a new expense.
func (s *ExpenseService) Add(e models.Expense) (models.Expense, error) {
	e = clean(e)
	if err := validate(e); err != nil {
		return models.Expense{}, err
	}
	return s.repo.Create(e), nil
}

// List returns all expenses.
func (s *ExpenseService) List() []models.Expense {
	return s.repo.GetAll()
}

// Get returns one expense or ErrNotFound.
func (s *ExpenseService) Get(id int) (models.Expense, error) {
	e, ok := s.repo.GetByID(id)
	if !ok {
		return models.Expense{}, ErrNotFound
	}
	return e, nil
}

// Update cleans, validates and replaces an expense, or returns ErrNotFound.
func (s *ExpenseService) Update(id int, e models.Expense) (models.Expense, error) {
	e = clean(e)
	if err := validate(e); err != nil {
		return models.Expense{}, err
	}
	updated, ok := s.repo.Update(id, e)
	if !ok {
		return models.Expense{}, ErrNotFound
	}
	return updated, nil
}

// Delete removes an expense, or returns ErrNotFound.
func (s *ExpenseService) Delete(id int) error {
	if !s.repo.Delete(id) {
		return ErrNotFound
	}
	return nil
}

// Summary adds up the total spending and the total per category.
func (s *ExpenseService) Summary() Summary {
	sum := Summary{ByCategory: make(map[string]float64)}
	for _, e := range s.repo.GetAll() {
		sum.Count++
		sum.TotalAmount += e.Amount
		sum.ByCategory[e.Category] += e.Amount
	}
	sum.TotalAmount = roundCents(sum.TotalAmount)
	for category, total := range sum.ByCategory {
		sum.ByCategory[category] = roundCents(total)
	}
	return sum
}
