package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go-agent-cityevents/internal/adapters/db"
	"go-agent-cityevents/internal/adapters/ingest"
	"go-agent-cityevents/internal/adapters/llm"
	"go-agent-cityevents/internal/api"
	"go-agent-cityevents/internal/core/usecase"

	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("Starting Madrid City Events AI Agent & Web Service...")

	ctx := context.Background()

	// 1. Initialize MongoDB Client (Outbound Adapter)
	mongoURI := getEnv("MONGO_URI", "mongodb://root:example@localhost:27017/?directConnection=true")
	dbName := getEnv("MONGO_DB_NAME", "madrid_events")
	colName := getEnv("MONGO_COLLECTION_NAME", "events")

	fmt.Println("Connecting to MongoDB...")
	mongoClient, err := db.NewMongoClient(ctx, mongoURI, dbName, colName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Close(ctx)

	// Ensure vector search index exists
	_ = mongoClient.EnsureVectorIndex(ctx)

	// 2. Initialize LLM Embedder (Outbound Adapter)
	provider := getEnv("LLM_PROVIDER", "ollama")
	fmt.Printf("Initializing LLM Embedder (Provider: %s)...\n", provider)
	embedder, err := llm.NewEmbedder(ctx, provider)
	if err != nil {
		log.Fatalf("Failed to initialize embedder: %v", err)
	}

	// 3. Initialize Ingestion Fetcher (Outbound Adapter)
	fetcher := ingest.NewMadridFetcher()

	// 4. Initialize Use Case (Core Use Case)
	eventUseCase := usecase.NewEventUseCase(mongoClient, embedder, fetcher)

	// 5. Ensure data exists before serving search requests.
	skipIngest := getEnv("SKIP_INGEST", "true")
	shouldIngest, err := shouldRunIngestionWithCount(ctx, mongoClient, skipIngest)
	if err != nil {
		log.Fatalf("Failed to check event collection state: %v", err)
	}
	if shouldIngest {
		eventsURL := "https://datos.madrid.es/dataset/206974-0-agenda-eventos-culturales-100/resource/206974-0-agenda-eventos-culturales-100-json/download/206974-0-agenda-eventos-culturales-100-json.json"
		fmt.Println("Ingesting events into MongoDB...")
		err = eventUseCase.IngestEvents(ctx, eventsURL)
		if err != nil {
			log.Fatalf("Failed to run events ingestion: %v", err)
		}
	} else {
		fmt.Println("Using existing event data. Set SKIP_INGEST=false to force re-ingestion.")
	}

	// 6. Start API Server
	apiServer := api.NewServer(eventUseCase)
	port := getEnv("PORT", "8080")
	fmt.Printf("Web server running at http://localhost:%s\n", port)
	if err := apiServer.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func shouldRunIngestionWithCount(ctx context.Context, mongoClient interface {
	CountDocuments(context.Context, interface{}, ...*options.CountOptions) (int64, error)
}, skipIngest string) (bool, error) {
	if skipIngest != "true" {
		return true, nil
	}

	count, err := mongoClient.CountDocuments(ctx, map[string]interface{}{})
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func shouldRunIngestion(skipIngest string, existingCount int) bool {
	if skipIngest != "true" {
		return true
	}
	return existingCount == 0
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
