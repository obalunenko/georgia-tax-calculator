// ge-tax-calc is CLI for taxes calculations.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/obalunenko/logger"
	"github.com/urfave/cli/v3"
)

func main() {
	// Ctrl+C during a rate request cancels it, so the spinner stops and shows the cursor again.
	// Inside the Bubble Tea forms Ctrl+C is a key press handled by the forms.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	// The banner and the exit line belong to run and convert only, not to --help, --version, or unknown commands.
	app.Commands = commands()
	app.Version = printVersion(ctx)

	return &app
}
