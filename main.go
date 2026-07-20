package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Tomdooo/spajz/internal/api"
	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/internal/db"
	"github.com/Tomdooo/spajz/pkg/validatorx"
	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

const Version = "dev"

func main() {
	// Logger
	handler := log.New(os.Stderr)
	handler.SetTimeFormat("2006-01-02 15:04:05")
	handler.SetReportTimestamp(true)
	handler.SetReportCaller(false)

	log.SetDefault(handler)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	printLogo()
	slog.Info("Špajz is warming up", "version", Version)

	config.Initialize()

	// Init modules
	bucketConfigManager := config.GetBucketConfigManager()
	if err := bucketConfigManager.LoadBucketConfigs(); err != nil {
		log.Fatal("failed to load bucket configs", "error", err)
	}
	databaseManager := db.GetDatabaseManager()
	if err := databaseManager.InitBucketDatabases(); err != nil {
		log.Fatal("failed to initialize bucket databases", "error", err)
	}

	// ECHO
	e := echo.New()

	e.Logger = logger
	e.Validator = validatorx.NewValidator()

	// Middlewares
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit(2_097_152)) // 2 MB

	g := e.Group("")
	api.RegisterRoutes(g)

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(200, map[string]string{"version": Version, "status": "OK"})
	})

	if err := e.Start(":1323"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}

}

func printLogo() {
	fmt.Print(
		"\n _\\_/               _     " +
			"\n/ ___| _ __   __ _ (_)____" +
			"\n\\___ \\| '_ \\ / _` || |_  /" +
			"\n ___) | |_) | (_| || |/ / " +
			"\n|____/| .__/ \\__,_|/ /___|" +
			"\n      |_|        |__/     " +
			"\n\n")
}
