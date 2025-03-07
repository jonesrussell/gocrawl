package collector

// Logger defines the interface for logging operations
type Logger interface {
	Debug(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// ArticleProcessor defines the interface for processing articles
type ArticleProcessor interface {
	Process(article interface{}) error
}

// ContentProcessor defines the interface for processing content
type ContentProcessor interface {
	Process(content string) (string, error)
}
