package helpers

import (
	"context"
	"errors"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/config"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/database"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/repository"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/pkg/logger"
)

func BuildRepository(cfg *config.Config) (repository.URLRepository, func(), error) {
	switch cfg.Storage {
	case config.StorageMemory:
		logger.Logger.Infow("storage selected", "storage", config.StorageMemory)
		return repository.NewInMemoryRepository(), func() {}, nil
	case config.StoragePostgres:
		logger.Logger.Infow(
			"storage selected",
			"storage", config.StoragePostgres,
			"db_host", cfg.DBHost,
			"db_port", cfg.DBPort,
			"db_name", cfg.DBName,
			"db_user", cfg.DBUser,
		)

		dsn := database.BuildDSN(database.ConnectionConfig{
			Host:     cfg.DBHost,
			Port:     cfg.DBPort,
			User:     cfg.DBUser,
			Password: cfg.DBPassword,
			Name:     cfg.DBName,
			SSLMode:  cfg.DBSSLMode,
		})
		pool, err := database.NewPostgresPool(context.Background(), dsn)
		if err != nil {
			return nil, nil, err
		}
		return repository.NewPostgresRepository(pool), pool.Close, nil
	default:
		return nil, nil, errors.New("unsupported storage")
	}
}
