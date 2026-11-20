package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ossirdian/url_shortener/internal/config"
	envconfig "github.com/ossirdian/url_shortener/internal/config/env"
	logger "github.com/ossirdian/url_shortener/internal/handlers/middleware"
	"github.com/ossirdian/url_shortener/internal/handlers/redirect"
	"github.com/ossirdian/url_shortener/internal/handlers/url/create"
	"github.com/ossirdian/url_shortener/internal/handlers/url/delete"
	"github.com/ossirdian/url_shortener/internal/handlers/url/update"
	"github.com/ossirdian/url_shortener/internal/httpauth"
	"github.com/ossirdian/url_shortener/internal/migrator"
	urlrepository "github.com/ossirdian/url_shortener/internal/repository/url_shortener"
	urlservice "github.com/ossirdian/url_shortener/internal/service/url_shortener"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := envconfig.NewConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = envLocal
	}
	log := setupLogger(appEnv, cfg.Logger())
	log.Info("starting URL shortener", slog.String("env", appEnv))

	ctx := context.Background()
	if cfg.Postgres().AutoMigrate() {
		log.Info("applying database migrations")
		if err := migrator.Up(ctx, cfg.Postgres().DSN(), cfg.Postgres().MigrationsDir()); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
	}

	pool, err := pgxpool.New(ctx, cfg.Postgres().DSN())
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}

	urlRepo := urlrepository.NewRepository(pool)
	urlService := urlservice.NewService(urlRepo)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(httpauth.BasicAuth(cfg.Auth().Username(), cfg.Auth().Password()))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Post("/", create.New(log, urlService))
		r.Delete("/{alias}", delete.New(log, urlService))
		r.Patch("/{alias}", update.New(log, urlService))
	})
	router.Get("/{alias}", redirect.New(log, urlService))

	server := &http.Server{
		Addr:              cfg.HTTP().Addr(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Info("starting HTTP server", slog.String("address", server.Addr))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}

	return nil
}

func setupLogger(appEnv string, loggerConfig config.LoggerConfig) *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(loggerConfig.Level()) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	options := &slog.HandlerOptions{Level: level}
	switch appEnv {
	case envLocal:
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	case envProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, options))
	default:
		if loggerConfig.AsJSON() {
			return slog.New(slog.NewJSONHandler(os.Stdout, options))
		}
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}
}
