package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewEmbedderRejectsInvalidOllamaModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"nomic-embed-text"}]}`))
	}))
	defer server.Close()

	t.Run("single model allowed", func(t *testing.T) {
		t.Setenv("OLLAMA_URL", server.URL)
		t.Setenv("EMBEDDING_MODEL", "nomic-embed-text")
		_, err := NewEmbedder(context.Background(), "ollama")
		if err != nil {
			t.Fatalf("expected valid ollama model to initialize, got error: %v", err)
		}
	})

	t.Run("comma-separated models rejected", func(t *testing.T) {
		t.Setenv("OLLAMA_URL", server.URL)
		t.Setenv("EMBEDDING_MODEL", "nomic-embed-text,another-model")
		_, err := NewEmbedder(context.Background(), "ollama")
		if err == nil {
			t.Fatal("expected comma-separated model list to be rejected")
		}
	})
}

func TestNewEmbedderRejectsMissingOllamaModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"other-model"}]}`))
	}))
	defer server.Close()

	t.Setenv("OLLAMA_URL", server.URL)
	t.Setenv("EMBEDDING_MODEL", "missing-model")

	_, err := NewEmbedder(context.Background(), "ollama")
	if err == nil {
		t.Fatal("expected missing ollama model to fail initialization")
	}
}
