// internal/operation/insert.go

package operation

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// InsertOperation represents an insert operation.
type InsertOperation struct {
	Success int64
	Failure int64
}

func (op *InsertOperation) Execute(ctx context.Context, collection *mongo.Collection) error {
	// doc, _ := schema.GenerateJSONDocumentFromSchema(schemaMap)
	// var schema map[string]interface{}
	// err := json.Unmarshal([]byte(doc), &schema)
	// if err != nil {
	// 	op.Failure++
	// 	fmt.Println("Error unmarshaling JSON schema:", err)
	// 	return err
	// }
	// _, err = collection.InsertOne(ctx, schema)
	// if err != nil {
	// 	op.Failure++
	// 	fmt.Println(err)
	// }
	// op.Success++
	return nil
}
