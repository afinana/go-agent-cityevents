package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go-agent-cityevents/internal/a2a"
	"go-agent-cityevents/internal/core/ports"
)

type Server struct {
	useCase ports.EventUseCase
}

func NewServer(useCase ports.EventUseCase) *Server {
	return &Server{useCase: useCase}
}

type SearchRequest struct {
	Query string `json:"query"`
}

func (s *Server) Start(port string) error {
	// Serve static files from the "static" directory
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// API Endpoints
	http.HandleFunc("/api/search", s.handleSearch)
	http.HandleFunc("/api/db/status", s.handleDBStatus)
	http.HandleFunc("/api/history", s.handleHistory)

	// A2A Discovery Endpoint
	http.HandleFunc("/.well-known/agent-card.json", s.handleAgentCard)

	log.Printf("Server starting on port %s...", port)
	return http.ListenAndServe(":"+port, nil)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Received search request for query: '%s'", req.Query)

	if req.Query == "" {
		log.Printf("[WARN] Empty search query received")
		http.Error(w, "Query cannot be empty", http.StatusBadRequest)
		return
	}

	events, err := s.useCase.SearchEvents(r.Context(), req.Query, 5)
	if err != nil {
		log.Printf("[ERROR] Search failed for query '%s': %v", req.Query, err)
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Save query in MongoDB search history
	_ = s.useCase.SaveQuery(r.Context(), req.Query)

	log.Printf("[INFO] Search successful. Returning %d events", len(events))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (s *Server) handleAgentCard(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[INFO] Received request for A2A Agent Card")

	card := a2a.GetAgentCard()
	b, err := card.ToJSON()
	if err != nil {
		log.Printf("[ERROR] Failed to marshal agent card: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) handleDBStatus(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	err := s.useCase.Ping(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"connected": false, "error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"connected": true})
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		history, err := s.useCase.GetQueryHistory(r.Context(), 20)
		if err != nil {
			log.Printf("[ERROR] Failed to fetch search history: %v", err)
			http.Error(w, fmt.Sprintf("Failed to fetch search history: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)

	case http.MethodDelete:
		queryToDelete := r.URL.Query().Get("query")
		if queryToDelete != "" {
			err := s.useCase.DeleteQuery(r.Context(), queryToDelete)
			if err != nil {
				log.Printf("[ERROR] Failed to delete query '%s': %v", queryToDelete, err)
				http.Error(w, fmt.Sprintf("Failed to delete query: %v", err), http.StatusInternalServerError)
				return
			}
		} else {
			err := s.useCase.ClearQueryHistory(r.Context())
			if err != nil {
				log.Printf("[ERROR] Failed to clear search history: %v", err)
				http.Error(w, fmt.Sprintf("Failed to clear history: %v", err), http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

