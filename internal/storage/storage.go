package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Tomdooo/spajz/internal/buckets"
	"github.com/Tomdooo/spajz/internal/cache"
	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/internal/models"
	"github.com/Tomdooo/spajz/pkg/hashx"
)

const (
	STORAGE_DIR       = "storage"
	BLOB_FILENAME     = "file.blob"
	METADATA_FILENAME = "metadata.json"
)

var (
	ErrBucketNotExist = errors.New("Bucket does not exist.")
	ErrFileExist      = errors.New("File already exists.")
	ErrFileNotExist   = errors.New("File does not exist.")
)

var cacheManager = cache.GetCacheManager()
var bucketConfigManager = config.GetBucketConfigManager()

func Add(fileContext *models.FileRequestContext, r io.Reader) (*FileMeta, error) {
	// Verify bucket existance
	if bucketExists, err := buckets.Exists(fileContext.Bucket); err != nil {
		return nil, fmt.Errorf("Couldn't verify, if %q bucket exists: %w", fileContext.Bucket, err)
	} else if !bucketExists {
		return nil, ErrBucketNotExist
	}

	// Create temp file and calculate hash
	h := md5.New() // ETag hasher
	tee := io.TeeReader(r, h)

	tempFile, err := os.CreateTemp(config.TempDir, "spajz_file_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed creating temp file for %q in bucket %q: %w", fileContext.ObjectKey, fileContext.Bucket, err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, tee); err != nil {
		return nil, fmt.Errorf("failed to write data to temp file %q for %q in bucket %q: %w", tempFile.Name(), fileContext.ObjectKey, fileContext.Bucket, err)
	}
	tempFile.Close()

	etag := hex.EncodeToString(h.Sum(nil))

	// Create metadata file in destination directory
	hashedFilename := GetObjectHash(fileContext.ObjectKey)

	fileDir := getFileDir(fileContext.Bucket, hashedFilename)
	filePath := filepath.Join(fileDir, BLOB_FILENAME)
	metadataPath := filepath.Join(fileDir, METADATA_FILENAME)

	if _, err := os.Stat(filePath); err == nil { // verify existance of the file in storage
		return nil, ErrFileExist
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to verify existance of %q in bucket %q: %w", fileContext.ObjectKey, fileContext.Bucket, err)
	}

	if err := os.MkdirAll(fileDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directories for %q file in %q bucket: %w", fileContext.ObjectKey, fileContext.Bucket, err)
	}

	fileStats, err := os.Stat(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to obtain temp file stats of %q in bucket %q: %w", fileContext.ObjectKey, fileContext.Bucket, err)
	}

	contentType, err := detectContentType(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to detect content type of %q in bucket %q: %w", fileContext.ObjectKey, fileContext.Bucket, err)
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
		return nil, fmt.Errorf("failed to marshal metadata for %q in bucket %q: %w", fileContext.ObjectKey, fileContext.Bucket, err)
	}
	if err := os.WriteFile(metadataPath, metadataJson, 0o644); err != nil {
		return nil, fmt.Errorf("failed to create metadata file for %q in bucket %q: %w", fileContext.ObjectKey, fileContext.Bucket, err)
	}

	// Copy temp file to it's destination
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return nil, fmt.Errorf("failed to copy temp file into it's final destination for %q in bucket %q: %w", fileContext.ObjectKey, fileContext.Bucket, err)
	}

	return metadata, nil
}

func Get(fileContext *models.FileRequestContext) ([]byte, error) {
	fileDir := getFileDir(fileContext.Bucket, fileContext.ObjectHash)
	filePath := filepath.Join(fileDir, BLOB_FILENAME)

	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotExist
		}
		return nil, fmt.Errorf("failed to verify existance of %q in bucket %q: %w", fileContext.ObjectKey, fileContext.Bucket, err)
	}

	return file, nil
}

func GetMetadata(fileContext *models.FileRequestContext) (*FileMeta, error) {
	fileDir := getFileDir(fileContext.Bucket, fileContext.ObjectHash)
	filePath := filepath.Join(fileDir, METADATA_FILENAME)

	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotExist
		}
		return nil, fmt.Errorf("failed to verify existance of %q in bucket %q: %w", fileContext.ObjectKey, fileContext.Bucket, err)
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
		return nil, nil, err
	}
	metadata, err := GetMetadata(fileContext)
	if err != nil {
		return nil, nil, err
	}
	return file, metadata, nil
}

// NOTE: returns only partial metadata to satisfy HTTP response headers
func GetPresetVariant(ctx context.Context, fileContext *models.FileRequestContext, preset string) ([]byte, *FileMeta, bool, error) {
	// get image preset and work with that for ensured consistency
	presetConfig, err := bucketConfigManager.GetImagePreset(fileContext.Bucket, preset)
	if err != nil {
		if errors.Is(err, config.ErrBucketNotExist) {
			return nil, nil, false, ErrBucketNotExist
		}
		return nil, nil, false, err
	}
	// get image variant from cache, if exists return
	cached, err := cacheManager.GetFile(ctx, fileContext, preset)
	if err != nil && !errors.Is(err, cache.ErrFileNotExist) {
		if errors.Is(err, cache.ErrBucketNotExist) {
			return nil, nil, false, ErrBucketNotExist
		}
	}
	fmt.Println(cached.PresetConfigHash)
	fmt.Println(presetConfig.ConfigHash)
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
	file, err := imageGenerator.CreatePresetVariant(fileContext, presetConfig)
	if err != nil {
		return nil, nil, false, err
	}
	fileMeta := &FileMeta{
		ContentType: http.DetectContentType(file), // ?: Maybe do not detect, but hardcode
		Size:        int64(len(file)),
		Etag:        hex.EncodeToString(hashx.HashMD5(file)),
	}

	// save image variant to cache in separate goroutine
	go func(ctx context.Context, fileContext *models.FileRequestContext, presetConfig *config.ImagePreset, fileMeta *FileMeta, data []byte) {
		err := cacheManager.SaveFile(ctx, fileContext, presetConfig, fileMeta.ContentType, data)
		if err != nil {
			slog.Error("Failed to save file into database cache.",
				"bucket", fileContext.Bucket,
				"objectKey", fileContext.ObjectKey,
				"preset", presetConfig.Name,
				"error", err)
		}
		// fileRow, err := cacheManager.GetFile(ctx, fileContext, presetConfig.Name)
		// fmt.Println(err)
		// fmt.Println(fileRow)
	}(context.Background(), fileContext, presetConfig, fileMeta, file)

	return file, fileMeta, false, nil
}

func Delete(fileConfig *models.FileRequestContext) error {
	// verify bucket existance
	if bucketExists, err := buckets.Exists(fileConfig.Bucket); err != nil {
		return fmt.Errorf("Couldn't verify, if %q bucket exists: %w", fileConfig.Bucket, err)
	} else if !bucketExists {
		return ErrBucketNotExist
	}

	// TODO: verify file existance

	// delete folder
	fileDir := getFileDir(fileConfig.Bucket, fileConfig.ObjectHash)

	if err := os.RemoveAll(fileDir); err != nil {
		return err
	}
	return nil
}
