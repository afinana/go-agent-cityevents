package api

import (
	"encoding/json"
	"log"
	"net/http"

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

	log.Printf("Server starting on port %s...", port)
	return http.ListenAndServe(":"+port, nil)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query cannot be empty", http.StatusBadRequest)
		return
	}

	events, err := s.useCase.SearchEvents(r.Context(), req.Query, 5)
	if err != nil {
		log.Printf("Search failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
