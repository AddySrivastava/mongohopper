// internal/operation/update.go

package operation

import (
	"context"
	"dbhopper/schema"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UpdateOperation represents an update operation.
type UpdateOperation struct {
	Filter       bson.D
	UpdateFields bson.D
	Success      int64
	Failure      int64
}

func (op *UpdateOperation) Execute(ctx context.Context, collection *mongo.Collection, schemaMap schema.SchemaType) error {
	op.UpdateFields = bson.D{{Key: "$set", Value: op.UpdateFields}}

	_, err := collection.UpdateMany(ctx, op.Filter, op.UpdateFields)

	if err != nil {
		op.Failure++
		return err
	}

	op.Success++
	return nil
}
