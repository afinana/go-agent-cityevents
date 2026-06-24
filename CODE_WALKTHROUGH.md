# Code Walkthrough: Madrid Events AI Agent

This document provides a detailed, line-by-line explanation of the Go codebase for the Madrid City Events AI Agent. It is intended to help you understand exactly what the application does, how it fetches data, communicates with LLMs, and stores vectors into MongoDB.

---

## 1. `main.go`
This is the entry point of our Go application. It glues all the internal packages together.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go-agent-cityevents/internal/db"
	"go-agent-cityevents/internal/ingest"
	"go-agent-cityevents/internal/llm"
)
```
- `package main`: Declares that this file should compile as an executable program rather than a shared library.
- `import`: Pulls in standard library packages (`context`, `fmt`, `log`, `os`) and our custom packages (`db`, `ingest`, `llm`).

```go
func main() {
	fmt.Println("Starting Madrid City Events AI Agent...")

	// Create a background context that can be passed down to database and network requests.
	ctx := context.Background()
```
- `func main()`: The primary execution function of the app.
- `ctx := context.Background()`: Initializes a non-nil, empty context used to carry deadlines, cancellation signals, and request-scoped values across API boundaries.

```go
	// 1. Initialize MongoDB Client
	mongoURI := getEnv("MONGO_URI", "mongodb://root:example@localhost:27017")
	dbName := getEnv("MONGO_DB_NAME", "madrid_events")
	colName := getEnv("MONGO_COLLECTION_NAME", "events")
```
- We read environment variables using our custom `getEnv` helper. If the variables aren't set, it uses the hardcoded defaults.

```go
	fmt.Println("Connecting to MongoDB...")
	mongoClient, err := db.NewMongoClient(ctx, mongoURI, dbName, colName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Close(ctx)
```
- `db.NewMongoClient`: Calls our custom database package to establish a connection to MongoDB.
- `log.Fatalf`: If connection fails, prints the error and crashes the app.
- `defer mongoClient.Close(ctx)`: Ensures that whenever `main()` finishes, the MongoDB connection is safely terminated.

```go
	// 2. Initialize LLM Embedder
	provider := getEnv("LLM_PROVIDER", "ollama")
	fmt.Printf("Initializing LLM Embedder (Provider: %s)...\n", provider)
	embedder, err := llm.NewEmbedder(ctx, provider)
	if err != nil {
		log.Fatalf("Failed to initialize embedder: %v", err)
	}
```
- Checks whether to use `ollama` or `vertex` via environment variable and instantiates the proper LLM struct using our `llm` package factory function.

```go
	// 3. Ingest Events
	eventsURL := "https://datos.madrid.es/dataset/206974-0-agenda-eventos-culturales-100/resource/206974-0-agenda-eventos-culturales-100-json/download/206974-0-agenda-eventos-culturales-100-json.json"
	fmt.Println("Fetching Madrid cultural events...")
	events, err := ingest.FetchEvents(eventsURL)
	if err != nil {
		log.Fatalf("Failed to fetch events: %v", err)
	}
```
- Defines the URL pointing to the Madrid Open Data platform.
- Calls `ingest.FetchEvents` to perform an HTTP GET request and parse the JSON into Go structs.

```go
	// 4. Process and Store
	successCount := 0
	for i, event := range events {
		textToEmbed := fmt.Sprintf("Title: %s. Description: %s", event.Title, event.Description)
```
- Loops through every parsed `event`.
- `textToEmbed`: Formats the title and description into a single string that the LLM will digest to generate vectors.

```go
		embedding, err := embedder.EmbedText(ctx, textToEmbed)
		if err != nil {
			log.Printf("Failed to embed event %s: %v", event.SourceID, err)
			continue
		}
```
- Asks the LLM (Ollama/Vertex) to embed the combined string. It returns a slice of floats (`[]float32`).
- If it fails, we log it and `continue` to the next event.

```go
		event.Embedding = embedding

		err = mongoClient.InsertEvent(ctx, &event)
		if err != nil {
			log.Printf("Failed to insert event %s into MongoDB: %v", event.SourceID, err)
			continue
		}
```
- Assigns the generated array of floats back to the `Embedding` field on our `event` struct.
- Inserts that entire struct (which now includes the BSON fields) directly into MongoDB.

```go
		successCount++
		if (i+1)%10 == 0 {
			fmt.Printf("Processed %d/%d events...\n", i+1, len(events))
		}
	}

	fmt.Printf("Ingestion and Vectorization complete! Successfully stored %d events.\n", successCount)
}
```
- Keeps track of successful inserts and prints progress to the console every 10 events.

---

## 2. `internal/models/event.go`
This file contains the data structures used throughout the app.

```go
package models

// EventResponse matches the outer wrapper of the Madrid JSON OpenData.
type EventResponse struct {
	Graph []EventRaw `json:"@graph"`
}
```
- The Madrid JSON API wraps all data in a field called `@graph`. We use this struct to unpack it.

```go
// EventRaw represents an individual object coming directly from the JSON payload.
type EventRaw struct {
	ID          string `json:"@id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
```
- Maps the incoming JSON keys (`@id`, `title`, `description`) to Go fields.

```go
// Event represents our normalized internal data structure and the MongoDB schema.
type Event struct {
	ID          string    `bson:"_id,omitempty"`
	SourceID    string    `bson:"source_id"`
	Title       string    `bson:"title"`
	Description string    `bson:"description"`
	Embedding   []float32 `bson:"embedding,omitempty"`
}
```
- **`bson` tags:** These tell the MongoDB driver how to name the columns in the database.
- `Embedding []float32`: This is where the mathematical vector output by the LLM is stored. MongoDB requires numeric vector arrays to be either 32-bit floats or 64-bit floats.

---

## 3. `internal/ingest/madrid.go`
Handles the external HTTP requests to the Open Data endpoint.

```go
package ingest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-agent-cityevents/internal/models"
)

func FetchEvents(url string) ([]models.Event, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
```
- `http.Get(url)`: Performs a standard GET request to fetch the raw data.
- `defer resp.Body.Close()`: Ensures the network stream is closed after reading, preventing memory leaks.

```go
	var response models.EventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
```
- Reads the raw bytes from the HTTP response body and decodes them directly into our Go `EventResponse` struct using the `json` standard library.

```go
	var events []models.Event
	for _, raw := range response.Graph {
		events = append(events, models.Event{
			SourceID:    raw.ID,
			Title:       raw.Title,
			Description: raw.Description,
		})
	}
	return events, nil
}
```
- Iterates over the raw events parsed from JSON, maps them into our clean, normalized `models.Event` format, and returns the slice.

---

## 4. `internal/llm/embedder.go`
Manages the connection to Language Models (both Vertex AI and Ollama).

```go
type Embedder interface {
	EmbedText(ctx context.Context, text string) ([]float32, error)
}
```
- **Interface:** Allows our main application to not care whether we are using Google Vertex or local Ollama. It just wants a method that takes a string and returns `[]float32`.

```go
func NewEmbedder(ctx context.Context, provider string) (Embedder, error) {
	if Provider(provider) == ProviderVertex {
        // Vertex Initialization code using google.golang.org/genai
    }
	return &OllamaEmbedder{...}, nil
}
```
- This is a "factory function". Depending on the environment variable, it builds either a Google Vertex client or our custom local Ollama HTTP client.

```go
func (e *VertexEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	res, err := e.client.Models.EmbedContent(ctx, e.model, genai.Text(text), nil)
    // ...
	return res.Embeddings[0].Values, nil
}
```
- Uses the official `google.golang.org/genai` library. We invoke `EmbedContent` with the text and return the mathematical result.

```go
func (e *OllamaEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	reqBody := ollamaEmbedRequest{
		Model:  e.model,
		Prompt: text,
	}
	b, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/embeddings", e.url), bytes.NewReader(b))
    // ...
```
- Because Ollama runs locally, we manually construct an HTTP POST request targeting its `/api/embeddings` endpoint. We encode the requested model (e.g., `nomic-embed-text`) and the event text. It responds with the embedding floats.

---

## 5. `internal/db/mongo.go`
Manages insertion into the MongoDB database.

```go
package db

import (
	"context"
	"fmt"
	"go-agent-cityevents/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
```
- Imports the official `go.mongodb.org/mongo-driver`.

```go
func NewMongoClient(ctx context.Context, uri, dbName, colName string) (*MongoClient, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
    // ...
```
- Creates configuration using the URI (like `mongodb://localhost:27017`) and initiates the network connection to the database.

```go
func (m *MongoClient) InsertEvent(ctx context.Context, event *models.Event) error {
	_, err := m.collection.InsertOne(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	return nil
}
```
- Takes our Go `event` struct and writes it into the `events` collection inside MongoDB. The driver natively converts Go types (like `[]float32`) into their proper BSON equivalent (arrays of doubles/floats).
