package cache

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/internal/db"
	"github.com/Tomdooo/spajz/internal/models"
	"github.com/Tomdooo/spajz/pkg/hashx"
)

const MAX_DELETE_RUNS = 10
const MAX_IN_DATABASE_FILE_SIZE = 300 * 1024

var cacheManager *CacheManager
var once sync.Once

func GetCacheManager() *CacheManager {
	once.Do(func() {
		cacheManager = &CacheManager{
			databaseManager:     *db.GetDatabaseManager(),
			bucketConfigManager: config.GetBucketConfigManager(),
		}
	})
	return cacheManager
}

type CacheManager struct {
	databaseManager     db.DatabaseManager
	bucketConfigManager *config.BucketConfigManager
}

func (m *CacheManager) SaveFile(ctx context.Context, fileContext *models.FileRequestContext, presetConfig *config.ImagePreset, mimeType string, data []byte) error {
	if len(data) > MAX_IN_DATABASE_FILE_SIZE {
		// TODO: think of different log / err handling
		slog.Warn("File was too big to save into cache, skipping.", "objectKey", fileContext.ObjectKey, "preset", presetConfig.Name)
		return nil
	}

	database, err := m.databaseManager.GetDatabase(fileContext.Bucket)
	if err != nil {
		return fmt.Errorf("getting bucket database: %w", err)
	}
	bucketConfig, err := m.bucketConfigManager.GetConfig(fileContext.Bucket)
	if err != nil {
		return fmt.Errorf("getting bucket config: %w", err)
	}

	etag := hex.EncodeToString(hashx.HashMD5(data))

	// check for maximum size of bucket size, if needed clear the cache
	// TODO: maybe refactor to separate settings of cache sizes - in database vs disk
	isSpaceAvailable := false
	for run := 0; run <= MAX_DELETE_RUNS; run++ { // complete MAX_DELETE_RUNS + 1 to make final verification of available size
		sizeRow, err := database.Queries.GetInDatabaseCacheSize(ctx)
		if err != nil {
			return fmt.Errorf("getting maximum cache size: %w", err)
		}
		cacheSize := sizeRow.(int64)
		fileSize := int64(len(data))

		// on the last run just verify if there is available space or not
		if run == MAX_DELETE_RUNS {
			if cacheSize+fileSize < bucketConfig.Cache.MaxSizeBytes {
				isSpaceAvailable = true
			}
			break
		}

		if cacheSize+fileSize < bucketConfig.Cache.MaxSizeBytes {
			isSpaceAvailable = true
			break
		}

		deletedRows, err := database.Queries.DeleteOldestCachedWithBlob(ctx, int64(bucketConfig.Cache.CleanBatchSize))
		if err != nil {
			return fmt.Errorf("deleting oldest in-database cached files: %w", err)
		}
		count := len(deletedRows)

		if count != MAX_IN_DATABASE_FILE_SIZE { // TODO: check the magic here, wtf is happening?
			slog.Info("Deleted last batch of in-database cached files.", "deletedCount", count, "batchSize", bucketConfig.Cache.CleanBatchSize)
			if cacheSize+fileSize < bucketConfig.Cache.MaxSizeBytes {
				isSpaceAvailable = true
			}
			break
		} else {
			slog.Info("Deleted one batch of in-database cached files.", "batchSize", bucketConfig.Cache.CleanBatchSize)
		}
	}
	if !isSpaceAvailable {
		slog.Warn("Cache is still full after maximum delete runs, skipping cache write.", "bucket", fileContext.Bucket)
		return fmt.Errorf("saving file into in-database cache: %w", models.ErrNotEnoughSpace)
	}

	// TODO: save to disk or db logic
	params := db.InsertCacheParams{
		FileHash:         fileContext.ObjectHash,
		Preset:           presetConfig.Name,
		PresetConfigHash: presetConfig.ConfigHash,
		Data:             data,
		MimeType:         mimeType,
		FileSize:         int64(len(data)),
		Etag:             etag,
	}
	err = database.Queries.InsertCache(ctx, params)
	if err != nil {
		return fmt.Errorf("saving file into in-database cache: %w", err)
	}

	return nil
}

func (m *CacheManager) GetFile(ctx context.Context, fileContext *models.FileRequestContext, preset string) (*db.GetCachedRow, error) {
	database, err := m.databaseManager.GetDatabase(fileContext.Bucket)
	if err != nil {
		return nil, fmt.Errorf("getting bucket database: %w", err)
	}

	params := db.GetCachedParams{
		FileHash: fileContext.ObjectHash,
		Preset:   preset,
	}
	row, err := database.Queries.GetCached(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("getting cached file from database: %w: %w", models.ErrFileNotFound, err)
		}
		return nil, fmt.Errorf("getting cached file from database: %w", err)
	}
	return &row, nil
}

func (m *CacheManager) UpdateFileAccessTime(ctx context.Context, fileContext *models.FileRequestContext, preset string) error {
	database, err := m.databaseManager.GetDatabase(fileContext.Bucket)
	if err != nil {
		return fmt.Errorf("getting bucket database: %w", err)
	}

	params := db.UpdateAccessTimeParams{
		FileHash: fileContext.ObjectHash,
		Preset:   preset,
	}
	err = database.Queries.UpdateAccessTime(ctx, params)
	if err != nil {
		return fmt.Errorf("updating file access time in database: %w", err)
	}
	return nil
}

func (m *CacheManager) DeleteAllFilesByObjectHash(ctx context.Context, fileContext *models.FileRequestContext) error {
	database, err := m.databaseManager.GetDatabase(fileContext.Bucket)
	if err != nil {
		return fmt.Errorf("getting bucket database: %w", err)
	}
	err = database.Queries.DeleteCachedByObjectHash(ctx, fileContext.ObjectHash)
	if err != nil {
		return fmt.Errorf("deleting cached files by object hash: %w", err)
	}
	return nil
}

func (m *CacheManager) DeleteFile(ctx context.Context, fileContext *models.FileRequestContext, preset string) error {
	database, err := m.databaseManager.GetDatabase(fileContext.Bucket)
	if err != nil {
		return fmt.Errorf("getting bucket database: %w", err)
	}
	params := db.DeleteCacheItemParams{
		FileHash: fileContext.ObjectHash,
		Preset:   preset,
	}
	err = database.Queries.DeleteCacheItem(ctx, params)
	if err != nil {
		return fmt.Errorf("deleting cached file: %w", err)
	}
	return nil
}
