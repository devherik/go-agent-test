package agent

import (
	"context"
	"fmt"

	"go-agent-test/infrastructure"
)

type WeatherAgent struct {
	provider infrastructure.AIProvider
	registry *Registry
}

func NewWeatherAgent(provider infrastructure.AIProvider) *WeatherAgent {
	a := &WeatherAgent{
		provider: provider,
		registry: NewRegistry(),
	}

	// Register Tools
	a.registry.Register(GetWeatherDeclaration, GetWeatherTool)

	return a
}

func (a *WeatherAgent) Run(ctx context.Context, userInput string) (string, error) {
	history := []infrastructure.Message{
		{
			Role:  "user",
			Parts: []infrastructure.Part{{Type: infrastructure.PartTypeText, Text: userInput}},
		},
	}

	for {
		// The agent uses the provider to generate the next message.
		// Note: We are passing SDK-specific declarations here, but these could be abstracted further.
		msg, _, err := a.provider.Generate(ctx, history, a.registry.Declarations())
		if err != nil {
			return "", err
		}

		// Append model response to history
		history = append(history, msg)

		// Check for tool calls in the response parts
		var toolCalls []infrastructure.Part
		for _, p := range msg.Parts {
			if p.Type == infrastructure.PartTypeToolCall {
				toolCalls = append(toolCalls, p)
			}
		}

		// If no tool calls, return the text content
		if len(toolCalls) == 0 {
			for _, p := range msg.Parts {
				if p.Type == infrastructure.PartTypeText {
					return p.Text, nil
				}
			}
			return "I couldn't generate a text response.", nil
		}

		// Handle function calls
		var responseParts []infrastructure.Part
		for _, call := range toolCalls {
			result, err := a.registry.Call(ctx, call.ToolName, call.ToolArgs)
			if err != nil {
				return "", fmt.Errorf("error calling tool %q: %w", call.ToolName, err)
			}

			responseParts = append(responseParts, infrastructure.Part{
				Type:       infrastructure.PartTypeToolResponse,
				ToolName:   call.ToolName,
				ToolResult: result,
			})
		}

		// Append function responses to history
		history = append(history, infrastructure.Message{
			Role:  "user",
			Parts: responseParts,
		})
	}
}
