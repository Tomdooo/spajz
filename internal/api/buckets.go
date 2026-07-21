package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Tomdooo/spajz/internal/buckets"
	"github.com/Tomdooo/spajz/internal/models"
	"github.com/Tomdooo/spajz/pkg/echox"
	"github.com/labstack/echo/v5"
)

type BucketsHandler struct{}

func NewBucketsHandler() *BucketsHandler {
	return &BucketsHandler{}
}

type BucketDto struct {
	Bucket string `param:"bucket" validate:"required,min=3,max=63,spajz_bucket"`
}

func (h *BucketsHandler) Create(c *echo.Context) error {
	// parse request dto
	dto := new(BucketDto)
	if err := c.Bind(dto); err != nil {
		return err
	}
	if err := c.Validate(dto); err != nil {
		return err
	}
	// create bucket
	defaultApiKey, err := buckets.Create(dto.Bucket)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrBucketAlreadyExists):
			return echox.ErrorResponse(c, http.StatusConflict, "Bucket already exists.", err)
		default:
			slog.Error("Failed to create bucket.",
				"bucket", dto.Bucket,
				"error", err)
			return err
		}
	}
	header := c.Response().Header()
	header.Set("Location", "/"+dto.Bucket)

	type CreateBucketResponse struct {
		DefaultApiKey string `json:"default_api_key"`
	}
	resBody := new(CreateBucketResponse{
		DefaultApiKey: defaultApiKey,
	})
	return c.JSON(http.StatusOK, resBody)
}

func (h *BucketsHandler) Delete(c *echo.Context) error {
	// parse request dto
	dto := new(BucketDto)
	if err := c.Bind(dto); err != nil {
		return err
	}
	if err := c.Validate(dto); err != nil {
		return err
	}

	// delete
	if err := buckets.Delete(dto.Bucket); err != nil {
		switch {
		case errors.Is(err, models.ErrBucketNotFound):
			return echox.ErrorResponse(c, http.StatusNotFound, "No such bucket.", err)
		case errors.Is(err, models.ErrBucketNotEmpty):
			return echox.ErrorResponse(c, http.StatusConflict, "Bucket is not empty.", err)
		default:
			slog.Error("Failed to delete bucket.",
				"bucket", dto.Bucket,
				"error", err)
			return err
		}
	}

	return c.NoContent(http.StatusNoContent)
}

type GetResponseBody struct {
	Buckets []buckets.BucketEntry `json:"buckets"`
}

func (h *BucketsHandler) Get(c *echo.Context) error {
	bucketEntries, err := buckets.Get()
	if err != nil {
		slog.Error("Failed to list buckets.",
			"error", err)
		return err
	}

	resBody := GetResponseBody{
		Buckets: bucketEntries,
	}
	return c.JSON(http.StatusOK, resBody)
}
