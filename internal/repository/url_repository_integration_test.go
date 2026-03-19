//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/domain"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const createLinksSchemaSQL = `
CREATE TABLE IF NOT EXISTS links (
    id            BIGSERIAL PRIMARY KEY,
    original_url  TEXT NOT NULL UNIQUE,
    short_url     TEXT NOT NULL UNIQUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at  TIMESTAMPTZ,
    use_count     BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_links_last_used_at ON links (last_used_at);
CREATE INDEX IF NOT EXISTS idx_links_created_at ON links (created_at);
`

var loggerOnce sync.Once

func initIntegrationLogger() {
	loggerOnce.Do(func() {
		logger.InitLogger()
	})
}

func integrationDBURL() string {
	if v := strings.TrimSpace(os.Getenv("DATABASE_URL")); v != "" {
		return v
	}

	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "postgres")
	name := getEnvOrDefault("DB_NAME", "url_shortener")
	sslMode := getEnvOrDefault("DB_SSLMODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, name, sslMode)
}

func getEnvOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func setupIntegrationRepo(t *testing.T) (*PostgresRepository, *pgxpool.Pool) {
	t.Helper()
	initIntegrationLogger()

	dbURL := integrationDBURL()
	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	if _, err = pool.Exec(context.Background(), createLinksSchemaSQL); err != nil {
		pool.Close()
		require.NoError(t, err)
	}
	if _, err = pool.Exec(context.Background(), "TRUNCATE TABLE links RESTART IDENTITY"); err != nil {
		pool.Close()
		require.NoError(t, err)
	}

	return NewPostgresRepository(pool), pool
}

func TestPostgresRepositoryIntegration_CreateGetResolve(t *testing.T) {
	repo, pool := setupIntegrationRepo(t)
	defer pool.Close()

	ctx := context.Background()
	originalURL := "https://example.com/integration"
	shortCode := "AbCdEf123_"

	created, err := repo.Create(ctx, originalURL, shortCode)
	require.NoError(t, err)
	require.Equal(t, shortCode, created.ShortURL)

	got, err := repo.GetByOriginalURL(ctx, originalURL)
	require.NoError(t, err)
	require.Equal(t, shortCode, got.ShortURL)

	resolved, err := repo.ResolveByShortURL(ctx, shortCode)
	require.NoError(t, err)
	require.Equal(t, originalURL, resolved.OriginalURL)
	require.EqualValues(t, 1, resolved.UseCount)
	require.NotNil(t, resolved.LastUsedAt)
}

func TestPostgresRepositoryIntegration_CreateConflicts(t *testing.T) {
	repo, pool := setupIntegrationRepo(t)
	defer pool.Close()

	ctx := context.Background()
	_, err := repo.Create(ctx, "https://one.example", "aBcDeF123_")
	require.NoError(t, err)

	_, err = repo.Create(ctx, "https://one.example", "ZZZZZZZZZZ")
	require.ErrorIs(t, err, domain.ErrAlreadyExists)

	_, err = repo.Create(ctx, "https://two.example", "aBcDeF123_")
	require.ErrorIs(t, err, domain.ErrConflict)
}

func TestPostgresRepositoryIntegration_DeleteExpiredBatch(t *testing.T) {
	repo, pool := setupIntegrationRepo(t)
	defer pool.Close()

	ctx := context.Background()
	cutoff := time.Now().UTC().AddDate(-3, 0, 0)
	old := cutoff.Add(-24 * time.Hour)

	_, err := pool.Exec(ctx, `
INSERT INTO links (original_url, short_url, created_at, last_used_at, use_count)
VALUES
('https://old-created.example', 'oldCreate1_', $1, NULL, 0),
('https://old-used.example', 'oldUsed__1_', now(), $1, 5),
('https://fresh.example', 'fresh_____1', now(), now(), 1)
`, old)
	require.NoError(t, err)

	deleted, err := repo.DeleteExpiredBatch(ctx, cutoff, 10)
	require.NoError(t, err)
	require.EqualValues(t, 2, deleted)
}
