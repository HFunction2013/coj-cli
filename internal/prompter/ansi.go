package prompter

import "fmt"

// ANSI helpers. gh leans on survey/huh for this; we keep a minimal set so the
// package stays dependency-free while producing the same look.
const (
	esc         = "\033["
	hideCursor  = esc + "?25l"
	showCursor  = esc + "?25h"
	clearLine   = esc + "2K"
	cursorStart = "\r"
	bold        = esc + "1m"
	reset       = esc + "0m"
	cyan        = esc + "36m"
	green       = esc + "32m"
	dim         = esc + "2m"
)

func moveUp(n int) string { return fmt.Sprintf("%s%dA", esc, n) }
