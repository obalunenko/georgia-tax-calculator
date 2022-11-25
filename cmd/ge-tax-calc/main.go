package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	promptlist "github.com/manifoldco/promptui/list"
	log "github.com/obalunenko/logger"
	"github.com/urfave/cli/v2"

	"github.com/obalunenko/georgia-tax-calculator/internal/converter"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func main() {
	// TODO:
	// 	- get sum of taxes
	//  - read date input
	// 	- read amount input
	ctx := context.Background()

	ctx = log.ContextWithLogger(ctx, log.FromContext(ctx))

	date := time.Now()

	result, err := convert(ctx, convertParams{
		date:   date,
		code:   currencies.EUR,
		amount: 2800.28,
	})
	if err != nil {
		log.WithError(ctx, err).Fatal("Failed to convert")
	}

	fmt.Println(result)
}

func onExit(_ context.Context) cli.AfterFunc {
	return func(c *cli.Context) error {
		fmt.Println("Exit...")

		return nil
	}
}

func makeMenuItemsList(list []string, commands ...string) []string {
	items := make([]string, 0, len(list)+len(commands))

	items = append(items, list...)

	items = append(items, commands...)

	return items
}

const (
	exit     = "exit"
	abort    = "abort"
	back     = "back"
	pageSize = 30
)

func isExit(in string) bool {
	return strings.EqualFold(exit, in)
}

func isAbort(err error) bool {
	return strings.HasSuffix(err.Error(), abort)
}

func isBack(in string) bool {
	return strings.EqualFold(back, in)
}

func menu(ctx context.Context) cli.ActionFunc {
	return func(c *cli.Context) error {
		years := getYears(time.Now())

		items := makeMenuItemsList(years, exit)

		prompt := promptui.Select{
			Label:             "Years menu (exit' for exit)",
			Items:             items,
			Size:              pageSize,
			CursorPos:         0,
			IsVimMode:         false,
			HideHelp:          false,
			HideSelected:      false,
			Templates:         nil,
			Keys:              nil,
			Searcher:          searcher(items),
			StartInSearchMode: false,
			Pointer:           promptui.DefaultCursor,
			Stdin:             nil,
			Stdout:            nil,
		}

		return handleYearChoices(ctx, prompt)
	}
}

func handleYearChoices(ctx context.Context, opt promptui.Select) error {
	for {
		_, choice, err := opt.Run()
		if err != nil {
			if isAbort(err) {
				return nil
			}

			return fmt.Errorf("prompt failed: %w", err)
		}

		if isExit(choice) {
			return nil
		}

		err = menuMonth(ctx, choice)
		if err != nil {
			if errors.Is(err, errExit) {
				return nil
			}

			log.WithError(ctx, err).Error("Puzzle menu failed")

			continue
		}
	}
}

var errExit = errors.New(exit)

func menuMonth(ctx context.Context, year string) error {
	return nil
}

func searcher(items []string) promptlist.Searcher {
	return func(input string, index int) bool {
		itm := items[index]

		itm = strings.ReplaceAll(strings.ToLower(itm), " ", "")

		input = strings.ReplaceAll(strings.ToLower(input), " ", "")

		return strings.Contains(itm, input)
	}
}

func getYears(now time.Time) []string {
	var years []string

	const begin = 2016

	for i := begin; i < now.Year(); i++ {
		years = append(years, strconv.Itoa(i))
	}

	return years
}

func getMonths() []string {
	months := make([]string, 0, 12)

	for i := time.January; i < time.December; i++ {
		months = append(months, i.String())
	}

	return months
}

type convertParams struct {
	date   time.Time
	code   string
	amount float64
}

func convert(ctx context.Context, p convertParams) (string, error) {
	client := nbggovge.New()

	c := converter.NewConverter(client)

	resp, err := c.ConvertToGel(ctx, p.amount, p.code, p.date)
	if err != nil {
		return "", err
	}

	return resp.String(), nil

}
