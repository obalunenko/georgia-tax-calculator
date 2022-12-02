package main

import (
	"context"
	"fmt"
	"text/tabwriter"

	log "github.com/obalunenko/logger"
	"github.com/urfave/cli/v2"
)

func printHeader(_ context.Context) cli.BeforeFunc {
	const (
		padding  int  = 1
		minWidth int  = 0
		tabWidth int  = 0
		padChar  byte = ' '
	)

	return func(c *cli.Context) error {
		w := tabwriter.NewWriter(c.App.Writer, minWidth, tabWidth, padding, padChar, tabwriter.TabIndent)

		_, err := fmt.Fprintf(w, `

 ██████╗ ███████╗ ████████╗ █████╗ ██╗  ██╗      ██████╗ █████╗ ██╗      ██████╗
██╔════╝ ██╔════╝ ╚══██╔══╝██╔══██╗╚██╗██╔╝     ██╔════╝██╔══██╗██║     ██╔════╝
██║  ███╗█████╗█████╗██║   ███████║ ╚███╔╝█████╗██║     ███████║██║     ██║     
██║   ██║██╔══╝╚════╝██║   ██╔══██║ ██╔██╗╚════╝██║     ██╔══██║██║     ██║     
╚██████╔╝███████╗    ██║   ██║  ██║██╔╝ ██╗     ╚██████╗██║  ██║███████╗╚██████╗
 ╚═════╝ ╚══════╝    ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝      ╚═════╝╚═╝  ╚═╝╚══════╝ ╚═════╝
                                                                                

`)
		if err != nil {
			return fmt.Errorf("print version: %w", err)
		}

		return nil
	}
}

func notFound(ctx context.Context) cli.CommandNotFoundFunc {
	return func(c *cli.Context, command string) {
		if _, err := fmt.Fprintf(
			c.App.Writer,
			"Command [%s] not supported.\nTry --help flag to see how to use it\n",
			command,
		); err != nil {
			log.WithError(ctx, err).Fatal("Failed to print not found message")
		}
	}
}

func onExit(_ context.Context) cli.AfterFunc {
	return func(c *cli.Context) error {
		fmt.Println("Exit...")

		return nil
	}
}
