// internal/db/db.go

package db

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// Database represents a database interface.
type Database interface {
	Connect(uri string) error
	Disconnect() error
	Collection(name string, db string) *mongo.Collection
}
