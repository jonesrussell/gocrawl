package common

import "errors"

var (

	// ErrLoggerRequired is returned when CommandDeps.Logger is nil

	ErrLoggerRequired = errors.New("logger is required")

	// ErrConfigRequired is returned when CommandDeps.Config is nil

	ErrConfigRequired = errors.New("config is required")

	// ErrInvalidDeps is returned when dependencies fail validation

	ErrInvalidDeps = errors.New("invalid dependencies")
)
