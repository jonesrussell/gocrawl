package mappings

// ArticleMapping represents the Elasticsearch mapping for articles
type ArticleMapping struct {
	Settings ArticleSettings `json:"settings"`
	Mappings ArticleMappings `json:"mappings"`
}

// ArticleSettings defines index-level settings
type ArticleSettings struct {
	BaseSettings
}

// ArticleMappings defines the field mappings for articles
type ArticleMappings struct {
	Properties ArticleProperties `json:"properties"`
}

// ArticleProperties defines the properties for each field in the article mapping
type ArticleProperties struct {
	ID          Field `json:"id"`
	URL         Field `json:"url"`
	Title       Field `json:"title"`
	Content     Field `json:"content"`
	Author      Field `json:"author"`
	PublishedAt Field `json:"published_at"`
	CreatedAt   Field `json:"created_at"`
	UpdatedAt   Field `json:"updated_at"`
	Source      Field `json:"source"`
	Tags        Field `json:"tags"`
}

// Field represents an Elasticsearch field mapping
type Field struct {
	Type     string `json:"type,omitempty"`
	Analyzer string `json:"analyzer,omitempty"`
	Format   string `json:"format,omitempty"`
}

// NewArticleMapping creates a new article mapping with default settings
func NewArticleMapping() *ArticleMapping {
	return &ArticleMapping{
		Settings: ArticleSettings{
			BaseSettings: DefaultSettings(),
		},
		Mappings: ArticleMappings{
			Properties: ArticleProperties{
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
				Content: Field{
					Type:     "text",
					Analyzer: "standard",
				},
				Author: Field{
					Type: "keyword",
				},
				PublishedAt: Field{
					Type:   "date",
					Format: "strict_date_optional_time||epoch_millis",
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
				Tags: Field{
					Type: "keyword",
				},
			},
		},
	}
}

// GetJSON returns the article mapping as a JSON string
func (m *ArticleMapping) GetJSON() (string, error) {
	return ToJSON(m)
}

// Validate validates the article mapping configuration
func (m *ArticleMapping) Validate() error {
	return ValidateSettings(m.Settings.BaseSettings)
}
