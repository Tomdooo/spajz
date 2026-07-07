package echox

import (
	"crypto/subtle"
	"net/http"

	"github.com/Tomdooo/spajz/internal/config"
	"github.com/labstack/echo/v5"
)

func BindAndValidate[T any](c *echo.Context) (*T, error) {
	u := new(T)
	if err := c.Bind(u); err != nil {
		return nil, err
	}
	if err := c.Validate(u); err != nil {
		return nil, err
	}
	return u, nil
}

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

func StorageAuthMiddleware(configManager *config.BucketConfigManager) echo.MiddlewareFunc { // TODO: implement properly verification of API key
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			bucketName := c.Param("bucket")
			method := c.Request().Method

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
