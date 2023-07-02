// Package spinner provides helper for setup spinner for long-term operations.
package spinner

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

// StopFunc is a spinner stop func.
type StopFunc func()

// Start runs the displaying of spinner to handle long time operations. Returns stop func.
func Start(name, finishMsg string) StopFunc {
	const delayMs = 100

	s := spinner.New(
		spinner.CharSets[62],
		delayMs*time.Millisecond,
		spinner.WithFinalMSG(fmt.Sprintln(finishMsg)),
		spinner.WithHiddenCursor(true),
		spinner.WithColor("yellow"),
		spinner.WithWriter(os.Stderr),
		spinner.WithSuffix(fmt.Sprintln(name)),
	)

	s.Prefix = "in progress..."

	s.Start()

	return func() {
		s.Stop()
	}
}
