# ReqQL

## Features of this implementation

- **Type-Safe:** Heavily utilizes Go Generics (`[T any]`) to bind request structs to SQL generation logic.
- **SQL Injection Safe:** Powered by `goqu`, eliminating the need to manually concatenate SQL strings and manually manage arguments.
- **Highly Modular:** Uses functional options (`WithWhereFunc`, `WithLimitFunc`, `WithOrderFunc`) to keep query building clean and isolated.
- **Dialect Agnostic:** Leverages `goqu`'s dialect system, so you can easily adapt the generated SQL for PostgreSQL, MySQL, SQLite, etc.

## Installation

```bash
go get github.com/piprim/reqql/goqu
```

## Quick Start

Here is a complete example of how to map a search request struct to a dynamic, safe SQL query.

```go
package main

import (
	"fmt"
	"log"

	"github.com/doug-martin/goqu/v9"
	reqql "github.com/piprim/reqql/goqu"
)

// UserFilterRequest represents an incoming HTTP/gRPC request payload
type UserFilterRequest struct {
	Role   string
	Active bool
	Limit  uint
}

func main() {
	// 1. Initialize a new Queryer for the specific request type
	q := reqql.New[UserFilterRequest]()

	// 2. Define the base SELECT/FROM dataset
	q.WithQueryParserFunc(func(req *UserFilterRequest) (*goqu.SelectDataset, error) {
		// Use Prepared(true) to force parameterization (e.g. ? or $1)
		return goqu.From("users").Select("id", "name", "role").Prepared(true), nil
	})

	// 3. Define the WHERE conditions based on the struct
	q.WithWhereFunc(func(req *UserFilterRequest) goqu.Expression {
		ex := goqu.Ex{}

		// Dynamically add conditions based on request fields
		if req.Role != "" {
			ex["role"] = req.Role
		}
		if req.Active {
			ex["status"] = "active"
		}

		return ex
	})

	// 4. Define the LIMIT based on the struct
	q.WithLimitFunc(func(req *UserFilterRequest) any {
		if req.Limit > 0 {
			return req.Limit
		}
		return uint(10) // Default limit
	})

	// --- Executing the Queryer ---

	// Mocking an incoming request
	incomingRequest := &UserFilterRequest{
		Role:   "admin",
		Active: true,
		Limit:  5,
	}

	// Parse generates the final SQL string and the arguments slice
	sql, args, err := q.Parse(incomingRequest)
	if err != nil {
		log.Fatalf("Failed to parse query: %v", err)
	}

	fmt.Printf("Generated SQL: %s\n", sql)
	fmt.Printf("Arguments:     %v\n", args)

	// Output:
	// Generated SQL: SELECT "id", "name", "role" FROM "users" WHERE (("role" = ?) AND ("status" = ?)) LIMIT ?
	// Arguments:     [admin active 5]
}
```

## Advanced Execution

ReqQL provides a `Processor` wrapper that associates a `Queryer` with a specific database execution function (e.g., standard library `*sql.DB`, `sqlx`, or ORMs). Use `reqql.NewProcessor` to bind a query function and `Proceed` to execute it securely.

## Integrations

ReqQL is designed to be agnostic of your database driver. Below are common recipes for connecting ReqQL to popular database libraries.

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

// Usage:
// pool, _ := pgxpool.New(ctx, "postgres://...")
// processor := reqql.NewProcessor(queryer, PgxAdapter[User](pool))
// err := processor.Proceed(ctx, &request, &results)
```

### Using `sqlx`

`sqlx` is compatible out-of-the-box as its `SelectContext` method matches the `reqql.QueryFuncArgs` signature perfectly.

```go
import (
	"github.com/jmoiron/sqlx"
	"github.com/piprim/reqql"
)

// Usage:
// db := sqlx.MustConnect("postgres", "...")
// processor := reqql.NewProcessor(queryer, db.SelectContext)
// err := processor.Proceed(ctx, &request, &results)
```
