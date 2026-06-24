# Implementation Plan: Copilot 365 Style GUI & Search Service

This plan outlines the steps to introduce a modern, Copilot 365-style Web GUI and a backend search service to our existing Madrid City Events AI Agent.

## User Review Required

Please review the proposed architecture and design approach. Once approved, I will begin implementing the frontend and backend.

> [!IMPORTANT]
> The Copilot 365 aesthetic will be achieved using Vanilla HTML/CSS/JS (no heavy framework like React/Next.js) to keep the POC lightweight but visually premium. It will feature a sleek dark mode, chat/search interface, glassmorphism, and smooth micro-animations.

## Open Questions

- We will use standard Go `net/http` (Go 1.22 has enhanced routing) for the backend. Does this sound good, or would you prefer a framework like Gin or Echo?
- To perform `$vectorSearch` locally, the `mongodb-atlas-local` image requires creating a Vector Search Index. I will add logic to create this index automatically on application startup. Are you okay with the application attempting to create the index upon boot?

## Proposed Changes

### 1. Backend Search Service (Go)
#### [MODIFY] `main.go`
- Update the main entry point to start an HTTP server after the ingestion is complete (or optionally skip ingestion if data is already present).
- Serve the static frontend files.
#### [NEW] `internal/api/server.go`
- Implement HTTP handlers:
  - `GET /` -> Serves the HTML GUI.
  - `POST /api/search` -> Accepts a JSON payload `{ "query": "..." }`, generates an embedding using the `llm.Embedder`, and queries MongoDB.
#### [MODIFY] `internal/db/mongo.go`
- Add a `SearchEvents(ctx, queryEmbedding []float32, limit int)` method.
- This method will use MongoDB's `$vectorSearch` aggregation pipeline to find the most semantically similar events.
- Add an `EnsureVectorIndex(ctx)` method to automatically create the required Atlas Vector Search index on the `events` collection.

### 2. Frontend GUI (HTML/CSS/JS)
#### [NEW] `static/index.html`
- The semantic HTML structure for the Copilot 365-like interface. It will feature a sidebar, a main search/chat area, and a results pane.
#### [NEW] `static/styles.css`
- Premium aesthetics mimicking Copilot 365:
  - Curated dark mode color palette (deep grays, vibrant primary accents).
  - Modern typography (Inter or Segoe UI).
  - Hover effects, smooth transitions, and glassmorphism.
#### [NEW] `static/app.js`
- Vanilla JavaScript to handle the search form submission, call the `/api/search` endpoint, handle loading states with skeleton loaders, and dynamically render the event cards into the DOM.

## Verification Plan

### Automated Tests
- N/A for this step (focusing on UI/UX and endpoint wiring).

### Manual Verification
1. Re-run `docker-compose up -d` to ensure MongoDB and Ollama are running.
2. Run `go run main.go`.
3. Open a browser to `http://localhost:8080`.
4. Type a natural language query (e.g., "music festivals this weekend") into the chat/search bar.
5. Verify the backend successfully generates an embedding and retrieves the relevant events from MongoDB via `$vectorSearch`.
6. Visually inspect the UI to ensure it meets the "premium Copilot 365" design requirements.
