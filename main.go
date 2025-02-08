package main

import (
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gocolly/colly/v2"
	"go.uber.org/fx"
)

type Crawler struct {
	Collector *colly.Collector
	ESClient  *elasticsearch.Client
}

func NewCrawler() *Crawler {
	collector := colly.NewCollector()
	esClient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}
	return &Crawler{Collector: collector, ESClient: esClient}
}

func (c *Crawler) Start(url string) {
	c.Collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Println("Found link:", link)
		// Here you can add code to index the link in Elasticsearch
	})

	c.Collector.Visit(url)
}

func main() {
	app := fx.New(
		fx.Provide(NewCrawler),
		fx.Invoke(func(c *Crawler) {
			c.Start("http://example.com") // Start crawling from this URL
		}),
	)

	app.Run()
}
