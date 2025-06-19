package main

import (
	"github.com/urfave/cli/v3"
)

func commands() []*cli.Command {
	const (
		cmdRun     = "run"
		cmdConvert = "convert"
	)

	cmds := []*cli.Command{
		{
			Name:   cmdRun,
			Usage:  "Runs taxes calculations",
			Action: menuCalcTaxes,
		},
		{
			Name:   cmdConvert,
			Usage:  "Runs currency converter",
			Action: menuConvert,
		},
	}

	return cmds
}
