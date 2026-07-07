package buckets

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/Tomdooo/spajz/internal/config"
)

var ErrAlreadyExists = errors.New("Bucket already exists.")
var ErrBucketNotExist = errors.New("Bucket does not exist.")

var bucketConfigManager = config.GetBucketConfigManager()

func Create(bucket string) error {
	bucketDir := config.GetBucketDir(bucket)
	configFile := config.GetBucketConfigPath(bucket)

	if err := os.MkdirAll(bucketDir, 0o755); err != nil {
		return fmt.Errorf("failed to create bucket directory at %q: %w", bucketDir, err)
	}

	// Open file - O_CREATE = create if not exists, O_EXCL = return error if exists, O_WRONLY = write only
	f, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("failed to create bucket config file at %q: %w", configFile, err)
	}
	defer f.Close()

	defaultConfigToml, err := toml.Marshal(bucketConfigManager.GetDefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to marshal default bucket config for %q: %w", bucket, err)
	}

	if _, err := f.Write(defaultConfigToml); err != nil {
		return fmt.Errorf("failed to write bucket config to %q: %w", configFile, err)
	}

	return nil
}

func Get() {
	// TODO: implement
}

func Exists(bucket string) (bool, error) {
	configFile := config.GetBucketConfigPath(bucket)
	_, err := os.Stat(configFile)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("failed to verify existance of bucket config file at %q: %w", configFile, err)
	}
}

func Delete(bucket string) error {
	bucketPath := config.GetBucketDir(bucket)
	_, err := os.Stat(bucketPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrBucketNotExist
		}
		return err
	}
	// TODO: if bucket is not empty, throw 409 BucketNotEmpty
	err = os.RemoveAll(bucketPath)
	if err != nil {
		return err
	}
	return nil
}
