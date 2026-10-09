package models

// Expense represents one spending record made by a student.
type Expense struct {
	ID       int     `json:"id"`       // unique number identifying the expense
	Title    string  `json:"title"`    // short name, e.g. "Lunch"
	Amount   float64 `json:"amount"`   // money spent (float64 is simple, but can have rounding errors)
	Category string  `json:"category"` // grouping used for the summary, e.g. "Food"
	Date     string  `json:"date"`     // stored as text (YYYY-MM-DD) to keep the project simple
	Note     string  `json:"note"`     // optional extra detail
}
