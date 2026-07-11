package storage

import (
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/pkg/hashx"
)

func getStorageDir(bucket string) string {
	return filepath.Join(config.GetBucketDir(bucket), STORAGE_DIR)
}

func GetObjectHash(objectKey string) string {
	hash := hex.EncodeToString(hashx.HashSHA256([]byte(objectKey)))
	return hash
}

func getFileDir(bucket string, objectHash string) string {
	storageDir := getStorageDir(bucket)
	return filepath.Join(storageDir, objectHash[:2], objectHash[2:4], objectHash)
}

func detectContentType(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Pro detekci stačí prvních 512 bajtů
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	// io.EOF je v pořádku, pokud je soubor menší než 512 bajtů (např. mini ikona)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Analyzujeme pouze skutečně přečtené bajty
	return http.DetectContentType(buffer[:n]), nil
}
