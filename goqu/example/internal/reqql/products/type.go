package products

// Product is the domain model returned by all queries.
type Product struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// SortOrder specifies dynamic ordering. Set Asc to nil to skip sorting.
// Cols may be "price", "name", "stock", or "category.name".
type SortOrder struct {
	Cols string `json:"cols"`
	Asc  *bool  `json:"asc"`
}

var validSortCols = map[string]bool{
	"price":         true,
	"name":          true,
	"stock":         true,
	"category.name": true,
}

func (s SortOrder) IsValid() bool {
	if s.Asc == nil {
		return true
	}

	return validSortCols[s.Cols]
}

// Filter is the JSON request body for POST /products.
type Filter struct {
	CategoryName CategoryName `json:"categoryName"`
	StockFilter  StockFilter  `json:"stockFilter"`
	SortOrder    SortOrder    `json:"sortOrder"`
	LimitName    LimitName    `json:"limitName"`
}
