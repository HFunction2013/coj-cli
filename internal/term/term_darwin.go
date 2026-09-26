//go:build darwin || freebsd || netbsd || openbsd || dragonfly

package term

import (
	"bufio"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// ioctl request codes: BSDs (including macOS) use TIOCGETA / TIOCSETA.
//
// This is exactly why calling syscall.TCGETS/TCSETS directly failed to build
// on macOS: those two constants exist only on Linux.
const (
	getTermios = 0x40487413 // syscall.TIOCGETA
	setTermios = 0x80487414 // syscall.TIOCSETA
)

func isTerminal(fd int) bool {
	var st syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(getTermios), uintptr(unsafe.Pointer(&st)))
	return errno == 0
}

func readPassword(prompt string) (string, error) {
	fd := int(os.Stdin.Fd())
	if !isTerminal(fd) {
		return readLine(prompt)
	}

	var old syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(getTermios), uintptr(unsafe.Pointer(&old))); errno != 0 {
		return "", errno
	}
	restore := func() {
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
			uintptr(setTermios), uintptr(unsafe.Pointer(&old)))
	}

	noecho := old
	noecho.Lflag &^= syscall.ECHO
	noecho.Lflag |= syscall.ICANON | syscall.ISIG
	noecho.Iflag |= syscall.ICRNL
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(setTermios), uintptr(unsafe.Pointer(&noecho))); errno != 0 {
		return "", errno
	}
	defer restore()

	return readLine(prompt)
}

func readLine(prompt string) (string, error) {
	if prompt != "" {
		if _, err := os.Stderr.WriteString(prompt); err != nil {
			return "", err
		}
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if isTerminal(int(os.Stdin.Fd())) {
		os.Stderr.WriteString("\n")
	}
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func setRaw(fd int) (func(), error) {
	var old syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(getTermios), uintptr(unsafe.Pointer(&old))); errno != 0 {
		return nil, ErrNoRawMode
	}
	restore := func() {
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
			uintptr(setTermios), uintptr(unsafe.Pointer(&old)))
	}

	raw := old
	raw.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP |
		syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	raw.Oflag &^= syscall.OPOST
	raw.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Cflag &^= syscall.CSIZE | syscall.PARENB
	raw.Cflag |= syscall.CS8
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(setTermios), uintptr(unsafe.Pointer(&raw))); errno != 0 {
		return nil, ErrNoRawMode
	}
	return restore, nil
}
