package buckets

import (
	"path/filepath"

	"github.com/Tomdooo/spajz/internal/config"
)

func GetPath(bucket string) string {
	return filepath.Join(config.DataDir, bucket)
}
