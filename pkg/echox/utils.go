package echox

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Tomdooo/spajz/internal/models"
	"github.com/labstack/echo/v5"
)

type errorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorInfo struct {
	Status    int
	Code      string
	Message   string
	ShouldLog bool
}

var errorRegistry = []struct {
	Err  error
	Info errorInfo
}{
	{models.ErrBucketNotFound, errorInfo{Status: http.StatusNotFound, Code: models.CodeBucketNotFound, Message: "No such bucket."}},
	{models.ErrBucketAlreadyExists, errorInfo{Status: http.StatusConflict, Code: models.CodeBucketAlreadyExists, Message: "Bucket already exists."}},
	{models.ErrBucketNotEmpty, errorInfo{Status: http.StatusConflict, Code: models.CodeBucketNotEmpty, Message: "Bucket is not empty."}},

	{models.ErrFileNotFound, errorInfo{Status: http.StatusNotFound, Code: models.CodeFileNotFound, Message: "No such file."}},
	{models.ErrFileAlreadyExists, errorInfo{Status: http.StatusConflict, Code: models.CodeFileAlreadyExists, Message: "File already exists."}},
	{models.ErrFileNotSaved, errorInfo{Status: http.StatusInternalServerError, Code: models.CodeFileNotSaved, Message: "File could not be saved.", ShouldLog: true}},
	{models.ErrFileTooLarge, errorInfo{Status: http.StatusRequestEntityTooLarge, Code: models.CodeFileTooLarge, Message: "File is too large."}},
	{models.ErrFileNotProcessable, errorInfo{Status: http.StatusBadRequest, Code: models.CodeFileNotProcessable, Message: "File is not processable."}},
	{models.ErrNotEnoughSpace, errorInfo{Status: http.StatusInsufficientStorage, Code: models.CodeNotEnoughSpace, Message: "Not enough space.", ShouldLog: true}},

	{models.ErrPresetNotFound, errorInfo{Status: http.StatusBadRequest, Code: models.CodePresetNotFound, Message: "No such preset."}},
	{models.ErrPresetFormatInvalid, errorInfo{Status: http.StatusBadRequest, Code: models.CodePresetFormatInvalid, Message: "Preset format is not valid."}},
	{models.ErrUnsupportedFormat, errorInfo{Status: http.StatusBadRequest, Code: models.CodeUnsupportedFormat, Message: "Unsupported format.", ShouldLog: true}},
	{models.ErrImageNotProcessable, errorInfo{Status: http.StatusBadRequest, Code: models.CodeImageNotProcessable, Message: "Image is not processable."}},

	{models.ErrDatabaseAlreadyExists, errorInfo{Status: http.StatusConflict, Code: models.CodeDatabaseAlreadyExists, Message: "Database already exists.", ShouldLog: true}},
	{models.ErrInvalidURLFormat, errorInfo{Status: http.StatusBadRequest, Code: models.CodeInvalidURLFormat, Message: "Invalid URL format, please use /<bucket>/<file_path>@<preset>."}},
}

func lookupErrorInfo(err error) (errorInfo, bool) {
	for _, e := range errorRegistry {
		if errors.Is(err, e.Err) {
			return e.Info, true
		}
	}
	return errorInfo{}, false
}

func HandleError(c *echo.Context, err error, logFields ...any) error {
	if info, ok := lookupErrorInfo(err); ok {
		if info.ShouldLog {
			slog.Error(info.Message, append(logFields, "error", err)...)
		}
		re := errorResponse{
			Status:  info.Status,
			Message: info.Message,
			Code:    info.Code,
		}
		return c.JSON(re.Status, re)
	}

	slog.Error("Unhandled request error", append(logFields, "error", err)...)
	re := errorResponse{
		Status:  http.StatusInternalServerError,
		Message: "Internal server error.",
	}
	return c.JSON(re.Status, re)
}
