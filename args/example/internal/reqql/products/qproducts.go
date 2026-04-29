package products

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	reqql "github.com/piprim/reqql/args"
)

// colSQL maps SortOrder.Cols values to their SQL column references.
var colSQL = map[string]string{
	"price":         "p.price",
	"name":          "p.name",
	"stock":         "p.stock",
	"category.name": "c.name",
}

func execTmpl(tmpl *template.Template, name string, args []any) reqql.QueryFuncType[Filter] {
	return func(_ *Filter) (string, []any, error) {
		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, name, nil); err != nil {
			return "", nil, fmt.Errorf("execute template %q: %w", name, err)
		}

		return buf.String(), args, nil
	}
}

// getQueryParseFuncs returns a QueryFuncType per CategoryName.
// For CategoryNameAll, the filter's SortOrder is consulted: when sorting by
// category.name a JOIN is required even without a category filter, so the
// "joinCategory" template is used instead of "all".
func getQueryParseFuncs(tmpl *template.Template) map[CategoryName]reqql.QueryFuncType[Filter] {
	return map[CategoryName]reqql.QueryFuncType[Filter]{
		CategoryNameAll: func(f *Filter) (string, []any, error) {
			name := "all"
			if f.SortOrder.Asc != nil && f.SortOrder.Cols == "category.name" {
				name = "joinCategory"
			}

			return execTmpl(tmpl, name, nil)(f)
		},
		CategoryNameElectronics: execTmpl(tmpl, "withCategory", []any{string(CategoryNameElectronics)}),
		CategoryNameClothing:    execTmpl(tmpl, "withCategory", []any{string(CategoryNameClothing)}),
		CategoryNameFood:        execTmpl(tmpl, "withCategory", []any{string(CategoryNameFood)}),
	}
}

func getWhereFuncs() map[StockFilter]reqql.WhereFuncType[Filter] {
	return map[StockFilter]reqql.WhereFuncType[Filter]{
		StockFilterAll:        func(_ *Filter) (string, []any, error) { return "TRUE", nil, nil },
		StockFilterInStock:    func(_ *Filter) (string, []any, error) { return "p.stock > 0", nil, nil },
		StockFilterOutOfStock: func(_ *Filter) (string, []any, error) { return "p.stock = 0", nil, nil },
	}
}

// orderFunc builds the ORDER BY clause dynamically from the filter's SortOrder.
func orderFunc(f *Filter) (string, []any, error) {
	if f.SortOrder.Asc == nil {
		return "1", nil, nil
	}

	col, ok := colSQL[f.SortOrder.Cols]
	if !ok {
		return "", nil, fmt.Errorf("unknown sort column: %s", f.SortOrder.Cols)
	}

	if *f.SortOrder.Asc {
		return col + " ASC", nil, nil
	}

	return col + " DESC", nil, nil
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

	lf, ok := getLimitFuncs()[filter.LimitName]
	if !ok {
		return fmt.Errorf("unknown limit name: %s", filter.LimitName)
	}

	q := reqql.New[Filter]()
	q.WithQueryParserFunc(qf).
		WithWhereFunc(wf).
		WithOrderFunc(orderFunc).
		WithLimitFunc(lf)

	p := reqql.NewProcessor(q, queryFunc)

	return p.Proceed(ctx, filter, dest)
}
