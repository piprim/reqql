# ReqQL

*ReqQL* is a lightweight, generic-driven bridge that seamlessly turns your Go request structs (HTTP/gRPC) into dynamic SQL queries.
Stop maintaining bulky GraphQL schemas and let your Go structs safely and expressively define your database queries.

The project offers two distinct implementations depending on your needs for simplicity versus power, organized as standard Go sub-packages.

## Choose Your Flavor

### 1. [Args Implementation (`github.com/piprim/reqql/args`)](./args/README.md)
**Zero-dependency and lightweight.**
This implementation relies purely on standard Go strings and `[]any` slices to manage parameterized queries. It is ideal for developers who want a bare-metal approach with maximum control and no external dependencies (aside from `pkg/errors`).

- **Returns:** `(string, []any, error)`
- **Safety:** Uses standard SQL placeholders (e.g., `?` or `$1`) and manual argument collection.
- **Best for:** Small to medium projects where you want to keep the binary size minimal.
- **Import:** `import "github.com/piprim/reqql/args"`

### 2. [Goqu Implementation (`github.com/piprim/reqql/goqu`)](./goqu/README.md)
**Powerful and expressive.**
This implementation is built on top of the [goqu](https://github.com/doug-martin/goqu) SQL builder. It allows you to return high-level expressions and datasets instead of raw SQL strings.

- **Returns:** `*goqu.SelectDataset` and `goqu.Expression`.
- **Safety:** Automatically handles parameterization, identifier escaping, and dialect-specific formatting (PostgreSQL, MySQL, SQLite, etc.).
- **Best for:** Complex queries, cross-database support, and developers who prefer a fluent API over raw SQL segments.
- **Import:** `import "github.com/piprim/reqql/goqu"`

---

## Core Features (Both Versions)

- **Type-Safe:** Uses Go Generics (`[T any]`) to bind your request structs directly to your query logic.
- **SQL Injection Safe:** Both implementations are designed from the ground up to use parameterized queries.
- **Highly Modular:** Uses functional options (`WithWhereFunc`, `WithLimitFunc`, etc.) to keep your code clean, readable, and easy to test.
- **Agnostic Execution:** ReqQL focuses only on *generating* the query. You can execute the result with standard `database/sql`, `sqlx`, or `pgx`.
