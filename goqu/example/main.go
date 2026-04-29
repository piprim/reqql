package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"

	reqql "github.com/piprim/reqql/goqu"
	"github.com/piprim/reqql/goqu/example/internal/reqql/products"
)

//go:embed templates/*
var templatesFS embed.FS

const port = ":8183"

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE categories (
			id   INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE
		);
		CREATE TABLE products (
			id          INTEGER PRIMARY KEY,
			name        TEXT    NOT NULL,
			category_id INTEGER NOT NULL REFERENCES categories(id),
			price       REAL    NOT NULL,
			stock       INTEGER NOT NULL DEFAULT 0
		);
	`); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	if err := seedData(db); err != nil {
		return nil, err
	}

	return db, nil
}

func seedData(db *sql.DB) error {
	cats := []string{"electronics", "clothing", "food"}
	catIDs := make(map[string]int64, len(cats))

	for _, name := range cats {
		res, err := db.Exec("INSERT INTO categories (name) VALUES (?)", name)
		if err != nil {
			return fmt.Errorf("insert category %q: %w", name, err)
		}

		id, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("last insert id for category %q: %w", name, err)
		}

		catIDs[name] = id
	}

	type row struct {
		name     string
		category string
		price    float64
		stock    int
	}

	rows := []row{
		{"Laptop Pro", "electronics", 1299.99, 5},
		{"USB-C Hub", "electronics", 49.99, 0},
		{"Smartphone X", "electronics", 899.99, 12},
		{"Wireless Mouse", "electronics", 39.99, 15},
		{"T-Shirt", "clothing", 19.99, 30},
		{"Winter Jacket", "clothing", 129.99, 0},
		{"Running Shoes", "clothing", 89.99, 8},
		{"Jeans", "clothing", 59.99, 20},
		{"Olive Oil", "food", 12.99, 50},
		{"Pasta", "food", 2.99, 100},
		{"Coffee Beans", "food", 24.99, 0},
		{"Green Tea", "food", 9.99, 45},
	}

	for _, r := range rows {
		if _, err := db.Exec(
			"INSERT INTO products (name, category_id, price, stock) VALUES (?, ?, ?, ?)",
			r.name, catIDs[r.category], r.price, r.stock,
		); err != nil {
			return fmt.Errorf("insert product %q: %w", r.name, err)
		}
	}

	return nil
}

func makeQueryFunc(db *sql.DB) reqql.QueryFuncArgs {
	return func(ctx context.Context, dest any, query string, args ...any) error {
		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("query: %w", err)
		}

		defer rows.Close()

		result, ok := dest.(*[]products.Product)
		if !ok {
			return fmt.Errorf("dest must be *[]products.Product, got %T", dest)
		}

		*result = (*result)[:0]

		for rows.Next() {
			var p products.Product
			if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
				return fmt.Errorf("scan product: %w", err)
			}

			*result = append(*result, p)
		}

		return rows.Err()
	}
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func serveIndex(w http.ResponseWriter, _ *http.Request) {
	data, err := templatesFS.ReadFile("templates/index.html")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func handleProducts(queryFunc reqql.QueryFuncArgs) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var filter products.Filter
		if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
			jsonError(w, "invalid request", http.StatusBadRequest)

			return
		}

		if !filter.CategoryName.IsValid() || !filter.StockFilter.IsValid() ||
			!filter.SortOrder.IsValid() || !filter.LimitName.IsValid() {
			jsonError(w, "invalid filter value", http.StatusBadRequest)

			return
		}

		result := make([]products.Product, 0)
		if err := products.Proceed(r.Context(), &filter, &result, queryFunc); err != nil {
			log.Printf("products.Proceed: %v", err)
			jsonError(w, "internal error", http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	defer db.Close()

	queryFunc := makeQueryFunc(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", serveIndex)
	mux.HandleFunc("POST /products", handleProducts(queryFunc))

	log.Println("listening on " + port)
	log.Println("open your browser at http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, mux))
}
