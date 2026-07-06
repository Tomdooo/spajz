package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

const BUCKET_CONFIG_FILE_NAME = "bucket.toml"

var (
	ErrBucketNotExist = errors.New("Bucket not exist.")
	ErrPresetNotExist = errors.New("Preset not exist.")
)

var bucketConfigManager = new(BucketConfigManager{})

type BucketConfigMap map[string]*BucketConfig
type BucketConfigManager struct {
	configMap BucketConfigMap
	mu        sync.RWMutex
}

func GetBucketConfigManager() *BucketConfigManager {
	return bucketConfigManager
}

func (m *BucketConfigManager) GetDefaultConfig() *BucketConfig {
	bucketSection := BucketSection{
		IsPublic: true,
	}
	cacheSection := CacheSection{
		Enabled:        true,
		MaxSizeGB:      1,
		CleanBatchSize: 20,
	}
	return &BucketConfig{
		Bucket: bucketSection,
		Cache:  cacheSection,
	}
}

func (m *BucketConfigManager) LoadBucketConfigs() error {
	dirEntries, err := os.ReadDir(DataDir)
	if err != nil {
		return fmt.Errorf("failed to read buckets directory: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.configMap = make(BucketConfigMap)

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() || strings.HasPrefix(dirEntry.Name(), ".") {
			continue
		}

		// bucketConfigPath := GetBucketConfigPath(dirEntry.Name())

		// var bucketConfig *BucketConfig
		// _, err := toml.DecodeFile(bucketConfigPath, &bucketConfig)
		// if err != nil {
		// 	if os.IsNotExist(err) {
		// 		slog.Debug("folder is not bucket", "folder", dirEntry.Name())
		// 		continue
		// 	}
		// 	return fmt.Errorf("failed to read bucket config in folder '%s': %w", dirEntry.Name(), err)
		// }

		// bucketConfig.ProcessPresets()

		err := m.loadBucket(dirEntry.Name())
		if err != nil {
			if errors.Is(err, ErrBucketNotExist) {
				slog.Debug("folder is not bucket", "folder", dirEntry.Name())
				continue
			}
			return err
		}
	}

	count := len(m.configMap)
	slog.Info("Buckets loaded successfully", "count", count)
	return nil
}

func (m *BucketConfigManager) LoadBucket(bucket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.loadBucket(bucket)
}

func (m *BucketConfigManager) loadBucket(bucket string) error {
	bucketConfigPath := GetBucketConfigPath(bucket)

	var bucketConfig *BucketConfig
	_, err := toml.DecodeFile(bucketConfigPath, &bucketConfig)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrBucketNotExist
		}
		return fmt.Errorf("failed to decode bucket config for '%s': %w", bucket, err)
	}

	bucketConfig.ProcessPresets()
	m.configMap[bucket] = bucketConfig

	return nil
}

func (m *BucketConfigManager) GetConfig(bucket string) (*BucketConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config := m.configMap[bucket]
	if config == nil {
		return nil, ErrBucketNotExist
	}
	return config, nil
}

func (m *BucketConfigManager) BucketsCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.configMap)
}

func (m *BucketConfigManager) GetImagePreset(bucket, preset string) (*ImagePreset, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	bucketConfig := m.configMap[bucket]
	if bucketConfig == nil {
		return nil, ErrBucketNotExist
	}
	presetConfig := bucketConfig.Presets.Image[preset]
	if presetConfig == nil {
		return nil, ErrPresetNotExist
	}
	return presetConfig, nil
}
