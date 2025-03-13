# Mongohopper

## **Overview**
mongohopper performs **concurrent read and write operations** on a MongoDB collection using worker-based execution.
It allows configuring **MongoDB connection settings, concurrency levels, and operation types** via command-line arguments and a JSON schema file.

## **Features**
- Configurable **MongoDB URI, database, and collection** settings
- Supports **concurrent workers** for parallel execution
- Allows **custom read preferences** (e.g., `primary`, `secondaryPreferred`)
- Uses **schema.json** to define **data constraints and automated operations**
- Supports **find, insert, and update** operations with adjustable ratios
- Implements **Role-Based Access Control (RBAC)** for restricted views

---

## **Installation**

### **1️⃣ Clone the Repository**
```sh
git clone https://github.com/AddySrivastava/mongohopper.git
cd mongohopper
```

### **2️⃣ Install Dependencies**
```sh
go mod tidy
```

### **3️⃣ Build the Application**
```sh
go build -o mongohopper
```
For **Windows**, use:
```sh
go build -o mongohopper.exe
```

### **4️⃣ Run the Application**
```sh
./mongohopper -uri="mongodb://localhost:27017" -db="testdb" -collection="users" -workers=10 -requests=1000 -readPreference="primary"
```
For **Windows**:
```sh
mongohopper.exe -uri="mongodb://localhost:27017" -db="testdb" -collection="users" -workers=10 -requests=1000 -readPreference="primary"
```

---

## **Configuration**

### **Command-Line Arguments**
| Flag             | Default Value                 | Description                                  |
|-----------------|-----------------------------|----------------------------------------------|
| `-uri`         | `mongodb://localhost:27017`   | MongoDB connection URI                      |
| `-db`          | `testdb`                      | Database name                               |
| `-collection`  | `testcollection`              | Collection name                             |
| `-workers`     | `10`                          | Number of concurrent workers                |
| `-requests`    | `1000`                        | Total number of requests                    |
| `-readPreference` | `primary`                  | Read preference (`primary`, `secondaryPreferred`, etc.) |

---

## **Schema Configuration (schema.json)**
The tool reads a **JSON schema file (`schema.json`)**, defining:
- **Field types & constraints** (e.g., unique fields, number ranges).
- **Automated operations** (insert, update, find) with execution ratios.

### **Example schema.json**
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
```

### **How Operations Work**
- **Find Operations (25%)**: Queries documents matching specified fields.
- **Insert Operations (25%)**: Inserts new documents into the collection.
- **Update Operations (50%)**: Updates specific fields for matching documents.

---

## **Testing**
To run tests, use:
```sh
go test ./...
```

---

## **Cross-Platform Build**
To build for different platforms:

### **Linux**
```sh
GOOS=linux GOARCH=amd64 go build -o mongohopper-linux
```

### **Windows**
```sh
GOOS=windows GOARCH=amd64 go build -o mongohopper.exe
```

### **macOS**
```sh
GOOS=darwin GOARCH=amd64 go build -o mongohopper-mac
```


