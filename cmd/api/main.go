package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	postgresrepo "github.com/ryuta-github/golang-api-server/internal/infrastructure/postgres"
	presentation "github.com/ryuta-github/golang-api-server/internal/presentation/http"
	"github.com/ryuta-github/golang-api-server/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return err
	}

	// Explicit dependency injection: infrastructure -> use case -> presentation.
	repository := postgresrepo.NewTaskRepository(pool)
	taskUseCase := usecase.NewTask(repository)
	taskHandler := presentation.NewTaskHandler(taskUseCase)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.RequestID(), middleware.Recover(), middleware.Logger())
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	taskHandler.Register(e.Group("/api/v1/tasks"))

	errCh := make(chan error, 1)
	go func() { errCh <- e.Start(":" + port) }()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return e.Shutdown(shutdownCtx)
	}
}
