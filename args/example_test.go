package reqql_test

import (
	"context"
	"fmt"

	reqql "github.com/piprim/reqql/args"
)

func ExampleQueryer_Parse() {
	type ProductFilter struct {
		Category string
		MaxPrice float64
	}

	q := reqql.New[ProductFilter]()
	q.WithQueryParserFunc(func(_ *ProductFilter) (string, []any, error) {
		return "SELECT id, name, price FROM products", nil, nil
	}).WithWhereFunc(func(f *ProductFilter) (string, []any, error) {
		return "category = ? AND price <= ?", []any{f.Category, f.MaxPrice}, nil
	}).WithOrderFunc(func(_ *ProductFilter) (string, []any, error) {
		return "price ASC", nil, nil
	})

	sql, args, err := q.Parse(&ProductFilter{Category: "books", MaxPrice: 29.99})
	if err != nil {
		fmt.Printf("error: %v\n", err)

		return
	}

	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT id, name, price FROM products
	// WHERE category = ? AND price <= ?
	// ORDER BY price ASC
	// LIMIT ALL
	// OFFSET 0
	// [books 29.99]
}

func ExampleProceed() {
	type Filter struct {
		Status string
	}

	q := reqql.New[Filter]()
	q.WithQueryParserFunc(func(_ *Filter) (string, []any, error) {
		return "SELECT id FROM orders", nil, nil
	}).WithWhereFunc(func(f *Filter) (string, []any, error) {
		return "status = ?", []any{f.Status}, nil
	})

	var capturedSQL string
	mockDB := func(_ context.Context, _ any, query string, args ...any) error {
		capturedSQL = query

		return nil
	}

	_ = reqql.Proceed(context.Background(), q, &Filter{Status: "shipped"}, nil, mockDB)
	fmt.Println(capturedSQL)
	// Output:
	// SELECT id FROM orders
	// WHERE status = ?
	// ORDER BY 1
	// LIMIT ALL
	// OFFSET 0
}

func ExampleNewProcessor() {
	type Filter struct {
		UserID int
	}

	q := reqql.New[Filter]()
	q.WithQueryParserFunc(func(_ *Filter) (string, []any, error) {
		return "SELECT id, name FROM items", nil, nil
	}).WithWhereFunc(func(f *Filter) (string, []any, error) {
		return "user_id = ?", []any{f.UserID}, nil
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
