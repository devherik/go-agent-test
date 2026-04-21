package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/devherik/go-agent-test/agent"
	"github.com/devherik/go-agent-test/infrastructure"

	"google.golang.org/genai"
)

// MockAIProvider allows us to simulate LLM behavior deterministically.
type MockAIProvider struct {
	responses []infrastructure.Message
	callCount int
}

func (m *MockAIProvider) Generate(ctx context.Context, history []infrastructure.Message, tools []*genai.FunctionDeclaration) (infrastructure.Message, infrastructure.RequestAnalysis, error) {
	if m.callCount >= len(m.responses) {
		return infrastructure.Message{}, infrastructure.RequestAnalysis{}, fmt.Errorf("no more responses mocked")
	}
	res := m.responses[m.callCount]
	m.callCount++
	return res, infrastructure.RequestAnalysis{Model: "mock-model"}, nil
}

func (m *MockAIProvider) EmbedText(ctx context.Context, text string) ([]float32, infrastructure.RequestAnalysis, error) {
	return []float32{0.1, 0.2}, infrastructure.RequestAnalysis{}, nil
}

func (m *MockAIProvider) GenerateStream(ctx context.Context, prompt string, systemInstruction string) (<-chan string, <-chan infrastructure.RequestAnalysis, <-chan error) {
	return nil, nil, nil
}

func (m *MockAIProvider) GenerateBlocking(ctx context.Context, prompt string, systemInstruction string) (string, infrastructure.RequestAnalysis, error) {
	return "mock response", infrastructure.RequestAnalysis{}, nil
}

func TestWeatherAgent_Run_Success(t *testing.T) {
	// 1. Setup Mock responses
	mock := &MockAIProvider{
		responses: []infrastructure.Message{
			// Turn 1: Model decides to call the weather tool
			{
				Role: "model",
				Parts: []infrastructure.Part{
					{
						Type:     infrastructure.PartTypeToolCall,
						ToolName: "get_weather",
						ToolArgs: map[string]any{"location": "London, UK"},
					},
				},
			},
			// Turn 2: Model provides final answer after receiving tool result
			{
				Role: "model",
				Parts: []infrastructure.Part{
					{
						Type: infrastructure.PartTypeText,
						Text: "The weather in London is currently 25°C and sunny.",
					},
				},
			},
		},
	}

	// 2. Initialize Agent with Mock
	a := agent.NewWeatherAgent(mock)

	// 3. Run
	result, err := a.Run(context.Background(), "How is the weather in London?")

	// 4. Assertions
	if err != nil {
		t.Fatalf("agent run failed: %v", err)
	}

	expected := "The weather in London is currently 25°C and sunny."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	if mock.callCount != 2 {
		t.Errorf("expected 2 calls to LLM, got %d", mock.callCount)
	}
}

func TestWeatherAgent_Run_ToolError(t *testing.T) {
	mock := &MockAIProvider{
		responses: []infrastructure.Message{
			{
				Role: "model",
				Parts: []infrastructure.Part{
					{
						Type:     infrastructure.PartTypeToolCall,
						ToolName: "non_existent_tool",
						ToolArgs: map[string]any{},
					},
				},
			},
		},
	}

	a := agent.NewWeatherAgent(mock)
	_, err := a.Run(context.Background(), "Trigger error")

	if err == nil {
		t.Error("expected error due to missing tool, got nil")
	}
}
