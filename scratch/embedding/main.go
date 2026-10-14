// Listing 7: composition instead of inheritance.
// Run with: go run ./scratch/embedding
package main

import (
	"fmt"

	"expense-api/models"
)

// Timestamps is a small struct that can be reused by embedding it.
type Timestamps struct{ CreatedAt string }

// AuditedExpense embeds models.Expense and Timestamps. Their fields are
// promoted, so they can be used as if they belonged to AuditedExpense.
type AuditedExpense struct {
	models.Expense // embedded: its fields and methods are promoted
	Timestamps
}

func main() {
	a := AuditedExpense{}
	a.Title = "Lunch" // accessed directly, with no inheritance
	a.CreatedAt = "2026-10-13"
	fmt.Printf("%s created at %s\n", a.Title, a.CreatedAt)
	fmt.Printf("full value: %+v\n", a)
}
