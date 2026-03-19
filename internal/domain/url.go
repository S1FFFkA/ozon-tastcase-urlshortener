package domain

import (
	"net/url"
	"strings"
)

func NormalizeOriginalURL(raw string) (string, error) { // Нормализуем url : Убираем лишнии пробелы, приводим к нижнему регистру хост и протокол.
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidURL
	}

	parsedURL, err := url.ParseRequestURI(raw)
	if err != nil {
		return "", ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", ErrInvalidURL
	}

	if parsedURL.Host == "" {
		return "", ErrInvalidURL
	}

	parsedURL.Host = strings.ToLower(parsedURL.Host)
	parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)

	return parsedURL.String(), nil
}
