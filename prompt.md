# Go AI Agent: Madrid City Events Vectorization

Please create a Proof of Concept (POC) for an Artificial Intelligence agent in Go. The application must fulfill the following requirements:

## 1. Data Ingestion
- Fetch the Madrid city cultural events dataset from the following JSON open data endpoint:
  `https://datos.madrid.es/dataset/206974-0-agenda-eventos-culturales-100/resource/206974-0-agenda-eventos-culturales-100-json/download/206974-0-agenda-eventos-culturales-100-json.json`
- Parse the JSON response to extract the essential details for each event (such as title, description, location, dates, etc.).

## 2. LLM and Embedding Generation
- Use the official Google AI SDK for Go to integrate LLM capabilities.
- The agent must support using either **Google Vertex AI** or **Ollama** for generating text embeddings of the event data.
- **Ollama must be configured as the default LLM provider**.

## 3. Vector Database (MongoDB)
- Persist the parsed events along with their generated vector embeddings into MongoDB.
- Ensure the MongoDB setup and schema take advantage of MongoDB's Vector Search extensions to store the embeddings.

## 4. Infrastructure (Docker Compose)
- Create a `docker-compose.yml` file to spin up the required infrastructure.
- The docker-compose file must include:
  - A MongoDB container configured with vector search capabilities.
  - (Optional but recommended) An Ollama container to serve the local embedding model, ensuring a fully local development experience by default.

## 5. Documentation
- Create a detailed `README.md` explaining the application in depth. It should include:
  - An overview of the architecture and workflow.
  - Prerequisites and required environment variables.
  - Detailed instructions on how to start the infrastructure using Docker Compose.
  - Instructions on how to build and run the Go agent.
  - Explanations on how to configure and switch between Ollama and Vertex AI.
