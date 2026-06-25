package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-agent-cityevents/internal/core/domain"
)

type mockEventUseCase struct {
	searchFunc func(ctx context.Context, query string, limit int) ([]*domain.Event, error)
	pingFunc   func(ctx context.Context) error
	saveFunc   func(ctx context.Context, query string) error
	histFunc   func(ctx context.Context, limit int) ([]domain.QueryHistoryItem, error)
	delFunc    func(ctx context.Context, query string) error
	clearFunc  func(ctx context.Context) error
}

func (m *mockEventUseCase) IngestEvents(ctx context.Context, url string) error {
	return nil
}

func (m *mockEventUseCase) SearchEvents(ctx context.Context, query string, limit int) ([]*domain.Event, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, query, limit)
	}
	return nil, nil
}

func (m *mockEventUseCase) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func (m *mockEventUseCase) SaveQuery(ctx context.Context, query string) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, query)
	}
	return nil
}

func (m *mockEventUseCase) GetQueryHistory(ctx context.Context, limit int) ([]domain.QueryHistoryItem, error) {
	if m.histFunc != nil {
		return m.histFunc(ctx, limit)
	}
	return nil, nil
}

func (m *mockEventUseCase) DeleteQuery(ctx context.Context, query string) error {
	if m.delFunc != nil {
		return m.delFunc(ctx, query)
	}
	return nil
}

func (m *mockEventUseCase) ClearQueryHistory(ctx context.Context) error {
	if m.clearFunc != nil {
		return m.clearFunc(ctx)
	}
	return nil
}

func TestHandleDBStatus_Connected(t *testing.T) {
	mockUC := &mockEventUseCase{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}
	server := NewServer(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/db/status", nil)
	w := httptest.NewRecorder()

	server.handleDBStatus(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body["connected"] != true {
		t.Errorf("expected connected to be true, got %+v", body)
	}
}

func TestHandleDBStatus_Disconnected(t *testing.T) {
	mockUC := &mockEventUseCase{
		pingFunc: func(ctx context.Context) error {
			return errors.New("connection failed")
		},
	}
	server := NewServer(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/db/status", nil)
	w := httptest.NewRecorder()

	server.handleDBStatus(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body["connected"] != false {
		t.Errorf("expected connected to be false, got %+v", body)
	}
	if body["error"] != "connection failed" {
		t.Errorf("expected error field to be 'connection failed', got %+v", body)
	}
}

func TestHandleSearch_Success(t *testing.T) {
	savedQuery := ""
	mockUC := &mockEventUseCase{
		searchFunc: func(ctx context.Context, query string, limit int) ([]*domain.Event, error) {
			return []*domain.Event{
				{ID: "1", Title: "Rock Fest"},
			}, nil
		},
		saveFunc: func(ctx context.Context, query string) error {
			savedQuery = query
			return nil
		},
	}
	server := NewServer(mockUC)

	reqBody, _ := json.Marshal(SearchRequest{Query: "rock"})
	req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.handleSearch(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var events []*domain.Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if len(events) != 1 || events[0].Title != "Rock Fest" {
		t.Errorf("unexpected events: %+v", events)
	}
	if savedQuery != "rock" {
		t.Errorf("expected query 'rock' to be saved, got %q", savedQuery)
	}
}

func TestHandleSearch_Error(t *testing.T) {
	mockUC := &mockEventUseCase{
		searchFunc: func(ctx context.Context, query string, limit int) ([]*domain.Event, error) {
			return nil, errors.New("ollama error")
		},
	}
	server := NewServer(mockUC)

	reqBody, _ := json.Marshal(SearchRequest{Query: "rock"})
	req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.handleSearch(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	respStr := buf.String()

	expectedError := "Search failed: ollama error\n"
	if respStr != expectedError {
		t.Errorf("expected response %q, got %q", expectedError, respStr)
	}
}

func TestHandleHistory_Get(t *testing.T) {
	now := time.Now()
	mockUC := &mockEventUseCase{
		histFunc: func(ctx context.Context, limit int) ([]domain.QueryHistoryItem, error) {
			return []domain.QueryHistoryItem{
				{Query: "teatro", Timestamp: now},
			}, nil
		},
	}
	server := NewServer(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/history", nil)
	w := httptest.NewRecorder()

	server.handleHistory(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var history []domain.QueryHistoryItem
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if len(history) != 1 || history[0].Query != "teatro" {
		t.Errorf("unexpected history: %+v", history)
	}
}

func TestHandleHistory_DeleteSingle(t *testing.T) {
	deletedQuery := ""
	mockUC := &mockEventUseCase{
		delFunc: func(ctx context.Context, query string) error {
			deletedQuery = query
			return nil
		},
	}
	server := NewServer(mockUC)

	req := httptest.NewRequest(http.MethodDelete, "/api/history?query=teatro", nil)
	w := httptest.NewRecorder()

	server.handleHistory(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if deletedQuery != "teatro" {
		t.Errorf("expected query 'teatro' to be deleted, got %q", deletedQuery)
	}
}

func TestHandleHistory_DeleteAll(t *testing.T) {
	cleared := false
	mockUC := &mockEventUseCase{
		clearFunc: func(ctx context.Context) error {
			cleared = true
			return nil
		},
	}
	server := NewServer(mockUC)

	req := httptest.NewRequest(http.MethodDelete, "/api/history", nil)
	w := httptest.NewRecorder()

	server.handleHistory(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if !cleared {
		t.Error("expected clear query history to be called")
	}
}
