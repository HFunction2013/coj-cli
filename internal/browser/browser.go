// Package browser opens pages outside the terminal.
//
// gh models this as a one-method interface (internal/browser: Browser with
// Browse(string) error) precisely so tests can assert "we opened this URL"
// without launching anything. We do the same.
package browser

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Browser opens a URL.
type Browser interface {
	Browse(url string) error
}

// New returns a Browser that shells out to the platform opener.
// The launcher argument lets config override the command.
func New(launcher string, stdout, stderr *os.File) Browser {
	return &realBrowser{launcher: launcher, stdout: stdout, stderr: stderr}
}

type realBrowser struct {
	launcher string
	stdout   *os.File
	stderr   *os.File
}

func (b *realBrowser) Browse(url string) error {
	name, args := b.command()
	args = append(args, url)
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		// gh treats a failed launch as non-fatal: print the URL so the user
		// can copy it, and let the command continue.
		fmt.Fprintf(b.stderr, "Could not open browser: %v\n", err)
		fmt.Fprintf(b.stderr, "Open this URL manually:\n  %s\n", url)
		return nil
	}
	go func() { _ = cmd.Wait() }()
	fmt.Fprintf(b.stderr, "Opening %s in your browser.\n", url)
	return nil
}

func (b *realBrowser) command() (string, []string) {
	if b.launcher != "" {
		return b.launcher, nil
	}
	switch runtime.GOOS {
	case "darwin":
		return "open", nil
	case "windows":
		return "cmd", []string{"/c", "start"}
	default:
		for _, c := range []string{"xdg-open", "x-www-browser", "www-browser", "wslview", "sensible-browser"} {
			if _, err := exec.LookPath(c); err == nil {
				return c, nil
			}
		}
		return "xdg-open", nil
	}
}

// Stub records URLs instead of opening them (tests, --no-browser).
type Stub struct {
	URLs []string
}

func (s *Stub) Browse(url string) error {
	s.URLs = append(s.URLs, url)
	return nil
}

var _ Browser = (*Stub)(nil)

// PageURL joins a site and a page path into a full URL.
// The CandyOJ frontend uses hash routing.
func PageURL(host, page string) string {
	host = strings.TrimRight(host, "/")
	if !strings.HasPrefix(page, "/") {
		page = "/" + page
	}
	return host + "/#" + page
}
