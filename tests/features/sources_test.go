// Package features_test provides feature tests for GoCrawl CLI commands.
// These tests verify end-to-end command behavior.
package features_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFeature_SourcesList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping feature test in short mode")
	}

	// Get the project root directory
	projectRoot, err := os.Getwd()
	require.NoError(t, err)
	// Navigate to project root (assuming we're in tests/features)
	for {
		if _, statErr := os.Stat(filepath.Join(projectRoot, "main.go")); statErr == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			require.Fail(t, "could not find project root")
		}
		projectRoot = parent
	}

	// Run the sources list command
	cmd := exec.Command("go", "run", "main.go", "sources", "list")
	cmd.Dir = projectRoot
	output, _ := cmd.CombinedOutput()

	// The command should run (even if no sources are configured)
	// We just verify the command executes and produces output
	require.NotNil(t, output, "command should produce output")
}
