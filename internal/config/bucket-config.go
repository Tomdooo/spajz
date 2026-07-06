package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

var (
	ErrBucketNotExist = errors.New("Bucket not exist.")
	ErrPresetNotExist = errors.New("Preset not exist.")
)

var bucketConfigManager = new(BucketConfigManager{})

type BucketConfigMap map[string]*BucketConfig
type BucketConfigManager struct {
	configMap BucketConfigMap
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
	bucketsDir := filepath.Join(DataDir)
	dirEntries, err := os.ReadDir(bucketsDir)
	if err != nil {
		return fmt.Errorf("failed to read buckets directory: %w", err)
	}

	m.configMap = make(BucketConfigMap)

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() || strings.HasPrefix(dirEntry.Name(), ".") {
			continue
		}

		bucketConfigPath := filepath.Join(bucketsDir, dirEntry.Name(), "bucket.toml")
		// file, err := os.ReadFile(bucketConfigPath)
		// if err != nil {
		// 	if os.IsNotExist(err) {
		// 		slog.Info("folder '%s' is not bucket", dirEntry.Name())
		// 		continue
		// 	}
		// 	return fmt.Errorf("failed to read bucket config in folder '%s': %w", dirEntry.Name(), err)
		// }

		var bucketConfig *BucketConfig
		_, err := toml.DecodeFile(bucketConfigPath, &bucketConfig)
		if err != nil {
			if os.IsNotExist(err) {
				slog.Debug("folder is not bucket", "folder", dirEntry.Name())
				continue
			}
			return fmt.Errorf("failed to read bucket config in folder '%s': %w", dirEntry.Name(), err)
		}

		bucketConfig.Presets.Image = make(ImagePresetMap)
		for _, imagePreset := range bucketConfig.Presets.RawImagePresets {
			bucketConfig.Presets.Image[imagePreset.Name] = &imagePreset
		}
		bucketConfig.Presets.Video = make(VideoPresetMap)
		for _, videoPreset := range bucketConfig.Presets.RawVideoPresets {
			bucketConfig.Presets.Video[videoPreset.Name] = &videoPreset
		}

		m.configMap[dirEntry.Name()] = bucketConfig
	}

	slog.Info("successfully loaded buckets")
	return nil
}

func (m *BucketConfigManager) GetConfig(bucket string) (*BucketConfig, error) {
	config := m.configMap[bucket]
	if config == nil {
		return nil, ErrBucketNotExist
	}
	return config, nil
}

func (m *BucketConfigManager) GetImagePreset(bucket, preset string) (*ImagePreset, error) {
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
