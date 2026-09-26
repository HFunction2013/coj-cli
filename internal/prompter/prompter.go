// Package prompter provides interactive prompts, modelled on gh.
//
// The shape follows gh's internal/prompter closely:
//
//   - Prompter is an interface, not a concrete type, so commands depend on
//     behaviour and tests can substitute a mock (see prompter_mock.go, which
//     mirrors gh's moq-generated mock).
//   - MultiSelectWithSearch is the workhorse: rather than materialising every
//     candidate up front, the list carries a "Search" sentinel; choosing it
//     asks for a query and calls searchFunc again. That keeps a 190-endpoint
//     backend responsive and avoids pulling thousands of rows.
//   - Prompts never run when the terminal cannot support them. gh gates on
//     IsStdinTTY() && IsStdoutTTY() and a neverPrompt switch; we do the same.
package prompter

import (
	"fmt"
	"slices"
	"strings"
)

// MultiSelectSearchResult is what a search callback returns.
// Keys are the values eventually returned by MultiSelectWithSearch,
// Labels are what the user sees.
type MultiSelectSearchResult struct {
	Keys        []string
	Labels      []string
	MoreResults int
	Err         error
}

// Prompter is the interface commands depend on.
//
// The generic half mirrors go-gh's prompter so that the calling style is
// identical to gh; the CandyOJ-specific half adds prompts that only make
// sense here.
type Prompter interface {
	// Select prompts for a single option.
	Select(prompt string, defaultValue string, options []string) (int, error)

	// MultiSelect prompts for one or more options.
	MultiSelect(prompt string, defaults []string, options []string) ([]int, error)

	// MultiSelectWithSearch is MultiSelect with an added search option to the
	// list, prompting the user for text input to filter the options via the
	// searchFunc. Items selected in the search are persisted in the list after
	// subsequent searches. Items passed in persistentOptions are always shown
	// in the list, even when not selected. Unlike MultiSelect, it returns the
	// selected option strings, not their indices, since the list of options is
	// dynamic.
	MultiSelectWithSearch(prompt, searchPrompt string, defaults []string, persistentOptions []string, searchFunc func(string) MultiSelectSearchResult) ([]string, error)

	// Input prompts for a string value.
	Input(prompt string, defaultValue string) (string, error)

	// Password prompts for a secret, without echo.
	Password(prompt string) (string, error)

	// Confirm asks a yes/no question.
	Confirm(prompt string, defaultValue bool) (bool, error)

	// ---- CandyOJ-specific ----

	// ConfirmDeletion asks the user to type requiredValue, so that a
	// destructive call (batch delete, one-key password reset) cannot be
	// triggered by a stray Enter.
	ConfirmDeletion(requiredValue string) error
}

// New returns a Prompter bound to the given streams.
// When the terminal cannot prompt, it returns a Prompter that fails fast
// instead of blocking — gh behaves the same way under --no-prompt.
func New(stdin FileReader, stdout, stderr FileWriter, canPrompt bool) Prompter {
	return &terminalPrompter{
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
		canPrompt: canPrompt,
	}
}

// FileReader is the subset of *os.File needed for input.
type FileReader interface {
	Fd() uintptr
	Read(p []byte) (int, error)
}

// FileWriter is the subset of *os.File needed for output.
type FileWriter interface {
	Fd() uintptr
	Write(p []byte) (int, error)
}

// ErrNotInteractive means the terminal cannot prompt.
// Callers surface it as an actionable message rather than hanging.
var ErrNotInteractive = fmt.Errorf("not an interactive terminal; pass the required flags or re-run in a TTY")

// ErrAborted means the user pressed Ctrl+C.
var ErrAborted = fmt.Errorf("aborted")

type terminalPrompter struct {
	stdin     FileReader
	stdout    FileWriter
	stderr    FileWriter
	canPrompt bool
}

func (p *terminalPrompter) guard() error {
	if !p.canPrompt {
		return ErrNotInteractive
	}
	return nil
}

func (p *terminalPrompter) Select(prompt, defaultValue string, options []string) (int, error) {
	if err := p.guard(); err != nil {
		return defaultIndex(options, defaultValue), err
	}
	if len(options) == 0 {
		return 0, fmt.Errorf("no options to select from")
	}
	if len(options) == 1 {
		return 0, nil
	}
	start := 0
	if i := slices.Index(options, defaultValue); i >= 0 {
		start = i
	}
	sel := newSelector(p.stdin, p.stdout, selector{
		Message: prompt,
		Options: options,
		Cursor:  start,
	})
	return sel.runSingle()
}

func (p *terminalPrompter) MultiSelect(prompt string, defaults []string, options []string) ([]int, error) {
	if err := p.guard(); err != nil {
		return defaultIndices(options, defaults), err
	}
	if len(options) == 0 {
		return nil, fmt.Errorf("no options to select from")
	}
	sel := newSelector(p.stdin, p.stdout, selector{
		Message: prompt,
		Options: options,
	})
	sel.checked = map[int]bool{}
	for _, d := range defaults {
		if i := slices.Index(options, d); i >= 0 {
			sel.checked[i] = true
		}
	}
	return sel.runMulti()
}

func (p *terminalPrompter) MultiSelectWithSearch(prompt, searchPrompt string, defaults []string, persistentOptions []string, searchFunc func(string) MultiSelectSearchResult) ([]string, error) {
	if err := p.guard(); err != nil {
		return defaults, err
	}
	return multiSelectWithSearch(p, prompt, searchPrompt, defaults, persistentOptions, searchFunc)
}

func (p *terminalPrompter) Input(prompt, defaultValue string) (string, error) {
	if err := p.guard(); err != nil {
		return defaultValue, err
	}
	return readLine(p.stdin, p.stdout, prompt, defaultValue, false)
}

func (p *terminalPrompter) Password(prompt string) (string, error) {
	if err := p.guard(); err != nil {
		return "", err
	}
	return readLine(p.stdin, p.stdout, prompt, "", true)
}

func (p *terminalPrompter) Confirm(prompt string, defaultValue bool) (bool, error) {
	if err := p.guard(); err != nil {
		return defaultValue, nil
	}
	hint := "y/N"
	if defaultValue {
		hint = "Y/n"
	}
	answer, err := readLine(p.stdin, p.stdout, fmt.Sprintf("%s (%s)", prompt, hint), "", false)
	if err != nil {
		return defaultValue, nil
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	}
	return defaultValue, nil
}

func (p *terminalPrompter) ConfirmDeletion(requiredValue string) error {
	if err := p.guard(); err != nil {
		return err
	}
	for {
		answer, err := readLine(p.stdin, p.stdout,
			fmt.Sprintf("Type %q to confirm:", requiredValue), "", false)
		if err != nil {
			return err
		}
		if answer == requiredValue {
			return nil
		}
		fmt.Fprintf(p.stdout, "%sYou entered: %q%s\n", dim, answer, reset)
	}
}

func defaultIndex(options []string, def string) int {
	if i := slices.Index(options, def); i >= 0 {
		return i
	}
	return 0
}

func defaultIndices(options, defaults []string) []int {
	out := []int{}
	for _, d := range defaults {
		if i := slices.Index(options, d); i >= 0 {
			out = append(out, i)
		}
	}
	return out
}
