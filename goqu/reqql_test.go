package reqql

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
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
	// goqu default dialect might use double quotes or different casing depending on dialect.
	// Default is "default" which uses double quotes.

	if sql != expectedSQL {
		t.Errorf("Expected SQL:\n%s\nGot:\n%s", expectedSQL, sql)
	}

	// With default dialect and literal values, args might be empty if goqu inlines them.
	// Actually, goqu.Ex usually inlines if not using Prepared(true).
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
