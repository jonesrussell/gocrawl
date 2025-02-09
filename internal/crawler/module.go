package crawler

import (
	"go.uber.org/fx"
)

// Module provides the crawler as an Fx module
var Module = fx.Module("crawler",
	fx.Provide(
		NewCrawler,
	),
)
