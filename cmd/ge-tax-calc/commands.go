package main

import (
	"github.com/urfave/cli/v2"
)

func commands() []*cli.Command {
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
			Action:                 menuCalcTaxes,
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
			Action:                 menuConvert,
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
