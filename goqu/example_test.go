package reqql_test

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	reqql "github.com/piprim/reqql/goqu"
)

func ExampleQueryer_Parse() {
	type ProductFilter struct {
		Category string
	}

	q := reqql.New[ProductFilter]()
	q.WithQueryParserFunc(func(_ *ProductFilter) (*goqu.SelectDataset, error) {
		return goqu.From("products").
			Select("id", "name", "price").
			Prepared(true), nil
	}).WithWhereFunc(func(f *ProductFilter) goqu.Expression {
		return goqu.C("category").Eq(f.Category)
	}).WithOrderFunc(func(_ *ProductFilter) []exp.OrderedExpression {
		return []exp.OrderedExpression{goqu.C("price").Asc()}
	})

	sql, args, err := q.Parse(&ProductFilter{Category: "books"})
	if err != nil {
		fmt.Printf("error: %v\n", err)

		return
	}

	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT "id", "name", "price" FROM "products" WHERE ("category" = ?) ORDER BY "price" ASC
	// [books]
}

func ExampleQueryer_Parse_multiWhere() {
	type Filter struct {
		Category string
		InStock  bool
	}

	q := reqql.New[Filter]()
	q.WithQueryParserFunc(func(_ *Filter) (*goqu.SelectDataset, error) {
		return goqu.From("products").Select("id", "name").Prepared(true), nil
	}).WithWhereFunc(func(f *Filter) goqu.Expression {
		return goqu.C("category").Eq(f.Category)
	}).WithWhereFunc(func(f *Filter) goqu.Expression {
		if !f.InStock {
			return nil
		}

		return goqu.C("stock").Gt(0)
	})

	sql, args, _ := q.Parse(&Filter{Category: "books", InStock: true})
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT "id", "name" FROM "products" WHERE (("category" = ?) AND ("stock" > ?))
	// [books 0]
}

func ExampleQueryer_Parse_multiOrder() {
	type Filter struct{}

	q := reqql.New[Filter]()
	q.WithQueryParserFunc(func(_ *Filter) (*goqu.SelectDataset, error) {
		return goqu.From("products").Select("id", "name", "price").Prepared(true), nil
	}).WithOrderFunc(func(_ *Filter) []exp.OrderedExpression {
		return []exp.OrderedExpression{
			goqu.C("category").Asc(),
			goqu.C("price").Desc(),
		}
	})

	sql, _, _ := q.Parse(&Filter{})
	fmt.Println(sql)
	// Output:
	// SELECT "id", "name", "price" FROM "products" ORDER BY "category" ASC, "price" DESC
}

func ExampleProceed() {
	type Filter struct {
		Status string
	}

	q := reqql.New[Filter]()
	q.WithQueryParserFunc(func(_ *Filter) (*goqu.SelectDataset, error) {
		return goqu.From("orders").Select("id").Prepared(true), nil
	}).WithWhereFunc(func(f *Filter) goqu.Expression {
		return goqu.C("status").Eq(f.Status)
	})

	var capturedSQL string
	mockDB := func(_ context.Context, _ any, query string, args ...any) error {
		capturedSQL = query

		return nil
	}

	_ = reqql.Proceed(context.Background(), q, &Filter{Status: "shipped"}, nil, mockDB)
	fmt.Println(capturedSQL)
	// Output:
	// SELECT "id" FROM "orders" WHERE ("status" = ?)
}

func ExampleNewProcessor() {
	type Filter struct {
		UserID int
	}

	q := reqql.New[Filter]()
	q.WithQueryParserFunc(func(_ *Filter) (*goqu.SelectDataset, error) {
		return goqu.From("items").Select("id").Prepared(true), nil
	}).WithWhereFunc(func(f *Filter) goqu.Expression {
		return goqu.C("user_id").Eq(f.UserID)
	})

	var capturedArgs []any
	mockDB := func(_ context.Context, _ any, query string, args ...any) error {
		capturedArgs = args

		return nil
	}

	proc := reqql.NewProcessor(q, mockDB)
	_ = proc.Proceed(context.Background(), &Filter{UserID: 42}, nil)
	fmt.Println(capturedArgs)
	// Output:
	// [42]
}
