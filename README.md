# MongoDB Concurrent Worker Tool  

## Overview  
This tool allows users to perform concurrent read and write operations on a **MongoDB collection** using multiple workers. The operations are defined in a **JSON schema configuration file**, allowing flexible execution of **find, insert, and update** operations with a predefined ratio.  

## Features  
- **Customizable MongoDB connection parameters** via command-line arguments.  
- **Concurrent workers** to perform operations in parallel.  
- **Configurable data schema (schema.json)** defining field types and constraints.  
- **Automated operations** (find, insert, update) based on specified ratios.  
- Supports **MongoDB Read Preferences** (e.g., `primary`, `secondaryPreferred`).  

## Usage  

### **Command-Line Arguments**  
The tool accepts the following flags:  

| Flag             | Default Value                 | Description                                  |
|-----------------|-----------------------------|----------------------------------------------|
| `-uri`         | `mongodb://localhost:27017`   | MongoDB connection URI                      |
| `-db`          | `testdb`                      | Database name                               |
| `-collection`  | `testcollection`              | Collection name                             |
| `-workers`     | `10`                          | Number of concurrent workers                |
| `-requests`    | `1000`                        | Total number of requests                    |
| `-readPreference` | `primary`                  | Read preference (`primary`, `secondaryPreferred`, etc.) |

### **Schema Configuration (schema.json)**  
The tool loads a schema definition from `schema.json`, which defines:  
- **Field constraints** (e.g., data types, unique fields).  
- **Operations** (insert, update, find) with execution ratios.  

#### **Example schema.json**  
```json
{
    "type": "object",
    "properties": {
      "_id": { "bsonType": "objectId", "description": "Unique identifier" },
      "username": { "bsonType": "string", "description": "Unique username", "unique": true },
      "email": { "bsonType": "string", "description": "Unique email address", "unique": true },
      "productId": { "bsonType": "int", "description": "Unique product ID", "unique": true },
      "age": { "bsonType": "int", "description": "Optional age", "minimum": 0, "maximum": 120 },
      "arrayField": {
        "bsonType": "array",
        "description": "Array field",
        "items": { "bsonType": "string" }
      }
    },
    "operations": [
      { "ratio": 50, "type": "update", "fields": [{"productId": 9283}], "updateFields": [{"foo": "bar"}] },
      { "ratio": 25, "type": "insert", "fields": [{"test": "value"}] },
      { "ratio": 25, "type": "find", "fields": [{"productId": 9299}] }
    ]
}
