package config

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"

	"github.com/Tomdooo/spajz/pkg/hashx"
)

type ImagePresetMap map[string]*ImagePreset
type VideoPresetMap map[string]*VideoPreset

type BucketConfig struct {
	Bucket   BucketSection  `toml:"bucket"`
	Cache    CacheSection   `toml:"cache"`
	Presets  PresetsSection `toml:"presets"`
	Database *sql.DB
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
	RawImagePresets []ImagePreset `toml:"image"`
	Image           ImagePresetMap
	RawVideoPresets []VideoPreset `toml:"video"`
	Video           VideoPresetMap
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

type VideoPreset struct { // NOTE: video presets are not handeled yet
	Name        string `toml:"name"`
	Codec       string `toml:"codec"`
	Resolution  string `toml:"resolution"`
	Fps         int    `toml:"fps"`
	BitrateKbps int    `toml:"bitrate_kbps"`
}

func (c *BucketConfig) processPresets() error {
	c.Presets.Image = make(ImagePresetMap)
	for _, imagePreset := range c.Presets.RawImagePresets {
		// calculate config hash
		configJson, err := json.Marshal(imagePreset)
		if err != nil {
			return err
		}
		hash := hex.EncodeToString(hashx.HashSHA256(configJson))
		imagePreset.ConfigHash = hash

		// apend to image preset map
		c.Presets.Image[imagePreset.Name] = &imagePreset
	}

	c.Presets.Video = make(VideoPresetMap)
	for _, videoPreset := range c.Presets.RawVideoPresets {
		c.Presets.Video[videoPreset.Name] = &videoPreset
	}
	return nil
}
