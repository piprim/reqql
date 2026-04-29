package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/piprim/reqql/args/example/internal/reqql/products"
)

func TestInitDB_SeedCount(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}

	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count); err != nil {
		t.Fatalf("count products: %v", err)
	}

	if count != 12 {
		t.Errorf("expected 12 seeded products, got %d", count)
	}
}

func TestInitDB_CategoryCount(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}

	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count); err != nil {
		t.Fatalf("count categories: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 categories, got %d", count)
	}
}

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	db, err := initDB()
	if err != nil {
		t.Fatalf("initDB: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	sqlTmpl, err := template.ParseFS(sqlFS, "sql/products.sql.tmpl")
	if err != nil {
		t.Fatalf("parse sql template: %v", err)
	}

	qf := makeQueryFunc(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", serveIndex)
	mux.HandleFunc("POST /products", handleProducts(sqlTmpl, qf))

	return httptest.NewServer(mux)
}

func postFilter(t *testing.T, srv *httptest.Server, f products.Filter) []products.Product {
	t.Helper()

	body, _ := json.Marshal(f)

	resp, err := http.Post(srv.URL+"/products", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /products: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result []products.Product
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return result
}

func TestHandleProducts_AllFilters(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	result := postFilter(t, srv, products.Filter{
		CategoryName: products.CategoryNameAll,
		StockFilter:  products.StockFilterAll,
		SortOrder:    products.SortOrder{},
		LimitName:    products.LimitNameAll,
	})

	if len(result) != 12 {
		t.Errorf("expected 12 products, got %d", len(result))
	}
}

func TestHandleProducts_CategoryElectronics(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	result := postFilter(t, srv, products.Filter{
		CategoryName: products.CategoryNameElectronics,
		StockFilter:  products.StockFilterAll,
		SortOrder:    products.SortOrder{},
		LimitName:    products.LimitNameAll,
	})

	if len(result) != 4 {
		t.Errorf("expected 4 electronics products, got %d", len(result))
	}
}

func TestHandleProducts_InStock(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	// 3 out-of-stock: USB-C Hub, Winter Jacket, Coffee Beans → 9 in stock
	result := postFilter(t, srv, products.Filter{
		CategoryName: products.CategoryNameAll,
		StockFilter:  products.StockFilterInStock,
		SortOrder:    products.SortOrder{},
		LimitName:    products.LimitNameAll,
	})

	if len(result) != 9 {
		t.Errorf("expected 9 in-stock products, got %d", len(result))
	}
}

func TestHandleProducts_Limit5(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	result := postFilter(t, srv, products.Filter{
		CategoryName: products.CategoryNameAll,
		StockFilter:  products.StockFilterAll,
		SortOrder:    products.SortOrder{},
		LimitName:    products.LimitName5,
	})

	if len(result) != 5 {
		t.Errorf("expected 5 products, got %d", len(result))
	}
}

func TestHandleProducts_InvalidCategory(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	body := bytes.NewBufferString(`{"categoryName":"bogus","stockFilter":"all","sortOrder":"none","limitName":"all"}`)

	resp, err := http.Post(srv.URL+"/products", "application/json", body)
	if err != nil {
		t.Fatalf("POST /products: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleProducts_MalformedJSON(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	body := bytes.NewBufferString(`not json`)

	resp, err := http.Post(srv.URL+"/products", "application/json", body)
	if err != nil {
		t.Fatalf("POST /products: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}
