// internal/operation/operation.go

package operation

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// Operation defines the interface for database operations.
type Operation interface {
	Execute(ctx context.Context, collection *mongo.Collection) error
}
