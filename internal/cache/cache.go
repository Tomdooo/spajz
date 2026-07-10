package cache

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/internal/db"
	"github.com/Tomdooo/spajz/internal/models"
	"github.com/Tomdooo/spajz/pkg/hashx"
)

var (
	ErrBucketNotExist = errors.New("Bucket does not exist.")
	ErrFileNotExist   = errors.New("File does not exist.")
)

var cacheManager *CacheManager
var once sync.Once

func GetCacheManager() *CacheManager {
	once.Do(func() {
		cacheManager = &CacheManager{
			databaseManager: *db.GetDatabaseManager(),
		}
	})
	return cacheManager
}

type CacheManager struct {
	databaseManager db.DatabaseManager
}

func (m *CacheManager) SaveFile(ctx context.Context, fileContext *models.FileRequestContext, presetConfig *config.ImagePreset, mimeType string, data []byte) error {
	database, err := m.databaseManager.GetDatabase(fileContext.Bucket)
	if err != nil {
		if errors.Is(err, db.ErrBucketNotExist) {
			return ErrBucketNotExist
		}
	}

	etag := hex.EncodeToString(hashx.HashMD5(data))

	// TODO: check for maximum bucket size, if needed clear the cache
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
		return err
	}

	return nil
}

func (m *CacheManager) GetFile(ctx context.Context, fileContext *models.FileRequestContext, preset string) (*db.GetCachedRow, error) {
	database, err := m.databaseManager.GetDatabase(fileContext.Bucket)
	if err != nil {
		if errors.Is(err, db.ErrBucketNotExist) {
			return nil, ErrBucketNotExist
		}
	}

	params := db.GetCachedParams{
		FileHash: fileContext.ObjectHash,
		Preset:   preset,
	}
	row, err := database.Queries.GetCached(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFileNotExist
		}
		return nil, fmt.Errorf("failed to fetch cached file '%q' in variant '%q' from database for bucket '%q': %w", fileContext.Filename, preset, fileContext.Bucket, err)
	}
	return &row, nil
}

func (m *CacheManager) UpdateFileAccessTime(ctx context.Context, fileContext *models.FileRequestContext, preset string) error {
	database, err := m.databaseManager.GetDatabase(fileContext.Bucket)
	if err != nil {
		if errors.Is(err, db.ErrBucketNotExist) {
			return ErrBucketNotExist
		}
	}

	params := db.UpdateAccessTimeParams{
		FileHash: fileContext.ObjectHash,
		Preset:   preset,
	}
	err = database.Queries.UpdateAccessTime(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrFileNotExist
		}
		return fmt.Errorf("failed to update cached access time for %q@%q in bucket bucket %q: %w", fileContext.Filename, preset, fileContext.Bucket, err)
	}
	return nil
}

// TODO: delete
