// Package iostreams centralises terminal state, following gh's
// pkg/iostreams: commands ask CanPrompt() rather than probing the terminal
// themselves, so behaviour stays consistent and testable.
package iostreams

import (
	"os"

	"github.com/HFunction2013/coj-cli/internal/browser"
	"github.com/HFunction2013/coj-cli/internal/prompter"
	"github.com/HFunction2013/coj-cli/internal/term"
)

// IOStreams carries the process streams plus terminal capability flags.
type IOStreams struct {
	In     *os.File
	Out    *os.File
	ErrOut *os.File

	stdinTTYOverride  *bool
	stdoutTTYOverride *bool
	neverPrompt       bool
	browser           browser.Browser
	launcher          string
}

// System returns streams bound to the real terminal.
func System() *IOStreams {
	io := &IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	// GH_PROMPT_DISABLED is the precedent: any value disables prompting.
	// We accept both COJ_PROMPT_DISABLED and the gh spelling.
	if _, disabled := os.LookupEnv("COJ_PROMPT_DISABLED"); disabled {
		io.SetNeverPrompt(true)
	} else if _, disabled := os.LookupEnv("GH_PROMPT_DISABLED"); disabled {
		io.SetNeverPrompt(true)
	}
	return io
}

// Test returns streams detached from any terminal.
func Test() *IOStreams {
	io := System()
	io.SetStdinTTY(false)
	io.SetStdoutTTY(false)
	return io
}

func (s *IOStreams) SetStdinTTY(v bool)  { s.stdinTTYOverride = &v }
func (s *IOStreams) SetStdoutTTY(v bool) { s.stdoutTTYOverride = &v }

func (s *IOStreams) IsStdinTTY() bool {
	if s.stdinTTYOverride != nil {
		return *s.stdinTTYOverride
	}
	return term.IsTerminal(int(s.In.Fd()))
}

func (s *IOStreams) IsStdoutTTY() bool {
	if s.stdoutTTYOverride != nil {
		return *s.stdoutTTYOverride
	}
	return term.IsTerminal(int(s.Out.Fd()))
}

// CanPrompt reports whether interactive prompting is allowed.
// gh requires stdin AND stdout to be a terminal, and honours neverPrompt.
func (s *IOStreams) CanPrompt() bool {
	if s.neverPrompt {
		return false
	}
	return s.IsStdinTTY() && s.IsStdoutTTY()
}

func (s *IOStreams) SetNeverPrompt(v bool) { s.neverPrompt = v }
func (s *IOStreams) GetNeverPrompt() bool  { return s.neverPrompt }

// Prompter builds a Prompter bound to these streams.
func (s *IOStreams) Prompter() prompter.Prompter {
	return prompter.New(s.In, s.Out, s.ErrOut, s.CanPrompt())
}

// Browser returns the opener for these streams.
// gh keeps Browser behind an interface for the same reason: tests assert on
// the URL without launching anything.
func (s *IOStreams) Browser() browser.Browser {
	if s.browser != nil {
		return s.browser
	}
	return browser.New(s.launcher, s.Out, s.ErrOut)
}

// SetBrowser installs a custom opener (tests, --no-browser).
func (s *IOStreams) SetBrowser(b browser.Browser) { s.browser = b }

// SetLauncher overrides the browser command from config.
func (s *IOStreams) SetLauncher(cmd string) { s.launcher = cmd }
