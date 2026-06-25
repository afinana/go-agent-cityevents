package db

import (
	"context"
	"fmt"
	"time"

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
	Link        string    `bson:"link"`
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
		{Key: "createSearchIndexes", Value: m.collection.Name()},
		{Key: "indexes", Value: []any{
			bson.D{
				{Key: "name", Value: "vector_index"},
				{Key: "definition", Value: bson.D{
					{Key: "mappings", Value: bson.D{
						{Key: "dynamic", Value: true},
						{Key: "fields", Value: bson.D{
							{Key: "embedding", Value: bson.D{
								{Key: "dimensions", Value: 768}, // For nomic-embed-text
								{Key: "similarity", Value: "cosine"},
								{Key: "type", Value: "knnVector"},
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
		{
			{Key: "$vectorSearch", Value: bson.D{
				{Key: "index", Value: "vector_index"},
				{Key: "path", Value: "embedding"},
				{Key: "queryVector", Value: queryEmbedding},
				{Key: "numCandidates", Value: limit * 10},
				{Key: "limit", Value: limit},
			}},
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "embedding", Value: 0},
				{Key: "score", Value: bson.D{{Key: "$meta", Value: "vectorSearchScore"}}},
			}},
		},
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
			Link:        r.Link,
		})
	}

	return domainEvents, nil
}

// CountDocuments returns the number of stored events in the collection.
func (m *MongoClient) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return m.collection.CountDocuments(ctx, filter, opts...)
}

// InsertEvent persists the domain Event into MongoDB.
func (m *MongoClient) InsertEvent(ctx context.Context, event *domain.Event) error {
	dbEvent := mongoEvent{
		SourceID:    event.SourceID,
		Title:       event.Title,
		Description: event.Description,
		Embedding:   event.Embedding,
		Link:        event.Link,
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

// Ping pings the MongoDB client connection.
func (m *MongoClient) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

// SaveQuery saves a search query to the database, updating the timestamp if it already exists.
func (m *MongoClient) SaveQuery(ctx context.Context, queryText string) error {
	col := m.collection.Database().Collection("queries")
	filter := bson.D{{Key: "query", Value: queryText}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "query", Value: queryText},
			{Key: "timestamp", Value: time.Now()},
		}},
	}
	opts := options.Update().SetUpsert(true)
	_, err := col.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetQueryHistory returns recent queries sorted by newest first.
func (m *MongoClient) GetQueryHistory(ctx context.Context, limit int) ([]domain.QueryHistoryItem, error) {
	col := m.collection.Database().Collection("queries")
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(int64(limit))
	cursor, err := col.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dbQueries []struct {
		Query     string    `bson:"query"`
		Timestamp time.Time `bson:"timestamp"`
	}
	if err := cursor.All(ctx, &dbQueries); err != nil {
		return nil, err
	}

	var history []domain.QueryHistoryItem
	for _, q := range dbQueries {
		history = append(history, domain.QueryHistoryItem{
			Query:     q.Query,
			Timestamp: q.Timestamp,
		})
	}
	return history, nil
}

// DeleteQuery removes a specific query from the history.
func (m *MongoClient) DeleteQuery(ctx context.Context, queryText string) error {
	col := m.collection.Database().Collection("queries")
	_, err := col.DeleteOne(ctx, bson.D{{Key: "query", Value: queryText}})
	return err
}

// ClearQueryHistory clears all stored queries.
func (m *MongoClient) ClearQueryHistory(ctx context.Context) error {
	col := m.collection.Database().Collection("queries")
	_, err := col.DeleteMany(ctx, bson.D{})
	return err
}
