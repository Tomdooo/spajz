package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/Tomdooo/spajz/internal/api"
	"github.com/Tomdooo/spajz/pkg/echox"
	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	// err := buckets.Create("test")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// http.HandleFunc("/upload", api.UploadHandler)
	// http.ListenAndServe(":8080", nil)

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit(2_097_152)) // 2 MB
	e.Validator = echox.NewValidator()
	logger := log.New(os.Stderr)
	e.Logger = slog.New(logger)

	e.GET("/", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Hello, World!"})
	})

	g := e.Group("")
	api.RegisterRoutes(g)

	if err := e.Start(":1323"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
