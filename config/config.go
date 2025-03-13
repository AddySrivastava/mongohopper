// internal/config/config.go

package config

import (
	"flag"
)

// Config holds the application configuration.
type Config struct {
	URI            string
	Database       string
	Collection     string
	Workers        int
	Requests       int
	ReadPreference string
}

// ParseConfig parses command-line flags and returns a Config.
func ParseConfig() Config {
	uri := flag.String("uri", "mongodb://localhost:27017", "MongoDB connection URI")
	dbName := flag.String("db", "testdb", "Database name")
	collectionName := flag.String("collection", "testcollection", "Collection name")
	numWorkers := flag.Int("workers", 10, "Number of concurrent workers")
	numRequests := flag.Int("requests", 1000, "Total number of requests")
	readPreference := flag.String("readPreference", "primary", "Read Preference (eg: primary,secondaryPreferred)")

	flag.Parse()

	return Config{
		URI:            *uri,
		Database:       *dbName,
		Collection:     *collectionName,
		Workers:        *numWorkers,
		Requests:       *numRequests,
		ReadPreference: *readPreference,
	}
}
