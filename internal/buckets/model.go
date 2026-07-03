package buckets

type BucketConfig struct {
	Storage StorageConfig `toml:"storage"`
	Cache   CacheConfig   `toml:"cache"`
}

type StorageConfig struct {
	MaxSize int `toml:"max_size"`
}

type CacheConfig struct {
	MaxSize int `toml:"max_size"`
}
