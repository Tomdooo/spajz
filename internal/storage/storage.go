package storage

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Tomdooo/spajz/internal/buckets"
	"github.com/Tomdooo/spajz/internal/config"
)

const (
	STORAGE_DIR       = "storage"
	BLOB_FILENAME     = "file.blob"
	METADATA_FILENAME = "metadata.json"
)

var (
	ErrBucketNotExist = errors.New("Bucket doesn't exist.")
	ErrFileExist      = errors.New("File already exists.")
	ErrFileNotExist   = errors.New("File not exists.")
)

func Add(bucket string, filename string, r io.Reader) (*FileMeta, error) {
	// Verify bucket existance
	if bucketExists, err := buckets.Exists(bucket); err != nil {
		return nil, fmt.Errorf("Couldn't verify, if %q bucket exists: %w", bucket, err)
	} else if !bucketExists {
		return nil, ErrBucketNotExist
	}

	// Create temp file and calculate hash
	h := md5.New() // ETag hasher
	tee := io.TeeReader(r, h)

	tempFile, err := os.CreateTemp(config.TempDir, "spajz_file_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed creating temp file for %q in bucket %q: %w", filename, bucket, err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, tee); err != nil {
		return nil, fmt.Errorf("failed to write data to temp file %q for %q in bucket %q: %w", tempFile.Name(), filename, bucket, err)
	}
	tempFile.Close()

	etag := hex.EncodeToString(h.Sum(nil))

	// Create metadata file in destination directory
	hashedFilename := getFilenameHash(filename)

	fileDir := getFileDir(bucket, hashedFilename)
	filePath := filepath.Join(fileDir, BLOB_FILENAME)
	metadataPath := filepath.Join(fileDir, METADATA_FILENAME)

	if _, err := os.Stat(filePath); err == nil { // verify existance of the file in storage
		return nil, ErrFileExist
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to verify existance of %q in bucket %q: %w", filename, bucket, err)
	}

	if err := os.MkdirAll(fileDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directories for %q file in %q bucket: %w", filename, bucket, err)
	}

	fileStats, err := os.Stat(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to obtain temp file stats of %q in bucket %q: %w", filename, bucket, err)
	}

	contentType, err := detectContentType(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to detect content type of %q in bucket %q: %w", filename, bucket, err)
	}
	metadata := &FileMeta{
		Id:          hashedFilename,
		Bucket:      bucket,
		ObjectKey:   filename,
		Filename:    filepath.Base(filename),
		Size:        fileStats.Size(),
		ContentType: contentType,
		Ext:         filepath.Ext(filename),
		Etag:        etag,
	}
	metadataJson, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata for %q in bucket %q: %w", filename, bucket, err)
	}
	if err := os.WriteFile(metadataPath, metadataJson, 0o644); err != nil {
		return nil, fmt.Errorf("failed to create metadata file for %q in bucket %q: %w", filename, bucket, err)
	}

	// Copy temp file to it's destination
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return nil, fmt.Errorf("failed to copy temp file into it's final destination for %q in bucket %q: %w", filename, bucket, err)
	}

	return metadata, nil
}

func Get(bucket string, filename string) ([]byte, error) {
	hash := getFilenameHash(filename)
	fileDir := getFileDir(bucket, hash)
	filePath := filepath.Join(fileDir, BLOB_FILENAME)

	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotExist
		}
		return nil, fmt.Errorf("failed to verify existance of %q in bucket %q: %w", filename, bucket, err)
	}

	return file, nil
}

func GetMetadata(bucket string, filename string) (*FileMeta, error) {
	hash := getFilenameHash(filename)
	fileDir := getFileDir(bucket, hash)
	filePath := filepath.Join(fileDir, METADATA_FILENAME)

	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotExist
		}
		return nil, fmt.Errorf("failed to verify existance of %q in bucket %q: %w", filename, bucket, err)
	}

	var metadata *FileMeta
	if err := json.Unmarshal(file, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

func GetWithMetadata(bucket string, filename string) ([]byte, *FileMeta, error) {
	file, err := Get(bucket, filename)
	if err != nil {
		return nil, nil, err
	}
	metadata, err := GetMetadata(bucket, filename)
	if err != nil {
		return nil, nil, err
	}
	return file, metadata, nil
}

func Update() {
	// TODO: implement
}
func Delete(bucket string, filename string) error {
	// verify bucket existance
	if bucketExists, err := buckets.Exists(bucket); err != nil {
		return fmt.Errorf("Couldn't verify, if %q bucket exists: %w", bucket, err)
	} else if !bucketExists {
		return ErrBucketNotExist
	}

	// delete folder
	hash := getFilenameHash(filename)
	fileDir := getFileDir(bucket, hash)

	if err := os.RemoveAll(fileDir); err != nil {
		return err
	}
	return nil
}
