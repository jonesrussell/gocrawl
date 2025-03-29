// Package testutil provides utilities for testing Cobra commands.
package testutil

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// ExecuteCommand executes a Cobra command for testing.
func ExecuteCommand(t *testing.T, root *cobra.Command, args ...string) (string, error) {
	t.Helper()

	// Create buffers for output
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)

	// Set args and execute
	root.SetArgs(args)
	err := root.Execute()

	return buf.String(), err
}

// ExecuteCommandWithInput executes a Cobra command with input for testing.
func ExecuteCommandWithInput(
	t *testing.T,
	root *cobra.Command,
	input string,
	args ...string,
) (string, error) {
	t.Helper()

	// Create a pipe for input
	r, w, err := os.Pipe()
	require.NoError(t, err)

	// Write input
	_, err = w.WriteString(input)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	// Create buffer for output
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetIn(r)

	// Set args and execute
	root.SetArgs(args)
	err = root.Execute()

	// Read output
	out, readErr := io.ReadAll(buf)
	require.NoError(t, readErr)

	return string(out), err
}

// ExecuteCommandC executes a Cobra command and returns the command instance,
// output, and error.
func ExecuteCommandC(
	t *testing.T,
	root *cobra.Command,
	args ...string,
) (*cobra.Command, string, error) {
	t.Helper()

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err := root.Execute()
	return root, buf.String(), err
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

// ExecuteCommandWithContextC executes a Cobra command with a context.
// Returns the command instance, output, and error.
func ExecuteCommandWithContextC(
	t *testing.T,
	ctx context.Context,
	cmd *cobra.Command,
	args ...string,
) (*cobra.Command, string, error) {
	t.Helper()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(ctx)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return cmd, buf.String(), err
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
