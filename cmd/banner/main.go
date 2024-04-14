package main

import (
	"banner/internal/config"
	"banner/internal/database/driver"
	"banner/internal/database/repository/pgsql"
	"banner/internal/http-server/handler/banner"
	"banner/internal/http-server/handler/banner/create"
	"banner/internal/http-server/handler/banner/delete"
	"banner/internal/http-server/handler/banner/update"
	userBanner "banner/internal/http-server/handler/banner/user"
	"banner/internal/http-server/middleware/logger"
	"banner/internal/http-server/middleware/validator"
	"fmt"

	"banner/pkg/lib/logger/slogpretty"
	"banner/pkg/lib/sl"
	"context"
	"errors"
	"net/http"

	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg, scr := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("starting app", slog.String("env", cfg.Env))

	log.Debug("debug messages are enabled")

	dataSourceName := fmt.Sprintf(
		"host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, scr.PostgresPassword, cfg.DBname, cfg.SSLmode,
	)

	sqlxConfig := &driver.SQLXConfig{
		DriverName:     cfg.DriverName,
		DataSourceName: dataSourceName,
		MaxOpenConns:   cfg.MaxOpenConns,
		MaxIdleConns:   cfg.MaxIdleConns,
		MaxLifetime:    cfg.MaxLifetime,
	}

	db, err := sqlxConfig.NewSQLXDatabase(log)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	bannerRepository := pgsql.NewBannerRepository(db)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(validator.New(log))
	router.Use(logger.New(log))
	//TODO: auth middleware

	router.Get("/banner", banner.New(log, bannerRepository))
	router.Post("/banner", create.New(log, bannerRepository))
	router.Delete("/banner/{id}", delete.New(log, bannerRepository))
	router.Patch("/banner/{id}", update.New(log, bannerRepository))
	router.Get("/user_banner", userBanner.New(log, bannerRepository))

	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("shutting server", sl.Err(err))
				return
			}
			log.Error("failed to start server", sl.Err(err))
		}
	}()

	log.Info("server started")
	sign := <-done
	log.Info("stopping server", slog.String("signal", sign.String()))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))
		return
	}

	if err := db.Close(); err != nil {
		log.Error("failed to close storage", sl.Err(err))
		return
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettyLogger()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}

func setupPrettyLogger() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
