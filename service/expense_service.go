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
	ErrInvalidFilter  = errors.New("invalid filter")
)

// Filter limits which expenses are listed or summarised. Empty fields are
// ignored, so an empty Filter matches every expense.
type Filter struct {
	Category string // only this category (upper/lower case does not matter)
	From     string // earliest date to include, YYYY-MM-DD
	To       string // latest date to include, YYYY-MM-DD
}

// Summary holds the totals returned by the /summary endpoint.
// Category, From and To show which filter was used, and are left out of the
// JSON when empty.
type Summary struct {
	Category    string             `json:"category,omitempty"`
	From        string             `json:"from,omitempty"`
	To          string             `json:"to,omitempty"`
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
	if e.Date != "" && !isDate(e.Date) {
		return fmt.Errorf("%w: date must be in YYYY-MM-DD format", ErrInvalidExpense)
	}
	return nil
}

// isDate reports whether s is a real date in YYYY-MM-DD format.
func isDate(s string) bool {
	_, err := time.Parse(time.DateOnly, s)
	return err == nil
}

// cleaned returns a copy of the filter with spaces trimmed from each field.
func (f Filter) cleaned() Filter {
	f.Category = strings.TrimSpace(f.Category)
	f.From = strings.TrimSpace(f.From)
	f.To = strings.TrimSpace(f.To)
	return f
}

// check reports an error if a date is badly formatted or the range is backwards.
func (f Filter) check() error {
	if f.From != "" && !isDate(f.From) {
		return fmt.Errorf("%w: from must be in YYYY-MM-DD format", ErrInvalidFilter)
	}
	if f.To != "" && !isDate(f.To) {
		return fmt.Errorf("%w: to must be in YYYY-MM-DD format", ErrInvalidFilter)
	}
	if f.From != "" && f.To != "" && f.From > f.To {
		return fmt.Errorf("%w: from must not be after to", ErrInvalidFilter)
	}
	return nil
}

// matches reports whether an expense passes the filter. Dates in YYYY-MM-DD
// format sort the same way as text, so they can be compared as strings.
func (f Filter) matches(e models.Expense) bool {
	if f.Category != "" && !strings.EqualFold(e.Category, f.Category) {
		return false
	}
	if (f.From != "" || f.To != "") && e.Date == "" {
		return false // an expense without a date cannot be inside a date range
	}
	if f.From != "" && e.Date < f.From {
		return false
	}
	if f.To != "" && e.Date > f.To {
		return false
	}
	return true
}

// filter checks f and returns only the expenses that match it.
func (s *ExpenseService) filter(f Filter) ([]models.Expense, error) {
	if err := f.check(); err != nil {
		return nil, err
	}
	list := make([]models.Expense, 0)
	for _, e := range s.repo.GetAll() {
		if f.matches(e) {
			list = append(list, e)
		}
	}
	return list, nil
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

// List returns the expenses that match the filter, sorted by ID.
// An empty Filter returns every expense.
func (s *ExpenseService) List(f Filter) ([]models.Expense, error) {
	return s.filter(f.cleaned())
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

// Summary adds up the total spending and the total per category for the
// expenses that match the filter, for example one month of spending.
func (s *ExpenseService) Summary(f Filter) (Summary, error) {
	f = f.cleaned()
	list, err := s.filter(f)
	if err != nil {
		return Summary{}, err
	}
	sum := Summary{
		Category:   f.Category,
		From:       f.From,
		To:         f.To,
		ByCategory: make(map[string]float64),
	}
	for _, e := range list {
		sum.Count++
		sum.TotalAmount += e.Amount
		sum.ByCategory[e.Category] += e.Amount
	}
	sum.TotalAmount = roundCents(sum.TotalAmount)
	for category, total := range sum.ByCategory {
		sum.ByCategory[category] = roundCents(total)
	}
	return sum, nil
}
