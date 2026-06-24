package ingest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMadridFetcher_FetchEvents(t *testing.T) {
	// Sample JSON matching the structure of Madrid cultural events
	jsonResponse := `{
		"@graph": [
			{
				"@id": "https://datos.madrid.es/egob/kos/entidadesYRecursos/Actividades/12345",
				"title": "Concierto de Verano",
				"description": "Un concierto al aire libre en el parque del Retiro."
			},
			{
				"@id": "https://datos.madrid.es/egob/kos/entidadesYRecursos/Actividades/67890",
				"title": "Exposición de Arte Moderno",
				"description": "Muestra retrospectiva de artistas locales."
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, jsonResponse)
	}))
	defer server.Close()

	fetcher := NewMadridFetcher()
	events, err := fetcher.FetchEvents(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Verify mappings
	e1 := events[0]
	if e1.SourceID != "https://datos.madrid.es/egob/kos/entidadesYRecursos/Actividades/12345" {
		t.Errorf("expected SourceID to be mapped, got %s", e1.SourceID)
	}
	if e1.Title != "Concierto de Verano" {
		t.Errorf("expected Title to be mapped, got %s", e1.Title)
	}
	if e1.Description != "Un concierto al aire libre en el parque del Retiro." {
		t.Errorf("expected Description to be mapped, got %s", e1.Description)
	}

	e2 := events[1]
	if e2.SourceID != "https://datos.madrid.es/egob/kos/entidadesYRecursos/Actividades/67890" {
		t.Errorf("expected SourceID to be mapped, got %s", e2.SourceID)
	}
}
