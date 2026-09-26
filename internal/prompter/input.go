package prompter

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/HFunction2013/coj-cli/internal/term"
)

// readLine prompts for one line. When secret is true the terminal echo is
// disabled first (password / token entry).
func readLine(in FileReader, out FileWriter, prompt, defaultValue string, secret bool) (string, error) {
	if prompt != "" {
		if defaultValue != "" && !secret {
			fmt.Fprintf(out, "%s%s?%s %s %s(%s)%s ", cyan, bold, reset, prompt, dim, defaultValue, reset)
		} else {
			fmt.Fprintf(out, "%s%s?%s %s ", cyan, bold, reset, prompt)
		}
	}

	if secret {
		fd := int(in.Fd())
		restore, err := term.SetRaw(fd)
		if err == nil {
			defer func() {
				fmt.Fprint(out, "\r\n")
				restore()
			}()
			return readSecretLine(in)
		}
		// No raw mode: fall through and read in the clear rather than fail.
	}

	// bufio over the raw FileReader; it only needs Read.
	r := bufio.NewReader(io.Reader(in))
	line, err := r.ReadString('\n')
	fmt.Fprint(out, "\r\n")
	if err != nil && line == "" {
		return "", err
	}
	v := strings.TrimRight(line, "\r\n")
	if v == "" {
		return defaultValue, nil
	}
	return v, nil
}

func readSecretLine(in FileReader) (string, error) {
	var buf []byte
	one := make([]byte, 1)
	for {
		n, err := in.Read(one)
		if err != nil || n == 0 {
			if len(buf) == 0 {
				return "", err
			}
			break
		}
		c := one[0]
		if c == '\r' || c == '\n' {
			break
		}
		if c == 0x03 { // Ctrl+C
			return "", ErrAborted
		}
		if c == 0x7f || c == 0x08 { // backspace
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
			}
			continue
		}
		buf = append(buf, c)
	}
	return string(buf), nil
}
