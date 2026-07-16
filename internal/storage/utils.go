package storage

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/pkg/hashx"
)

func GetObjectHash(objectKey string) string {
	hash := hex.EncodeToString(hashx.HashSHA256([]byte(objectKey)))
	return hash
}

func getFileDir(bucket string, objectHash string) string {
	storageDir := config.GetStorageDir(bucket)
	return filepath.Join(storageDir, objectHash[:2], objectHash[2:4], objectHash)
}

func detectContentType(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// Pro detekci stačí prvních 512 bajtů
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	// io.EOF je v pořádku, pokud je soubor menší než 512 bajtů (např. mini ikona)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("reading file: %w", err)
	}

	// Analyzujeme pouze skutečně přečtené bajty
	return http.DetectContentType(buffer[:n]), nil
}

// detect content type and if not fully detectable, replace with one from request
func determinContentType(filename string, contentTypeClaim string) (string, error) {
	determinedContentType, err := detectContentType(filename)
	if err != nil {
		return "", fmt.Errorf("detecting content type of temp file %s: %w", filename, err)
	}
	if strings.HasPrefix(determinedContentType, "text/plain") {
		// types that cannot be just simple plain text = is text, but user says it is binary
		isBinaryClaim := strings.HasPrefix(contentTypeClaim, "image/") ||
			strings.HasPrefix(contentTypeClaim, "video/") ||
			strings.HasPrefix(contentTypeClaim, "audio/") ||
			strings.HasPrefix(contentTypeClaim, "application/zip") ||
			strings.HasPrefix(contentTypeClaim, "application/pdf") ||
			strings.HasPrefix(contentTypeClaim, "application/octet-stream")

		if !isBinaryClaim && contentTypeClaim != "" {
			// user uploaded non-binary, we trust him
			determinedContentType = contentTypeClaim
		} else {
			// we dont trust the user or there is not more specific content type
			determinedContentType = "application/octet-stream"
		}
	}
	return determinedContentType, nil
}
