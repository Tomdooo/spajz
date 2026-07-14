package config

import (
	"path/filepath"
)

var STORAGE_DIR = "storage"

func GetBucketDir(bucket string) string {
	return filepath.Join(DataDir, bucket)
}

func GetBucketConfigPath(bucket string) string {
	return filepath.Join(DataDir, bucket, BUCKET_CONFIG_FILE_NAME)
}

func GetStorageDir(bucket string) string {
	return filepath.Join(GetBucketDir(bucket), STORAGE_DIR)
}
