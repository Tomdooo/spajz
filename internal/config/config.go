package config

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

var DataDir string
var TempDir string

func Initialize() {
	DataDir = os.Getenv("SPAJZ_DATA")
	if DataDir == "" {
		log.Fatal("SPAJZ_DATA environment variable not set!")
	}

	TempDir = filepath.Join(DataDir, ".tmp")
	err := os.MkdirAll(filepath.Join(TempDir), 0o755)
	if err != nil {
		log.Fatal("Error creating data folder.")
	}

	if err := bucketConfigManager.LoadBucketConfigs(); err != nil {
		log.Fatal("failed to load bucket configs", "error", err)
	}

}
