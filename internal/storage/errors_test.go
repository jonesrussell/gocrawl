package storage_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	t.Run("ErrInvalidHits", func(t *testing.T) {
		err := storage.ErrInvalidHits
		require.ErrorIs(t, err, storage.ErrInvalidHits)
		require.Equal(t, "invalid response format: hits not found", err.Error())
	})

	t.Run("ErrInvalidHitsArray", func(t *testing.T) {
		err := storage.ErrInvalidHitsArray
		require.ErrorIs(t, err, storage.ErrInvalidHitsArray)
		require.Equal(t, "invalid response format: hits array not found", err.Error())
	})

	t.Run("ErrMissingURL", func(t *testing.T) {
		err := storage.ErrMissingURL
		require.ErrorIs(t, err, storage.ErrMissingURL)
		require.Equal(t, "elasticsearch URL is required", err.Error())
	})

	t.Run("ErrInvalidScrollID", func(t *testing.T) {
		err := storage.ErrInvalidScrollID
		require.ErrorIs(t, err, storage.ErrInvalidScrollID)
		require.Equal(t, "invalid scroll ID", err.Error())
	})

	t.Run("Error wrapping", func(t *testing.T) {
		wrappedErr := errors.New("wrapped error")
		err := fmt.Errorf("failed with: %w", wrappedErr)
		require.ErrorIs(t, err, wrappedErr)
	})
}
