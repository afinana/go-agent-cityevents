package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"go-agent-cityevents/internal/core/ports"
	"google.golang.org/genai"
)

// Provider type represents the supported LLM/embedding providers.
type Provider string

const (
	ProviderOllama Provider = "ollama"
	ProviderVertex Provider = "vertex"
)

// NewEmbedder returns an implementation of ports.Embedder based on the provider config.
func NewEmbedder(ctx context.Context, provider string) (ports.Embedder, error) {
	if Provider(provider) == ProviderVertex {
		location := getEnv("GCP_LOCATION", "us-central1")
		projectID := getEnv("GCP_PROJECT_ID", "")
		model := getEnv("EMBEDDING_MODEL", "text-embedding-004")

		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			Backend:  genai.BackendVertexAI,
			Project:  projectID,
			Location: location,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create vertex client: %w", err)
		}

		return &VertexEmbedder{
			client: client,
			model:  model,
		}, nil
	}

	// Default to Ollama
	return &OllamaEmbedder{
		url:   getEnv("OLLAMA_URL", "http://localhost:11434"),
		model: getEnv("EMBEDDING_MODEL", "nomic-embed-text"),
	}, nil
}

// -- Vertex Embedder --

type VertexEmbedder struct {
	client *genai.Client
	model  string
}

func (e *VertexEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	res, err := e.client.Models.EmbedContent(ctx, e.model, genai.Text(text), nil)
	if err != nil {
		return nil, fmt.Errorf("vertex embedding error: %w", err)
	}

	if len(res.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned from vertex")
	}

	return res.Embeddings[0].Values, nil
}

// -- Ollama Embedder --

type OllamaEmbedder struct {
	url   string
	model string
}

type ollamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

func (e *OllamaEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	reqBody := ollamaEmbedRequest{
		Model:  e.model,
		Prompt: text,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/embeddings", e.url), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama error: status %d", resp.StatusCode)
	}

	var res ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Embedding, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
