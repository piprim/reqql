package products

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	reqql "github.com/piprim/reqql/args"
)

func getQueryParseFuncs(tmpl *template.Template) map[CategoryName]reqql.QueryFuncType[Filter] {
	execTmpl := func(name string, args []any) reqql.QueryFuncType[Filter] {
		return func(_ *Filter) (string, []any, error) {
			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, name, nil); err != nil {
				return "", nil, fmt.Errorf("execute template %q: %w", name, err)
			}

			return buf.String(), args, nil
		}
	}

	return map[CategoryName]reqql.QueryFuncType[Filter]{
		CategoryNameAll:         execTmpl("all", nil),
		CategoryNameElectronics: execTmpl("withCategory", []any{string(CategoryNameElectronics)}),
		CategoryNameClothing:    execTmpl("withCategory", []any{string(CategoryNameClothing)}),
		CategoryNameFood:        execTmpl("withCategory", []any{string(CategoryNameFood)}),
	}
}

func getWhereFuncs() map[StockFilter]reqql.WhereFuncType[Filter] {
	return map[StockFilter]reqql.WhereFuncType[Filter]{
		StockFilterAll:        func(_ *Filter) (string, []any, error) { return "TRUE", nil, nil },
		StockFilterInStock:    func(_ *Filter) (string, []any, error) { return "p.stock > 0", nil, nil },
		StockFilterOutOfStock: func(_ *Filter) (string, []any, error) { return "p.stock = 0", nil, nil },
	}
}

func getOrderFuncs() map[SortOrder]reqql.OrderFuncType[Filter] {
	return map[SortOrder]reqql.OrderFuncType[Filter]{
		SortOrderNone:      func(_ *Filter) (string, []any, error) { return "1", nil, nil },
		SortOrderPriceAsc:  func(_ *Filter) (string, []any, error) { return "p.price ASC", nil, nil },
		SortOrderPriceDesc: func(_ *Filter) (string, []any, error) { return "p.price DESC", nil, nil },
	}
}

func getLimitFuncs() map[LimitName]reqql.LimitFuncType[Filter] {
	return map[LimitName]reqql.LimitFuncType[Filter]{
		LimitNameAll: func(_ *Filter) (string, []any, error) { return "-1", nil, nil },
		LimitName5:   func(_ *Filter) (string, []any, error) { return "5", nil, nil },
		LimitName10:  func(_ *Filter) (string, []any, error) { return "10", nil, nil },
		LimitName20:  func(_ *Filter) (string, []any, error) { return "20", nil, nil },
	}
}

// Proceed assembles a reqql.Queryer from filter, executes it via queryFunc,
// and stores results into dest (must be *[]Product).
func Proceed(ctx context.Context, tmpl *template.Template, filter *Filter, dest any, queryFunc reqql.QueryFuncArgs) error {
	qf, ok := getQueryParseFuncs(tmpl)[filter.CategoryName]
	if !ok {
		return fmt.Errorf("unknown category name: %s", filter.CategoryName)
	}

	wf, ok := getWhereFuncs()[filter.StockFilter]
	if !ok {
		return fmt.Errorf("unknown stock filter: %s", filter.StockFilter)
	}

	of, ok := getOrderFuncs()[filter.SortOrder]
	if !ok {
		return fmt.Errorf("unknown sort order: %s", filter.SortOrder)
	}

	lf, ok := getLimitFuncs()[filter.LimitName]
	if !ok {
		return fmt.Errorf("unknown limit name: %s", filter.LimitName)
	}

	q := reqql.New[Filter]()
	q.WithQueryParserFunc(qf).
		WithWhereFunc(wf).
		WithOrderFunc(of).
		WithLimitFunc(lf)

	p := reqql.NewProcessor(q, queryFunc)

	return p.Proceed(ctx, filter, dest)
}
