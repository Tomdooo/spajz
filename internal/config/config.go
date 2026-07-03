package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var DataDir string
var TempDir string

func init() {
	DataDir = os.Getenv("STOROS_DATA")
	if DataDir == "" {
		fmt.Println("STOROS_DATA environment variable not set!")
		os.Exit(1)
	}

	TempDir = filepath.Join(DataDir, ".tmp")
	err := os.MkdirAll(filepath.Join(TempDir), 0o755)
	if err != nil {
		fmt.Println("Error creating data folder.")
		os.Exit(1)
	}

}
