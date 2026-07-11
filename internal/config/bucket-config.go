package config

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

const BUCKET_CONFIG_FILE_NAME = "bucket.toml"

var (
	ErrBucketNotExist = errors.New("Bucket not exist.")
	ErrPresetNotExist = errors.New("Preset not exist.")
)

var bucketConfigManager *BucketConfigManager
var once sync.Once

type BucketConfigMap map[string]*BucketConfig
type BucketConfigManager struct {
	configMap BucketConfigMap
	mu        sync.RWMutex
}

func GetBucketConfigManager() *BucketConfigManager {
	once.Do(func() {
		bucketConfigManager = new(BucketConfigManager{})
	})
	return bucketConfigManager
}

func (m *BucketConfigManager) GetDefaultConfig() *BucketConfig {
	apiKeys := make(ApiKeys, 0)
	apiKeys = append(apiKeys, ApiKey{
		Name:         "default",
		Key:          "spajz-default-api-key",
		AllowReading: true,
		AllowUpload:  true,
		AllowDelete:  true,
	})
	bucketSection := BucketSection{
		AllowPublicReading: true,
		AllowPublicUpload:  false,
		AllowPublicDelete:  false,
		ApiKeys:            apiKeys,
	}
	cacheSection := CacheSection{
		// Enabled:        true,
		MaxSizeGB:      1,
		CleanBatchSize: 20,
	}
	return &BucketConfig{
		Bucket: bucketSection,
		Cache:  cacheSection,
	}
}

func (m *BucketConfigManager) LoadBucketConfigs() error {
	fmt.Println(DataDir)
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

	bucketConfig.Cache.MaxSizeBytes = int64(bucketConfig.Cache.MaxSizeGB) * 1024 * 1024 * 1024

	bucketConfig.processPresets()

	m.configMap[bucket] = bucketConfig

	return nil
}

func (m *BucketConfigManager) getConfig(bucket string) (*BucketConfig, error) {
	config := m.configMap[bucket]
	if config == nil {
		return nil, ErrBucketNotExist
	}
	return config, nil
}
func (m *BucketConfigManager) GetConfig(bucket string) (*BucketConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getConfig(bucket)
}

func (m *BucketConfigManager) BucketsCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.configMap)
}

func (m *BucketConfigManager) GetImagePreset(bucket, preset string) (*ImagePreset, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	bucketConfig, err := m.getConfig(bucket)
	if err != nil {
		return nil, err
	}
	presetConfig := bucketConfig.Presets.Image[preset]
	if presetConfig == nil {
		return nil, ErrPresetNotExist
	}
	return presetConfig, nil
}

func (m *BucketConfigManager) VerifyApiKey(bucket, key string) (valid bool, apiKeyConfig *ApiKey, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	bucketConfig, err := m.getConfig(bucket)
	if err != nil {
		return false, nil, err
	}
	var validApiKey ApiKey
	for _, apiKey := range bucketConfig.Bucket.ApiKeys {
		if apiKey.Key == key {
			validApiKey = apiKey
			break
		}
	}
	return true, &validApiKey, nil
}

func (m *BucketConfigManager) GetBucketList() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return slices.Collect(maps.Keys(m.configMap))
}

func GetBucketList() ([]string, error) {
	dirEntries, err := os.ReadDir(DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read buckets directory: %w", err)
	}

	buckets := []string{}
	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() || strings.HasPrefix(dirEntry.Name(), ".") {
			continue
		}

		configPath := GetBucketConfigPath(dirEntry.Name())
		_, err := os.Stat(configPath)
		if err != nil {
			if !os.IsNotExist(err) {
				// TODO: better error handling
				slog.Error("failed to read file stats to verify it's existance", "filePath", configPath, "error", err)
			}
			continue
		}

		buckets = append(buckets, dirEntry.Name())
	}
	return buckets, nil
}
