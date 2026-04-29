package products

import (
	"context"
	"strings"
	"testing"
	"text/template"
)

const testTmplSrc = `{{define "all"}}SELECT p.id, p.name, p.price, p.stock FROM products p{{end}}` +
	`{{define "withCategory"}}{{template "all" .}} JOIN categories c ON p.category_id = c.id AND c.name = ?{{end}}`

func mustParseTmpl(t *testing.T) *template.Template {
	t.Helper()

	tmpl, err := template.New("products.sql.tmpl").Parse(testTmplSrc)
	if err != nil {
		t.Fatalf("template parse: %v", err)
	}

	return tmpl
}

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
		SortOrder:    SortOrderNone,
		LimitName:    LimitNameAll,
	}

	if err := Proceed(context.Background(), tmpl, filter, nil, mockQF); err != nil {
		t.Fatalf("Proceed: %v", err)
	}

	if !strings.Contains(capturedSQL, "SELECT p.id, p.name, p.price, p.stock FROM products p") {
		t.Errorf("unexpected SQL: %s", capturedSQL)
	}

	if strings.Contains(capturedSQL, "JOIN") {
		t.Errorf("expected no JOIN for CategoryNameAll, got: %s", capturedSQL)
	}

	if len(capturedArgs) != 0 {
		t.Errorf("expected no args for CategoryNameAll, got: %v", capturedArgs)
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
		SortOrder:    SortOrderPriceAsc,
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
		t.Errorf("expected price ASC in ORDER BY, got: %s", capturedSQL)
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
		SortOrder:    SortOrderPriceDesc,
		LimitName:    LimitName5,
	}

	if err := Proceed(context.Background(), tmpl, filter, nil, mockQF); err != nil {
		t.Fatalf("Proceed: %v", err)
	}

	if !strings.Contains(capturedSQL, "p.stock = 0") {
		t.Errorf("expected stock = 0 in WHERE, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "p.price DESC") {
		t.Errorf("expected price DESC in ORDER BY, got: %s", capturedSQL)
	}

	if !strings.Contains(capturedSQL, "LIMIT 5") {
		t.Errorf("expected LIMIT 5, got: %s", capturedSQL)
	}
}
