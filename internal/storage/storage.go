package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Tomdooo/storos/internal/buckets"
	"github.com/Tomdooo/storos/internal/config"
	"github.com/Tomdooo/storos/pkg/hashx"
)

const STORAGE_DIR = "storage"
const BLOB_FILENAME = "file.blob"
const METADATA_FILENAME = "metadata.json"

var (
	ErrBucketNotExist = errors.New("Bucket doesn't exist.")
	ErrFileExist      = errors.New("File already exists.")
)

func getStorageDir(bucket string) string {
	return filepath.Join(buckets.GetPath(bucket), STORAGE_DIR)
}

func Add(bucket string, r io.Reader, filename string) error {
	if bucketExists, err := buckets.Exists(bucket); err != nil {
		return fmt.Errorf("Couldn't verify, if %q bucket exists: %w", bucket, err)
	} else if !bucketExists {
		return ErrBucketNotExist
	}

	// Create temp file and calculate hash
	h := sha256.New()
	tee := io.TeeReader(r, h)

	tempFile, err := os.CreateTemp(config.TempDir, "storos_file_*.tmp")
	if err != nil {
		return fmt.Errorf("failed creating temp file for %q in bucket %q: %w", filename, bucket, err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, tee); err != nil {
		return fmt.Errorf("failed to write data to temp file %q for %q in bucket %q: %w", tempFile.Name(), filename, bucket, err)
	}
	tempFile.Close()

	etag := hex.EncodeToString(h.Sum(nil))

	// Create metadata file in destination directory
	hashedFilename := hex.EncodeToString(hashx.HashSHA256([]byte(filename)))

	storageDir := getStorageDir(bucket)
	fileDir := filepath.Join(storageDir, hashedFilename[:2], hashedFilename[2:4], hashedFilename)
	filePath := filepath.Join(fileDir, BLOB_FILENAME)
	metadataPath := filepath.Join(fileDir, METADATA_FILENAME)

	if _, err := os.Stat(filePath); err == nil { // verify existance of the file in storage
		return ErrFileExist
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to verify existance of %q in bucket %q: %w", filename, bucket, err)
	}

	if err := os.MkdirAll(fileDir, 0o755); err != nil {
		return fmt.Errorf("failed to create storage directories for %q file in %q bucket: %w", filename, bucket, err)
	}

	fileStats, err := os.Stat(tempFile.Name())
	if err == nil {
		return fmt.Errorf("failed to obtain temp file stats of %q in bucket %q: %w", filename, bucket, err)
	}
	metadata := FileMeta{
		Filename: filename,
		Ext:      filepath.Ext(filename),
		Size:     fileStats.Size(),
		Etag:     etag,
	}
	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata for %q in bucket %q: %w", filename, bucket, err)
	}
	if err := os.WriteFile(metadataPath, metadataJson, 0o644); err != nil {
		return fmt.Errorf("failed to create metadata file for %q in bucket %q: %w", filename, bucket, err)
	}

	// Copy temp file to it's destination
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return fmt.Errorf("failed to copy temp file into it's final destination for %q in bucket %q: %w", filename, bucket, err)
	}

	return nil
}
func Update() {}
func Delete() {}
