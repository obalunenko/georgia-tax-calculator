package main

import (
	"context"
	"fmt"
	"text/tabwriter"

	log "github.com/obalunenko/logger"
	"github.com/urfave/cli/v3"
)

func printHeader(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	const (
		padding  int  = 1
		minWidth int  = 0
		tabWidth int  = 0
		padChar  byte = ' '
	)

	w := tabwriter.NewWriter(cmd.Writer, minWidth, tabWidth, padding, padChar, tabwriter.TabIndent)

	_, err := fmt.Fprintf(w, `

 ██████╗ ███████╗ ████████╗ █████╗ ██╗  ██╗      ██████╗ █████╗ ██╗      ██████╗
██╔════╝ ██╔════╝ ╚══██╔══╝██╔══██╗╚██╗██╔╝     ██╔════╝██╔══██╗██║     ██╔════╝
██║  ███╗█████╗█████╗██║   ███████║ ╚███╔╝█████╗██║     ███████║██║     ██║     
██║   ██║██╔══╝╚════╝██║   ██╔══██║ ██╔██╗╚════╝██║     ██╔══██║██║     ██║     
╚██████╔╝███████╗    ██║   ██║  ██║██╔╝ ██╗     ╚██████╗██║  ██║███████╗╚██████╗
 ╚═════╝ ╚══════╝    ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝      ╚═════╝╚═╝  ╚═╝╚══════╝ ╚═════╝
                                                                                

`)
	if err != nil {
		return ctx, fmt.Errorf("print version: %w", err)
	}

	return ctx, nil
}

func notFound(ctx context.Context, cmd *cli.Command, command string) {
	if _, err := fmt.Fprintf(
		cmd.Writer,
		"Command [%s] not supported.\nTry --help flag to see how to use it\n",
		command,
	); err != nil {
		log.WithError(ctx, err).Fatal("Failed to print not found message")
	}
}

func onExit(_ context.Context, _ *cli.Command) error {
	fmt.Println("Exit...")

	return nil
}
