// internal/operation/update.go

package operation

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UpdateOperation represents an update operation.
type UpdateOperation struct {
	Filter  bson.D
	Update  bson.D
	Success int64
	Failure int64
}

func (op *UpdateOperation) Execute(ctx context.Context, collection *mongo.Collection) error {
	op.Update = bson.D{{Key: "$set", Value: op.Update}}

	_, err := collection.UpdateMany(ctx, op.Filter, op.Update)

	if err != nil {
		op.Failure++
		return err
	}

	op.Success++
	return nil
}
