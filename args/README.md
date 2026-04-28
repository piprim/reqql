# ReqQL

## Features of this implementation

- **Type-Safe:** Heavily utilizes Go Generics (`[T any]`) to bind request structs to SQL generation logic.
- **SQL Injection Safe:** Functions return `(string, []any, error)` so placeholders (like `?` or `$1`) and arguments are safely passed directly to the database driver.
- **Highly Modular:** Uses functional options (`WithWhereFunc`, `WithLimitFunc`, `WithOrderFunc`) to keep query building clean and isolated.
- **Lightweight & Agnostic:** Does not depend on heavy external SQL builders. It generates standard queries that you can execute with `database/sql`, `sqlx`, or `pgx`.

## Installation

To install this version (parameterized arguments), use:

```bash
go get github.com/piprim/reqql/args
```

## Quick Start

Here is a complete example of how to map a search request struct to a dynamic, safe SQL query.

```go
package main

import (
	"fmt"
	"log"

	reqql "github.com/piprim/reqql/args"
)

// UserFilterRequest represents an incoming HTTP/gRPC request payload
type UserFilterRequest struct {
	Role   string
	Active bool
	Limit  string
}

func main() {
	// 1. Initialize a new Queryer for the specific request type
	q := reqql.New[UserFilterRequest]()

	// 2. Define the base SELECT/FROM query
	q.WithQueryParserFunc(func(_ *UserFilterRequest) (string, []any, error) {
		return "SELECT id, name, role FROM users", nil, nil
	})

	// 3. Define the WHERE conditions based on the struct
	// Note how we return the SQL string with placeholders AND the arguments slice
	q.WithWhereFunc(func(req *UserFilterRequest) (string, []any, error) {
		query := "role = ? AND status = ?"

		status := "inactive"
		if req.Active {
			status = "active"
		}

		args := []any{req.Role, status}
		return query, args, nil
	})

	// 4. Define the LIMIT
	q.WithLimitFunc(func(req *UserFilterRequest) (string, []any, error) {
		if req.Limit != "" {
			// Using parameters even for limits is a good security practice
			return "?", []any{req.Limit}, nil
		}

		return "10", nil, nil
	})

	// --- Executing the Queryer ---

	// Mocking an incoming request
	incomingRequest := &UserFilterRequest{
		Role:   "admin",
		Active: true,
		Limit:  "5",
	}

	// Parse generates the final SQL string and safely merges the arguments slice
	sql, args, err := q.Parse(incomingRequest)
	if err != nil {
		log.Fatalf("Failed to parse query: %v", err)
	}

	fmt.Printf("Generated SQL:\n%s\n\n", sql)
	fmt.Printf("Arguments: %v\n", args)

	// Output:
	// Generated SQL:
	// SELECT id, name, role FROM users
	// WHERE role = ? AND status = ?
	// ORDER BY 1
	// LIMIT ?
	// OFFSET 0
	// 
	// Arguments: [admin active 5]
}
```

## Advanced Execution

ReqQL provides a `Processor` wrapper that associates a `Queryer` with a specific database execution function (e.g., standard library `*sql.DB`, `sqlx`, or ORMs). Use `reqql.NewProcessor` to bind a query function and `Proceed` to execute it securely.

## Integrations

### PostgreSQL with `pgxpool` (pgx v5)

With `pgx` v5, you can use native row collection (`pgx.CollectRows` and `pgx.RowToStructByName`) to build an adapter without needing external scanning libraries. Because `CollectRows` requires knowing the destination type, the adapter function itself is generic.

```go
import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/piprim/reqql"
)

// PgxAdapter bridges pgxpool and ReqQL using native pgx v5 row collection.
// T represents the destination struct type (e.g., User).
func PgxAdapter[T any](pool *pgxpool.Pool) reqql.QueryFuncArgs {
	return func(ctx context.Context, dest any, query string, args ...any) error {
		// 1. Execute the query
		rows, err := pool.Query(ctx, query, args...)
		if err != nil {
			return err
		}

		// 2. Collect rows natively into a slice of T
		results, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
		if err != nil {
			return err
		}

		// 3. Assign the typed slice back to the dynamic 'dest' pointer
		if destPtr, ok := dest.(*[]T); ok {
			*destPtr = results
			return nil
		}

		return fmt.Errorf("invalid dest type: expected *[]%T", *new(T))
	}
}

func main() {
	// 1. Setup your pool and queryer
	// pool, _ := pgxpool.New(ctx, "postgres://...")
	q := reqql.New[MyRequest]()

	// 2. Create a processor. We instantiate the adapter with the 'User' struct.
	processor := reqql.NewProcessor(q, PgxAdapter[User](pool))

	// 3. Execute!
	var users []User
	err := processor.Proceed(context.Background(), &myRequest, &users)
}
```

### Using `sqlx`

`sqlx` is compatible out-of-the-box as its `SelectContext` method matches the `reqql.QueryFuncArgs` signature perfectly.

```go
import (
	"github.com/jmoiron/sqlx"
	"github.com/piprim/reqql"
)

// db := sqlx.MustConnect("postgres", "...")
// processor := reqql.NewProcessor(queryer, db.SelectContext)
// err := processor.Proceed(ctx, &request, &results)
```
