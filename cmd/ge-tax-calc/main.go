// ge-tax-calc is CLI for taxes calculations.
package main

import (
	"context"
	"os"

	log "github.com/obalunenko/logger"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx := context.Background()

	ctx = log.ContextWithLogger(ctx, log.FromContext(ctx))

	if err := newApp(ctx).Run(ctx, os.Args); err != nil {
		log.WithError(ctx, err).Fatal("Run failed")
	}
}

// newApp builds the root command. CommandNotFound is left unset on purpose:
// urfave/cli then reports an unknown command ("No help topic for ...") and exits with status 3,
// while a custom handler has no way to return an error and the CLI would exit with 0.
func newApp(ctx context.Context) *cli.Command {
	app := cli.Command{}
	app.Name = "ge-tax-calc"
	app.Description = "Helper tool for preparing tax declarations in Georgia.\n" +
		"It gets the income amount in the received currency, converts it to GEL at the\n" +
		"official NBG rate on the date of income, and calculates the tax amount for the\n" +
		"selected tax category."
	app.Usage = `A command line tool helper for preparing tax declaration in Georgia `
	app.Authors = []any{
		"Oleg Balunenko <oleg.balunenko@gmail.com>",
	}

	app.Commands = commands()
	app.Version = printVersion(ctx)
	app.Before = printHeader
	app.After = onExit

	return &app
}
