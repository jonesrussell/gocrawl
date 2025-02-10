package storage

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	t.Run("ErrInvalidHits", func(t *testing.T) {
		err := ErrInvalidHits
		assert.True(t, errors.Is(err, ErrInvalidHits))
		assert.Equal(t, "invalid response format: hits not found", err.Error())
	})

	t.Run("ErrInvalidHitsArray", func(t *testing.T) {
		err := ErrInvalidHitsArray
		assert.True(t, errors.Is(err, ErrInvalidHitsArray))
		assert.Equal(t, "invalid response format: hits array not found", err.Error())
	})

	t.Run("ErrMissingURL", func(t *testing.T) {
		err := ErrMissingURL
		assert.True(t, errors.Is(err, ErrMissingURL))
		assert.Equal(t, "elasticsearch URL is required", err.Error())
	})

	t.Run("ErrInvalidScrollID", func(t *testing.T) {
		err := ErrInvalidScrollID
		assert.True(t, errors.Is(err, ErrInvalidScrollID))
		assert.Equal(t, "invalid scroll ID", err.Error())
	})

	t.Run("Error wrapping", func(t *testing.T) {
		wrappedErr := errors.New("wrapped error")
		err := fmt.Errorf("failed with: %w", wrappedErr)
		assert.True(t, errors.Is(err, wrappedErr))
	})
}
