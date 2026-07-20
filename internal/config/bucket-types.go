package config

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Tomdooo/spajz/internal/models"
	"github.com/Tomdooo/spajz/pkg/hashx"
	"github.com/Tomdooo/spajz/pkg/validatorx"
)

type ImagePresetMap map[string]*ImagePreset

// type VideoPresetMap map[string]*VideoPreset

type BucketConfig struct {
	Name      string
	Bucket    BucketSection  `toml:"bucket"`
	Cache     CacheSection   `toml:"cache"`
	Presets   PresetsSection `toml:"presets"`
	CreatedAt time.Time
	Database  *sql.DB
}

func (c *BucketConfig) ProcessPresets() error {
	return c.Presets.ProcessPresets(c.Name)
}

type BucketSection struct {
	// MaxFileSize int    `toml:"max_file_size_bytes"`
	AllowPublicUpload  bool    `toml:"allow_public_upload"`
	AllowPublicReading bool    `toml:"allow_public_reading"`
	AllowPublicDelete  bool    `toml:"allow_public_delete"`
	ApiKeys            ApiKeys `toml:"api_keys"`
}

type ApiKey struct {
	Name         string `toml:"name"`
	Key          string `toml:"key"`
	AllowUpload  bool   `toml:"allow_upload"`
	AllowReading bool   `toml:"allow_reading"`
	AllowDelete  bool   `toml:"allow_delete"`
}
type ApiKeys = []ApiKey

type CacheSection struct {
	// Enabled        bool    `toml:"enabled"`
	MaxSizeGB      float64 `toml:"max_size_gb"`
	MaxSizeBytes   int64
	CleanBatchSize int `toml:"clean_batch_size"`
}

type PresetsSection struct {
	mu              sync.RWMutex
	RawImagePresets []ImagePreset `toml:"image"`
	Image           ImagePresetMap
	// RawVideoPresets []VideoPreset `toml:"video"`
	// Video           VideoPresetMap
}

type ImagePreset struct {
	Name       string `toml:"name" json:"name"`
	Width      int    `toml:"width" json:"width"`
	Height     int    `toml:"height" json:"height"`
	Format     string `toml:"format" json:"format"`
	Enlarge    bool   `toml:"enlarge" json:"enlarge"`
	Embed      bool   `toml:"embed" json:"embed"`
	Quality    int    `toml:"quality" json:"quality"`
	ConfigHash string
}

// type VideoPreset struct { // NOTE: video presets are not handeled yet
// 	Name        string `toml:"name"`
// 	Codec       string `toml:"codec"`
// 	Resolution  string `toml:"resolution"`
// 	Fps         int    `toml:"fps"`
// 	BitrateKbps int    `toml:"bitrate_kbps"`
// }

func (c *PresetsSection) ProcessPresets(bucket string) error {
	c.mu.Lock() // ?: with presets refactor move into public method
	defer c.mu.Unlock()
	c.Image = make(ImagePresetMap)
	for _, imagePreset := range c.RawImagePresets {
		// validate preset name
		if !validatorx.PresetRegex.MatchString(imagePreset.Name) {
			slog.Error("Image preset name does not have valid format.", "bucket", bucket, "name", imagePreset.Name, "requiredFormat", validatorx.PresetRegex.String())
			continue
		}
		// calculate config hash
		configJson, err := json.Marshal(imagePreset)
		if err != nil {
			return fmt.Errorf("failed to marshal json: %w", err)
		}
		hash := hex.EncodeToString(hashx.HashSHA256(configJson))
		imagePreset.ConfigHash = hash

		// apend to image preset map
		c.Image[imagePreset.Name] = &imagePreset
	}

	// c.Presets.Video = make(VideoPresetMap)
	// for _, videoPreset := range c.Presets.RawVideoPresets {
	// 	c.Presets.Video[videoPreset.Name] = &videoPreset
	// }
	return nil
}

func (c *PresetsSection) GetImagePresetList() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	list := []string{}
	for _, preset := range c.Image {
		list = append(list, preset.Name)
	}
	return list
}

func (c *PresetsSection) GetImagePreset(preset string) (*ImagePreset, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.getImagePreset(preset)
}

func (c *PresetsSection) getImagePreset(preset string) (*ImagePreset, error) {
	presetConfig := c.Image[preset]
	if presetConfig == nil {
		return nil, models.ErrPresetNotFound
	}
	return presetConfig, nil
}
