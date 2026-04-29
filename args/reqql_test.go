package reqql

import (
	"context"
	"testing"
)

type MyInput struct {
	ID   int
	Name string
}

func TestQueryer_Parse_Parameterized(t *testing.T) {
	q := New[MyInput]()
	q.WithQueryParserFunc(func(i *MyInput) (string, []any, error) {
		return "SELECT id, name FROM users", nil, nil
	})
	q.WithWhereFunc(func(i *MyInput) (string, []any, error) {
		return "id = ? AND name = ?", []any{i.ID, i.Name}, nil
	})

	input := &MyInput{ID: 42, Name: "Alice"}
	sql, args, err := q.Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedSQL := "SELECT id, name FROM users\nWHERE id = ? AND name = ?\nORDER BY 1\nLIMIT ALL\nOFFSET 0"
	if sql != expectedSQL {
		t.Errorf("Expected SQL:\n%q\nGot:\n%q", expectedSQL, sql)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	} else {
		if args[0] != 42 || args[1] != "Alice" {
			t.Errorf("Expected args [42, Alice], got %v", args)
		}
	}
}

func TestQueryer_Parse_MultiWhere(t *testing.T) {
	q := New[MyInput]()
	q.WithQueryParserFunc(func(_ *MyInput) (string, []any, error) {
		return "SELECT id FROM users", nil, nil
	})
	q.WithWhereFunc(func(i *MyInput) (string, []any, error) {
		return "id = ?", []any{i.ID}, nil
	})
	q.WithWhereFunc(func(i *MyInput) (string, []any, error) {
		return "name = ?", []any{i.Name}, nil
	})

	sql, args, err := q.Parse(&MyInput{ID: 7, Name: "Bob"})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedSQL := "SELECT id FROM users\nWHERE (id = ?) AND (name = ?)\nORDER BY 1\nLIMIT ALL\nOFFSET 0"
	if sql != expectedSQL {
		t.Errorf("Expected SQL:\n%q\nGot:\n%q", expectedSQL, sql)
	}

	if len(args) != 2 || args[0] != 7 || args[1] != "Bob" {
		t.Errorf("Expected args [7, Bob], got %v", args)
	}
}

func TestQueryer_Parse_TrueFuncSkipped(t *testing.T) {
	q := New[MyInput]()
	q.WithQueryParserFunc(func(_ *MyInput) (string, []any, error) {
		return "SELECT id FROM users", nil, nil
	})
	q.WithWhereFunc(WhereTrueFunc[MyInput])
	q.WithWhereFunc(func(i *MyInput) (string, []any, error) {
		return "id = ?", []any{i.ID}, nil
	})

	sql, _, err := q.Parse(&MyInput{ID: 1})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// WhereTrueFunc should be skipped; only the real predicate remains (no parens).
	expectedSQL := "SELECT id FROM users\nWHERE id = ?\nORDER BY 1\nLIMIT ALL\nOFFSET 0"
	if sql != expectedSQL {
		t.Errorf("Expected SQL:\n%q\nGot:\n%q", expectedSQL, sql)
	}
}

func TestQueryer_Parse_NoWhereDefaultsToTrue(t *testing.T) {
	q := New[MyInput]()
	q.WithQueryParserFunc(func(_ *MyInput) (string, []any, error) {
		return "SELECT id FROM users", nil, nil
	})

	sql, _, err := q.Parse(&MyInput{})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedSQL := "SELECT id FROM users\nWHERE TRUE\nORDER BY 1\nLIMIT ALL\nOFFSET 0"
	if sql != expectedSQL {
		t.Errorf("Expected SQL:\n%q\nGot:\n%q", expectedSQL, sql)
	}
}

func TestProceed_Parameterized(t *testing.T) {
	q := New[MyInput]()
	q.WithWhereFunc(func(i *MyInput) (string, []any, error) {
		return "id = ?", []any{i.ID}, nil
	})

	var called bool
	var capturedArgs []any

	mockQueryFunc := func(ctx context.Context, dest any, query string, args ...any) error {
		called = true
		capturedArgs = args

		return nil
	}

	err := Proceed(context.Background(), q, &MyInput{ID: 123}, nil, mockQueryFunc)
	if err != nil {
		t.Fatalf("Proceed failed: %v", err)
	}

	if !called {
		t.Error("mockQueryFunc was not called")
	}

	if len(capturedArgs) != 1 || capturedArgs[0] != 123 {
		t.Errorf("Expected capturedArgs [123], got %v", capturedArgs)
	}
}
