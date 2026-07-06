package config

import "path/filepath"

func GetBucketDir(bucket string) string {
	return filepath.Join(DataDir, bucket)
}

func GetBucketConfigPath(bucket string) string {
	return filepath.Join(DataDir, bucket, BUCKET_CONFIG_FILE_NAME)
}
