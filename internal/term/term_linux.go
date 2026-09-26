//go:build linux

package term

import (
	"bufio"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// ioctl request codes: Linux uses TCGETS / TCSETS.
const (
	getTermios = syscall.TCGETS
	setTermios = syscall.TCSETS
)

func isTerminal(fd int) bool {
	var st syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(getTermios), uintptr(unsafe.Pointer(&st)), 0, 0, 0)
	return errno == 0
}

func readPassword(prompt string) (string, error) {
	fd := int(os.Stdin.Fd())
	if !isTerminal(fd) {
		return readLine(prompt)
	}

	// read current attributes so we can restore them
	var old syscall.Termios
	if _, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(getTermios), uintptr(unsafe.Pointer(&old)), 0, 0, 0); errno != 0 {
		return "", errno
	}
	restore := func() {
		syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd),
			uintptr(setTermios), uintptr(unsafe.Pointer(&old)), 0, 0, 0)
	}

	// disable echo, keep everything else
	noecho := old
	noecho.Lflag &^= syscall.ECHO
	noecho.Lflag |= syscall.ICANON | syscall.ISIG
	noecho.Iflag |= syscall.ICRNL
	if _, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(setTermios), uintptr(unsafe.Pointer(&noecho)), 0, 0, 0); errno != 0 {
		return "", errno
	}
	// always restore echo, even if reading fails
	defer restore()

	line, err := readLine(prompt)
	if err != nil {
		return "", err
	}
	return line, nil
}

func readLine(prompt string) (string, error) {
	if prompt != "" {
		if _, err := os.Stderr.WriteString(prompt); err != nil {
			return "", err
		}
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	// in raw mode we must emit the newline ourselves, since input is not echoed
	if isTerminal(int(os.Stdin.Fd())) {
		os.Stderr.WriteString("\n")
	}
	if err != nil {
		if line == "" {
			return "", err
		}
		// io.EOF but we did read something: accept it
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func setRaw(fd int) (func(), error) {
	// save the current state
	var old syscall.Termios
	if _, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(getTermios), uintptr(unsafe.Pointer(&old)), 0, 0, 0); errno != 0 {
		return nil, ErrNoRawMode
	}
	restore := func() {
		syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd),
			uintptr(setTermios), uintptr(unsafe.Pointer(&old)), 0, 0, 0)
	}

	// equivalent of cfmakeraw
	raw := old
	raw.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP |
		syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	raw.Oflag &^= syscall.OPOST
	raw.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Cflag &^= syscall.CSIZE | syscall.PARENB
	raw.Cflag |= syscall.CS8
	if _, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(setTermios), uintptr(unsafe.Pointer(&raw)), 0, 0, 0); errno != 0 {
		return nil, ErrNoRawMode
	}
	return restore, nil
}
