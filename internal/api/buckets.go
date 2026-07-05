package api

import (
	"errors"
	"net/http"

	"github.com/Tomdooo/spajz/internal/buckets"
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
		if errors.Is(err, buckets.ErrAlreadyExists) {
			return c.NoContent(http.StatusConflict)
		}
		return err
	}
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
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
