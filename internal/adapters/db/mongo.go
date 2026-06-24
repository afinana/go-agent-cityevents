package db

import (
	"context"
	"fmt"

	"go-agent-cityevents/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoEvent represents the MongoDB BSON document layout for an event.
type mongoEvent struct {
	ID          string    `bson:"_id,omitempty"`
	SourceID    string    `bson:"source_id"`
	Title       string    `bson:"title"`
	Description string    `bson:"description"`
	Embedding   []float32 `bson:"embedding,omitempty"`
}

// MongoClient implements ports.EventRepository using MongoDB.
type MongoClient struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMongoClient establishes a connection to MongoDB and returns a MongoClient instance.
func NewMongoClient(ctx context.Context, uri, dbName, colName string) (*MongoClient, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	collection := client.Database(dbName).Collection(colName)

	return &MongoClient{
		client:     client,
		collection: collection,
	}, nil
}

// EnsureVectorIndex creates the vector search index on the 'embedding' field.
func (m *MongoClient) EnsureVectorIndex(ctx context.Context) error {
	cmd := bson.D{
		{"createSearchIndexes", m.collection.Name()},
		{"indexes", []bson.D{
			{
				{"name", "vector_index"},
				{"definition", bson.D{
					{"mappings", bson.D{
						{"dynamic", true},
						{"fields", bson.D{
							{"embedding", bson.D{
								{"dimensions", 768}, // For nomic-embed-text
								{"similarity", "cosine"},
								{"type", "knnVector"},
							}},
						}},
					}},
				}},
			},
		}},
	}
	err := m.client.Database(m.collection.Database().Name()).RunCommand(ctx, cmd).Err()
	if err != nil {
		fmt.Printf("Warning: couldn't auto-create vector index (it may already exist): %v\n", err)
	}
	return nil
}

// SearchEvents uses $vectorSearch to find the most semantically similar events.
func (m *MongoClient) SearchEvents(ctx context.Context, queryEmbedding []float32, limit int) ([]*domain.Event, error) {
	pipeline := mongo.Pipeline{
		{{
			"$vectorSearch", bson.D{
				{"index", "vector_index"},
				{"path", "embedding"},
				{"queryVector", queryEmbedding},
				{"numCandidates", limit * 10},
				{"limit", limit},
			},
		}},
		{{
			"$project", bson.D{
				{"embedding", 0},
				{"score", bson.D{{"$meta", "vectorSearchScore"}}},
			},
		}},
	}

	cursor, err := m.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []mongoEvent
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	var domainEvents []*domain.Event
	for _, r := range results {
		domainEvents = append(domainEvents, &domain.Event{
			ID:          r.ID,
			SourceID:    r.SourceID,
			Title:       r.Title,
			Description: r.Description,
		})
	}

	return domainEvents, nil
}

// InsertEvent persists the domain Event into MongoDB.
func (m *MongoClient) InsertEvent(ctx context.Context, event *domain.Event) error {
	dbEvent := mongoEvent{
		SourceID:    event.SourceID,
		Title:       event.Title,
		Description: event.Description,
		Embedding:   event.Embedding,
	}
	if event.ID != "" {
		dbEvent.ID = event.ID
	}

	_, err := m.collection.InsertOne(ctx, dbEvent)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	return nil
}

// Close closes the MongoDB client connection.
func (m *MongoClient) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
