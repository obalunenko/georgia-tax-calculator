package main

import (
	"bytes"
	"context"
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

	ctx := context.Background()

	app := newApp(ctx)
	app.Writer = io.Discard

	err := app.Run(ctx, []string{"ge-tax-calc", "foo"})
	require.Error(t, err)

	assert.Equal(t, 3, exitCode)
	assert.Contains(t, stderr.String(), "No help topic for 'foo'")
}

func TestNewAppCommands(t *testing.T) {
	app := newApp(context.Background())

	names := make([]string, 0, len(app.Commands))
	for _, c := range app.Commands {
		names = append(names, c.Name)
	}

	assert.Equal(t, []string{"run", "convert"}, names)
	assert.Nil(t, app.CommandNotFound, "a custom handler would make unknown commands exit with 0")
}
