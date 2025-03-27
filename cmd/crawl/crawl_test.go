package crawl_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/crawl"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCommandValidation(t *testing.T) {
	t.Parallel()

	// Create root command with proper setup
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}
	cmd := crawl.Command()
	rootCmd.AddCommand(cmd)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "valid source name",
			args:    []string{"test-source"},
			wantErr: false,
		},
		{
			name:    "too many arguments",
			args:    []string{"test-source", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd.SetContext(t.Context())

			// Validate arguments against the crawl command
			err := cmd.ValidateArgs(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
