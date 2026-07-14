package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	"github.com/Tomdooo/spajz/internal/buckets"
	"github.com/Tomdooo/spajz/internal/cache"
	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/internal/models"
	"github.com/Tomdooo/spajz/pkg/hashx"
)

const (
	BLOB_FILENAME     = "file.blob"
	METADATA_FILENAME = "metadata.json"
)

var cacheManager = cache.GetCacheManager()
var bucketConfigManager = config.GetBucketConfigManager()

func Add(fileContext *models.FileRequestContext, r io.Reader) (*FileMeta, error) {
	// Verify bucket existance
	if bucketExists, err := buckets.Exists(fileContext.Bucket); err != nil {
		return nil, fmt.Errorf("verifying existance of bucket: %w", err)
	} else if !bucketExists {
		return nil, models.ErrBucketNotFound
	}

	// Create temp file and calculate hash
	h := md5.New() // ETag hasher
	tee := io.TeeReader(r, h)

	tempFile, err := os.CreateTemp(config.TempDir, "spajz_file_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	defer func() { // close temp file after unexpected earlier method end
		tempFileName := tempFile.Name()
		err := tempFile.Close()
		if err != nil && !errors.Is(err, os.ErrClosed) {
			slog.Error("Failed to close temp file after uploading.",
				"bucket", fileContext.Bucket,
				"objectKey", fileContext.ObjectKey,
				"tempFileName", tempFileName,
				"error", err)
		}
	}()
	defer func() { // delete temp file after method end
		tempFileName := tempFile.Name()
		err := os.Remove(tempFileName)
		if err != nil {
			slog.Error("Failed to delete temp file after uploading.",
				"bucket", fileContext.Bucket,
				"objectKey", fileContext.ObjectKey,
				"tempFileName", tempFileName,
				"error", err)
		}
	}()

	if _, err := io.Copy(tempFile, tee); err != nil {
		return nil, fmt.Errorf("failed to write data into temp file %s: %w", tempFile.Name(), err)
	}
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file %s: %w", tempFile.Name(), err)
	}

	etag := hex.EncodeToString(h.Sum(nil))

	// Create metadata file in destination directory
	hashedFilename := GetObjectHash(fileContext.ObjectKey)

	fileDir := getFileDir(fileContext.Bucket, hashedFilename)
	filePath := filepath.Join(fileDir, BLOB_FILENAME)
	metadataPath := filepath.Join(fileDir, METADATA_FILENAME)

	if _, err := os.Stat(filePath); err == nil { // verify existance of the file in storage
		return nil, models.ErrFileAlreadyExists
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("verifying existance of storage file: %w", err)
	}

	if err := os.MkdirAll(fileDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating storage directiories for file: %w", err)
	}

	fileStats, err := os.Stat(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("getting temp file stats (%s): %w", tempFile.Name(), err)
	}

	contentType, err := detectContentType(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("detecting content type of temp file %s: %w", tempFile.Name(), err)
	}
	metadata := &FileMeta{
		Id:          hashedFilename,
		Bucket:      fileContext.Bucket,
		ObjectKey:   fileContext.ObjectKey,
		Filename:    fileContext.Filename,
		Size:        fileStats.Size(),
		ContentType: contentType,
		Ext:         filepath.Ext(fileContext.Filename),
		Etag:        etag,
	}
	metadataJson, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling metadata for new storage file: %w", err)
	}
	if err := os.WriteFile(metadataPath, metadataJson, 0o644); err != nil {
		return nil, fmt.Errorf("creating metadata file for storage file: %w", err)
	}

	// Copy temp file to it's destination
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return nil, fmt.Errorf("copying temp file (%s) into it's final destination: %w", tempFile.Name(), err)
	}

	return metadata, nil
}

func Get(fileContext *models.FileRequestContext) ([]byte, error) {
	fileDir := getFileDir(fileContext.Bucket, fileContext.ObjectHash)
	filePath := filepath.Join(fileDir, BLOB_FILENAME)

	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, models.ErrFileNotFound
		}
		return nil, fmt.Errorf("reading storage file: %w", err)
	}

	return file, nil
}

func GetMetadata(fileContext *models.FileRequestContext) (*FileMeta, error) {
	fileDir := getFileDir(fileContext.Bucket, fileContext.ObjectHash)
	filePath := filepath.Join(fileDir, METADATA_FILENAME)

	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, models.ErrFileNotFound
		}
		return nil, fmt.Errorf("reading metadata of storage file: %w", err)
	}

	var metadata *FileMeta
	if err := json.Unmarshal(file, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

func GetWithMetadata(fileContext *models.FileRequestContext) ([]byte, *FileMeta, error) {
	file, err := Get(fileContext)
	if err != nil {
		return nil, nil, fmt.Errorf("getting storage file: %w", err)
	}
	metadata, err := GetMetadata(fileContext)
	if err != nil {
		return nil, nil, fmt.Errorf("getting metadata of storage file: %w", err)
	}
	return file, metadata, nil
}

// NOTE: returns only partial metadata to satisfy HTTP response headers
func GetPresetVariant(ctx context.Context, fileContext *models.FileRequestContext, preset string) (file []byte, fileMeta *FileMeta, isCacheHit bool, err error) {
	// get image preset and work with that for ensured consistency
	presetConfig, err := bucketConfigManager.GetImagePreset(fileContext.Bucket, preset)
	if err != nil {
		return nil, nil, false, fmt.Errorf("getting bucket image preset: %w", err)
	}
	// get image variant from cache, if exists with same preset config return
	cached, err := cacheManager.GetFile(ctx, fileContext, preset)
	if err != nil && !errors.Is(err, models.ErrFileNotFound) {
		return nil, nil, false, fmt.Errorf("getting cached file: %w", err)
	}
	if cached != nil && cached.PresetConfigHash == presetConfig.ConfigHash {
		fileMeta := &FileMeta{
			ContentType: cached.MimeType,
			Size:        cached.FileSize,
			Etag:        cached.Etag,
		}
		// update last accessed time
		go func(ctx context.Context, fileContext *models.FileRequestContext, preset string) {
			err := cacheManager.UpdateFileAccessTime(ctx, fileContext, preset)
			if err != nil {
				slog.Error("Failed to update accessed time of file in database.",
					"bucket", fileContext.Bucket,
					"objectKey", fileContext.ObjectKey,
					"preset", presetConfig.Name,
					"error", err)
			}
		}(context.Background(), fileContext, preset)
		return cached.Data, fileMeta, true, nil
	}

	// generate image variant
	file, err = imageGenerator.CreatePresetVariant(fileContext, presetConfig)
	if err != nil {
		return nil, nil, false, fmt.Errorf("creating image variant from preset: %w", err)
	}
	fileMeta = &FileMeta{
		ContentType: http.DetectContentType(file), // ?: Maybe do not detect, but hardcode
		Size:        int64(len(file)),
		Etag:        hex.EncodeToString(hashx.HashMD5(file)),
	}

	// save image variant to cache in separate goroutine
	go func(ctx context.Context, fileContext *models.FileRequestContext, presetConfig *config.ImagePreset, fileMeta *FileMeta, data []byte) {
		err := cacheManager.SaveFile(ctx, fileContext, presetConfig, fileMeta.ContentType, data)
		if err != nil {
			slog.Error("Failed to save image variant into cache.",
				"bucket", fileContext.Bucket,
				"objectKey", fileContext.ObjectKey,
				"preset", presetConfig.Name,
				"error", err)
		}
	}(context.Background(), fileContext, presetConfig, fileMeta, file)

	return file, fileMeta, false, nil
}

func Exists(fileConfig *models.FileRequestContext) (bool, error) {
	fileDir := getFileDir(fileConfig.Bucket, fileConfig.ObjectHash)
	metaFilePath := filepath.Join(fileDir, METADATA_FILENAME)

	_, err := os.Stat(metaFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("verifying existance of storage file: %w", err)
	}
	return true, nil
}

func Delete(ctx context.Context, fileConfig *models.FileRequestContext) error {
	// verify bucket existance
	if bucketExists, err := buckets.Exists(fileConfig.Bucket); err != nil {
		return fmt.Errorf("verifying bucket existance: %w", err)
	} else if !bucketExists {
		return models.ErrBucketNotFound
	}

	// verify file existance
	exists, err := Exists(fileConfig)
	if err != nil {
		return fmt.Errorf("verifying storage file existance: %w", err)
	}
	if !exists {
		return models.ErrFileNotFound
	}

	// cache Delete
	if err := cacheManager.DeleteAllFilesByObjectHash(ctx, fileConfig); err != nil {
		return fmt.Errorf("deleting all storage file cached variants: %w", err)
	}

	// delete folder
	fileDir := getFileDir(fileConfig.Bucket, fileConfig.ObjectHash)

	if err := os.RemoveAll(fileDir); err != nil {
		return fmt.Errorf("deleting storage file folder: %w", err)
	}

	parentDir := filepath.Dir(fileDir)
	fmt.Println(parentDir)
	if err := removeIfEmpty(parentDir); err != nil {
		return fmt.Errorf("cleaning parent directory %s: %w", parentDir, err)
	}

	grandParentDir := filepath.Dir(parentDir)
	fmt.Println(grandParentDir)
	if err := removeIfEmpty(grandParentDir); err != nil {
		return fmt.Errorf("cleaning grand parent directory %s: %w", parentDir, err)
	}

	return nil
}

// removeIfEmpty attempts to remove the specified directory.
// It returns nil if the deletion succeeds, if the directory does not exist,
// or if the directory cannot be removed because it is not empty.
// Any other filesystem errors (e.g., permission denied, I/O errors) are returned.
func removeIfEmpty(dir string) error {
	err := os.Remove(dir)
	if err == nil {
		return nil
	}

	// If the directory does not exist, our goal is already achieved.
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	// Unwrap the underlying system error from *fs.PathError to inspect the errno.
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		sysErr := pathErr.Err

		// ENOTEMPTY: standard Unix/Linux "directory not empty"
		// EEXIST: returned by some specific filesystems when a directory is not empty
		// Windows/Other: fallback string comparison check for cross-platform reliability
		if sysErr == syscall.ENOTEMPTY || sysErr == syscall.EEXIST || sysErr.Error() == "directory not empty" {
			return nil
		}
	}

	// Return any other critical errors (e.g., permission issues, read-only filesystem).
	return err
}
