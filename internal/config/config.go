package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var DataDir string
var TempDir string

func init() {
	DataDir = os.Getenv("SPAJZ_DATA")
	if DataDir == "" {
		fmt.Println("SPAJZ_DATA environment variable not set!")
		os.Exit(1)
	}

	TempDir = filepath.Join(DataDir, ".tmp")
	err := os.MkdirAll(filepath.Join(TempDir), 0o755)
	if err != nil {
		fmt.Println("Error creating data folder.")
		os.Exit(1)
	}

	if err := LoadBucketConfigs(); err != nil {
		log.Fatalf("failed to load bucket configs: %w", err)
	}

	fmt.Println(Buckets["test"])

}
