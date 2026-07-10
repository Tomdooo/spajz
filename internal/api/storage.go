package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Tomdooo/spajz/internal/models"
	"github.com/Tomdooo/spajz/internal/storage"
	"github.com/labstack/echo/v5"
)

type StorageHandler struct{}

func NewStorageHandler() *StorageHandler {
	return &StorageHandler{}
}

type S3LikeDto struct {
	Bucket    string `param:"bucket" validate:"required"`
	ObjectKey string `param:"*" validate:"required"`
}

func (h *StorageHandler) Head(c *echo.Context) error {
	dto := new(S3LikeDto)
	if err := c.Bind(dto); err != nil {
		return err
	}
	if err := c.Validate(dto); err != nil {
		return err
	}
	fileContext := models.NewFileRequestContext(dto.Bucket, dto.ObjectKey, storage.GetObjectHash(dto.ObjectKey))

	metadata, err := storage.GetMetadata(fileContext)
	if err != nil {
		if errors.Is(err, storage.ErrBucketNotExist) || errors.Is(err, storage.ErrFileNotExist) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return err
	}

	setFileFullHeaders(c, metadata)
	return c.NoContent(http.StatusOK)
}

func (h *StorageHandler) Get(c *echo.Context) error {
	dto := new(S3LikeDto)
	if err := c.Bind(dto); err != nil {
		return err
	}
	if err := c.Validate(dto); err != nil {
		return err
	}

	var file []byte
	var metadata *storage.FileMeta
	var err error

	if strings.Contains(dto.ObjectKey, "@") { // Preset
		ctx := c.Request().Context()
		splittedObjectKey := strings.Split(dto.ObjectKey, "@")
		if len(splittedObjectKey) > 2 {
			return echo.NewHTTPError(http.StatusBadRequest, "Bad URL format, please use /<bucket>/<file_path>@<preset>.")
		}
		objectKey := splittedObjectKey[0]
		preset := splittedObjectKey[1]

		fileContext := models.NewFileRequestContext(dto.Bucket, objectKey, storage.GetObjectHash(objectKey))

		var isCacheHit bool
		file, metadata, isCacheHit, err = storage.GetPresetVariant(ctx, fileContext, preset)

		if isCacheHit {
			c.Response().Header().Set("Cache", "HIT") // TODO: verify format of the header
		} else {
			c.Response().Header().Set("Cache", "MISS")
		}

	} else { // Base image
		fileContext := models.NewFileRequestContext(dto.Bucket, dto.ObjectKey, storage.GetObjectHash(dto.ObjectKey))
		file, metadata, err = storage.GetWithMetadata(fileContext)
	}

	if err != nil {
		if errors.Is(err, storage.ErrPresetNotExist) || errors.Is(err, storage.ErrUnsupportedFormat) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		} else if errors.Is(err, storage.ErrFileNotExist) || errors.Is(err, storage.ErrBucketNotExist) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return err
	}

	setFileFullHeaders(c, metadata)
	return c.Blob(http.StatusOK, metadata.ContentType, file)
}

// type FormUploadDto struct {
// 	Bucket   string `form:"bucket" validate:"required"`
// 	Filename string `form:"filename" validate:"required"`
// }

// func (h *StorageHandler) FormUpload(c *echo.Context) error {
// 	dto := new(FormUploadDto)
// 	if err := c.Bind(dto); err != nil {
// 		return err
// 	}
// 	if err := c.Validate(dto); err != nil {
// 		return err
// 	}

// 	fileHeader, err := c.FormFile("file")
// 	if err != nil {
// 		return err
// 	}
// 	fileReader, err := fileHeader.Open()
// 	if err != nil {
// 		return err
// 	}
// 	defer fileReader.Close()

// 	metadata, err := storage.Add(dto.Bucket, dto.Filename, fileReader)
// 	if err != nil {
// 		if errors.Is(err, storage.ErrBucketNotExist) {
// 			return c.NoContent(http.StatusBadRequest)
// 		} else if errors.Is(err, storage.ErrFileExist) {
// 			return c.NoContent(http.StatusConflict)
// 		}

// 		return err
// 	}

// 	setFileEtagHeader(c, metadata)
// 	return c.NoContent(http.StatusOK)
// }

func (h *StorageHandler) Upload(c *echo.Context) error {
	bucket := c.Param("bucket")
	objectKey := c.Param("*")
	if bucket == "" || objectKey == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing bucket or object key")
	}

	fileReader := c.Request().Body
	defer fileReader.Close()

	fileContext := models.NewFileRequestContext(bucket, objectKey, storage.GetObjectHash(objectKey))
	metadata, err := storage.Add(fileContext, fileReader)
	if err != nil {
		if errors.Is(err, storage.ErrBucketNotExist) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		} else if errors.Is(err, storage.ErrFileExist) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return err
	}

	setFileEtagHeader(c, metadata)
	return c.NoContent(http.StatusOK)
}

func (h *StorageHandler) Delete(c *echo.Context) error {
	dto := new(S3LikeDto)
	if err := c.Bind(dto); err != nil {
		return err
	}
	if err := c.Validate(dto); err != nil {
		return err
	}

	fileContext := models.NewFileRequestContext(dto.Bucket, dto.ObjectKey, storage.GetObjectHash(dto.ObjectKey))
	if err := storage.Delete(fileContext); err != nil {
		if errors.Is(err, storage.ErrBucketNotExist) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
