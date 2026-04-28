# ReqQL

*ReqQL* is a lightweight, generic-driven bridge that seamlessly turns your Go request structs (HTTP/gRPC) into dynamic SQL queries.  
Stop maintaining bulky GraphQL schemas and let your Go structs safely and expressively define your database queries.

The project offers two distinct implementations depending on your needs for simplicity and flexibility versus power and expressivity, organized as standalone Go packages.

## Choose Your Flavor

### 1. Args Implementation

**Zero-dependency and lightweight.**
This implementation relies purely on standard Go strings and `[]any` slices to manage parameterized queries.  
It is ideal for developers who want a bare-metal approach with maximum control and no external dependencies (aside from `pkg/errors`).

> Take a look at the [Args Implementation Documentation](./args/README.md)

### 2. Goqu Implementation

**Powerful and expressive.**
This implementation is built on top of the [goqu](https://github.com/doug-martin/goqu) SQL builder.  
It allows you to return high-level expressions and datasets instead of raw SQL strings.

> Take a look at the [Goqu Implementation Documentation](./goqu/README.md)

---

## Core Features (Both Versions)

- **Type-Safe:** Uses Go Generics (`[T any]`) to bind your request structs directly to your query logic.
- **SQL Injection Safe:** Both implementations are designed from the ground up to use parameterized queries.
- **Highly Modular:** Uses functional options (`WithWhereFunc`, `WithLimitFunc`, etc.) to keep your code clean, readable, and easy to test.
- **Agnostic Execution:** ReqQL focuses only on *generating* the query. You can execute the result with standard `database/sql`, `sqlx`, or `pgx`.
