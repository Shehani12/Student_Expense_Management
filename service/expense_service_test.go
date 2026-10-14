package service

import (
	"errors"
	"fmt"
	"testing"

	"expense-api/models"
	"expense-api/repository"
)

func newTestService() *ExpenseService {
	return NewExpenseService(repository.NewInMemoryExpenseRepository())
}

// mustAdd adds an expense and stops the test if it fails.
func mustAdd(t *testing.T, svc *ExpenseService, e models.Expense) models.Expense {
	t.Helper()
	created, err := svc.Add(e)
	if err != nil {
		t.Fatalf("Add(%+v) failed: %v", e, err)
	}
	return created
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
	mustAdd(t, svc, models.Expense{Title: "Lunch", Amount: 10, Category: "Food"})
	mustAdd(t, svc, models.Expense{Title: "Snack", Amount: 5, Category: " Food "})
	mustAdd(t, svc, models.Expense{Title: "Bus", Amount: 3, Category: "Transport"})

	sum := mustSummary(t, svc, Filter{})
	if sum.Count != 3 || sum.TotalAmount != 18 || sum.ByCategory["Food"] != 15 {
		t.Errorf("unexpected summary: %+v", sum)
	}
}

func TestAddValidation(t *testing.T) {
	tests := []struct {
		name string
		e    models.Expense
	}{
		{"blank title", models.Expense{Title: "   ", Amount: 5, Category: "Food"}},
		{"zero amount", models.Expense{Title: "Lunch", Amount: 0, Category: "Food"}},
		{"negative amount", models.Expense{Title: "Lunch", Amount: -1, Category: "Food"}},
		{"empty category", models.Expense{Title: "Lunch", Amount: 5}},
		{"bad date format", models.Expense{Title: "Lunch", Amount: 5, Category: "Food", Date: "12/10/2026"}},
		{"impossible date", models.Expense{Title: "Lunch", Amount: 5, Category: "Food", Date: "2026-02-30"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newTestService().Add(tt.e)
			if !errors.Is(err, ErrInvalidExpense) {
				t.Errorf("expected ErrInvalidExpense, got %v", err)
			}
		})
	}
}

func TestAddTrimsTextFields(t *testing.T) {
	svc := newTestService()
	e := mustAdd(t, svc, models.Expense{Title: "  Lunch ", Amount: 5, Category: " Food", Date: " 2026-10-12 "})
	if e.Title != "Lunch" || e.Category != "Food" || e.Date != "2026-10-12" {
		t.Errorf("fields were not trimmed: %+v", e)
	}
}

func TestUpdate(t *testing.T) {
	svc := newTestService()
	created := mustAdd(t, svc, models.Expense{Title: "Lunch", Amount: 5, Category: "Food"})

	updated, err := svc.Update(created.ID, models.Expense{Title: "Dinner", Amount: 8, Category: "Food"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.ID != created.ID || updated.Title != "Dinner" || updated.Amount != 8 {
		t.Errorf("unexpected updated expense: %+v", updated)
	}

	if _, err := svc.Update(99, models.Expense{Title: "X", Amount: 1, Category: "Y"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if _, err := svc.Update(created.ID, models.Expense{}); !errors.Is(err, ErrInvalidExpense) {
		t.Errorf("expected ErrInvalidExpense, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	svc := newTestService()
	created := mustAdd(t, svc, models.Expense{Title: "Lunch", Amount: 5, Category: "Food"})

	if err := svc.Delete(created.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := svc.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("expense still exists after delete")
	}
	if err := svc.Delete(created.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound on second delete, got %v", err)
	}
}

// TestSummaryRoundsFloatTotals shows the float64 problem with money and the fix.
func TestSummaryRoundsFloatTotals(t *testing.T) {
	a, b := 0.1, 0.2
	if a+b == 0.3 {
		t.Fatalf("expected a float64 rounding error, but 0.1 + 0.2 == 0.3")
	}

	svc := newTestService()
	mustAdd(t, svc, models.Expense{Title: "Pen", Amount: 0.1, Category: "Stationery"})
	mustAdd(t, svc, models.Expense{Title: "Pencil", Amount: 0.2, Category: "Stationery"})
	sum := mustSummary(t, svc, Filter{})
	if sum.TotalAmount != 0.3 || sum.ByCategory["Stationery"] != 0.3 {
		t.Errorf("expected totals of 0.3 after rounding, got %+v", sum)
	}
}

// mustSummary gets a summary and stops the test if it fails.
func mustSummary(t *testing.T, svc *ExpenseService, f Filter) Summary {
	t.Helper()
	sum, err := svc.Summary(f)
	if err != nil {
		t.Fatalf("Summary(%+v) failed: %v", f, err)
	}
	return sum
}

// newFilterTestService returns a service holding expenses across two months.
func newFilterTestService(t *testing.T) *ExpenseService {
	svc := newTestService()
	mustAdd(t, svc, models.Expense{Title: "Lunch", Amount: 10, Category: "Food", Date: "2026-09-28"})
	mustAdd(t, svc, models.Expense{Title: "Bus", Amount: 3, Category: "Transport", Date: "2026-10-01"})
	mustAdd(t, svc, models.Expense{Title: "Dinner", Amount: 12, Category: "Food", Date: "2026-10-15"})
	mustAdd(t, svc, models.Expense{Title: "Book", Amount: 25, Category: "Study"}) // no date
	return svc
}

func TestListFilter(t *testing.T) {
	svc := newFilterTestService(t)
	tests := []struct {
		name    string
		f       Filter
		wantIDs []int
	}{
		{"no filter returns all", Filter{}, []int{1, 2, 3, 4}},
		{"category ignores case", Filter{Category: "food"}, []int{1, 3}},
		{"from is inclusive", Filter{From: "2026-10-01"}, []int{2, 3}},
		{"to is inclusive", Filter{To: "2026-10-01"}, []int{1, 2}},
		{"one month", Filter{From: "2026-10-01", To: "2026-10-31"}, []int{2, 3}},
		{"category and month", Filter{Category: "Food", From: "2026-10-01", To: "2026-10-31"}, []int{3}},
		{"no match", Filter{Category: "Rent"}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list, err := svc.List(tt.f)
			if err != nil {
				t.Fatalf("List failed: %v", err)
			}
			gotIDs := make([]int, 0, len(list))
			for _, e := range list {
				gotIDs = append(gotIDs, e.ID)
			}
			if fmt.Sprint(gotIDs) != fmt.Sprint(tt.wantIDs) {
				t.Errorf("got IDs %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestInvalidFilter(t *testing.T) {
	svc := newFilterTestService(t)
	for _, f := range []Filter{
		{From: "01/10/2026"},
		{To: "2026-13-01"},
		{From: "2026-10-31", To: "2026-10-01"}, // backwards range
	} {
		if _, err := svc.List(f); !errors.Is(err, ErrInvalidFilter) {
			t.Errorf("List(%+v): expected ErrInvalidFilter, got %v", f, err)
		}
		if _, err := svc.Summary(f); !errors.Is(err, ErrInvalidFilter) {
			t.Errorf("Summary(%+v): expected ErrInvalidFilter, got %v", f, err)
		}
	}
}

func TestSummaryForOneMonth(t *testing.T) {
	svc := newFilterTestService(t)
	sum := mustSummary(t, svc, Filter{From: "2026-10-01", To: "2026-10-31"})
	if sum.Count != 2 || sum.TotalAmount != 15 || sum.ByCategory["Food"] != 12 || sum.ByCategory["Transport"] != 3 {
		t.Errorf("unexpected October summary: %+v", sum)
	}
	if sum.From != "2026-10-01" || sum.To != "2026-10-31" {
		t.Errorf("summary should report the date range used, got from=%q to=%q", sum.From, sum.To)
	}
}
