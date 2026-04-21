# Go Gemini Agent Test

A robust, production-ready Go implementation of an Agentic AI system using the Google GenAI SDK. This project demonstrates Clean Architecture, Dependency Inversion, and TDD principles in an AI context.

## 🏗 Architecture

The project follows **Clean Architecture** to ensure that the core business logic (the Agent) is decoupled from external infrastructure (the Gemini SDK).

### Directory Structure
- `agent/`: **The Core.** Contains the orchestration logic, tool registry, and business-facing agent.
- `infrastructure/`: **The Shell.** Contains concrete implementations of AI adapters and third-party SDK integrations.
- `test/`: **Verification.** Contains unit and integration tests using a Mocking strategy.

## 🧠 Core Concepts

### 1. The Agentic Loop
The agent implements a recursive orchestration loop. Instead of a single prompt/response, it:
1.  Sends the current state to the LLM.
2.  Parses the model's intent (Text vs. Tool Call).
3.  If a **Tool Call** is requested, it executes the corresponding Go function from the `Registry`.
4.  Feeds the result back to the model and repeats until a final answer is generated.

### 2. Dependency Inversion (DIP)
The `WeatherAgent` does not depend on the Gemini SDK. It depends on the `AIProvider` interface. This allows for:
-   **Swapability**: Easy migration to OpenAI, Anthropic, or local models.
-   **Testability**: Mocking the LLM for deterministic unit tests.

### 3. Tool Registry
Tools are defined as standard Go functions and registered with a JSON Schema declaration. This metadata is sent to the LLM so it "knows" how to call your internal functions.

## 🚀 Getting Started

### Prerequisites
- Go 1.23+
- Google Gemini API Key

### Installation
```bash
go mod download
```

### Running the Agent
```bash
# Set your API key in main.go or via env (recommended)
go run main.go
```

## 🧪 Testing

We use a **TDD (Test-Driven Development)** approach. Our tests use a `MockAIProvider` to simulate multi-turn LLM conversations without hitting real API quotas.

```bash
# Run all tests
go test -v ./test/...
```

## 🛠 Tech Stack
- **Language**: Go 1.25
- **SDK**: `google.golang.org/genai` (The latest Google GenAI SDK)
- **Patterns**: Singleton Registry, Adapter Pattern, Dependency Injection.

---
*Created by Antigravity AI Assistant.*
