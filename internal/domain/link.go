package domain

import "time"

type Link struct {
	ID          int64
	OriginalURL string
	ShortURL    string
	CreatedAt   time.Time
	LastUsedAt  *time.Time
	UseCount    int64
}
