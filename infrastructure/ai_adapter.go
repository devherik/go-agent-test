package infrastructure

import (
	"context"

	"google.golang.org/genai"
)

// LLM Interaction
type RequestAnalysis struct {
	Model              string `json:"model" binding:"required"`
	PromptTokens       int    `json:"prompt_tokens" binding:"required"`
	CompletationTokens int    `json:"completion_tokens" binding:"required"`
	TotalTokens        int    `json:"total_tokens" binding:"required"`
}

type PartType string

const (
	PartTypeText         PartType = "text"
	PartTypeToolCall     PartType = "tool_call"
	PartTypeToolResponse PartType = "tool_response"
)

type Part struct {
	Type PartType
	Text string
	// ToolCall data
	ToolName string
	ToolArgs map[string]any
	// ToolResponse data
	ToolResult any
}

type Message struct {
	Role  string
	Parts []Part
}

type LLMResponse struct {
	Content  string          `json:"content" binding:"required"`
	Analysis RequestAnalysis `json:"analysis" binding:"required"`
}

// AIProvider is a composite interface for full AI capabilities.
type AIProvider interface {
	LLMGenerator
	Embedder
}

// Embedder defines the contract for turning text into vectors.
type Embedder interface {
	EmbedText(ctx context.Context, text string) ([]float32, RequestAnalysis, error)
}

// LLMGenerator defines the contract for text generation.
type LLMGenerator interface {
	// Generate handles agentic workflows with history and tools.
	Generate(ctx context.Context, history []Message, tools []*genai.FunctionDeclaration) (Message, RequestAnalysis, error)

	// GenerateStream creates a response based on a prompt (legacy/simple version).
	GenerateStream(ctx context.Context, prompt string, systemInstruction string) (<-chan string, <-chan RequestAnalysis, <-chan error)

	// GenerateBlocking is a simple non-streaming version for background tasks.
	GenerateBlocking(ctx context.Context, prompt string, systemInstruction string) (string, RequestAnalysis, error)
}
