package config

type ImagePresetMap map[string]*ImagePreset
type VideoPresetMap map[string]*VideoPreset

type BucketConfig struct {
	Bucket  BucketSection  `toml:"bucket"`
	Cache   CacheSection   `toml:"cache"`
	Presets PresetsSection `toml:"presets"`
}

type BucketSection struct {
	IsPublic bool `toml:"is_public"`
	// MaxFileSize int    `toml:"max_file_size_bytes"`
}

type CacheSection struct {
	Enabled        bool    `toml:"enabled"`
	MaxSizeGB      float64 `toml:"max_size_gb"`
	CleanBatchSize int     `toml:"clean_batch_size"`
}

type PresetsSection struct {
	RawImagePresets []ImagePreset `toml:"image"`
	Image           ImagePresetMap
	RawVideoPresets []VideoPreset `toml:"video"`
	Video           VideoPresetMap
}

type ImagePreset struct {
	Name    string `toml:"name"`
	Width   int    `toml:"width"`
	Height  int    `toml:"height"`
	Format  string `toml:"format"`
	Enlarge bool   `toml:"enlarge"`
	Quality int    `toml:"quality"`
}

type VideoPreset struct {
	Name        string `toml:"name"`
	Codec       string `toml:"codec"`
	Resolution  string `toml:"resolution"`
	Fps         int    `toml:"fps"`
	BitrateKbps int    `toml:"bitrate_kbps"`
}

// NOTE: does not uses mutex lock
func (c *BucketConfig) ProcessPresets() {
	c.Presets.Image = make(ImagePresetMap)
		for _, imagePreset := range c.Presets.RawImagePresets {
			c.Presets.Image[imagePreset.Name] = &imagePreset
		}
		c.Presets.Video = make(VideoPresetMap)
		for _, videoPreset := range c.Presets.RawVideoPresets {
			c.Presets.Video[videoPreset.Name] = &videoPreset
		}
}
