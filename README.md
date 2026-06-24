# Go AI Agent: Madrid City Events

This is a Proof of Concept (POC) Go application that acts as an Artificial Intelligence agent. It is designed to ingest cultural events from the Madrid City Open Data platform, generate text embeddings for the event descriptions using a Large Language Model (LLM), and persist those vectors in a MongoDB database equipped with vector search capabilities.

## Architecture and Agentic Workflow

The application was designed based on a multi-agent structure:
1. **[Architect](Architect.md):** Defined the robust integration of MongoDB vector capabilities and Google Gen AI SDK within a containerized environment.
2. **[Developer](Developer.md):** Implemented the ingestion pipeline, the abstraction layer for LLM providers, and MongoDB storage in Go.
3. **[QA Tester](QATester.md):** Validates the data flow, handles testing toggles between LLMs, and monitors data persistence integrity.

## Prerequisites

- **Go 1.22+**
- **Docker & Docker Compose** (to run MongoDB and Ollama locally)
- *(Optional)* Google Cloud Platform account with Vertex AI API enabled (if you prefer to use Vertex instead of Ollama).

## Setup & Infrastructure

1. **Environment Variables**
   Copy the example environment configuration:
   ```bash
   cp .env.example .env
   ```

2. **Start the Infrastructure**
   Spin up the local MongoDB instance and Ollama service using Docker Compose:
   ```bash
   docker-compose up -d
   ```

3. **Pull the Local Embedding Model**
   Since the default LLM provider is Ollama, you'll need to pull the chosen embedding model:
   ```bash
   docker exec -it <project_name>_ollama_1 ollama run nomic-embed-text
   # OR simply run locally if Ollama is installed on host:
   # ollama pull nomic-embed-text
   ```

## Running the Application

To run the ingestion and vectorization agent, execute the main Go file:

```bash
go run main.go
```

The agent will:
1. Connect to MongoDB.
2. Fetch the JSON payload containing the cultural events from the Madrid Open Data portal.
3. Iterate through each event, contact the LLM provider (Ollama by default) to generate vector embeddings based on the title and description.
4. Insert the final document structure including the vector data into the MongoDB database.

## Switching to Google Vertex AI

To switch the underlying LLM provider from Ollama to Google Vertex AI:

1. Edit your `.env` file and set `LLM_PROVIDER=vertex`.
2. Ensure you have defined your `GCP_PROJECT_ID` and `GCP_LOCATION`.
3. Set your Google Application Credentials in your shell environment:
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your/service-account.json"
   # OR run: gcloud auth application-default login
   ```
4. Re-run `go run main.go`. The Google Gen AI SDK (`google.golang.org/genai`) will seamlessly route embedding requests to Vertex AI.
