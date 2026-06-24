package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go-agent-cityevents/internal/core/domain"
)

// EventResponse matches the outer wrapper of the Madrid JSON OpenData.
type EventResponse struct {
	Graph []EventRaw `json:"@graph"`
}

// EventRaw represents an individual object coming directly from the JSON payload.
type EventRaw struct {
	ID          string `json:"@id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// MadridFetcher implements ports.EventFetcher for fetching events from Madrid's open data platform.
type MadridFetcher struct{}

// NewMadridFetcher creates a new instance of MadridFetcher.
func NewMadridFetcher() *MadridFetcher {
	return &MadridFetcher{}
}

// FetchEvents fetches cultural events from the given URL using http.Client with context support.
func (f *MadridFetcher) FetchEvents(ctx context.Context, url string) ([]domain.Event, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response EventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	var events []domain.Event
	for _, raw := range response.Graph {
		events = append(events, domain.Event{
			SourceID:    raw.ID,
			Title:       raw.Title,
			Description: raw.Description,
		})
	}

	return events, nil
}
