package repository

import (
	"sync"
	"testing"

	"expense-api/models"
)

// TestConcurrentCreate starts many goroutines that add expenses at the same
// time, like many HTTP requests arriving together. The mutex must make sure
// no expense is lost and every ID is different.
func TestConcurrentCreate(t *testing.T) {
	repo := NewInMemoryExpenseRepository()
	const workers = 100

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			repo.Create(models.Expense{Title: "Lunch", Amount: 5, Category: "Food"})
		}()
	}
	wg.Wait() // block until every goroutine has finished

	all := repo.GetAll()
	if len(all) != workers {
		t.Fatalf("expected %d expenses, got %d", workers, len(all))
	}
	for i, e := range all {
		if e.ID != i+1 {
			t.Fatalf("expected IDs 1..%d in order, found ID %d at position %d", workers, e.ID, i)
		}
	}
}
