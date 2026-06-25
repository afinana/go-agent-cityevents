package ports

import (
	"context"

	"go-agent-cityevents/internal/core/domain"
)

// EventRepository defines the outbound port for persisting and querying events.
type EventRepository interface {
	InsertEvent(ctx context.Context, event *domain.Event) error
	SearchEvents(ctx context.Context, queryEmbedding []float32, limit int) ([]*domain.Event, error)
	EnsureVectorIndex(ctx context.Context) error
	Ping(ctx context.Context) error
	SaveQuery(ctx context.Context, query string) error
	GetQueryHistory(ctx context.Context, limit int) ([]domain.QueryHistoryItem, error)
	DeleteQuery(ctx context.Context, query string) error
	ClearQueryHistory(ctx context.Context) error
}

// Embedder defines the outbound port for generating vector embeddings.
type Embedder interface {
	EmbedText(ctx context.Context, text string) ([]float32, error)
}

// EventFetcher defines the outbound port for fetching external event data.
type EventFetcher interface {
	FetchEvents(ctx context.Context, url string) ([]domain.Event, error)
}

// EventUseCase defines the inbound port for orchestrating the events ingestion use case.
type EventUseCase interface {
	IngestEvents(ctx context.Context, url string) error
	SearchEvents(ctx context.Context, query string, limit int) ([]*domain.Event, error)
	Ping(ctx context.Context) error
	SaveQuery(ctx context.Context, query string) error
	GetQueryHistory(ctx context.Context, limit int) ([]domain.QueryHistoryItem, error)
	DeleteQuery(ctx context.Context, query string) error
	ClearQueryHistory(ctx context.Context) error
}
