package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"

	"dbhopper/config"
	"dbhopper/operation"
	"dbhopper/schema"

	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var layout = "2006-01-02" // or define your format: "2006-01-02 15:04:05"

type OperationPlan struct {
	Id     string
	Filter bson.D
	Ratio  int
	Type   string
	Update bson.D
}

func loadJSONArrayFromFile(path string) ([]map[string]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed random generator once globally

	cfg := config.ParseConfig()

	clientOptions := options.Client().ApplyURI(cfg.URI)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(cfg.Database).Collection(cfg.Collection)

	jobSchema, err := schema.ParseSchema("schema.json")
	if err != nil {
		log.Fatalf("Failed to parse schema: %v", err)
	}

	var operationPlans []OperationPlan
	totalRatio := 0 // Precompute total ratio

	for _, op := range jobSchema.Operations {
		operationId := uuid.New()

		var filters []map[string]interface{}

		if op.FiterSource == "" && (op.Type == "find" || op.Type == "update") {
			panic("find or update operation must have filter source")
		}

		filters, err = loadJSONArrayFromFile(op.FiterSource)
		if err != nil {
			log.Fatalf("Failed to load filters: %v", err)
		}

		if err != nil {
			log.Fatalf("Failed to load updates: %v", err)
		}

		totalRatio += op.Ratio

		filterDoc := bson.D{}

		for i := range filters {
			for k, v := range filters[i] {
				filterDoc = append(filterDoc, bson.E{Key: k, Value: v})
			}

			//check if the AppendDate is true and then append the date filter in the filterDoc
			if op.AppendDate {

				startDateISO, err := time.Parse(layout, op.StartDate)
				if err != nil {
					panic("invalid start date format")
				}

				endDateISO, err := time.Parse(layout, op.EndDate)
				if err != nil {
					panic("invalid end date format")
				}

				filterDoc = append(filterDoc, bson.E{Key: op.AppendDateField, Value: bson.D{{Key: "$gte", Value: startDateISO}, {Key: "$lte", Value: endDateISO}}})
			}

			updateDoc := bson.D{bson.E{Key: "_mongohopper_update", Value: time.Now()}}

			operationPlans = append(operationPlans, OperationPlan{
				Id:     operationId.String(),
				Filter: filterDoc,
				Update: updateDoc,
				Ratio:  op.Ratio,
				Type:   op.Type,
			})

		}
	}

	runLoadTest(cfg, collection, operationPlans, totalRatio)
}

func runLoadTest(cfg config.Config, collection *mongo.Collection, operationPlans []OperationPlan, totalRatio int) {
	var wg sync.WaitGroup

	requestsPerWorker := cfg.Requests / cfg.Workers
	remainingRequests := cfg.Requests % cfg.Workers

	var (
		insertLatencies, findLatencies, updateLatencies, deleteLatencies []time.Duration
		insertCount, findCount, updateCount, deleteCount                 int
		mu                                                               sync.Mutex // Protect shared resources
	)

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
				ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)

				startTime := time.Now()
				op, err := selectOperation(ctx, &operationPlans, collection, totalRatio)
				if err != nil {
					log.Printf("Worker %d, Request %d operation selection error: %v", workerID, j, err)
					continue
				}

				latency := time.Since(startTime)

				mu.Lock()
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
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	stop <- true
	printStats(insertLatencies, findLatencies, updateLatencies, deleteLatencies, insertCount, findCount, updateCount, deleteCount)
}

func selectOperation(ctx context.Context, operationPlans *[]OperationPlan, collection *mongo.Collection, totalRatio int) (operation.Operation, error) {
	randVal := rand.Intn(totalRatio)
	currentRatio := 0

	for _, operation := range *operationPlans {
		currentRatio += operation.Ratio
		if randVal < currentRatio {
			return processOperation(ctx, collection, operation)
		}
	}
	return nil, fmt.Errorf("no operation selected")
}

func processOperation(ctx context.Context, collection *mongo.Collection, opPlan OperationPlan) (operation.Operation, error) {
	switch opPlan.Type {
	case "find":
		op := &operation.FindOperation{Filter: opPlan.Filter}
		return op, op.Execute(ctx, collection)
	case "update":
		op := &operation.UpdateOperation{Filter: opPlan.Filter, Update: opPlan.Update}
		return op, op.Execute(ctx, collection)
	case "delete":
		op := &operation.DeleteOperation{Filter: opPlan.Filter}
		return op, op.Execute(ctx, collection)
	case "insert":
		op := &operation.InsertOperation{}
		return op, op.Execute(ctx, collection)
	default:
		return nil, fmt.Errorf("unsupported operation type: %s", opPlan.Type)
	}
}

func printPeriodicStats(stop chan bool, insertLatencies, findLatencies, updateLatencies, deleteLatencies *[]time.Duration, insertCount, findCount, updateCount, deleteCount *int) {
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
	table.SetHeader([]string{"Operation", "Count", "Avg Latency", "50th %ile", "95th %ile", "99th %ile", "Max Latency", "Time"})

	addOperationStats(table, "Insert", insertLatencies, insertCount)
	addOperationStats(table, "Find", findLatencies, findCount)
	addOperationStats(table, "Update", updateLatencies, updateCount)
	addOperationStats(table, "Delete", deleteLatencies, deleteCount)

	table.Render()
}

func addOperationStats(table *tablewriter.Table, operationName string, latencies []time.Duration, count int) {
	now := time.Now()

	if len(latencies) == 0 {
		table.Append([]string{operationName, fmt.Sprintf("%d", count), "N/A", "N/A", "N/A", "N/A", "N/A", now.String()})
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
	percentile50 := latencies[len(latencies)/2]
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
		now.String(),
	})
}
