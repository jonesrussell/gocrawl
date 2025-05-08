package mappings

// PageMapping represents the Elasticsearch mapping for pages
type PageMapping struct {
	Settings PageSettings `json:"settings"`
	Mappings PageMappings `json:"mappings"`
}

// PageSettings defines index-level settings
type PageSettings struct {
	BaseSettings
}

// PageMappings defines the field mappings for pages
type PageMappings struct {
	Properties PageProperties `json:"properties"`
}

// PageProperties defines the properties for each field in the page mapping
type PageProperties struct {
	ID          Field `json:"id"`
	URL         Field `json:"url"`
	Title       Field `json:"title"`
	Description Field `json:"description"`
	Content     Field `json:"content"`
	Language    Field `json:"language"`
	CreatedAt   Field `json:"created_at"`
	UpdatedAt   Field `json:"updated_at"`
	Source      Field `json:"source"`
	Links       Field `json:"links"`
	Headers     Field `json:"headers"`
	Images      Field `json:"images"`
}

// NewPageMapping creates a new page mapping with default settings
func NewPageMapping() *PageMapping {
	return &PageMapping{
		Settings: PageSettings{
			BaseSettings: DefaultSettings(),
		},
		Mappings: PageMappings{
			Properties: PageProperties{
				ID: Field{
					Type: "keyword",
				},
				URL: Field{
					Type: "keyword",
				},
				Title: Field{
					Type:     "text",
					Analyzer: "standard",
				},
				Description: Field{
					Type:     "text",
					Analyzer: "standard",
				},
				Content: Field{
					Type:     "text",
					Analyzer: "standard",
				},
				Language: Field{
					Type: "keyword",
				},
				CreatedAt: Field{
					Type:   "date",
					Format: "strict_date_optional_time||epoch_millis",
				},
				UpdatedAt: Field{
					Type:   "date",
					Format: "strict_date_optional_time||epoch_millis",
				},
				Source: Field{
					Type: "keyword",
				},
				Links: Field{
					Type: "keyword",
				},
				Headers: Field{
					Type:     "text",
					Analyzer: "standard",
				},
				Images: Field{
					Type: "keyword",
				},
			},
		},
	}
}

// GetJSON returns the page mapping as a JSON string
func (m *PageMapping) GetJSON() (string, error) {
	return ToJSON(m)
}

// Validate validates the page mapping configuration
func (m *PageMapping) Validate() error {
	return ValidateSettings(m.Settings.BaseSettings)
}
