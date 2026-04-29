package reqql

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type MyInput struct {
	ID   int
	Name string
}

func TestQueryer_Parse(t *testing.T) {
	q := New[MyInput]()
	q.WithQueryParserFunc(func(i *MyInput) (*goqu.SelectDataset, error) {
		return goqu.From("users").Select("id", "name"), nil
	})
	q.WithWhereFunc(func(i *MyInput) goqu.Expression {
		return goqu.Ex{"id": i.ID, "name": i.Name}
	})

	input := &MyInput{ID: 1, Name: "John"}
	sql, args, err := q.Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedSQL := `SELECT "id", "name" FROM "users" WHERE (("id" = 1) AND ("name" = 'John'))`
	if sql != expectedSQL {
		t.Errorf("Expected SQL:\n%s\nGot:\n%s", expectedSQL, sql)
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 args with default inlining, got %d", len(args))
	}
}

func TestQueryer_Prepared(t *testing.T) {
	q := New[MyInput]()
	q.WithQueryParserFunc(func(i *MyInput) (*goqu.SelectDataset, error) {
		return goqu.From("users").Select("id", "name").Prepared(true), nil
	})
	q.WithWhereFunc(func(i *MyInput) goqu.Expression {
		return goqu.Ex{"id": i.ID}
	})

	input := &MyInput{ID: 42}
	sql, args, err := q.Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedSQL := `SELECT "id", "name" FROM "users" WHERE ("id" = ?)`
	if sql != expectedSQL {
		t.Errorf("Expected SQL:\n%s\nGot:\n%s", expectedSQL, sql)
	}

	if len(args) != 1 || args[0] != int64(42) {
		t.Errorf("Expected arg [42] (int64), got %v (%T)", args, args[0])
	}
}

func TestQueryer_Parse_MultiWhere(t *testing.T) {
	q := New[MyInput]()
	q.WithQueryParserFunc(func(_ *MyInput) (*goqu.SelectDataset, error) {
		return goqu.From("users").Select("id").Prepared(true), nil
	})
	q.WithWhereFunc(func(i *MyInput) goqu.Expression {
		return goqu.C("id").Eq(i.ID)
	})
	q.WithWhereFunc(func(i *MyInput) goqu.Expression {
		return goqu.C("name").Eq(i.Name)
	})

	sql, args, err := q.Parse(&MyInput{ID: 7, Name: "Bob"})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedSQL := `SELECT "id" FROM "users" WHERE (("id" = ?) AND ("name" = ?))`
	if sql != expectedSQL {
		t.Errorf("Expected SQL:\n%s\nGot:\n%s", expectedSQL, sql)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d: %v", len(args), args)
	}
}

func TestQueryer_MultiColumnOrder(t *testing.T) {
	q := New[MyInput]()
	q.WithQueryParserFunc(func(_ *MyInput) (*goqu.SelectDataset, error) {
		return goqu.From("users").Select("id", "name").Prepared(true), nil
	})
	q.WithOrderFunc(func(_ *MyInput) []exp.OrderedExpression {
		return []exp.OrderedExpression{
			goqu.C("name").Asc(),
			goqu.C("id").Desc(),
		}
	})

	sql, _, err := q.Parse(&MyInput{})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedSQL := `SELECT "id", "name" FROM "users" ORDER BY "name" ASC, "id" DESC`
	if sql != expectedSQL {
		t.Errorf("Expected SQL:\n%s\nGot:\n%s", expectedSQL, sql)
	}
}
