package main

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestNewAppUnknownCommandExitsNonZero(t *testing.T) {
	oldExiter, oldErrWriter := cli.OsExiter, cli.ErrWriter

	t.Cleanup(func() {
		cli.OsExiter, cli.ErrWriter = oldExiter, oldErrWriter
	})

	exitCode := -1
	cli.OsExiter = func(code int) { exitCode = code }

	var stderr bytes.Buffer

	cli.ErrWriter = &stderr

	ctx := t.Context()

	app := newApp(ctx)
	app.Writer = io.Discard

	err := app.Run(ctx, []string{"ge-tax-calc", "foo"})
	require.Error(t, err)

	assert.Equal(t, 3, exitCode)
	assert.Contains(t, stderr.String(), "No help topic for 'foo'")
}

func TestNewAppCommands(t *testing.T) {
	app := newApp(t.Context())

	names := make([]string, 0, len(app.Commands))
	for _, c := range app.Commands {
		names = append(names, c.Name)
	}

	assert.Equal(t, []string{"run", "convert"}, names)
	assert.Nil(t, app.CommandNotFound, "a custom handler would make unknown commands exit with 0")
}

func TestNewAppHelpHasNoBannerOrExitLine(t *testing.T) {
	var out bytes.Buffer

	ctx := t.Context()

	app := newApp(ctx)
	app.Writer = &out

	require.NoError(t, app.Run(ctx, []string{"ge-tax-calc", "--help"}))

	assert.Contains(t, out.String(), "COMMANDS:")
	assert.NotContains(t, out.String(), "██", "no banner for --help")
	assert.NotContains(t, out.String(), "Exit...", "no exit line for --help")
}

func TestCommandsHaveBannerAndExitLine(t *testing.T) {
	for _, c := range commands() {
		assert.NotNil(t, c.Before, c.Name)
		assert.NotNil(t, c.After, c.Name)
	}
}

func TestOnExitWritesToCommandWriter(t *testing.T) {
	var out bytes.Buffer

	require.NoError(t, onExit(t.Context(), &cli.Command{Writer: &out}))

	assert.Equal(t, "Exit...\n", out.String())
}
