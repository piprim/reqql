package products

import (
	"context"
	"strings"
	"testing"
	"text/template"
)

const testTmplSrc = `{{define "all"}}SELECT p.id, p.name, p.price, p.stock FROM products p{{end}}` +
	`{{define "joinCategory"}}{{template "all" .}} JOIN categories c ON p.category_id = c.id{{end}}` +
	`{{define "withCategory"}}{{template "joinCategory" .}} AND c.name = ?{{end}}`

func mustParseTmpl(t *testing.T) *template.Template {
	t.Helper()

	tmpl, err := template.New("products.sql.tmpl").Parse(testTmplSrc)
	if err != nil {
		t.Fatalf("template parse: %v", err)
	}

	return tmpl
}

func boolPtr(b bool) *bool { return &b }

func TestProceed_AllProducts(t *testing.T) {
	tmpl := mustParseTmpl(t)

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

	if err := Proceed(context.Background(), tmpl, filter, nil, mockQF); err != nil {
		t.Fatalf("Proceed: %v", err)
	}

	if !strings.Contains(capturedSQL, "SELECT p.id, p.name, p.price, p.stock FROM products p") {
		t.Errorf("unexpected SQL: %s", capturedSQL)
	}

	if strings.Contains(capturedSQL, "JOIN") {
		t.Errorf("expected no JOIN for CategoryNameAll without category.name sort, got: %s", capturedSQL)
	}

	if len(capturedArgs) != 0 {
		t.Errorf("expected no args, got: %v", capturedArgs)
	}
}

func TestProceed_WithCategory(t *testing.T) {
	tmpl := mustParseTmpl(t)

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

	if err := Proceed(context.Background(), tmpl, filter, nil, mockQF); err != nil {
		t.Fatalf("Proceed: %v", err)
	}

	if !strings.Contains(capturedSQL, "JOIN categories c") {
		t.Errorf("expected JOIN for CategoryNameElectronics, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "p.stock > 0") {
		t.Errorf("expected stock > 0 in WHERE, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "p.price ASC") {
		t.Errorf("expected p.price ASC in ORDER BY, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "LIMIT 10") {
		t.Errorf("expected LIMIT 10, got: %s", capturedSQL)
	}

	if len(capturedArgs) != 1 || capturedArgs[0] != "electronics" {
		t.Errorf("expected args [\"electronics\"], got: %v", capturedArgs)
	}
}

func TestProceed_StockOutOfStock(t *testing.T) {
	tmpl := mustParseTmpl(t)

	var capturedSQL string

	mockQF := func(_ context.Context, _ any, query string, args ...any) error {
		capturedSQL = query

		return nil
	}

	filter := &Filter{
		CategoryName: CategoryNameAll,
		StockFilter:  StockFilterOutOfStock,
		SortOrder:    SortOrder{Cols: "price", Asc: boolPtr(false)},
		LimitName:    LimitName5,
	}

	if err := Proceed(context.Background(), tmpl, filter, nil, mockQF); err != nil {
		t.Fatalf("Proceed: %v", err)
	}

	if !strings.Contains(capturedSQL, "p.stock = 0") {
		t.Errorf("expected stock = 0 in WHERE, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "p.price DESC") {
		t.Errorf("expected p.price DESC in ORDER BY, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "LIMIT 5") {
		t.Errorf("expected LIMIT 5, got: %s", capturedSQL)
	}
}

func TestProceed_SortByCategoryName(t *testing.T) {
	tmpl := mustParseTmpl(t)

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

	if err := Proceed(context.Background(), tmpl, filter, nil, mockQF); err != nil {
		t.Fatalf("Proceed: %v", err)
	}

	if !strings.Contains(capturedSQL, "JOIN categories c") {
		t.Errorf("expected JOIN for category.name sort, got: %s", capturedSQL)
	}

	if strings.Contains(capturedSQL, "c.name = ?") {
		t.Errorf("expected no category filter in ON clause for CategoryNameAll, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "c.name ASC") {
		t.Errorf("expected c.name ASC in ORDER BY, got: %s", capturedSQL)
	}

	if len(capturedArgs) != 0 {
		t.Errorf("expected no args for CategoryNameAll + category.name sort, got: %v", capturedArgs)
	}
}
