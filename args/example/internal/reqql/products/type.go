package products

// Product is the domain model returned by all queries.
type Product struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// Filter is the JSON request body for POST /products.
type Filter struct {
	CategoryName CategoryName `json:"categoryName"`
	StockFilter  StockFilter  `json:"stockFilter"`
	SortOrder    SortOrder    `json:"sortOrder"`
	LimitName    LimitName    `json:"limitName"`
}
