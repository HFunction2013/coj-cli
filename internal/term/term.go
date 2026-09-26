// Package term reads passwords across platforms, with echo disabled.
//
// Why not call syscall directly: the terminal control constants are
// platform-specific —
//
//	Linux            TCGETS / TCSETS
//	macOS / *BSD     TIOCGETA / TIOCSETA
//	Windows          GetConsoleMode / SetConsoleMode
//
// So the implementation is split per platform (term_linux.go,
// term_darwin.go, term_windows.go …) and selected by build tags,
// without taking on any third-party dependency.
//
// If you would rather use golang.org/x/term, see the "Cross-platform notes"
// section of the README; replacing this package takes a single call.
package term

import "fmt"

// ReadPassword prompts for and reads one line of input.
// Echo is disabled in a terminal; non-interactive environments (pipe, CI)
// fall back to reading a line of plain input.
func ReadPassword(prompt string) (string, error) {
	return readPassword(prompt)
}

// IsTerminal reports whether fd is attached to a terminal.
func IsTerminal(fd int) bool {
	return isTerminal(fd)
}

// ErrNoTerminal is returned when input cannot be read safely.
var ErrNoTerminal = fmt.Errorf("not an interactive terminal")

// ErrNoRawMode means raw mode is unavailable on this platform/environment.
var ErrNoRawMode = fmt.Errorf("raw mode unavailable")

// SetRaw switches the terminal to raw mode (no echo, no line buffering) and
// returns a function that restores the previous state. On failure it returns
// ErrNoRawMode and the caller should fall back to non-interactive prompts.
func SetRaw(fd int) (restore func(), err error) {
	return setRaw(fd)
}
