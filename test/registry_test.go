package test


import (
	"context"
	"testing"

	"go-agent-test/agent"

	"google.golang.org/genai"
)

func TestRegistry(t *testing.T) {
	r := agent.NewRegistry()

	decl := &genai.FunctionDeclaration{
		Name: "test_tool",
	}

	r.Register(decl, func(ctx context.Context, args map[string]any) (any, error) {
		return "ok", nil
	})

	res, err := r.Call(context.Background(), "test_tool", nil)
	if err != nil {
		t.Fatalf("failed to call tool: %v", err)
	}

	if res != "ok" {
		t.Errorf("expected ok, got %v", res)
	}

	decls := r.Declarations()
	if len(decls) != 1 || decls[0].Name != "test_tool" {
		t.Errorf("unexpected declarations: %v", decls)
	}
}
