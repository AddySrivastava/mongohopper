// internal/db/mongodb/mongodb.go

package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB implements the Database interface for MongoDB.
type MongoDB struct {
	client *mongo.Client
}

// Connect connects to MongoDB.
func (m *MongoDB) Connect(uri string) error {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}
	m.client = client
	return nil
}

// Disconnect disconnects from MongoDB.
func (m *MongoDB) Disconnect() error {
	if m.client == nil {
		return nil
	}
	if err := m.client.Disconnect(context.Background()); err != nil {
		return err
	}
	return nil
}

// Collection returns a MongoDB collection.
func (m *MongoDB) Collection(name string, db string) *mongo.Collection {
	return m.client.Database(db).Collection(name)
}

// NewMongoDB creates a new MongoDB instance.
func NewMongoDB() *MongoDB {
	return &MongoDB{}
}
