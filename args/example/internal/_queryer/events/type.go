//go:generate go-enum --marshal --values

package qevents

// ENUM(all, visio, notVisio)
type TypeName string

// ENUM(all, next, previous)
type FilterName string

// ENUM(all, 1, 2, 3, 4, 5)
type LimitName string

// ENUM(none, last)
type OrderName string

type Filter struct {
	TypeName   TypeName   `json:"typeName"`
	FilterName FilterName `json:"filterName"`
	LimitName  LimitName  `json:"limitName"`
	OrderName  OrderName  `json:"orderName"`
}
