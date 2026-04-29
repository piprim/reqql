// Package reqql assembles parameterized SQL SELECT queries from a typed input
// struct using a set of composable builder functions.
//
// # Overview
//
// A [Queryer] is constructed with [New] and configured by chaining five
// builder functions, each responsible for one SQL clause:
//
//   - [QueryFuncType] — the "SELECT … FROM …" fragment
//   - [WhereFuncType] — the WHERE predicate
//   - [OrderFuncType] — the ORDER BY expression
//   - [LimitFuncType] — the LIMIT value
//   - [OffsetFuncType] — the OFFSET value
//
// Every builder function receives a pointer to the typed input and returns
// a SQL fragment string plus any positional arguments (`?` placeholders)
// for that clause.  [Queryer.Parse] concatenates the fragments into the
// final query:
//
//	<query>
//	WHERE <where>
//	ORDER BY <order>
//	LIMIT <limit>
//	OFFSET <offset>
//
// Arguments are collected in the order query → where → order → limit →
// offset, which matches the positional placeholders left-to-right in the
// assembled SQL.
//
// All five builder functions are initialized to safe no-ops by [New], so
// only the functions relevant to the domain need to be overridden.
//
// # Quick Start
//
// Build and inspect a query:
//
//	type ProductFilter struct {
//	    Category string
//	    MaxPrice float64
//	}
//
//	q := reqql.New[ProductFilter]()
//	q.WithQueryParserFunc(func(_ *ProductFilter) (string, []any, error) {
//	    return "SELECT id, name, price FROM products", nil, nil
//	}).WithWhereFunc(func(f *ProductFilter) (string, []any, error) {
//	    return "category = ? AND price <= ?", []any{f.Category, f.MaxPrice}, nil
//	}).WithOrderFunc(func(_ *ProductFilter) (string, []any, error) {
//	    return "price ASC", nil, nil
//	})
//
//	sql, args, err := q.Parse(&ProductFilter{Category: "books", MaxPrice: 29.99})
//	// sql:  "SELECT id, name, price FROM products\nWHERE category = ? AND price <= ?\nORDER BY price ASC\nLIMIT ALL\nOFFSET 0"
//	// args: ["books", 29.99]
//
// # Executing Queries
//
// Pass the assembled SQL and arguments to a database query function via the
// top-level [Proceed]:
//
//	err := reqql.Proceed(ctx, q, filter, &results, db.QueryContext)
//
// For a fixed queryer + executor pair, wrap them with [NewProcessor] and
// call its [Processor.Proceed] method:
//
//	processor := reqql.NewProcessor(q, db.QueryContext)
//	err := processor.Proceed(ctx, filter, &results)
//
// [Processor.Proceed] is the recommended entry point in production code
// because it keeps the queryer and the database function together and
// eliminates repetition at call sites.
//
// # Domain Pattern
//
// In practice, domain packages wrap [Queryer] with enum-driven maps so
// that each valid input value selects a pre-built builder function:
//
//	type Status string
//
//	const (
//	    StatusAll    Status = "all"
//	    StatusActive Status = "active"
//	)
//
//	func buildWhereFuncs() map[Status]reqql.WhereFuncType[Filter] {
//	    return map[Status]reqql.WhereFuncType[Filter]{
//	        StatusAll:    func(_ *Filter) (string, []any, error) { return "TRUE", nil, nil },
//	        StatusActive: func(_ *Filter) (string, []any, error) { return "active = ?", []any{true}, nil },
//	    }
//	}
//
// This keeps SQL fragments close to the domain types and makes each clause
// independently testable.
package reqql
