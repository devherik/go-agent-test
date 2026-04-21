package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/genai"

	"go-agent-test/agent"
	"go-agent-test/infrastructure"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error to load .env file: %v\n", err)
		return
	}

	key := os.Getenv("GEMINI_API_KEY")

	if key == "" {
		fmt.Printf("Error to create genai client: GEMINI_API_KEY not found in environment variables\n")
		return
	}

	// 1. Initialize the concrete SDK client (Infrastructure)
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  key,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		fmt.Printf("Error to create genai client: %v\n", err)
		return
	}

	// 2. Initialize the Adapter (Interface implementation)
	provider := infrastructure.NewGeminiAdapter(client, "gemini-2.5-flash")

	// 3. Initialize the Agent with the abstraction (Dependency Injection)
	agent := agent.NewWeatherAgent(provider)

	// 4. Execute the business logic
	result, err := agent.Run(ctx, "What is the weather like in Timóteo, MG, Brazil?")
	if err != nil {
		fmt.Printf("Agent error: %v\n", err)
		return
	}

	fmt.Println("Agent Response:", result)
}
