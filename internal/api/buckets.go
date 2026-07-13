package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Tomdooo/spajz/internal/buckets"
	"github.com/Tomdooo/spajz/internal/models"
	"github.com/labstack/echo/v5"
)

type BucketsHandler struct{}

func NewBucketsHandler() *BucketsHandler {
	return &BucketsHandler{}
}

type BucketDto struct {
	Bucket string `param:"bucket" validate:"required"`
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
	if err := buckets.Create(dto.Bucket); err != nil {
		switch {
		case errors.Is(err, models.ErrBucketAlreadyExists):
			return echo.NewHTTPError(http.StatusConflict, "No such bucket.")
		default:
			slog.Error("Failed to create bucket.",
				"bucket", dto.Bucket,
				"error", err)
			return err
		}
	}
	header := c.Response().Header()
	header.Set("Location", "/"+dto.Bucket)
	return c.NoContent(http.StatusOK)
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
			return echo.NewHTTPError(http.StatusNotFound, "No such bucket.")
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
