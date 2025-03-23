package crawl_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/cmd/crawl"
	"github.com/stretchr/testify/assert"
)

// Since ValidateArgs is not exported, we'll test through the Command's Args function
func TestCommandValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "with source name",
			args:    []string{"test-source"},
			wantErr: false,
		},
		{
			name:    "too many args",
			args:    []string{"test-source", "extra-arg"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := crawl.Command()
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
