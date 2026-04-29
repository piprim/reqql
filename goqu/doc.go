// Package reqql assembles parameterized SQL SELECT queries from a typed input
// struct using a set of composable builder functions backed by the
// github.com/doug-martin/goqu/v9 SQL builder.
//
// # Overview
//
// A [Queryer] is constructed with [New] and configured by chaining five
// builder functions, each responsible for one SQL clause:
//
//   - [QueryFuncType] — returns a *goqu.SelectDataset (the base SELECT … FROM …)
//   - [WhereFuncType] — returns a goqu.Expression for the WHERE predicate
//   - [OrderFuncType] — returns an exp.OrderedExpression for ORDER BY
//   - [LimitFuncType] — returns a uint limit value (or nil for no limit)
//   - [OffsetFuncType] — returns a uint offset value
//
// Unlike the plain-SQL sibling package (github.com/piprim/reqql/args), SQL
// generation is delegated entirely to goqu's dialect engine.  All clause
// functions operate on goqu types; [Queryer.Parse] assembles the final dataset
// with the contributions of all five functions and calls [goqu.SelectDataset.ToSQL].
//
// Use [goqu.SelectDataset.Prepared] on the base dataset to get parameterized
// output (`?` placeholders) instead of inlined literals.
//
// # Dialect
//
// Select a dialect by calling [goqu.Dialect] and importing the corresponding
// blank-import driver (e.g. `_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"`).
// Without a blank import the default goqu dialect is used (ANSI double-quoted
// identifiers).
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
//	q.WithQueryParserFunc(func(_ *ProductFilter) (*goqu.SelectDataset, error) {
//	    return goqu.Dialect("sqlite3").
//	        From(goqu.T("products")).
//	        Select("id", "name", "price").
//	        Prepared(true), nil
//	}).WithWhereFunc(func(f *ProductFilter) goqu.Expression {
//	    return goqu.And(
//	        goqu.C("category").Eq(f.Category),
//	        goqu.C("price").Lte(f.MaxPrice),
//	    )
//	}).WithOrderFunc(func(_ *ProductFilter) exp.OrderedExpression {
//	    return goqu.C("price").Asc()
//	})
//
//	sql, args, err := q.Parse(&ProductFilter{Category: "books", MaxPrice: 29.99})
//	// sql:  `SELECT "id", "name", "price" FROM "products" WHERE (("category" = ?) AND ("price" <= ?)) ORDER BY "price" ASC`
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
//	        StatusAll: func(_ *Filter) goqu.Expression { return nil },
//	        StatusActive: func(_ *Filter) goqu.Expression {
//	            return goqu.C("active").IsTrue()
//	        },
//	    }
//	}
//
// This keeps query predicates close to the domain types and makes each clause
// independently testable.
package reqql
