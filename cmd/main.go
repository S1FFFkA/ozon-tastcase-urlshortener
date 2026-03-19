package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/config"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/handler"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/service"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/worker"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/pkg/helpers"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	logger.InitLogger()
	defer logger.Sync()

	cfg, err := config.FromEnv()
	if err != nil {
		logger.Logger.Fatalw("failed to parse config", "error", err)
	}

	repo, cleanup, err := helpers.BuildRepository(cfg)
	if err != nil {
		logger.Logger.Fatalw("failed to initialize repository", "error", err, "storage", cfg.Storage)
	}
	defer cleanup()

	urlService := service.NewURLService(repo)
	urlHandler := handler.NewURLHandler(urlService, cfg.BaseURL)
	expiredLinksCleaner := worker.NewExpiredLinksCleaner(repo, cfg.CleanupBatch, cfg.RetentionYears)
	scheduler, err := worker.NewExpiredLinksCleanupScheduler(expiredLinksCleaner, cfg.CleanupHour, cfg.CleanupMinute)
	if err != nil {
		logger.Logger.Fatalw("failed to initialize cleanup scheduler", "error", err)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	urlHandler.RegisterRoutes(router)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	go func() {
		logger.Logger.Infow("http server started", "addr", cfg.HTTPAddr, "base_url", cfg.BaseURL, "storage", cfg.Storage)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Logger.Fatalw("http server failed", "error", serveErr)
		}
	}()
	scheduler.Start()
	logger.Logger.Infow(
		"cleanup scheduler started",
		"schedule", "daily",
		"hour", cfg.CleanupHour,
		"minute", cfg.CleanupMinute,
		"batch_size", cfg.CleanupBatch,
		"retention_years", cfg.RetentionYears,
	)

	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-stopCtx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err = server.Shutdown(shutdownCtx); err != nil {
		logger.Logger.Errorw("graceful shutdown failed", "error", err)
		return
	}
	if err = scheduler.Shutdown(); err != nil {
		logger.Logger.Errorw("scheduler graceful shutdown failed", "error", err)
	}

	logger.Logger.Infow("server stopped gracefully")
}
