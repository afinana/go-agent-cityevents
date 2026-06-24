package usecase

import (
	"context"
	"errors"
	"testing"

	"go-agent-cityevents/internal/core/domain"
)

type mockRepository struct {
	inserted []*domain.Event
	fail     bool
}

func (m *mockRepository) InsertEvent(ctx context.Context, event *domain.Event) error {
	if m.fail {
		return errors.New("db error")
	}
	m.inserted = append(m.inserted, event)
	return nil
}

type mockEmbedder struct {
	embedMap map[string][]float32
	failText string
}

func (m *mockEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	if m.failText != "" && text == m.failText {
		return nil, errors.New("embedding error")
	}
	if val, ok := m.embedMap[text]; ok {
		return val, nil
	}
	return []float32{0.1, 0.2, 0.3}, nil
}

type mockFetcher struct {
	events []domain.Event
	err    error
}

func (m *mockFetcher) FetchEvents(ctx context.Context, url string) ([]domain.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.events, nil
}

func TestIngestEvents_Success(t *testing.T) {
	ctx := context.Background()

	events := []domain.Event{
		{SourceID: "1", Title: "Event 1", Description: "Description 1"},
		{SourceID: "2", Title: "Event 2", Description: "Description 2"},
	}

	fetcher := &mockFetcher{events: events}
	embedder := &mockEmbedder{
		embedMap: map[string][]float32{
			"Title: Event 1. Description: Description 1": {1.0, 1.1},
			"Title: Event 2. Description: Description 2": {2.0, 2.1},
		},
	}
	repo := &mockRepository{}

	uc := NewEventUseCase(repo, embedder, fetcher)
	err := uc.IngestEvents(ctx, "http://example.com/api")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repo.inserted) != 2 {
		t.Errorf("expected 2 inserted events, got %d", len(repo.inserted))
	}

	if repo.inserted[0].SourceID != "1" || len(repo.inserted[0].Embedding) != 2 || repo.inserted[0].Embedding[0] != 1.0 {
		t.Errorf("unexpected first inserted event: %+v", repo.inserted[0])
	}
}

func TestIngestEvents_EmbedderFailureContinues(t *testing.T) {
	ctx := context.Background()

	events := []domain.Event{
		{SourceID: "1", Title: "Event 1", Description: "Description 1"},
		{SourceID: "2", Title: "Event 2", Description: "Description 2"},
	}

	fetcher := &mockFetcher{events: events}
	embedder := &mockEmbedder{
		failText: "Title: Event 1. Description: Description 1",
		embedMap: map[string][]float32{
			"Title: Event 2. Description: Description 2": {2.0, 2.1},
		},
	}
	repo := &mockRepository{}

	uc := NewEventUseCase(repo, embedder, fetcher)
	err := uc.IngestEvents(ctx, "http://example.com/api")
	if err != nil {
		t.Fatalf("expected no error overall, got %v", err)
	}

	// First event should have failed to embed, so only second event should be inserted
	if len(repo.inserted) != 1 {
		t.Fatalf("expected 1 inserted event, got %d", len(repo.inserted))
	}

	if repo.inserted[0].SourceID != "2" {
		t.Errorf("expected Event 2 to be inserted, got %s", repo.inserted[0].SourceID)
	}
}
