package domain

import "testing"

import "github.com/stretchr/testify/require"

func TestValidateShortCode_Valid(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateShortCode("Abcdef123_"))
}

func TestValidateShortCode_Invalid(t *testing.T) {
	t.Parallel()

	require.Error(t, ValidateShortCode("short"))
	require.Error(t, ValidateShortCode("Abcdef123!"))
}
