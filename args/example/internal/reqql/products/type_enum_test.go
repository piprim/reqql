package products

import (
	"testing"
)

func TestParseCategoryName(t *testing.T) {
	for _, v := range []string{"all", "electronics", "clothing", "food"} {
		if _, err := ParseCategoryName(v); err != nil {
			t.Errorf("ParseCategoryName(%q) unexpected error: %v", v, err)
		}
	}

	if _, err := ParseCategoryName("invalid"); err == nil {
		t.Error("ParseCategoryName(\"invalid\") expected error, got nil")
	}
}

func TestCategoryNameIsValid(t *testing.T) {
	if !CategoryNameAll.IsValid() {
		t.Error("CategoryNameAll.IsValid() = false, want true")
	}

	if CategoryName("bogus").IsValid() {
		t.Error("CategoryName(\"bogus\").IsValid() = true, want false")
	}
}

func TestParseStockFilter(t *testing.T) {
	for _, v := range []string{"all", "inStock", "outOfStock"} {
		if _, err := ParseStockFilter(v); err != nil {
			t.Errorf("ParseStockFilter(%q) unexpected error: %v", v, err)
		}
	}

	if _, err := ParseStockFilter("invalid"); err == nil {
		t.Error("ParseStockFilter(\"invalid\") expected error, got nil")
	}
}

func TestParseSortOrder(t *testing.T) {
	for _, v := range []string{"none", "priceAsc", "priceDesc"} {
		if _, err := ParseSortOrder(v); err != nil {
			t.Errorf("ParseSortOrder(%q) unexpected error: %v", v, err)
		}
	}

	if _, err := ParseSortOrder("invalid"); err == nil {
		t.Error("ParseSortOrder(\"invalid\") expected error, got nil")
	}
}

func TestParseLimitName(t *testing.T) {
	for _, v := range []string{"all", "5", "10", "20"} {
		if _, err := ParseLimitName(v); err != nil {
			t.Errorf("ParseLimitName(%q) unexpected error: %v", v, err)
		}
	}

	if _, err := ParseLimitName("invalid"); err == nil {
		t.Error("ParseLimitName(\"invalid\") expected error, got nil")
	}
}
