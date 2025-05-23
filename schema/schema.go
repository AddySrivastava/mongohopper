// internal/schema/schema.go

package schema

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
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
	Ratio           int    `json:"ratio"`
	Type            string `json:"type"`
	FiterSource     string `json:"fiterSource"` // or map[string]string if always strings.
	AppendDate      bool   `json:"appendDate"`
	AppendDateField string `json:"appendDateField"`
	StartDate       string `json:"startDate"`
	EndDate         string `json:"endDate"`
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
	Monotonic   bool          `json:"monotonic,omitempty"`
	Minimum     int64         `json:"minimum,omitempty"`
	Maximum     int64         `json:"maximum,omitempty"`
	Items       *Property     `json:"items,omitempty"`
	Values      []interface{} `json:"values,omitempty"`
	currentMax  interface{}
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

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(crypto_rand.Reader, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// generateValue generates a random value based on the schema's bsonType.
func generateValue(prop Property) interface{} {

	/*
		If monotonic we need to keep track of minimum and currenMax and generate a randomValue for that field for find ops accordingly
	*/

	if prop.Monotonic {
		return GetNextMonotonicValue(prop)
	}

	if prop.Unique {
		if prop.BSONType == "string" {
			value, err := generateRandomBytes(12)

			if err != nil {
				fmt.Printf("error = %s", err)
			}

			randomString := hex.EncodeToString(value)

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
		if prop.Minimum != 0 && prop.Maximum != 0 {
			return rand.Int63n(prop.Maximum-prop.Minimum+1) + int64(prop.Minimum)
		}
		return rand.Intn(10000)
	case "array":
		itemValue := generateValue(*prop.Items)
		return []interface{}{itemValue, itemValue} // Create an array of 2 elements.
	case "enum":
		randomIndex := rand.Intn(len(prop.Values))
		return prop.Values[randomIndex]
	case "date":
		return time.Now().UTC()
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

func bytesToInt64(b []byte) int64 {
	if len(b) < 8 {
		// Pad with zeros if the slice is less than 8 bytes
		padded := make([]byte, 8)
		copy(padded[8-len(b):], b)
		b = padded
	}
	return int64(binary.BigEndian.Uint64(b))
}

/*
Get monotonic value for the property
*/
func GetNextMonotonicValue(prop Property) interface{} {

	if prop.BSONType != "int64" {
		return nil
	}

	if prop.Maximum < prop.currentMax.(int64) {
		return prop.Maximum
	}

	value, err := generateRandomBytes(3)

	if err != nil {
		return nil
	}

	randomValue := time.Now().UnixNano() + bytesToInt64(value)

	if prop.currentMax == 0 {
		prop.Minimum = randomValue
	}

	prop.currentMax = randomValue

	return prop.currentMax

}
