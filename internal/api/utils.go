package api

import (
	"strconv"

	"github.com/Tomdooo/spajz/internal/storage"
	"github.com/labstack/echo/v5"
)

func setFileFullHeaders(c *echo.Context, metadata *storage.FileMeta) {
	headers := c.Response().Header()
	headers.Set("Content-Type", metadata.ContentType)
	headers.Set("Content-Length", strconv.FormatInt(metadata.Size, 10))
	headers.Set("ETag", "\""+metadata.Etag+"\"")
}

func setFileEtagHeader(c *echo.Context, metadata *storage.FileMeta) {
	c.Response().Header().Set("ETag", "\""+metadata.Etag+"\"")
}
