package storage

import (
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Tomdooo/spajz/internal/buckets"
	"github.com/Tomdooo/spajz/pkg/hashx"
)

func getStorageDir(bucket string) string {
	return filepath.Join(buckets.GetPath(bucket), STORAGE_DIR)
}

func getFilenameHash(filename string) string {
	hash := hex.EncodeToString(hashx.HashSHA256([]byte(filename)))
	return hash
}

func getFileDir(bucket string, hash string) string {
	storageDir := getStorageDir(bucket)
	return filepath.Join(storageDir, hash[:2], hash[2:4], hash)
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
