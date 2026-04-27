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
