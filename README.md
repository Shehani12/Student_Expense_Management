# Student Expense Management System

A REST API web service written in Go (standard library only) for recording and
summarising student expenses. Built for CSP3341 Programming Languages and Paradigms.

## Run

```
go run .
```

The server listens on http://localhost:8080. Press Ctrl+C to stop it; requests
already in progress are allowed to finish first.

## Test

```
go test ./... -v
```

## Endpoints

| Method | Endpoint         | Description                       |
|--------|------------------|-----------------------------------|
| POST   | /expenses        | Add an expense                    |
| GET    | /expenses        | List all expenses                 |
| GET    | /expenses/{id}   | Get one expense                   |
| PUT    | /expenses/{id}   | Update an expense                 |
| DELETE | /expenses/{id}   | Delete an expense                 |
| GET    | /summary         | Total spent and total per category|

Example request body:

```json
{"title": "Lunch", "amount": 12.50, "category": "Food", "date": "2026-10-13", "note": "Campus cafe"}
```

`title`, `amount` (greater than 0) and `category` are required. `date` is
optional but must use the YYYY-MM-DD format. Data is kept in memory, so it
resets when the server restarts.

## Structure

```
main.go                      starts the server, graceful shutdown
handlers/expense_handler.go  HTTP + JSON layer (routes, status codes)
handlers/middleware.go       request logging middleware
service/expense_service.go   business rules and validation
repository/                  storage interface + in-memory implementation
models/expense.go            the Expense struct shared by all layers
```
