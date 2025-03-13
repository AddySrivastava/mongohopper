// internal/schema/schema.go

package schema

import (
	crypto_rand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Operation represents the operations part of the schema (If needed).
type Operation struct {
	Ratio        int                      `json:"ratio"`
	Type         string                   `json:"type"`
	Fields       []map[string]interface{} `json:"fields"` // or map[string]string if always strings.
	UpdateFields []map[string]interface{} `json:"updates"`
}

// Schema represents the complete schema (If needed)
type SchemaType struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Operations []Operation         `json:"operations"`
}

// Property represents the properties of the schema.
type Property struct {
	BSONType    string        `json:"bsonType"`
	Description string        `json:"description"`
	Unique      bool          `json:"unique,omitempty"`
	Minimum     int64         `json:"minimum,omitempty"`
	Maximum     int64         `json:"maximum,omitempty"`
	Items       *Property     `json:"items,omitempty"`
	Values      []interface{} `json:"values,omitempty"`
}

// ParseSchema reads and unmarshals the JSON schema.
func ParseSchema(filePath string) (SchemaType, error) {

	var schemaConfig SchemaType
	schemaBytes, err := os.ReadFile(filePath)
	if err != nil {
		return SchemaType{}, err
	}

	err = json.Unmarshal(schemaBytes, &schemaConfig)
	if err != nil {
		return SchemaType{}, err
	}

	return schemaConfig, nil
}

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(crypto_rand.Reader, b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateValue generates a random value based on the schema's bsonType.
func generateValue(prop Property) interface{} {

	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	if prop.Unique {
		if prop.BSONType == "string" {
			randomString, err := generateRandomString(12)
			if err != nil {
				fmt.Println("Error generating random bytes:", err)
				return err
			}
			return fmt.Sprintf("random_%s", randomString)
		} else {
			return errors.New("unique type is only allowed with string")
		}
	}

	switch prop.BSONType {
	case "objectId":
		return primitive.NewObjectID()
	case "string":
		return fmt.Sprintf("random%d_bar%d", rand.Intn(1000), rand.Intn(1000))
	case "int":
		max := prop.Maximum
		min := prop.Minimum
		if min != 0 && max != 0 {
			return rand.Int63n(max-min+1) + int64(min)
		}
		return rand.Intn(10000)
	case "array":
		itemValue := generateValue(*prop.Items)
		return []interface{}{itemValue, itemValue} // Create an array of 2 elements.
	case "enum":
		randomIndex := rand.Intn(len(prop.Values))
		return prop.Values[randomIndex]
	default:
		return nil
	}
}

// generateJSONDocumentFromSchema generates a JSON document based on the schema.
func GenerateJSONDocumentFromSchema(schema SchemaType) ([]byte, error) {

	doc := make(map[string]interface{})
	for key, propMap := range schema.Properties {
		doc[key] = generateValue(propMap)
	}

	jsonData, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}
