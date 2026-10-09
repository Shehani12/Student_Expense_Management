package repository

import (
	"sort"
	"sync"

	"expense-api/models"
)

// ExpenseRepository describes how expenses are stored. Any type that has
// these methods satisfies the interface automatically (implicit interfaces).
type ExpenseRepository interface {
	Create(e models.Expense) models.Expense
	GetAll() []models.Expense
	GetByID(id int) (models.Expense, bool)
	Update(id int, e models.Expense) (models.Expense, bool)
	Delete(id int) bool
}

// InMemoryExpenseRepository keeps expenses in a map held in memory.
// The mutex makes it safe when several requests arrive at the same time.
type InMemoryExpenseRepository struct {
	mu       sync.RWMutex
	expenses map[int]models.Expense
	nextID   int
}

// NewInMemoryExpenseRepository creates an empty repository.
func NewInMemoryExpenseRepository() *InMemoryExpenseRepository {
	return &InMemoryExpenseRepository{
		expenses: make(map[int]models.Expense),
		nextID:   1,
	}
}

// Create assigns a new ID, stores the expense and returns it.
func (r *InMemoryExpenseRepository) Create(e models.Expense) models.Expense {
	r.mu.Lock()
	defer r.mu.Unlock() // unlocks when the function ends
	e.ID = r.nextID
	r.nextID++
	r.expenses[e.ID] = e
	return e
}

// GetAll returns every expense, sorted by ID.
func (r *InMemoryExpenseRepository) GetAll() []models.Expense {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]models.Expense, 0, len(r.expenses))
	for _, e := range r.expenses {
		list = append(list, e)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list
}

// GetByID returns one expense; the bool says whether it was found.
func (r *InMemoryExpenseRepository) GetByID(id int) (models.Expense, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.expenses[id]
	return e, ok
}

// Update replaces an existing expense, keeping its ID.
func (r *InMemoryExpenseRepository) Update(id int, e models.Expense) (models.Expense, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.expenses[id]; !ok {
		return models.Expense{}, false
	}
	e.ID = id
	r.expenses[id] = e
	return e, true
}

// Delete removes an expense and reports whether it existed.
func (r *InMemoryExpenseRepository) Delete(id int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.expenses[id]; !ok {
		return false
	}
	delete(r.expenses, id)
	return true
}
