package echox

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

var (
	// bucketRegex enforces AWS S3 naming rules (excluding length):
	// - Must start and end with a lowercase letter or number.
	// - Can contain lowercase letters, numbers, and hyphens in between.
	bucketRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
)

// CustomValidator wraps the go-playground validator to integrate with the Echo framework.
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator initializes the validator instance and registers custom Spajz rules.
func NewValidator() *CustomValidator {
	v := validator.New()

	// Register validator for Bucket names (AWS S3 compliant characters only)
	err := v.RegisterValidation("spajz_bucket", func(fl validator.FieldLevel) bool {
		return bucketRegex.MatchString(fl.Field().String())
	})
	if err != nil {
		panic(fmt.Sprintf("failed to register spajz_bucket validator: %v", err))
	}

	// Register validator for Object Keys to prevent path traversal attacks
	err = v.RegisterValidation("spajz_objectkey", func(fl validator.FieldLevel) bool {
		key := fl.Field().String()
		if key == "" {
			return false
		}
		// Deny relative path parent directory tokens and Windows directory separators
		if strings.Contains(key, "..") || strings.Contains(key, `\`) {
			return false
		}
		return true
	})
	if err != nil {
		panic(fmt.Sprintf("failed to register spajz_objectkey validator: %v", err))
	}

	return &CustomValidator{validator: v}
}

// Validate implements the echo.Validator interface.
// It automatically wraps validation errors inside an HTTP 400 Bad Request error.
func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		// Wrap the validation error so Echo can handle it as a 400 Bad Request
		return echo.ErrBadRequest.Wrap(err)
	}
	return nil
}
