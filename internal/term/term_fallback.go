//go:build !linux && !darwin && !freebsd && !netbsd && !openbsd && !dragonfly && !windows

package term

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
)

// Fallback: try POSIX stty to disable echo, else read in the clear.
func isTerminal(fd int) bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func readPassword(prompt string) (string, error) {
	if !isTerminal(int(os.Stdin.Fd())) {
		return readLine(prompt)
	}
	// stty is widely available on POSIX systems; failure is not fatal
	if err := exec.Command("stty", "-echo").Run(); err == nil {
		defer exec.Command("stty", "echo").Run()
	}
	return readLine(prompt)
}

func readLine(prompt string) (string, error) {
	if prompt != "" {
		if _, err := os.Stderr.WriteString(prompt); err != nil {
			return "", err
		}
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if isTerminal(int(os.Stdin.Fd())) && prompt != "" {
		os.Stderr.WriteString("\n")
	}
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func setRaw(fd int) (func(), error) {
	return nil, ErrNoRawMode
}
