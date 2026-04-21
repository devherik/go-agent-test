package agent

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/genai"
)

var GetWeatherDeclaration = &genai.FunctionDeclaration{
	Name:        "get_weather",
	Description: "Get the current weather in a given location",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"location": {
				Type:        genai.TypeString,
				Description: "The city and state, e.g. San Francisco, CA",
			},
		},
		Required: []string{"location"},
	},
}

func GetWeatherTool(ctx context.Context, args map[string]any) (any, error) {
	location, ok := args["location"].(string)
	if !ok {
		return nil, errors.New("missing location argument")
	}
	// Logic: Call OpenWeather API or a Database
	fmt.Printf("Executing GetWeatherTool for location: %s\n", location)
	return fmt.Sprintf("The weather in %s is 25°C", location), nil
}
