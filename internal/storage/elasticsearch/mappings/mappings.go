package mappings

import (
	"encoding/json"
	"errors"
	"fmt"
)

// IndexMapping represents a generic Elasticsearch index mapping
type IndexMapping interface {
	// GetJSON returns the mapping as a JSON string
	GetJSON() (string, error)
	// Validate validates the mapping configuration
	Validate() error
}

// BaseSettings defines common index-level settings
type BaseSettings struct {
	NumberOfShards   int `json:"number_of_shards"`
	NumberOfReplicas int `json:"number_of_replicas"`
}

// DefaultSettings returns the default index settings
func DefaultSettings() BaseSettings {
	return BaseSettings{
		NumberOfShards:   1,
		NumberOfReplicas: 1,
	}
}

// ValidateSettings validates the index settings
func ValidateSettings(settings BaseSettings) error {
	if settings.NumberOfShards < 1 {
		return errors.New("number_of_shards must be greater than 0")
	}
	if settings.NumberOfReplicas < 0 {
		return errors.New("number_of_replicas must be greater than or equal to 0")
	}
	return nil
}

// ToJSON converts any mapping to a JSON string with proper indentation
func ToJSON(mapping any) (string, error) {
	data, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal mapping to JSON: %w", err)
	}
	return string(data), nil
}
