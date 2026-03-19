package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/domain"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/pkg/logger"
	"github.com/stretchr/testify/require"
)

type workerRepoMock struct {
	deleteFn func(ctx context.Context, cutoff time.Time, limit int) (int64, error)
	calls    int
	limits   []int
}

var workerLoggerOnce sync.Once

func initWorkerTestLogger() {
	workerLoggerOnce.Do(func() {
		logger.InitLogger()
	})
}

func (m *workerRepoMock) GetByOriginalURL(context.Context, string) (*domain.Link, error) {
	return nil, nil
}

func (m *workerRepoMock) ResolveByShortURL(context.Context, string) (*domain.Link, error) {
	return nil, nil
}

func (m *workerRepoMock) Create(context.Context, string, string) (*domain.Link, error) {
	return nil, nil
}

func (m *workerRepoMock) DeleteExpiredBatch(ctx context.Context, cutoff time.Time, limit int) (int64, error) {
	m.calls++
	m.limits = append(m.limits, limit)
	return m.deleteFn(ctx, cutoff, limit)
}

func TestCleanup_DeletesInBatchesUntilZero(t *testing.T) {
	t.Parallel()
	initWorkerTestLogger()

	responses := []int64{5, 2, 0}
	idx := 0
	repo := &workerRepoMock{
		deleteFn: func(_ context.Context, _ time.Time, _ int) (int64, error) {
			val := responses[idx]
			idx++
			return val, nil
		},
	}

	cleaner := NewExpiredLinksCleaner(repo, 50, 3)
	cleaner.Cleanup(context.Background())

	require.Equal(t, 3, repo.calls)
	for _, limit := range repo.limits {
		require.Equal(t, 50, limit)
	}
}

func TestCleanup_StopsOnRepositoryError(t *testing.T) {
	t.Parallel()
	initWorkerTestLogger()

	repo := &workerRepoMock{
		deleteFn: func(_ context.Context, _ time.Time, _ int) (int64, error) {
			return 0, errors.New("db error")
		},
	}

	cleaner := NewExpiredLinksCleaner(repo, 10, 3)
	cleaner.Cleanup(context.Background())

	require.Equal(t, 1, repo.calls)
}

func TestNewExpiredLinksCleanupScheduler_CreateAndShutdown(t *testing.T) {
	t.Parallel()
	initWorkerTestLogger()

	repo := &workerRepoMock{
		deleteFn: func(_ context.Context, _ time.Time, _ int) (int64, error) {
			return 0, nil
		},
	}
	cleaner := NewExpiredLinksCleaner(repo, 10, 3)

	scheduler, err := NewExpiredLinksCleanupScheduler(cleaner, -1, 99)
	require.NoError(t, err)

	require.NoError(t, scheduler.Shutdown())
}
