package repository

import (
	"context"
	"time"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/domain"
)

type URLRepository interface {
	GetByOriginalURL(ctx context.Context, originalURL string) (*domain.Link, error)
	ResolveByShortURL(ctx context.Context, shortURL string) (*domain.Link, error)
	Create(ctx context.Context, originalURL, shortURL string) (*domain.Link, error)
	DeleteExpiredBatch(ctx context.Context, cutoff time.Time, limit int) (int64, error)
}
