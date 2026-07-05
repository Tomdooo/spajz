package api

import (
	"errors"
	"net/http"

	"github.com/Tomdooo/spajz/internal/storage"
	"github.com/labstack/echo/v5"
)

type StorageHandler struct{}

func NewStorageHandler() *StorageHandler {
	return &StorageHandler{}
}

type S3LikeDto struct {
	Bucket   string `param:"bucket" validate:"required"`
	Filename string `param:"*" validate:"required"`
}

func (h *StorageHandler) Head(c *echo.Context) error {
	dto := new(S3LikeDto)
	if err := c.Bind(dto); err != nil {
		return err
	}
	if err := c.Validate(dto); err != nil {
		return err
	}

	metadata, err := storage.GetMetadata(dto.Bucket, dto.Filename)
	if err != nil {
		if errors.Is(err, storage.ErrFileNotExist) || errors.Is(err, storage.ErrBucketNotExist) {
			return c.NoContent(http.StatusNotFound)
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

	file, metadata, err := storage.GetWithMetadata(dto.Bucket, dto.Filename)
	if err != nil {
		if errors.Is(err, storage.ErrFileNotExist) || errors.Is(err, storage.ErrBucketNotExist) {
			return c.NoContent(http.StatusNotFound)
		}
		return err
	}

	setFileFullHeaders(c, metadata)
	return c.Blob(http.StatusOK, metadata.ContentType, file)
}

type FormUploadDto struct {
	Bucket   string `form:"bucket" validate:"required"`
	Filename string `form:"filename" validate:"required"`
}

func (h *StorageHandler) FormUpload(c *echo.Context) error {
	dto := new(FormUploadDto)
	if err := c.Bind(dto); err != nil {
		return err
	}
	if err := c.Validate(dto); err != nil {
		return err
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return err
	}
	fileReader, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()

	metadata, err := storage.Add(dto.Bucket, dto.Filename, fileReader)
	if err != nil {
		if errors.Is(err, storage.ErrBucketNotExist) {
			return c.NoContent(http.StatusBadRequest)
		} else if errors.Is(err, storage.ErrFileExist) {
			return c.NoContent(http.StatusConflict)
		}

		return err
	}

	setFileEtagHeader(c, metadata)
	return c.NoContent(http.StatusOK)
}

func (h *StorageHandler) S3LikeUpload(c *echo.Context) error {
	bucket := c.Param("bucket")
	filename := c.Param("*")
	if bucket == "" || filename == "" {
		return c.String(http.StatusBadRequest, "Missing bucket or object key")
	}

	fileReader := c.Request().Body
	defer fileReader.Close()

	metadata, err := storage.Add(bucket, filename, fileReader)
	if err != nil {
		if errors.Is(err, storage.ErrBucketNotExist) {
			return c.NoContent(http.StatusBadRequest)
		} else if errors.Is(err, storage.ErrFileExist) {
			return c.NoContent(http.StatusConflict)
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

	if err := storage.Delete(dto.Bucket, dto.Filename); err != nil {
		if errors.Is(err, storage.ErrBucketNotExist) {
			return c.NoContent(http.StatusNotFound)
		}
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
