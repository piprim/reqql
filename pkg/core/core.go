package core

import "context"

type QueryFuncArgs = func(ctx context.Context, dest any, query string, args ...any) error
type QueryFuncArg = func(ctx context.Context, dest any, query string, arg any) error

type QueryFunc interface{ QueryFuncArgs | QueryFuncArg }
