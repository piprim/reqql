package products

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	"github.com/doug-martin/goqu/v9/exp"

	reqql "github.com/piprim/reqql/goqu"
)

var productColumns = []any{
	goqu.T("p").Col("id"),
	goqu.T("p").Col("name"),
	goqu.T("p").Col("price"),
	goqu.T("p").Col("stock"),
}

// colIdentifiers maps SortOrder.Cols values to their goqu column identifiers.
var colIdentifiers = map[string]exp.IdentifierExpression{
	"price":         goqu.T("p").Col("price"),
	"name":          goqu.T("p").Col("name"),
	"stock":         goqu.T("p").Col("stock"),
	"category.name": goqu.T("c").Col("name"),
}

// queryParseFunc builds the base dataset.
// A categories JOIN is added when a specific category is filtered OR when
// sorting by category.name (so that c.name is available in ORDER BY).
func queryParseFunc(f *Filter) (*goqu.SelectDataset, error) {
	ds := goqu.Dialect("sqlite3").
		From(goqu.T("products").As("p")).
		Select(productColumns...).
		Prepared(true)

	categoryJoin := goqu.T("p").Col("category_id").Eq(goqu.T("c").Col("id"))

	if f.CategoryName != CategoryNameAll {
		return ds.Join(
			goqu.T("categories").As("c"),
			goqu.On(goqu.And(
				categoryJoin,
				goqu.T("c").Col("name").Eq(string(f.CategoryName)),
			)),
		), nil
	}

	if f.SortOrder.Asc != nil && f.SortOrder.Cols == "category.name" {
		return ds.Join(
			goqu.T("categories").As("c"),
			goqu.On(categoryJoin),
		), nil
	}

	return ds, nil
}

func getWhereFuncs() map[StockFilter]reqql.WhereFuncType[Filter] {
	return map[StockFilter]reqql.WhereFuncType[Filter]{
		StockFilterAll:        func(_ *Filter) goqu.Expression { return nil },
		StockFilterInStock:    func(_ *Filter) goqu.Expression { return goqu.T("p").Col("stock").Gt(0) },
		StockFilterOutOfStock: func(_ *Filter) goqu.Expression { return goqu.T("p").Col("stock").Eq(0) },
	}
}

// orderFunc builds the ORDER BY expression dynamically from the filter's SortOrder.
func orderFunc(f *Filter) []exp.OrderedExpression {
	if f.SortOrder.Asc == nil {
		return nil
	}

	col, ok := colIdentifiers[f.SortOrder.Cols]
	if !ok {
		return nil
	}

	if *f.SortOrder.Asc {
		return []exp.OrderedExpression{col.Asc()}
	}

	return []exp.OrderedExpression{col.Desc()}
}

func getLimitFuncs() map[LimitName]reqql.LimitFuncType[Filter] {
	return map[LimitName]reqql.LimitFuncType[Filter]{
		LimitNameAll: func(_ *Filter) any { return nil },
		LimitName5:   func(_ *Filter) any { return uint(5) },
		LimitName10:  func(_ *Filter) any { return uint(10) },
		LimitName20:  func(_ *Filter) any { return uint(20) },
	}
}

// Proceed assembles a reqql.Queryer from filter, executes it via queryFunc,
// and stores results into dest (must be *[]Product).
func Proceed(ctx context.Context, filter *Filter, dest any, queryFunc reqql.QueryFuncArgs) error {
	if !filter.CategoryName.IsValid() {
		return fmt.Errorf("unknown category name: %s", filter.CategoryName)
	}

	wf, ok := getWhereFuncs()[filter.StockFilter]
	if !ok {
		return fmt.Errorf("unknown stock filter: %s", filter.StockFilter)
	}

	lf, ok := getLimitFuncs()[filter.LimitName]
	if !ok {
		return fmt.Errorf("unknown limit name: %s", filter.LimitName)
	}

	q := reqql.New[Filter]()
	q.WithQueryParserFunc(queryParseFunc).
		WithWhereFunc(wf).
		WithOrderFunc(orderFunc).
		WithLimitFunc(lf)

	p := reqql.NewProcessor(q, queryFunc)

	return p.Proceed(ctx, filter, dest)
}
