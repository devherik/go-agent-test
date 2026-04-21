package infrastructure

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type GeminiAdapter struct {
	client *genai.Client
	model  string
}

func NewGeminiAdapter(client *genai.Client, model string) *GeminiAdapter {
	return &GeminiAdapter{
		client: client,
		model:  model,
	}
}

func (a *GeminiAdapter) EmbedText(ctx context.Context, text string) ([]float32, RequestAnalysis, error) {
	resp, err := a.client.Models.EmbedContent(ctx, a.model, []*genai.Content{
		genai.NewContentFromText(text, "user"),
	}, nil)
	if err != nil {
		return nil, RequestAnalysis{}, err
	}

	if len(resp.Embeddings) == 0 {
		return nil, RequestAnalysis{}, fmt.Errorf("no embeddings returned")
	}

	analysis := RequestAnalysis{
		Model: a.model,
	}

	return resp.Embeddings[0].Values, analysis, nil
}

func (a *GeminiAdapter) Generate(ctx context.Context, history []Message, tools []*genai.FunctionDeclaration) (Message, RequestAnalysis, error) {
	contents := make([]*genai.Content, len(history))
	for i, msg := range history {
		contents[i] = mapToGenAIContent(msg)
	}

	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{{FunctionDeclarations: tools}},
	}

	resp, err := a.client.Models.GenerateContent(ctx, a.model, contents, config)
	if err != nil {
		return Message{}, RequestAnalysis{}, err
	}

	if len(resp.Candidates) == 0 {
		return Message{}, RequestAnalysis{}, fmt.Errorf("no candidates returned")
	}

	resultMsg := mapFromGenAIContent(resp.Candidates[0].Content)

	analysis := RequestAnalysis{
		Model: a.model,
	}
	if resp.UsageMetadata != nil {
		analysis.PromptTokens = int(resp.UsageMetadata.PromptTokenCount)
		analysis.CompletationTokens = int(resp.UsageMetadata.CandidatesTokenCount)
		analysis.TotalTokens = int(resp.UsageMetadata.TotalTokenCount)
	}

	return resultMsg, analysis, nil
}

func (a *GeminiAdapter) GenerateBlocking(ctx context.Context, prompt string, systemInstruction string) (string, RequestAnalysis, error) {
	msg, analysis, err := a.Generate(ctx, []Message{
		{Role: "user", Parts: []Part{{Type: PartTypeText, Text: prompt}}},
	}, nil)
	if err != nil {
		return "", RequestAnalysis{}, err
	}
	return msg.Parts[0].Text, analysis, nil
}

func (a *GeminiAdapter) GenerateStream(ctx context.Context, prompt string, systemInstruction string) (<-chan string, <-chan RequestAnalysis, <-chan error) {
	contentChan := make(chan string)
	analysisChan := make(chan RequestAnalysis)
	errChan := make(chan error, 1)

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
		},
	}

	go func() {
		defer close(contentChan)
		defer close(analysisChan)
		defer close(errChan)

		stream := a.client.Models.GenerateContentStream(ctx, a.model, []*genai.Content{
			genai.NewContentFromText(prompt, "user"),
		}, config)

		var lastAnalysis RequestAnalysis
		for resp, err := range stream {
			if err != nil {
				errChan <- err
				return
			}

			if resp.UsageMetadata != nil {
				lastAnalysis = RequestAnalysis{
					Model:              a.model,
					PromptTokens:       int(resp.UsageMetadata.PromptTokenCount),
					CompletationTokens: int(resp.UsageMetadata.CandidatesTokenCount),
					TotalTokens:        int(resp.UsageMetadata.TotalTokenCount),
				}
			}

			contentChan <- resp.Text()
		}

		analysisChan <- lastAnalysis
	}()

	return contentChan, analysisChan, errChan
}

// Helpers

func mapToGenAIContent(msg Message) *genai.Content {
	parts := make([]*genai.Part, len(msg.Parts))
	for i, p := range msg.Parts {
		switch p.Type {
		case PartTypeText:
			parts[i] = genai.NewPartFromText(p.Text)
		case PartTypeToolCall:
			parts[i] = genai.NewPartFromFunctionCall(p.ToolName, p.ToolArgs)
		case PartTypeToolResponse:
			resp := map[string]any{"result": p.ToolResult}
			parts[i] = genai.NewPartFromFunctionResponse(p.ToolName, resp)
		}
	}
	return &genai.Content{
		Role:  msg.Role,
		Parts: parts,
	}
}

func mapFromGenAIContent(content *genai.Content) Message {
	if content == nil {
		return Message{}
	}
	parts := make([]Part, 0, len(content.Parts))
	for _, p := range content.Parts {
		if p.Text != "" {
			parts = append(parts, Part{Type: PartTypeText, Text: p.Text})
		} else if p.FunctionCall != nil {
			parts = append(parts, Part{
				Type:     PartTypeToolCall,
				ToolName: p.FunctionCall.Name,
				ToolArgs: p.FunctionCall.Args,
			})
		} else if p.FunctionResponse != nil {
			parts = append(parts, Part{
				Type:       PartTypeToolResponse,
				ToolName:   p.FunctionResponse.Name,
				ToolResult: p.FunctionResponse.Response["result"],
			})
		}
	}
	return Message{
		Role:  content.Role,
		Parts: parts,
	}
}
