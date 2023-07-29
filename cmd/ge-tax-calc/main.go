// ge-tax-calc is CLI for taxes calculations.
package main

import (
	"context"
	"os"

	log "github.com/obalunenko/logger"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx := context.Background()

	ctx = log.ContextWithLogger(ctx, log.FromContext(ctx))

	app := cli.NewApp()
	app.Name = "ge-tax-calc"
	app.Description = "Helper tool for preparing tax declarations in Georgia.\n" +
		"It get income amount in received currency, converts it to GEL according to \n" +
		"official rates on date of income and calculates tax amount \n" +
		"according to selected taxes category."
	app.Usage = `A command line tool helper for preparing tax declaration in Georgia `
	app.Authors = []*cli.Author{
		{
			Name:  "Oleg Balunenko",
			Email: "oleg.balunenko@gmail.com",
		},
	}

	app.CommandNotFound = notFound
	app.Commands = commands(ctx)
	app.Version = printVersion(ctx)
	app.Before = printHeader
	app.After = onExit

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.WithError(ctx, err).Fatal("Run failed")
	}
}
