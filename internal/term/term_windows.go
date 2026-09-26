//go:build windows

package term

import (
	"bufio"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	stdInputHandle  = ^uintptr(9) // -10, STD_INPUT_HANDLE
	enableEchoInput = 0x0004
	enableLineInput = 0x0002
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	// GetStdHandle returns 0 rather than -1 on Windows, so validate the handle
	procGetStdHandle   = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

func isTerminal(fd int) bool {
	var mode uint32
	return getConsoleMode(&mode)
}

func getConsoleMode(mode *uint32) bool {
	handle, _, _ := procGetStdHandle.Call(stdInputHandle)
	if handle == 0 || handle == ^uintptr(0) {
		return false
	}
	ret, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(mode)))
	return ret != 0
}

func readPassword(prompt string) (string, error) {
	if !isTerminal(0) {
		return readLine(prompt)
	}

	var mode uint32
	handle, _, _ := procGetStdHandle.Call(stdInputHandle)
	if handle == 0 {
		return readLine(prompt)
	}
	if ret, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode))); ret == 0 {
		return readLine(prompt)
	}

	// clear ENABLE_ECHO_INPUT, keep line input mode
	noecho := mode &^ enableEchoInput
	if ret, _, _ := procSetConsoleMode.Call(handle, uintptr(noecho)); ret != 0 {
		defer procSetConsoleMode.Call(handle, uintptr(mode))
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
	if isTerminal(0) {
		os.Stderr.WriteString("\r\n")
	}
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func setRaw(fd int) (func(), error) {
	// The Windows console has no POSIX raw mode;
	// disabling line input and echo is enough to read key by key.
	var mode uint32
	handle, _, _ := procGetStdHandle.Call(stdInputHandle)
	if handle == 0 {
		return nil, ErrNoRawMode
	}
	if ret, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode))); ret == 0 {
		return nil, ErrNoRawMode
	}
	rawMode := mode &^ (enableEchoInput | enableLineInput)
	if ret, _, _ := procSetConsoleMode.Call(handle, uintptr(rawMode)); ret == 0 {
		return nil, ErrNoRawMode
	}
	restore := func() {
		procSetConsoleMode.Call(handle, uintptr(mode))
	}
	return restore, nil
}
