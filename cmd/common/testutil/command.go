// Package testutil provides utilities for testing Cobra commands
package testutil

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// ExecuteCommand executes a Cobra command and returns its output and error
func ExecuteCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

// ExecuteCommandC executes a Cobra command and returns the command instance, output, and error
func ExecuteCommandC(t *testing.T, cmd *cobra.Command, args ...string) (*cobra.Command, string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	c, err := cmd.ExecuteC()
	return c, buf.String(), err
}

// ExecuteCommandWithContext executes a Cobra command with a context and returns its output and error
func ExecuteCommandWithContext(t *testing.T, ctx context.Context, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	return buf.String(), err
}

// ExecuteCommandWithContextC executes a Cobra command with a context and returns the command instance, output, and error
func ExecuteCommandWithContextC(t *testing.T, ctx context.Context, cmd *cobra.Command, args ...string) (*cobra.Command, string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	cmd.SetContext(ctx)

	c, err := cmd.ExecuteC()
	return c, buf.String(), err
}

// RequireCommandSuccess executes a command and requires it to succeed
func RequireCommandSuccess(t *testing.T, cmd *cobra.Command, args ...string) string {
	t.Helper()
	output, err := ExecuteCommand(t, cmd, args...)
	require.NoError(t, err)
	return output
}

// RequireCommandFailure executes a command and requires it to fail
func RequireCommandFailure(t *testing.T, cmd *cobra.Command, args ...string) string {
	t.Helper()
	output, err := ExecuteCommand(t, cmd, args...)
	require.Error(t, err)
	return output
}
