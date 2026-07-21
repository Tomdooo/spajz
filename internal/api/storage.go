package api

import (
	"net/http"
	"strings"

	"github.com/Tomdooo/spajz/internal/models"
	"github.com/Tomdooo/spajz/internal/storage"
	"github.com/Tomdooo/spajz/pkg/echox"
	"github.com/Tomdooo/spajz/pkg/validatorx"
	"github.com/labstack/echo/v5"
)

type StorageHandler struct{}

func NewStorageHandler() *StorageHandler {
	return &StorageHandler{}
}

func (h *StorageHandler) Head(c *echo.Context) error {
	dto := new(models.S3LikeDto)
	if err := c.Bind(dto); err != nil {
		return err
	}
	fileContext := models.NewFileRequestContext(dto.Bucket, dto.ObjectKey, storage.GetObjectHash(dto.ObjectKey))

	metadata, err := storage.GetMetadata(fileContext)
	if err != nil {
		return echox.HandleError(c, err, "bucket", fileContext.Bucket, "objectKey", fileContext.ObjectKey)
	}

	setFileFullHeaders(c, metadata)
	return c.NoContent(http.StatusOK)
}

func (h *StorageHandler) Get(c *echo.Context) error {
	dto := new(models.S3LikeDto)
	if err := c.Bind(dto); err != nil {
		return err
	}

	var file []byte
	var metadata *storage.FileMeta
	var err error

	if strings.Contains(dto.ObjectKey, "@") { // Preset
		ctx := c.Request().Context()
		splittedObjectKey := strings.Split(dto.ObjectKey, "@")
		if len(splittedObjectKey) > 2 {
			return echox.HandleError(c, models.ErrInvalidURLFormat)
		}
		objectKey := splittedObjectKey[0]
		preset := splittedObjectKey[1]
		if valid := validatorx.PresetRegex.MatchString(preset); !valid { // validate preset name
			return echox.HandleError(c, models.ErrPresetFormatInvalid)
		}

		fileContext := models.NewFileRequestContext(dto.Bucket, objectKey, storage.GetObjectHash(objectKey))

		var isCacheHit bool
		file, metadata, isCacheHit, err = storage.GetPresetVariant(ctx, fileContext, preset)

		if err != nil {
			return echox.HandleError(c, err, "bucket", fileContext.Bucket, "objectKey", fileContext.ObjectKey, "preset", preset)
		} else {
			if isCacheHit {
				c.Response().Header().Set("X-Cache", "HIT")
			} else {
				c.Response().Header().Set("X-Cache", "MISS")
			}
		}
	} else { // Base image
		fileContext := models.NewFileRequestContext(dto.Bucket, dto.ObjectKey, storage.GetObjectHash(dto.ObjectKey))
		file, metadata, err = storage.GetWithMetadata(fileContext)

		if err != nil {
			return echox.HandleError(c, err, "bucket", fileContext.Bucket, "objectKey", fileContext.ObjectKey)
		}
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

	fileReader := c.Request().Body
	defer fileReader.Close()

	contentType := c.Request().Header.Get("Content-Type")

	fileContext := models.NewFileRequestContext(bucket, objectKey, storage.GetObjectHash(objectKey))
	metadata, err := storage.Add(fileContext, contentType, fileReader)
	if err != nil {
		return echox.HandleError(c, err, "bucket", fileContext.Bucket, "objectKey", fileContext.ObjectKey)
	}

	setFileEtagHeader(c, metadata)
	return c.NoContent(http.StatusOK)
}

func (h *StorageHandler) Delete(c *echo.Context) error {
	dto := new(models.S3LikeDto)
	if err := c.Bind(dto); err != nil {
		return err
	}

	ctx := c.Request().Context()
	fileContext := models.NewFileRequestContext(dto.Bucket, dto.ObjectKey, storage.GetObjectHash(dto.ObjectKey))
	if err := storage.Delete(ctx, fileContext); err != nil {
		return echox.HandleError(c, err, "bucket", fileContext.Bucket, "objectKey", fileContext.ObjectKey)
	}

	return c.NoContent(http.StatusNoContent)
}
