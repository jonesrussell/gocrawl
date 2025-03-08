// Package mapping provides types and utilities for Elasticsearch index mappings.
package mapping

import (
	"encoding/json"
)

// Mapping represents an Elasticsearch index mapping.
type Mapping struct {
	Settings Settings              `json:"settings,omitempty"`
	Mappings map[string]Properties `json:"mappings,omitempty"`
}

// Settings represents index settings.
type Settings struct {
	NumberOfShards   int `json:"number_of_shards,omitempty"`
	NumberOfReplicas int `json:"number_of_replicas,omitempty"`
}

// Properties represents field mappings.
type Properties map[string]Field

// Field represents a field mapping.
type Field struct {
	Type       string     `json:"type,omitempty"`
	Analyzer   string     `json:"analyzer,omitempty"`
	Format     string     `json:"format,omitempty"`
	Properties Properties `json:"properties,omitempty"`
	Fields     Properties `json:"fields,omitempty"`
	Dynamic    *bool      `json:"dynamic,omitempty"`
}

// NewMapping creates a new mapping with default settings.
func NewMapping() *Mapping {
	return &Mapping{
		Settings: Settings{
			NumberOfShards:   1,
			NumberOfReplicas: 1,
		},
		Mappings: make(map[string]Properties),
	}
}

// AddType adds a new type mapping.
func (m *Mapping) AddType(typeName string, properties Properties) {
	m.Mappings[typeName] = properties
}

// AddField adds a field to a type mapping.
func (m *Mapping) AddField(typeName, fieldName string, field Field) {
	if _, ok := m.Mappings[typeName]; !ok {
		m.Mappings[typeName] = make(Properties)
	}
	m.Mappings[typeName][fieldName] = field
}

// JSON returns the mapping as a JSON string.
func (m *Mapping) JSON() (string, error) {
	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// NewTextField creates a text field mapping.
func NewTextField(analyzer string) Field {
	return Field{
		Type:     "text",
		Analyzer: analyzer,
	}
}

// NewKeywordField creates a keyword field mapping.
func NewKeywordField() Field {
	return Field{
		Type: "keyword",
	}
}

// NewDateField creates a date field mapping.
func NewDateField(format string) Field {
	if format == "" {
		format = "strict_date_optional_time||epoch_millis"
	}
	return Field{
		Type:   "date",
		Format: format,
	}
}

// NewObjectField creates an object field mapping.
func NewObjectField(properties Properties) Field {
	return Field{
		Type:       "object",
		Properties: properties,
	}
}

// NewNestedField creates a nested field mapping.
func NewNestedField(properties Properties) Field {
	return Field{
		Type:       "nested",
		Properties: properties,
	}
}

// DisableDynamic disables dynamic mapping for a field.
func (f Field) DisableDynamic() Field {
	disabled := false
	f.Dynamic = &disabled
	return f
}

// WithFields adds multi-fields to a field mapping.
func (f Field) WithFields(fields Properties) Field {
	f.Fields = fields
	return f
}
