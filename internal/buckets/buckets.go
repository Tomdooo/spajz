package buckets

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/internal/db"
	"github.com/Tomdooo/spajz/internal/models"
)

var bucketConfigManager = config.GetBucketConfigManager()
var databaseManager = db.GetDatabaseManager()

func Create(bucket string) error {
	bucketDir := config.GetBucketDir(bucket)
	configFile := config.GetBucketConfigPath(bucket)

	if err := os.MkdirAll(bucketDir, 0o755); err != nil {
		return fmt.Errorf("creating bucket directory: %w", err)
	}

	// Open file - O_CREATE = create if not exists, O_EXCL = return error if exists, O_WRONLY = write only
	f, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return models.ErrBucketAlreadyExists
		}
		return fmt.Errorf("opening bucket config file: %w", err)
	}
	defer f.Close()

	defaultConfigToml, err := toml.Marshal(bucketConfigManager.GetDefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to marshal default bucket config: %w", err)
	}

	if _, err := f.Write(defaultConfigToml); err != nil {
		return fmt.Errorf("writing default config into file: %w", err)
	}

	// load bucket
	bucketConfigManager.LoadBucket(bucket)
	databaseManager.InitDatabase(bucket)

	return nil
}

type BucketEntry struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func Get() ([]BucketEntry, error) {
	bucketList := bucketConfigManager.GetBucketList()
	bucketEntries := []BucketEntry{}
	for _, bucket := range bucketList {
		createdAt, err := bucketConfigManager.GetCreatedAt(bucket)
		if err != nil {
			return nil, fmt.Errorf("getting bucket (%s) created at value: %w", bucket, err)
		}

		bucketEntry := BucketEntry{
			Name:      bucket,
			CreatedAt: createdAt,
		}
		bucketEntries = append(bucketEntries, bucketEntry)
	}
	return bucketEntries, nil
}

func Exists(bucket string) (bool, error) {
	configFile := config.GetBucketConfigPath(bucket)
	_, err := os.Stat(configFile)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("verifying existance of bucket config file: %w", err)
	}
}

func Delete(bucket string) error {
	bucketPath := config.GetBucketDir(bucket)
	_, err := os.Stat(bucketPath)
	if err != nil {
		if os.IsNotExist(err) {
			return models.ErrBucketNotFound
		}
		return fmt.Errorf("verifying existance of bucket: %w", err)
	}
	// TODO: if bucket is not empty, throw 409 BucketNotEmpty

	bucketConfigManager.UnloadBucket(bucket)
	databaseManager.DestroyDatabase(bucket)

	err = os.RemoveAll(bucketPath)
	if err != nil {
		return fmt.Errorf("deleting bucket folder: %w", err)
	}
	return nil
}
