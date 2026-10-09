package service

import (
	"errors"
	"fmt"
	"strings"

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

// validate checks the rules for an expense.
func validate(e models.Expense) error {
	if strings.TrimSpace(e.Title) == "" {
		return fmt.Errorf("%w: title must not be empty", ErrInvalidExpense)
	}
	if e.Amount <= 0 {
		return fmt.Errorf("%w: amount must be greater than 0", ErrInvalidExpense)
	}
	if strings.TrimSpace(e.Category) == "" {
		return fmt.Errorf("%w: category must not be empty", ErrInvalidExpense)
	}
	return nil
}

// Add validates and stores a new expense.
func (s *ExpenseService) Add(e models.Expense) (models.Expense, error) {
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

// Update validates and replaces an expense, or returns ErrNotFound.
func (s *ExpenseService) Update(id int, e models.Expense) (models.Expense, error) {
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
	return sum
}
