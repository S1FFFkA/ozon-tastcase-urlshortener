package service

import (
	"testing"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestGenerateShortCode_DeterministicAndValid(t *testing.T) {
	t.Parallel()

	originalURL := "https://example.com/path?q=1"
	nonce := 7

	code1 := GenerateShortCode(originalURL, nonce)
	code2 := GenerateShortCode(originalURL, nonce)

	require.Equal(t, code1, code2, "expected deterministic code")

	require.NoError(t, domain.ValidateShortCode(code1), "expected valid short code")
}

func TestGenerateShortCode_DifferentNonceProducesDifferentCode(t *testing.T) {
	t.Parallel()

	originalURL := "https://example.com/path?q=1"

	code1 := GenerateShortCode(originalURL, 0)
	code2 := GenerateShortCode(originalURL, 1)

	require.NotEqual(t, code1, code2, "expected different codes for different nonce")
}
