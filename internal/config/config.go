package config

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

var DataDir string
var TempDir string
var MasterKey string

func Initialize() {
	// Get env variables
	DataDir = os.Getenv("SPAJZ_DATA")
	if DataDir == "" {
		log.Fatal("SPAJZ_DATA environment variable not set!")
	}

	MasterKey = os.Getenv("SPAJZ_MASTER_KEY")
	if MasterKey == "" {
		log.Warn("SPAJZ_MASTER_KEY environment variable is not set.")
	}

	// Compute temp dir path
	TempDir = filepath.Join(DataDir, ".tmp")
	err := os.MkdirAll(filepath.Join(TempDir), 0o755)
	if err != nil {
		log.Fatal("Error creating data folder: %w", err)
	}

}
