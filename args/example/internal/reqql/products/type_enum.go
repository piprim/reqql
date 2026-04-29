package products

import (
	"errors"
	"fmt"
)

// CategoryName ---------------------------------------------------------------

type CategoryName string

const (
	CategoryNameAll         CategoryName = "all"
	CategoryNameElectronics CategoryName = "electronics"
	CategoryNameClothing    CategoryName = "clothing"
	CategoryNameFood        CategoryName = "food"
)

var ErrInvalidCategoryName = errors.New("not a valid CategoryName")

func ParseCategoryName(s string) (CategoryName, error) {
	switch CategoryName(s) {
	case CategoryNameAll, CategoryNameElectronics, CategoryNameClothing, CategoryNameFood:
		return CategoryName(s), nil
	}

	return "", fmt.Errorf("%s is %w", s, ErrInvalidCategoryName)
}

func (x CategoryName) IsValid() bool {
	_, err := ParseCategoryName(string(x))

	return err == nil
}

// StockFilter ----------------------------------------------------------------

type StockFilter string

const (
	StockFilterAll        StockFilter = "all"
	StockFilterInStock    StockFilter = "inStock"
	StockFilterOutOfStock StockFilter = "outOfStock"
)

var ErrInvalidStockFilter = errors.New("not a valid StockFilter")

func ParseStockFilter(s string) (StockFilter, error) {
	switch StockFilter(s) {
	case StockFilterAll, StockFilterInStock, StockFilterOutOfStock:
		return StockFilter(s), nil
	}

	return "", fmt.Errorf("%s is %w", s, ErrInvalidStockFilter)
}

func (x StockFilter) IsValid() bool {
	_, err := ParseStockFilter(string(x))

	return err == nil
}

// LimitName ------------------------------------------------------------------

type LimitName string

const (
	LimitNameAll LimitName = "all"
	LimitName5   LimitName = "5"
	LimitName10  LimitName = "10"
	LimitName20  LimitName = "20"
)

var ErrInvalidLimitName = errors.New("not a valid LimitName")

func ParseLimitName(s string) (LimitName, error) {
	switch LimitName(s) {
	case LimitNameAll, LimitName5, LimitName10, LimitName20:
		return LimitName(s), nil
	}

	return "", fmt.Errorf("%s is %w", s, ErrInvalidLimitName)
}

func (x LimitName) IsValid() bool {
	_, err := ParseLimitName(string(x))

	return err == nil
}
