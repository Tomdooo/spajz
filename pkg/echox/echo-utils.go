package echox

import "github.com/labstack/echo/v5"

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
