package main

import (
	"context"

	"github.com/urfave/cli/v2"
)

func commands(ctx context.Context) []*cli.Command {
	const (
		cmdRun     = "run"
		cmdConvert = "convert"
	)

	cmds := []*cli.Command{
		{
			Name:                   cmdRun,
			Aliases:                nil,
			Usage:                  "Runs taxes calculations",
			UsageText:              "",
			Description:            "",
			ArgsUsage:              "",
			Category:               "",
			BashComplete:           nil,
			Before:                 nil,
			After:                  nil,
			Action:                 menuCalcTaxes(ctx),
			OnUsageError:           nil,
			Subcommands:            nil,
			Flags:                  nil,
			SkipFlagParsing:        false,
			HideHelp:               false,
			HideHelpCommand:        false,
			Hidden:                 false,
			UseShortOptionHandling: false,
			HelpName:               "",
			CustomHelpTemplate:     "",
		},
		{
			Name:                   cmdConvert,
			Aliases:                nil,
			Usage:                  "Runs currency converter",
			UsageText:              "",
			Description:            "",
			ArgsUsage:              "",
			Category:               "",
			BashComplete:           nil,
			Before:                 nil,
			After:                  nil,
			Action:                 menuConvert(ctx),
			OnUsageError:           nil,
			Subcommands:            nil,
			Flags:                  nil,
			SkipFlagParsing:        false,
			HideHelp:               false,
			HideHelpCommand:        false,
			Hidden:                 false,
			UseShortOptionHandling: false,
			HelpName:               "",
			CustomHelpTemplate:     "",
		},
	}

	return cmds
}
