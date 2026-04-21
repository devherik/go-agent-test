package agent

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/genai"
)

type ToolFunc func(ctx context.Context, args map[string]any) (any, error)

type toolEntry struct {
	fn   ToolFunc
	decl *genai.FunctionDeclaration
}

type Registry struct {
	mu    sync.RWMutex
	tools map[string]toolEntry
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]toolEntry),
	}
}

func (r *Registry) Register(decl *genai.FunctionDeclaration, fn ToolFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[decl.Name] = toolEntry{
		fn:   fn,
		decl: decl,
	}
}

func (r *Registry) Call(ctx context.Context, name string, args map[string]any) (any, error) {
	r.mu.RLock()
	entry, ok := r.tools[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("tool %q not found", name)
	}
	return entry.fn(ctx, args)
}

func (r *Registry) Declarations() []*genai.FunctionDeclaration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	decls := make([]*genai.FunctionDeclaration, 0, len(r.tools))
	for _, entry := range r.tools {
		decls = append(decls, entry.decl)
	}
	return decls
}
