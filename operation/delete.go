// internal/operation/delete.go

package operation

import (
	"context"
	"dbhopper/schema"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// DeleteOperation represents a delete operation.
type DeleteOperation struct {
	Filter  bson.D
	Success int64
	Failure int64
}

func (op *DeleteOperation) Execute(ctx context.Context, collection *mongo.Collection, schemaMap schema.SchemaType) error {
	_, err := collection.DeleteMany(ctx, op.Filter)

	if err != nil {
		op.Failure++
		return err
	}

	op.Success++
	return nil
}
