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

	app := cli.Command{}
	app.Name = "ge-tax-calc"
	app.Description = "Helper tool for preparing tax declarations in Georgia.\n" +
		"It get income amount in received currency, converts it to GEL according to \n" +
		"official rates on date of income and calculates tax amount \n" +
		"according to selected taxes category."
	app.Usage = `A command line tool helper for preparing tax declaration in Georgia `
	app.Authors = []any{
		"Oleg Balunenko <oleg.balunenko@gmail.com>",
	}

	app.CommandNotFound = notFound
	app.Commands = commands()
	app.Version = printVersion(ctx)
	app.Before = printHeader
	app.After = onExit

	if err := app.Run(ctx, os.Args); err != nil {
		log.WithError(ctx, err).Fatal("Run failed")
	}
}
