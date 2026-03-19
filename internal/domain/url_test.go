package domain

import "testing"

import "github.com/stretchr/testify/require"

func TestNormalizeOriginalURL_Success(t *testing.T) {
	t.Parallel()

	got, err := NormalizeOriginalURL("  HTTPS://EXAMPLE.COM/path?q=1 ")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/path?q=1", got)
}

func TestNormalizeOriginalURL_Invalid(t *testing.T) {
	t.Parallel()

	_, err := NormalizeOriginalURL("not-a-url")
	require.Error(t, err)
}
