package repository

import (
	"context"
	"sync"
	"time"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/domain"
)

// Решил для реализации in-memory сделать максимально быстрый поиск через две мапы, где ключи будут
// OriginalURL и ShortURL, а значения будут доменная модель Link
type InMemoryRepository struct {
	mu            sync.RWMutex // Mutex для предотвращения гонки данных
	byOriginalURL map[string]*domain.Link
	byShortURL    map[string]*domain.Link
	nextID        int64
}

var _ URLRepository = (*InMemoryRepository)(nil)

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		byOriginalURL: make(map[string]*domain.Link),
		byShortURL:    make(map[string]*domain.Link),
		nextID:        1,
	}
}

func (r *InMemoryRepository) GetByOriginalURL(_ context.Context, originalURL string) (*domain.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, ok := r.byOriginalURL[originalURL]
	if !ok {
		return nil, domain.ErrNotFound
	}

	copyLink := *link
	return &copyLink, nil // Возвращаем копию для защиты внутреннего состояния
}

func (r *InMemoryRepository) ResolveByShortURL(_ context.Context, shortURL string) (*domain.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	link, ok := r.byShortURL[shortURL]
	if !ok {
		return nil, domain.ErrNotFound
	}

	now := time.Now().UTC()
	link.LastUsedAt = &now
	link.UseCount++

	copyLink := *link
	return &copyLink, nil
}

func (r *InMemoryRepository) Create(_ context.Context, originalURL, shortURL string) (*domain.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byOriginalURL[originalURL]; exists {
		return nil, domain.ErrAlreadyExists
	}

	if _, exists := r.byShortURL[shortURL]; exists {
		return nil, domain.ErrConflict
	}

	now := time.Now().UTC()
	link := &domain.Link{
		ID:          r.nextID,
		OriginalURL: originalURL,
		ShortURL:    shortURL,
		CreatedAt:   now,
		LastUsedAt:  nil,
		UseCount:    0,
	}
	r.nextID++

	r.byOriginalURL[originalURL] = link
	r.byShortURL[shortURL] = link

	copyLink := *link
	return &copyLink, nil
}

func (r *InMemoryRepository) DeleteExpiredBatch(_ context.Context, cutoff time.Time, limit int) (int64, error) {
	if limit <= 0 {
		return 0, domain.ErrBadRequest
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var deleted int64
	for shortURL, link := range r.byShortURL {
		if deleted >= int64(limit) {
			break
		}

		expired := false
		if link.LastUsedAt != nil {
			expired = link.LastUsedAt.Before(cutoff)
		} else {
			expired = link.CreatedAt.Before(cutoff)
		}
		if !expired {
			continue
		}

		delete(r.byShortURL, shortURL)            // В Go 1.25 новая map может сама себя сжимать при удалении достаточного колл-ва.
		delete(r.byOriginalURL, link.OriginalURL) // Поэтому map не переинициализирую после delete
		deleted++
	}

	return deleted, nil
}
