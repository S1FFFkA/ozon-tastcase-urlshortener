package repository

import (
	"context"
	"errors"
	"time"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/domain"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresDB interface { // Интерфейс для тестов
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type PostgresRepository struct {
	db postgresDB
}

var _ URLRepository = (*PostgresRepository)(nil)

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*domain.Link, error) {
	const query = `
SELECT id, original_url, short_url, created_at, last_used_at, use_count
FROM links
WHERE original_url = $1
`

	var (
		link       domain.Link
		lastUsedAt *time.Time
	)

	err := r.db.QueryRow(ctx, query, originalURL).Scan(
		&link.ID,
		&link.OriginalURL,
		&link.ShortURL,
		&link.CreatedAt,
		&lastUsedAt,
		&link.UseCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Logger.Warnw("link not found by original url", "original_url", originalURL, "error", err)
			return nil, domain.ErrNotFound
		}

		logger.Logger.Errorw("failed to get link by original url", "original_url", originalURL, "error", err)
		return nil, domain.ErrInternal
	}

	if lastUsedAt != nil {
		link.LastUsedAt = lastUsedAt
	}

	return &link, nil
}

func (r *PostgresRepository) ResolveByShortURL(ctx context.Context, shortURL string) (*domain.Link, error) {
	const query = `
UPDATE links
SET last_used_at = now(),
    use_count = use_count + 1
WHERE short_url = $1
RETURNING id, original_url, short_url, created_at, last_used_at, use_count
`

	var (
		link          domain.Link
		lastUsedAtUTC time.Time
	)

	err := r.db.QueryRow(ctx, query, shortURL).Scan(
		&link.ID,
		&link.OriginalURL,
		&link.ShortURL,
		&link.CreatedAt,
		&lastUsedAtUTC,
		&link.UseCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Logger.Warnw("link not found by short url", "short_url", shortURL, "error", err)
			return nil, domain.ErrNotFound
		}

		logger.Logger.Errorw("failed to resolve link by short url", "short_url", shortURL, "error", err)
		return nil, domain.ErrInternal
	}

	link.LastUsedAt = &lastUsedAtUTC

	return &link, nil
}

func (r *PostgresRepository) Create(ctx context.Context, originalURL, shortURL string) (*domain.Link, error) {
	const query = `
INSERT INTO links (original_url, short_url)
VALUES ($1, $2)
RETURNING id, original_url, short_url, created_at, last_used_at, use_count
`

	var (
		link       domain.Link
		lastUsedAt *time.Time
	)

	err := r.db.QueryRow(ctx, query, originalURL, shortURL).Scan(
		&link.ID,
		&link.OriginalURL,
		&link.ShortURL,
		&link.CreatedAt,
		&lastUsedAt,
		&link.UseCount,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				if pgErr.ConstraintName == "links_original_url_key" {
					logger.Logger.Warnw("link already exists by original url", "original_url", originalURL, "short_url", shortURL, "error", err)
					return nil, domain.ErrAlreadyExists
				}

				logger.Logger.Warnw("short url conflict on create", "original_url", originalURL, "short_url", shortURL, "error", err)
				return nil, domain.ErrConflict
			}
		}

		logger.Logger.Errorw("failed to create link", "original_url", originalURL, "short_url", shortURL, "error", err)
		return nil, domain.ErrInternal
	}

	link.LastUsedAt = lastUsedAt

	logger.Logger.Infow("short url created", "original_url", link.OriginalURL, "short_url", link.ShortURL)

	return &link, nil
}

func (r *PostgresRepository) DeleteExpiredBatch(ctx context.Context, cutoff time.Time, limit int) (int64, error) {
	if limit <= 0 {
		return 0, domain.ErrBadRequest
	}

	const query = `
WITH to_delete AS (
    SELECT id
    FROM links
    WHERE (last_used_at IS NOT NULL AND last_used_at < $1)
       OR (last_used_at IS NULL AND created_at < $1)
    ORDER BY id
    LIMIT $2
)
DELETE FROM links
WHERE id IN (SELECT id FROM to_delete)
`

	result, err := r.db.Exec(ctx, query, cutoff.UTC(), limit)
	if err != nil {
		logger.Logger.Errorw("failed to delete expired links batch", "cutoff", cutoff.UTC(), "limit", limit, "error", err)
		return 0, domain.ErrInternal
	}

	deletedCount := result.RowsAffected()
	logger.Logger.Infow("expired links batch deleted", "cutoff", cutoff.UTC(), "limit", limit, "deleted_count", deletedCount)

	return deletedCount, nil
}
