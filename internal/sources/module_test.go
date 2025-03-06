package sources_test

import (
	"os"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestModule(t *testing.T) {
	t.Parallel()

	// Create a temporary YAML file with source configuration
	tmpFile, err := os.CreateTemp(t.TempDir(), "sources*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`
sources:
  - name: test
    url: https://example.com
    index: test_index
`)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	// Create a symbolic link to sources.yml
	err = os.Symlink(tmpFile.Name(), "sources.yml")
	require.NoError(t, err)
	defer os.Remove("sources.yml")

	var s *sources.Sources
	app := fxtest.New(t,
		sources.Module,
		fx.Invoke(func(sources *sources.Sources) {
			s = sources
		}),
	)

	app.RequireStart().RequireStop()

	require.NotNil(t, s)
	require.Len(t, s.Sources, 1)
	require.Equal(t, "test", s.Sources[0].Name)
	require.Equal(t, "https://example.com", s.Sources[0].URL)
	require.Equal(t, "test_index", s.Sources[0].Index)
}
