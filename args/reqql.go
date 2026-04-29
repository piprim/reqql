package reqql

import (
	"context"
	"fmt"

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

// SelectFuncType is an alias for [QueryFuncType] kept for backward compatibility.
type SelectFuncType[T any] func(*T) (string, []any, error)

// QueryFuncType is a builder function that returns the SELECT … FROM … SQL
// fragment, its positional `?` arguments, and any error.
type QueryFuncType[T any] func(*T) (string, []any, error)

// FromFuncType is an alias for [QueryFuncType] kept for backward compatibility.
type FromFuncType[T any] func(*T) (string, []any, error)

// WhereFuncType is a builder function that returns the WHERE predicate SQL
// fragment, its positional `?` arguments, and any error.
type WhereFuncType[T any] func(*T) (string, []any, error)

// OrderFuncType is a builder function that returns the ORDER BY SQL fragment,
// its positional `?` arguments, and any error.
type OrderFuncType[T any] func(*T) (string, []any, error)

// LimitFuncType is a builder function that returns the LIMIT SQL fragment,
// its positional `?` arguments, and any error.
type LimitFuncType[T any] func(*T) (string, []any, error)

// OffsetFuncType is a builder function that returns the OFFSET SQL fragment,
// its positional `?` arguments, and any error.
type OffsetFuncType[T any] func(*T) (string, []any, error)

const (
	selectAll   = "*"
	fromDefault = "(VALUES (1, 'V1'), (2, 'V2')) AS data (id, value)"
	whereTrue   = "TRUE"
	noOrder     = "1"
	noLimit     = "ALL"
	noOffset    = "0"
	queryFormat = `
SELECT %s
FROM %s`
	sqlFormat = `%s
WHERE %s
ORDER BY %s
LIMIT %s
OFFSET %s`
)

// DefaultQueryParser is the default [QueryFuncType] used by [New].
// It returns "SELECT * FROM (VALUES …)" and is intended as a safe placeholder;
// real queryers should override it with [Queryer.WithQueryParserFunc].
func DefaultQueryParser[T any](_ *T) (sql string, args []any, err error) {
	return fmt.Sprintf(queryFormat, selectAll, fromDefault), nil, nil
}

// WhereTrueFunc is the default [WhereFuncType] used by [New].
// It returns "TRUE", which produces an unconditional WHERE clause.
func WhereTrueFunc[T any](_ *T) (string, []any, error) {
	return whereTrue, nil, nil
}

// NoOrderFunc is the default [OrderFuncType] used by [New].
// It returns "1" (constant), which is a no-op ORDER BY on all major databases.
func NoOrderFunc[T any](_ *T) (string, []any, error) {
	return noOrder, nil, nil
}

// NoLimitFunc is the default [LimitFuncType] used by [New].
// It returns "ALL", which disables the limit (standard SQL).
func NoLimitFunc[T any](_ *T) (string, []any, error) {
	return noLimit, nil, nil
}

// NoOffsetFunc is the default [OffsetFuncType] used by [New].
// It returns "0", which is a no-op offset.
func NoOffsetFunc[T any](_ *T) (string, []any, error) {
	return noOffset, nil, nil
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

// WithQueryParserFunc set the function that provides the "SELECT … FROM …" part of the queryer.
func (q *Queryer[InputType]) WithQueryParserFunc(f QueryFuncType[InputType]) *Queryer[InputType] {
	q.queryFunc = f

	return q
}

// WithWhereFunc set the function that provide the "where" part of the queryer.
func (q *Queryer[InputType]) WithWhereFunc(f WhereFuncType[InputType]) *Queryer[InputType] {
	q.whereFunc = f

	return q
}

// WithOrderFunc set the function that provide the "order" part of the queryer.
func (q *Queryer[InputType]) WithOrderFunc(f OrderFuncType[InputType]) *Queryer[InputType] {
	q.orderFunc = f

	return q
}

// WithLimitFunc set the function that provide the "limit" part of the queryer.
func (q *Queryer[InputType]) WithLimitFunc(f LimitFuncType[InputType]) *Queryer[InputType] {
	q.limitFunc = f

	return q
}

// WithOffsetFunc set the function that provide the "offset" part of the queryer.
func (q *Queryer[InputType]) WithOffsetFunc(f OffsetFuncType[InputType]) *Queryer[InputType] {
	q.offsetFunc = f

	return q
}

// SetArg sets a single opaque argument that is forwarded as-is to the query
// executor instead of the []any slice produced by the builder functions.
// Use this when the executor expects a named-parameter struct rather than
// positional arguments.  Calling SetArg is mutually exclusive with builder
// functions that return non-nil argument slices; [Proceed] returns an error
// if both are set simultaneously.
func (q *Queryer[InputType]) SetArg(arg any) *Queryer[InputType] {
	q.arg = arg

	return q
}

// Parse calls each of the five builder functions in order (query → where →
// order → limit → offset), concatenates the fragments into the final SQL, and
// collects all positional arguments in the same left-to-right order.
func (q *Queryer[InputType]) Parse(input *InputType) (sql string, args []any, err error) {
	query, qArgs, err := q.queryFunc(input)
	if err != nil {
		return "", nil, err
	}

	where, wArgs, err := q.whereFunc(input)
	if err != nil {
		return "", nil, err
	}

	order, oArgs, err := q.orderFunc(input)
	if err != nil {
		return "", nil, err
	}

	limit, lArgs, err := q.limitFunc(input)
	if err != nil {
		return "", nil, err
	}

	offset, offArgs, err := q.offsetFunc(input)
	if err != nil {
		return "", nil, err
	}

	// Concatenate all arguments
	args = append(args, qArgs...)
	args = append(args, wArgs...)
	args = append(args, oArgs...)
	args = append(args, lArgs...)
	args = append(args, offArgs...)

	sql = fmt.Sprintf(sqlFormat, query, where, order, limit, offset)

	return sql, args, nil
}

// Proceed calls [Queryer.Parse] on q with the given input, then dispatches the
// resulting SQL and arguments to queryFunc.  queryFunc may be a [QueryFuncArgs]
// (variadic) or [QueryFuncArg] (single any); [SetArg] controls which form is
// used for the single-arg variant.
func Proceed[InputType any, QF QueryFunc](
	ctx context.Context, q *Queryer[InputType],
	input *InputType, dest any, queryFunc QF) error {
	sql, args, err := q.Parse(input)
	if err != nil {
		return err
	}

	arg := any(args)
	if q.arg != nil {
		if len(args) > 0 {
			return errors.New("Parse query functions return not nil arguments while SetArg was used")
		}
		arg = q.arg
	}

	switch f := any(queryFunc).(type) {
	case QueryFuncArg:
		return f(ctx, dest, sql, arg)
	case QueryFuncArgs:
		switch a := arg.(type) {
		case []any:
			return f(ctx, dest, sql, a...)
		case nil:
			return f(ctx, dest, sql)
		default:
			return errors.Errorf("function need []any, %T given", a)
		}
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
