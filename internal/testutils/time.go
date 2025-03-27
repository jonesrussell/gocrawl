// Package testutils provides shared testing utilities across the application.
package testutils

import "time"

// MockTimeProvider implements TimeProvider for testing
type MockTimeProvider struct {
	currentTime time.Time
}

func (m *MockTimeProvider) Now() time.Time {
	return m.currentTime
}

func (m *MockTimeProvider) Advance(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
}
