package buckets

import (
	"crypto/rand"
	"encoding/hex"
	"path/filepath"

	"github.com/Tomdooo/spajz/internal/config"
)

func GetPath(bucket string) string {
	return filepath.Join(config.DataDir, bucket)
}

func generateRandomKey() (string, error) {
	bytes := make([]byte, 24) // 24 bajtů dává po hex encode 48 znaků
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "spajz_key_" + hex.EncodeToString(bytes), nil
}
