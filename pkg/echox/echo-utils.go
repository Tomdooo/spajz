package echox

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/internal/models"
	"github.com/Tomdooo/spajz/pkg/validatorx"
	"github.com/labstack/echo/v5"
)

func BucketsAuthMiddleware(masterKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		if masterKey == "" {
			return func(c *echo.Context) error {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "Bucket configuration API is disabled.")
			}
		}
		return func(c *echo.Context) error {
			requestMasterKey := c.Request().Header.Get("X-Spajz-Master-Key")
			if requestMasterKey == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing master key")
			}

			// secured comparation of keys
			if subtle.ConstantTimeCompare([]byte(requestMasterKey), []byte(config.MasterKey)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid master key.")
			}

			return next(c)
		}
	}
}

func StorageAuthMiddleware(configManager *config.BucketConfigManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			bucketName := c.Param("bucket")
			method := c.Request().Method

			// Validate bucket name before doing anything else
			if len(bucketName) < 3 || len(bucketName) > 63 || !validatorx.BucketRegex.MatchString(bucketName) {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid bucket name")
			}

			// get bucket config
			bucketConfig, err := configManager.GetConfig(bucketName)
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, "Bucket not found")
			}

			// allow public reading or public upload
			if (bucketConfig.Bucket.AllowPublicReading && (method == http.MethodGet || method == http.MethodHead)) ||
				(bucketConfig.Bucket.AllowPublicUpload && (method == http.MethodPut || method == http.MethodPost)) ||
				(bucketConfig.Bucket.AllowPublicDelete && method == http.MethodDelete) {
				return next(c)
			}

			// get API key
			apiKey := c.Request().Header.Get("X-Spajz-Key")
			if apiKey == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing API key")
			}

			// verify API key
			valid, apiKeyConfig, err := configManager.VerifyApiKey(bucketName, apiKey)
			if err != nil {
				return err
			}
			if !valid {
				return echo.NewHTTPError(http.StatusForbidden, "Invalid API key for this bucket")
			}

			if (apiKeyConfig.AllowReading && (method == http.MethodGet || method == http.MethodHead)) ||
				(apiKeyConfig.AllowUpload && (method == http.MethodPut || method == http.MethodPost)) ||
				(apiKeyConfig.AllowDelete && method == http.MethodDelete) {
				return next(c)
			}

			return echo.NewHTTPError(http.StatusForbidden, "The API key does not have permission for this")
		}
	}
}

func StorageValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			bucket := c.Param("bucket")
			objectKey := c.Param("*")

			params := models.S3LikeDto{
				Bucket:    bucket,
				ObjectKey: objectKey,
			}
			if err := c.Validate(&params); err != nil {
				fmt.Println(params)
				return err
			}
			return next(c)
		}
	}
}
