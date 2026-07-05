package buckets

import (
	"path/filepath"

	"github.com/Tomdooo/spajz/internal/config"
)

func getDefaultConfig() BucketConfig {
	var storage = StorageConfig{
		MaxSize: 0,
	}
	var cache = CacheConfig{
		MaxSize: 0,
	}
	var cfg = BucketConfig{
		Storage: storage,
		Cache:   cache,
	}
	return cfg
}

func GetPath(bucket string) string {
	return filepath.Join(config.DataDir, bucket)
}
