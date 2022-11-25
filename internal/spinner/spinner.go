package spinner

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
)

type StopFunc func()

// Start runs the displaying of spinner to handle long time operations. Returns stop func.
func Start() StopFunc {
	const delayMs = 100

	s := spinner.New(
		spinner.CharSets[62],
		delayMs*time.Millisecond,
		spinner.WithFinalMSG("done!"),
		spinner.WithHiddenCursor(true),
		spinner.WithColor("yellow"),
		spinner.WithWriter(os.Stderr),
	)

	s.Prefix = "in progress..."

	s.Start()

	return func() {
		s.Stop()
	}
}
