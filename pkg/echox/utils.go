package echox

import (
	"errors"

	"github.com/Tomdooo/spajz/internal/models"
	"github.com/labstack/echo/v5"
)

type responseError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ErrorResponse(c *echo.Context, status int, msg string, knownErr error) error {
	re := responseError{
		Status:  status,
		Message: msg,
	}
	switch {
	case errors.Is(knownErr, models.ErrBucketNotFound):
		re.Code = models.CodeBucketNotFound
	case errors.Is(knownErr, models.ErrBucketAlreadyExists):
		re.Code = models.CodeBucketAlreadyExists
	case errors.Is(knownErr, models.ErrBucketNotEmpty):
		re.Code = models.CodeBucketNotEmpty
	case errors.Is(knownErr, models.ErrFileNotFound):
		re.Code = models.CodeFileNotFound
	case errors.Is(knownErr, models.ErrFileNotSaved):
		re.Code = models.CodeFileNotSaved
	case errors.Is(knownErr, models.ErrFileAlreadyExists):
		re.Code = models.CodeFileAlreadyExists
	case errors.Is(knownErr, models.ErrNotEnoughSpace):
		re.Code = models.CodeNotEnoughSpace
	case errors.Is(knownErr, models.ErrFileTooLarge):
		re.Code = models.CodeFileTooLarge
	case errors.Is(knownErr, models.ErrFileNotProcessable):
		re.Code = models.CodeFileNotProcessable
	case errors.Is(knownErr, models.ErrPresetNotFound):
		re.Code = models.CodePresetNotFound
	case errors.Is(knownErr, models.ErrUnsupportedFormat):
		re.Code = models.CodeUnsupportedFormat
	case errors.Is(knownErr, models.ErrImageNotProcessable):
		re.Code = models.CodeImageNotProcessable
	case errors.Is(knownErr, models.ErrDatabaseAlreadyExists):
		re.Code = models.CodeDatabaseAlreadyExists
	}
	return c.JSON(status, re)
}
