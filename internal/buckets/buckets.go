package buckets

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/Tomdooo/spajz/internal/config"
)

const CONFIG_FILE_NAME = "bucket.toml"

var ErrAlreadyExists = errors.New("Bucket already exists.")

func Create(bucket string) error {
	bucketDir := filepath.Join(config.DataDir, bucket)
	configFile := filepath.Join(bucketDir, CONFIG_FILE_NAME)

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

	defaultConfigToml, err := toml.Marshal(getDefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to marshal default bucket config for %q: %w", bucket, err)
	}

	if _, err := f.Write(defaultConfigToml); err != nil {
		return fmt.Errorf("failed to write bucket config to %q: %w", configFile, err)
	}

	return nil
}

func Exists(bucket string) (bool, error) {
	configFile := filepath.Join(config.DataDir, bucket, CONFIG_FILE_NAME)
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
	bucketPath := GetPath(bucket)
	err := os.RemoveAll(bucketPath)
	if err != nil {
		return err
	}
	return nil
}
