package service

import (
	"context"
	"errors"
	"strings"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/domain"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/repository"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/pkg/logger"
)

const maxCreateAttempts = 10

type URLService struct {
	repository repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{
		repository: repo,
	}
}

func (s *URLService) Create(ctx context.Context, rawOriginalURL string) (*domain.Link, error) {
	normalizedURL, err := domain.NormalizeOriginalURL(rawOriginalURL)
	if err != nil {
		return nil, err
	}

	existing, err := s.repository.GetByOriginalURL(ctx, normalizedURL)
	if err == nil {
		return existing, nil
	}

	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	for nonce := 0; nonce < maxCreateAttempts; nonce++ {
		code := GenerateShortCode(normalizedURL, nonce)
		if err = domain.ValidateShortCode(code); err != nil {
			return nil, err
		}

		link, createErr := s.repository.Create(ctx, normalizedURL, code)
		if createErr == nil {
			logger.Logger.Infow("short url created in service", "original_url", normalizedURL, "short_url", code, "nonce", nonce)
			return link, nil
		}

		if errors.Is(createErr, domain.ErrAlreadyExists) {
			existing, getErr := s.repository.GetByOriginalURL(ctx, normalizedURL)
			if getErr == nil {
				return existing, nil
			}
			return nil, getErr
		}

		if errors.Is(createErr, domain.ErrConflict) {
			continue
		}

		return nil, createErr
	}

	return nil, domain.ErrConflict
}

func (s *URLService) Resolve(ctx context.Context, shortURL string) (*domain.Link, error) {
	shortURL = strings.TrimSpace(shortURL)
	if shortURL == "" {
		return nil, domain.ErrBadRequest
	}

	return s.repository.ResolveByShortURL(ctx, shortURL)
}
