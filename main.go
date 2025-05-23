package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"dbhopper/config"
	"dbhopper/operation"
	"dbhopper/schema"
)

// Policy struct for JSON file
type Values struct {
	Value string `json:"value"`
}

var Policies []string

// Function to load JSON data
func loadJSONData(filePath string, target interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(target)
}

func main() {
	cfg := config.ParseConfig()

	// Set client options
	clientOptions := options.Client().ApplyURI(cfg.URI)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(cfg.Database).Collection(cfg.Collection)

	docSchema, err := schema.ParseSchema("schema.json")
	if err != nil {
		log.Fatalf("Failed to parse schema: %v", err)
	}
	// Initialize policyNo values globally
	initializePolicies("values.json")

	runLoadTest(cfg, collection, docSchema)
}

// Function to initialize global policy list
func initializePolicies(dataPath string) {
	var policyRecords []Values
	if err := loadJSONData(dataPath, &policyRecords); err != nil {
		log.Fatal("Error loading data file:", err)
	}

	for _, p := range policyRecords {
		Policies = append(Policies, p.Value)
	}

	if len(Policies) == 0 {
		log.Fatal("No policies found in the JSON file!")
	}
}

func runLoadTest(cfg config.Config, collection *mongo.Collection, docSchema schema.SchemaType) {

	var wg sync.WaitGroup

	// To get the requests per worker
	requestsPerWorker := cfg.Requests / cfg.Workers

	// To get the remaining requests since requests can be an uneven distribution too
	remainingRequests := cfg.Requests % cfg.Workers

	insertLatencies := make([]time.Duration, 0, cfg.Requests/4)
	findLatencies := make([]time.Duration, 0, cfg.Requests/4)
	updateLatencies := make([]time.Duration, 0, cfg.Requests/4)
	deleteLatencies := make([]time.Duration, 0, cfg.Requests/4)

	insertCount, findCount, updateCount, deleteCount := 0, 0, 0, 0

	stop := make(chan bool)

	go printPeriodicStats(stop, &insertLatencies, &findLatencies, &updateLatencies, &deleteLatencies, &insertCount, &findCount, &updateCount, &deleteCount)

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			requests := requestsPerWorker
			if workerID == cfg.Workers-1 {
				requests += remainingRequests
			}
			for j := 0; j < requests; j++ {
				ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

				startTime := time.Now()

				op, err := selectOperation(ctx, docSchema, collection)

				if err != nil {
					log.Printf("Worker %d, Request %d operation selection error: %v", workerID, j, err)
					continue
				}

				latency := time.Since(startTime)

				switch op.(type) {
				case *operation.InsertOperation:
					insertCount++
					insertLatencies = append(insertLatencies, latency)
				case *operation.FindOperation:
					findCount++
					findLatencies = append(findLatencies, latency)
				case *operation.UpdateOperation:
					updateCount++
					updateLatencies = append(updateLatencies, latency)
				case *operation.DeleteOperation:
					deleteCount++
					deleteLatencies = append(deleteLatencies, latency)
				}
			}
		}(i)
	}

	wg.Wait()
	stop <- true
	printStats(insertLatencies, findLatencies, updateLatencies, deleteLatencies, insertCount, findCount, updateCount, deleteCount)
}

func selectOperation(ctx context.Context, docSchema schema.SchemaType, collection *mongo.Collection) (operation.Operation, error) {
	randVal := rand.Intn(100)
	operationList := docSchema.Operations

	totalRatio := calculateTotalRatio(operationList)

	currentRatio := 0

	for _, operation := range operationList {

		currentRatio += int(operation.Ratio)
		if currentRatio >= randVal {
			if randVal < currentRatio*100/totalRatio {
				return processOperation(ctx, operation, collection, docSchema)
			}
		}
	}
	return nil, nil
}

func calculateTotalRatio(operationList []schema.Operation) int {
	totalRatio := 0
	for _, operation := range operationList {
		totalRatio += int(operation.Ratio)
	}
	return totalRatio
}

func processOperation(ctx context.Context, operationMap schema.Operation, collection *mongo.Collection, docSchema schema.SchemaType) (operation.Operation, error) {

	rand.Seed(time.Now().UnixNano())

	queryValue := Policies[rand.Intn(len(Policies))]

	filterMap := bson.D{}
	for _, fieldMap := range operationMap.Fields {
		for key, _ := range fieldMap {
			filterMap = bson.D{{Key: key, Value: queryValue}}
		}
	}
	switch operationMap.Type {
	case "find":
		op := &operation.FindOperation{Filter: filterMap}
		return op, op.Execute(ctx, collection, docSchema)
	case "update":
		updateMap := bson.D{}
		for _, updateField := range operationMap.UpdateFields {
			for key, _ := range updateField {
				updateMap = bson.D{{Key: key, Value: queryValue}}
			}
		}
		op := &operation.UpdateOperation{Filter: filterMap, UpdateFields: updateMap}
		return op, op.Execute(ctx, collection, docSchema)
	case "delete":
		op := &operation.DeleteOperation{Filter: filterMap}
		return op, op.Execute(ctx, collection, docSchema)
	case "insert":
		op := &operation.InsertOperation{}
		return op, op.Execute(ctx, collection, docSchema)
	// case "findById":
	// 	op := &operation.FindByIdOperation{}
	// 	return op, op.Execute(ctx, collection, docSchema)
	default:
		return nil, fmt.Errorf("unsupported operation type: %s", operationMap.Type)
	}
}

func printPeriodicStats(stop chan bool, insertLatencies *[]time.Duration, findLatencies *[]time.Duration, updateLatencies *[]time.Duration, deleteLatencies *[]time.Duration, insertCount, findCount, updateCount, deleteCount *int) {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			printStats(*insertLatencies, *findLatencies, *updateLatencies, *deleteLatencies, *insertCount, *findCount, *updateCount, *deleteCount)
		case <-stop:
			return
		}
	}
}

func printStats(insertLatencies, findLatencies, updateLatencies, deleteLatencies []time.Duration, insertCount, findCount, updateCount, deleteCount int) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Operation", "Count", "Avg Latency", "50th %ile", "95th %ile", "99th %ile", "Max Latency"})

	addOperationStats(table, "Insert", insertLatencies, insertCount)
	addOperationStats(table, "Find", findLatencies, findCount)
	addOperationStats(table, "Update", updateLatencies, updateCount)
	addOperationStats(table, "Delete", deleteLatencies, deleteCount)

	table.Render()
}

func addOperationStats(table *tablewriter.Table, operationName string, latencies []time.Duration, count int) {
	if len(latencies) == 0 {
		table.Append([]string{operationName, fmt.Sprintf("%d", count), "N/A", "N/A", "N/A", "N/A", "N/A"})
		return
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	totalLatency := time.Duration(0)
	maxLatency := time.Duration(0)
	for _, latency := range latencies {
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	averageLatency := totalLatency / time.Duration(len(latencies))
	percentile50 := latencies[int(float64(len(latencies))*0.5)]
	percentile95 := latencies[int(float64(len(latencies))*0.95)]
	percentile99 := latencies[int(float64(len(latencies))*0.99)]

	table.Append([]string{
		operationName,
		fmt.Sprintf("%d", count),
		averageLatency.String(),
		percentile50.String(),
		percentile95.String(),
		percentile99.String(),
		maxLatency.String(),
	})
}
