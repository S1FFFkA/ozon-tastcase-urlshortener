package service

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

type serviceRepoMock struct {
	getByOriginalFn     func(ctx context.Context, originalURL string) (*domain.Link, error)
	resolveByShortURLFn func(ctx context.Context, shortURL string) (*domain.Link, error)
	createFn            func(ctx context.Context, originalURL, shortURL string) (*domain.Link, error)
	deleteExpiredFn     func(ctx context.Context, cutoff time.Time, limit int) (int64, error)
}

var serviceLoggerOnce sync.Once

func initServiceTestLogger() {
	serviceLoggerOnce.Do(func() {
		logger.InitLogger()
	})
}

func (m *serviceRepoMock) GetByOriginalURL(ctx context.Context, originalURL string) (*domain.Link, error) {
	return m.getByOriginalFn(ctx, originalURL)
}

func (m *serviceRepoMock) ResolveByShortURL(ctx context.Context, shortURL string) (*domain.Link, error) {
	return m.resolveByShortURLFn(ctx, shortURL)
}

func (m *serviceRepoMock) Create(ctx context.Context, originalURL, shortURL string) (*domain.Link, error) {
	return m.createFn(ctx, originalURL, shortURL)
}

func (m *serviceRepoMock) DeleteExpiredBatch(ctx context.Context, cutoff time.Time, limit int) (int64, error) {
	if m.deleteExpiredFn == nil {
		return 0, nil
	}
	return m.deleteExpiredFn(ctx, cutoff, limit)
}

func TestURLServiceCreate_ReturnsExistingLink(t *testing.T) {
	t.Parallel()
	initServiceTestLogger()

	expected := &domain.Link{ID: 1, OriginalURL: "https://example.com", ShortURL: "abcdefghij"}
	repo := &serviceRepoMock{
		getByOriginalFn: func(_ context.Context, _ string) (*domain.Link, error) {
			return expected, nil
		},
		resolveByShortURLFn: func(_ context.Context, _ string) (*domain.Link, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, _, _ string) (*domain.Link, error) {
			return nil, errors.New("should not be called")
		},
	}

	svc := NewURLService(repo)
	got, err := svc.Create(context.Background(), "https://example.com")
	require.NoError(t, err)
	require.Equal(t, expected.ShortURL, got.ShortURL)
}

func TestURLServiceCreate_RetryOnConflictAndSucceeds(t *testing.T) {
	t.Parallel()
	initServiceTestLogger()

	createCalls := 0
	repo := &serviceRepoMock{
		getByOriginalFn: func(_ context.Context, _ string) (*domain.Link, error) {
			return nil, domain.ErrNotFound
		},
		resolveByShortURLFn: func(_ context.Context, _ string) (*domain.Link, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, originalURL, shortURL string) (*domain.Link, error) {
			createCalls++
			if createCalls == 1 {
				return nil, domain.ErrConflict
			}
			return &domain.Link{OriginalURL: originalURL, ShortURL: shortURL}, nil
		},
	}

	svc := NewURLService(repo)
	got, err := svc.Create(context.Background(), "https://example.com")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotEmpty(t, got.ShortURL)
	require.Equal(t, 2, createCalls)
}

func TestURLServiceResolve_EmptyInput(t *testing.T) {
	t.Parallel()
	initServiceTestLogger()

	repo := &serviceRepoMock{
		getByOriginalFn: func(_ context.Context, _ string) (*domain.Link, error) { return nil, nil },
		resolveByShortURLFn: func(_ context.Context, _ string) (*domain.Link, error) {
			return nil, errors.New("should not be called")
		},
		createFn: func(_ context.Context, _, _ string) (*domain.Link, error) { return nil, nil },
	}

	svc := NewURLService(repo)
	_, err := svc.Resolve(context.Background(), "   ")
	require.ErrorIs(t, err, domain.ErrBadRequest)
}
