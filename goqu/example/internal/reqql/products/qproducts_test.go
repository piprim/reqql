package products

import (
	"context"
	"strings"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestProceed_AllProducts(t *testing.T) {
	var capturedSQL string

	var capturedArgs []any

	mockQF := func(_ context.Context, _ any, query string, args ...any) error {
		capturedSQL = query
		capturedArgs = args

		return nil
	}

	filter := &Filter{
		CategoryName: CategoryNameAll,
		StockFilter:  StockFilterAll,
		SortOrder:    SortOrder{Asc: nil},
		LimitName:    LimitNameAll,
	}

	if err := Proceed(context.Background(), filter, nil, mockQF); err != nil {
		t.Fatalf("Proceed: %v", err)
	}

	if !strings.Contains(capturedSQL, "products") {
		t.Errorf("expected products table in SQL, got: %s", capturedSQL)
	}

	if strings.Contains(capturedSQL, "JOIN") {
		t.Errorf("expected no JOIN for CategoryNameAll without category.name sort, got: %s", capturedSQL)
	}

	if strings.Contains(capturedSQL, "WHERE") {
		t.Errorf("expected no WHERE for StockFilterAll, got: %s", capturedSQL)
	}

	if len(capturedArgs) != 0 {
		t.Errorf("expected no args, got: %v", capturedArgs)
	}
}

func TestProceed_WithCategory(t *testing.T) {
	var capturedSQL string

	var capturedArgs []any

	mockQF := func(_ context.Context, _ any, query string, args ...any) error {
		capturedSQL = query
		capturedArgs = args

		return nil
	}

	filter := &Filter{
		CategoryName: CategoryNameElectronics,
		StockFilter:  StockFilterInStock,
		SortOrder:    SortOrder{Cols: "price", Asc: boolPtr(true)},
		LimitName:    LimitName10,
	}

	if err := Proceed(context.Background(), filter, nil, mockQF); err != nil {
		t.Fatalf("Proceed: %v", err)
	}

	if !strings.Contains(capturedSQL, "JOIN") {
		t.Errorf("expected JOIN for CategoryNameElectronics, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "categories") {
		t.Errorf("expected categories table in JOIN, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "stock") {
		t.Errorf("expected stock filter in WHERE, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "price` ASC") {
		t.Errorf("expected price ASC in ORDER BY, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "LIMIT") {
		t.Errorf("expected LIMIT clause, got: %s", capturedSQL)
	}

	if len(capturedArgs) == 0 || capturedArgs[0] != "electronics" {
		t.Errorf("expected first arg \"electronics\", got: %v", capturedArgs)
	}
}

func TestProceed_StockOutOfStock(t *testing.T) {
	var capturedSQL string

	var capturedArgs []any

	mockQF := func(_ context.Context, _ any, query string, args ...any) error {
		capturedSQL = query
		capturedArgs = args

		return nil
	}

	filter := &Filter{
		CategoryName: CategoryNameAll,
		StockFilter:  StockFilterOutOfStock,
		SortOrder:    SortOrder{Cols: "price", Asc: boolPtr(false)},
		LimitName:    LimitName5,
	}

	if err := Proceed(context.Background(), filter, nil, mockQF); err != nil {
		t.Fatalf("Proceed: %v", err)
	}

	if !strings.Contains(capturedSQL, "stock") {
		t.Errorf("expected stock filter in WHERE, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "price` DESC") {
		t.Errorf("expected price DESC in ORDER BY, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "LIMIT") {
		t.Errorf("expected LIMIT clause, got: %s", capturedSQL)
	}

	if len(capturedArgs) == 0 {
		t.Errorf("expected args for stock filter + limit, got none")
	}
}

func TestProceed_SortByCategoryName(t *testing.T) {
	var capturedSQL string

	var capturedArgs []any

	mockQF := func(_ context.Context, _ any, query string, args ...any) error {
		capturedSQL = query
		capturedArgs = args

		return nil
	}

	// CategoryNameAll + sort by category.name: needs JOIN but no category filter arg.
	filter := &Filter{
		CategoryName: CategoryNameAll,
		StockFilter:  StockFilterAll,
		SortOrder:    SortOrder{Cols: "category.name", Asc: boolPtr(true)},
		LimitName:    LimitNameAll,
	}

	if err := Proceed(context.Background(), filter, nil, mockQF); err != nil {
		t.Fatalf("Proceed: %v", err)
	}

	if !strings.Contains(capturedSQL, "JOIN") {
		t.Errorf("expected JOIN for category.name sort, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "categories") {
		t.Errorf("expected categories table in JOIN, got: %s", capturedSQL)
	}

	// No category filter value in the ON condition.
	if strings.Contains(capturedSQL, "c`.`name` = ?") {
		t.Errorf("expected no category filter in ON clause for CategoryNameAll, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "name` ASC") {
		t.Errorf("expected c.name ASC in ORDER BY, got: %s", capturedSQL)
	}

	if len(capturedArgs) != 0 {
		t.Errorf("expected no args for CategoryNameAll + category.name sort, got: %v", capturedArgs)
	}
}
