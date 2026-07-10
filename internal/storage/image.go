package storage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/internal/models"
	"github.com/h2non/bimg"
	"golang.org/x/sync/singleflight"
)

var ErrPresetNotExist = errors.New("Preset does not exist.")
var ErrUnsupportedFormat = errors.New("Unsupported format.")

var imageGenerator = NewImageGenerator()

type ImageGenerator struct {
	presetGroup         singleflight.Group
	bucketConfigManager *config.BucketConfigManager
}

func NewImageGenerator() *ImageGenerator {
	return &ImageGenerator{
		bucketConfigManager: config.GetBucketConfigManager(),
	}
}

func (ig *ImageGenerator) getBimgImageType(format string) (bimg.ImageType, error) {
	switch strings.ToLower(format) {
	case "webp":
		return bimg.WEBP, nil
	case "png":
		return bimg.PNG, nil
	case "jpg":
	case "jpeg":
		return bimg.JPEG, nil
	case "avif":
		return bimg.AVIF, nil
	case "gif":
		return bimg.GIF, nil
	case "heig":
		return bimg.HEIF, nil
	}

	return 0, ErrUnsupportedFormat
}

func (ig *ImageGenerator) CreatePresetVariant(fileContext *models.FileRequestContext, presetConfig *config.ImagePreset) ([]byte, error) {
	key := fileContext.Bucket + "@" + fileContext.ObjectHash + "@" + presetConfig.Name

	// Initialize singleflight
	v, err, _ := ig.presetGroup.Do(key, func() (any, error) {

		// Load original image
		originalImage, err := Get(fileContext)
		if err != nil {
			return nil, fmt.Errorf("failed to read original file: %w", err)
		}
		// Prepare preset config
		// presetConfig, err := ig.bucketConfigManager.GetImagePreset(bucket, preset)
		// if err != nil {
		// 	if errors.Is(err, config.ErrBucketNotExist) {
		// 		return nil, ErrBucketNotExist
		// 	} else if errors.Is(err, config.ErrPresetNotExist) {
		// 		return nil, ErrPresetNotExist
		// 	}
		// }

		format, err := ig.getBimgImageType(presetConfig.Format)
		if err != nil {
			return nil, err
		}

		// Initialize bimg transformation
		image := bimg.NewImage(originalImage)
		options := bimg.Options{
			Width:         presetConfig.Width,
			Height:        presetConfig.Height,
			Type:          format, // Vynutíme WebP pro web
			Quality:       presetConfig.Quality,
			Enlarge:       presetConfig.Enlarge,
			Embed:         true,
			StripMetadata: true,
		}

		// Process image
		transformedData, err := image.Process(options)
		if err != nil {
			return nil, fmt.Errorf("bimg processing failed: %w", err)
		}

		return transformedData, nil
	})

	// Check for singleflight errors
	if err != nil {
		return nil, err
	}

	return v.([]byte), nil
}
