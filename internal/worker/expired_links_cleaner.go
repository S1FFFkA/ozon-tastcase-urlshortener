package worker

import (
	"context"
	"time"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/repository"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/pkg/logger"
	"github.com/go-co-op/gocron/v2"
)

const (
	defaultCleanupBatchSize = 100
	defaultRetentionYears   = 3
	defaultCleanupHour      = 3
	defaultCleanupMinute    = 0
)

type ExpiredLinksCleaner struct {
	repo           repository.URLRepository
	batchSize      int
	retentionYears int
}

func NewExpiredLinksCleaner(repo repository.URLRepository, batchSize, retentionYears int) *ExpiredLinksCleaner {
	if batchSize <= 0 {
		batchSize = defaultCleanupBatchSize
	}
	if retentionYears <= 0 {
		retentionYears = defaultRetentionYears
	}

	return &ExpiredLinksCleaner{
		repo:           repo,
		batchSize:      batchSize,
		retentionYears: retentionYears,
	}
}

func (c *ExpiredLinksCleaner) Cleanup(ctx context.Context) {
	cutoff := time.Now().UTC().AddDate(-c.retentionYears, 0, 0)
	var totalDeleted int64

	for {
		deleted, err := c.repo.DeleteExpiredBatch(ctx, cutoff, c.batchSize)
		if err != nil {
			logger.Logger.Errorw("expired links cleanup failed", "error", err, "cutoff", cutoff, "batch_size", c.batchSize)
			return
		}

		totalDeleted += deleted
		if deleted == 0 {
			break
		}
	}

	logger.Logger.Infow("expired links cleanup completed", "cutoff", cutoff, "deleted_total", totalDeleted, "retention_years", c.retentionYears)
}

func NewExpiredLinksCleanupScheduler(cleaner *ExpiredLinksCleaner, hour, minute int) (gocron.Scheduler, error) {
	if hour < 0 || hour > 23 {
		hour = defaultCleanupHour
	}
	if minute < 0 || minute > 59 {
		minute = defaultCleanupMinute
	}

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	_, err = scheduler.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(uint(hour), uint(minute), 0),
			),
		),
		gocron.NewTask(func() {
			cleaner.Cleanup(context.Background())
		}),
	)
	if err != nil {
		shutdownErr := scheduler.Shutdown()
		if shutdownErr != nil {
			logger.Logger.Errorw("failed to shutdown scheduler after setup error", "error", shutdownErr)
		}
		return nil, err
	}

	return scheduler, nil
}
