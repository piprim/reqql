package qevents

import (
	"context"

	"ovya.fr/csite/app/dao/sqltemplate"
	e "ovya.fr/csite/app/errors"
	m "ovya.fr/csite/app/models"
	"ovya.fr/csite/app/queryer"
)

func getQueryParseFuncs[IT any]() map[TypeName]queryer.QueryFuncType[IT] {
	return map[TypeName]queryer.QueryFuncType[IT]{
		TypeNameAll:      queryParseVisio[IT],
		TypeNameVisio:    queryParseVisio[IT],
		TypeNameNotVisio: queryParseVisio[IT],
	}
}

func getWhereFuncs[IT any]() map[FilterName]queryer.WhereFuncType[IT] {
	return map[FilterName]queryer.WhereFuncType[IT]{
		FilterNameAll:      queryer.WhereTrueFunc[IT],
		FilterNameNext:     whereNextFunc[IT],
		FilterNamePrevious: wherePreviousFunc[IT],
	}
}

func queryParseVisio[IT any](_ *IT) (sql string, arg any, err e.PrivateErrorI) {
	tmplService := sqltemplate.GetService()
	sql, err = tmplService.GetRaw("events/events", "visio")

	if err != nil {
		return "", nil, err
	}

	return sql, nil, nil
}

func whereNextFunc[T any](_ *T) (string, e.PrivateErrorI) {
	return "date_debut::date >= CURRENT_DATE", nil
}

func wherePreviousFunc[T any](_ *T) (string, e.PrivateErrorI) {
	return "date_debut::date < CURRENT_DATE", nil
}

// Get returns the events from the given filter.
func Proceed[QF queryer.QueryFunc, OT any](ctx context.Context, filter *Filter, dest *OT, qf QF) e.PrivateErrorI {
	q := queryer.New[m.Nil]()
	q.
		WithQueryParserFunc(getQueryParseFuncs[m.Nil]()[filter.TypeName]).
		WithWhereFunc(getWhereFuncs[m.Nil]()[filter.FilterName])

	p := queryer.NewProcessor(q, qf)
	err := p.Proceed(ctx, nil, dest)

	return err
}
