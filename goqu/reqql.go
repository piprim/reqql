package reqql

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/pkg/errors"
)

// QueryFuncArgs is a database executor whose positional arguments are passed
// as a variadic slice (e.g. sql.DB.QueryContext).
type QueryFuncArgs = func(ctx context.Context, dest any, query string, args ...any) error

// QueryFuncArg is a database executor that receives all arguments as a single
// any value (e.g. sqlx NamedQuery-style functions).
type QueryFuncArg = func(ctx context.Context, dest any, query string, arg any) error

// QueryFunc is the type constraint satisfied by [QueryFuncArgs] and [QueryFuncArg].
type QueryFunc interface{ QueryFuncArgs | QueryFuncArg }

// NoInputType is an alias for any, used when the queryer input type is untyped.
type NoInputType = any

// QueryFuncType is a builder function that returns the base *goqu.SelectDataset
// (the SELECT … FROM … part) and any error.  Call [goqu.SelectDataset.Prepared]
// on the returned dataset to get parameterized output.
type QueryFuncType[T any] func(*T) (*goqu.SelectDataset, error)

// WhereFuncType is a builder function that returns the WHERE predicate as a
// goqu.Expression.  Return nil for an unconditional query (no WHERE clause).
type WhereFuncType[T any] func(*T) goqu.Expression

// OrderFuncType is a builder function that returns the ORDER BY expression.
// Return nil for no ordering.
type OrderFuncType[T any] func(*T) exp.OrderedExpression

// LimitFuncType is a builder function that returns the LIMIT value as a uint,
// or nil for no limit.
type LimitFuncType[T any] func(*T) any

// OffsetFuncType is a builder function that returns the OFFSET value.
// Return 0 for no offset.
type OffsetFuncType[T any] func(*T) uint

// DefaultQueryParser is the default [QueryFuncType] used by [New].
// It returns a simple SELECT id, value FROM data dataset as a safe placeholder;
// real queryers should override it with [Queryer.WithQueryParserFunc].
func DefaultQueryParser[T any](_ *T) (*goqu.SelectDataset, error) {
	return goqu.From("data").Select("id", "value"), nil
}

// WhereTrueFunc is the default [WhereFuncType] used by [New].
// It returns nil, which goqu interprets as no WHERE clause (unconditional query).
func WhereTrueFunc[T any](_ *T) goqu.Expression {
	return nil
}

// NoOrderFunc is the default [OrderFuncType] used by [New].
// It returns nil, which goqu interprets as no ORDER BY clause.
func NoOrderFunc[T any](_ *T) exp.OrderedExpression {
	return nil
}

// NoLimitFunc is the default [LimitFuncType] used by [New].
// It returns nil, which goqu interprets as no LIMIT clause.
func NoLimitFunc[T any](_ *T) any {
	return nil
}

// NoOffsetFunc is the default [OffsetFuncType] used by [New].
// It returns 0, which goqu interprets as no OFFSET clause.
func NoOffsetFunc[T any](_ *T) uint {
	return 0
}

// Processor is the queryer processor useful to proceed the
// queryer with always the same querying function.
type Processor[InputType any, QueryFuncType QueryFunc] struct {
	queryer   *Queryer[InputType]
	queryFunc QueryFuncType
}

// Proceed is equivalent to calling the package-level [Proceed] with the
// queryer and query function that were supplied to [NewProcessor].
func (p *Processor[InputType, QueryFuncType]) Proceed(ctx context.Context, input *InputType, dest any) error {
	return Proceed(ctx, p.queryer, input, dest, p.queryFunc)
}

// NewProcessor returns a [Processor] that pairs q with f so that callers only
// need to supply the per-call input and destination when using
// [Processor.Proceed].
func NewProcessor[InputType any, QueryFuncType QueryFunc](q *Queryer[InputType], f QueryFuncType) *Processor[InputType, QueryFuncType] {
	p := new(Processor[InputType, QueryFuncType])
	p.queryer = q
	p.queryFunc = f

	return p
}

type Filter[InputType any] struct {
	whereFunc  WhereFuncType[InputType]
	orderFunc  OrderFuncType[InputType]
	limitFunc  LimitFuncType[InputType]
	offsetFunc OffsetFuncType[InputType]
}

type Queryer[InputType any] struct {
	queryFunc QueryFuncType[InputType]
	Filter[InputType]
	arg any
}

// WithQueryParserFunc set the function that provides the base dataset (SELECT ... FROM ...).
func (q *Queryer[InputType]) WithQueryParserFunc(f QueryFuncType[InputType]) *Queryer[InputType] {
	q.queryFunc = f
	return q
}

// WithWhereFunc sets the function that provides the "where" expressions.
func (q *Queryer[InputType]) WithWhereFunc(f WhereFuncType[InputType]) *Queryer[InputType] {
	q.whereFunc = f
	return q
}

// WithOrderFunc sets the function that provides the "order" expressions.
func (q *Queryer[InputType]) WithOrderFunc(f OrderFuncType[InputType]) *Queryer[InputType] {
	q.orderFunc = f
	return q
}

// WithLimitFunc sets the function that provides the "limit".
func (q *Queryer[InputType]) WithLimitFunc(f LimitFuncType[InputType]) *Queryer[InputType] {
	q.limitFunc = f
	return q
}

// WithOffsetFunc sets the function that provides the "offset".
func (q *Queryer[InputType]) WithOffsetFunc(f OffsetFuncType[InputType]) *Queryer[InputType] {
	q.offsetFunc = f
	return q
}

// SetArg sets a single opaque argument forwarded as-is to the query executor
// instead of the []any slice produced by goqu.  Use this when the executor
// expects a named-parameter struct.  Typically not needed with goqu since
// args are generated natively by [Queryer.Parse].
func (q *Queryer[InputType]) SetArg(arg any) *Queryer[InputType] {
	q.arg = arg
	return q
}

// Parse applies each of the five builder functions to build up the goqu
// dataset, then calls [goqu.SelectDataset.ToSQL] to produce the final SQL
// string and positional arguments.
func (q *Queryer[InputType]) Parse(input *InputType) (sql string, args []any, err error) {
	ds, err := q.queryFunc(input)
	if err != nil {
		return "", nil, err
	}

	if w := q.whereFunc(input); w != nil {
		ds = ds.Where(w)
	}

	if o := q.orderFunc(input); o != nil {
		ds = ds.Order(o)
	}

	if l := q.limitFunc(input); l != nil {
		if limit, ok := l.(uint); ok {
			ds = ds.Limit(limit)
		}
	}

	if offset := q.offsetFunc(input); offset > 0 {
		ds = ds.Offset(offset)
	}

	return ds.ToSQL()
}

// Proceed calls [Queryer.Parse] on q with the given input, then dispatches the
// resulting SQL and arguments to queryFunc.  queryFunc may be a [QueryFuncArgs]
// (variadic) or [QueryFuncArg] (single any).
func Proceed[InputType any, QF QueryFunc](
	ctx context.Context, q *Queryer[InputType],
	input *InputType, dest any, queryFunc QF) error {
	sql, args, err := q.Parse(input)
	if err != nil {
		return err
	}

	switch f := any(queryFunc).(type) {
	case QueryFuncArg:
		// If q.arg is set, we prefer it for QueryFuncArg (backward compat)
		arg := any(args)
		if q.arg != nil {
			arg = q.arg
		}
		return f(ctx, dest, sql, arg)
	case QueryFuncArgs:
		return f(ctx, dest, sql, args...)
	}

	return errors.New("unpredictible error (wtf)")
}

// New constructs a [Queryer] with all five builder functions set to safe
// no-ops: [DefaultQueryParser], [WhereTrueFunc], [NoOrderFunc], [NoLimitFunc],
// and [NoOffsetFunc].  Override any subset with the With* builder methods.
func New[InputType any]() *Queryer[InputType] {
	q := Queryer[InputType]{}
	q.
		WithQueryParserFunc(DefaultQueryParser).
		WithWhereFunc(WhereTrueFunc).
		WithOrderFunc(NoOrderFunc).
		WithLimitFunc(NoLimitFunc).
		WithOffsetFunc(NoOffsetFunc)

	return &q
}
