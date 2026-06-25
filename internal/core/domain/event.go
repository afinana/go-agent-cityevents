package domain

import "time"

type Event struct {
	ID          string    `json:"id"`
	SourceID    string    `json:"sourceId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Embedding   []float32 `json:"embedding,omitempty"`
	Link        string    `json:"link"`
}

type QueryHistoryItem struct {
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
}
