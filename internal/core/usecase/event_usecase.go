package usecase

import (
	"context"
	"fmt"

	"go-agent-cityevents/internal/core/domain"
	"go-agent-cityevents/internal/core/ports"
)

type eventUseCase struct {
	repo     ports.EventRepository
	embedder ports.Embedder
	fetcher  ports.EventFetcher
}

// NewEventUseCase creates a new instance of EventUseCase.
func NewEventUseCase(repo ports.EventRepository, embedder ports.Embedder, fetcher ports.EventFetcher) ports.EventUseCase {
	return &eventUseCase{
		repo:     repo,
		embedder: embedder,
		fetcher:  fetcher,
	}
}

// IngestEvents fetches events, generates embeddings, and saves them to the repository.
func (u *eventUseCase) IngestEvents(ctx context.Context, url string) error {
	fmt.Println("Fetching Madrid cultural events...")
	events, err := u.fetcher.FetchEvents(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch events: %w", err)
	}

	fmt.Printf("Fetched %d events. Processing and generating embeddings...\n", len(events))

	successCount := 0
	for i, event := range events {
		textToEmbed := fmt.Sprintf("Title: %s. Description: %s", event.Title, event.Description)

		embedding, err := u.embedder.EmbedText(ctx, textToEmbed)
		if err != nil {
			fmt.Printf("Failed to embed event %s: %v\n", event.SourceID, err)
			continue
		}

		event.Embedding = embedding

		err = u.repo.InsertEvent(ctx, &event)
		if err != nil {
			fmt.Printf("Failed to insert event %s into MongoDB: %v\n", event.SourceID, err)
			continue
		}

		successCount++
		if (i+1)%10 == 0 {
			fmt.Printf("Processed %d/%d events...\n", i+1, len(events))
		}
	}

	fmt.Printf("Ingestion and Vectorization complete! Successfully stored %d events.\n", successCount)
	return nil
}

// SearchEvents embeds the query and searches the repository.
func (u *eventUseCase) SearchEvents(ctx context.Context, query string, limit int) ([]*domain.Event, error) {
	// Embed the query
	embedding, err := u.embedder.EmbedText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search in repo
	return u.repo.SearchEvents(ctx, embedding, limit)
}
