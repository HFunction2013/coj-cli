package prompter

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/HFunction2013/coj-cli/internal/term"
)

// maxRows caps the rendered list so a 500-item result set does not flood the
// screen. gh gets this from survey's page size; we implement the same window.
const maxRows = 12

type selector struct {
	Message string
	Options []string
	Cursor  int
}

type selectState struct {
	in  FileReader
	out FileWriter

	sel     selector
	checked map[int]bool

	filter  string
	rows    []int // indices into sel.Options after filtering
	cursor  int   // position within rows
	lines   int   // rows drawn last time, moved over before redraw
	restore func()
	buf     []byte
}

func newSelector(in FileReader, out FileWriter, s selector) *selectState {
	return &selectState{
		in:      in,
		out:     out,
		sel:     s,
		checked: nil,
		buf:     make([]byte, 1),
	}
}

func (st *selectState) indexInOptions(i int) int {
	if i < 0 || i >= len(st.rows) {
		return -1
	}
	return st.rows[i]
}

func (st *selectState) recompute() {
	st.rows = nil
	f := strings.ToLower(st.filter)
	for i, o := range st.sel.Options {
		if f == "" || strings.Contains(strings.ToLower(o), f) {
			st.rows = append(st.rows, i)
		}
	}
	if st.cursor >= len(st.rows) {
		st.cursor = len(st.rows) - 1
	}
	if st.cursor < 0 {
		st.cursor = 0
	}
}

// window returns the slice of rows to draw, keeping the cursor centred.
func (st *selectState) window() []int {
	total := len(st.rows)
	if total == 0 {
		return nil
	}
	start := 0
	if total > maxRows {
		start = st.cursor - maxRows/2
		if start < 0 {
			start = 0
		}
		if start+maxRows > total {
			start = total - maxRows
		}
	}
	end := start + maxRows
	if end > total {
		end = total
	}
	return st.rows[start:end]
}

func (st *selectState) writeLine(s string) {
	// raw mode disables OPOST, so "\n" alone would not return the column
	fmt.Fprint(st.out, clearLine+cursorStart+s+"\r\n")
	st.lines++
}

func (st *selectState) render(multi bool) {
	if st.lines > 0 {
		fmt.Fprint(st.out, moveUp(st.lines))
	}
	fmt.Fprint(st.out, hideCursor)
	st.lines = 0

	title := fmt.Sprintf("%s%s?%s %s", cyan, bold, reset, st.sel.Message)
	if st.filter != "" {
		title += fmt.Sprintf(" %s%s%s", green, st.filter, reset)
	} else {
		title += fmt.Sprintf(" %s(type to filter)%s", dim, reset)
	}
	st.writeLine(title)

	rows := st.window()
	if len(rows) == 0 {
		st.writeLine(fmt.Sprintf("%sNo matches%s", dim, reset))
	} else {
		for _, idx := range rows {
			prefix := "  "
			if idx == st.indexInOptions(st.cursor) {
				prefix = fmt.Sprintf("%s>%s ", cyan, reset)
			}
			mark := ""
			if multi {
				if st.checked[idx] {
					mark = fmt.Sprintf("%s[x]%s ", green, reset)
				} else {
					mark = "[ ] "
				}
			}
			st.writeLine(prefix + mark + st.sel.Options[idx])
		}
	}
	tip := "Use arrows to move, type to filter, Enter to select"
	if multi {
		tip = "Use arrows to move, Space to toggle, Enter to confirm"
	}
	st.writeLine(fmt.Sprintf("%s%s%s", dim, tip, reset))
}

func (st *selectState) readKey() (string, error) {
	n, err := st.in.Read(st.buf)
	if err != nil || n == 0 {
		return "", err
	}
	c := st.buf[0]
	switch c {
	case 0x03:
		return "ctrl-c", nil
	case 0x0d, 0x0a:
		return "enter", nil
	case 0x20:
		return "space", nil
	case 0x7f, 0x08:
		return "backspace", nil
	case 0x1b:
		seq := make([]byte, 2)
		if n, err := st.in.Read(seq); err == nil && n == 2 && seq[0] == '[' {
			switch seq[1] {
			case 'A':
				return "up", nil
			case 'B':
				return "down", nil
			case 'C':
				return "right", nil
			case 'D':
				return "left", nil
			}
		}
		return "esc", nil
	}
	if c < 0x20 {
		return "", nil
	}
	if c >= 0x80 {
		size := utf8.RuneLen(rune(c))
		if size > 1 {
			rest := make([]byte, size-1)
			st.in.Read(rest)
			b := append([]byte{c}, rest...)
			if r, _ := utf8.DecodeRune(b); r != utf8.RuneError {
				return "char:" + string(r), nil
			}
		}
	}
	return "char:" + string(rune(c)), nil
}

func (st *selectState) enter() error {
	st.restore, _ = term.SetRaw(int(st.in.Fd()))
	if st.restore == nil {
		return ErrNotInteractive
	}
	return nil
}

func (st *selectState) close() {
	fmt.Fprint(st.out, showCursor)
	if st.restore != nil {
		st.restore()
	}
}

func (st *selectState) runSingle() (int, error) {
	if err := st.enter(); err != nil {
		idx, err := fallbackSelect(st.sel, st.out, st.in, false)
		return idx, err
	}
	defer st.close()
	st.recompute()
	st.render(false)
	for {
		key, err := st.readKey()
		if err != nil {
			return 0, err
		}
		switch {
		case key == "ctrl-c":
			fmt.Fprint(st.out, "\r\n")
			return 0, ErrAborted
		case key == "enter":
			if len(st.rows) == 0 {
				st.render(false)
				continue
			}
			idx := st.indexInOptions(st.cursor)
			st.finish(st.sel.Options[idx])
			return idx, nil
		case key == "up":
			if st.cursor > 0 {
				st.cursor--
			}
		case key == "down":
			if st.cursor < len(st.rows)-1 {
				st.cursor++
			}
		case key == "backspace":
			if len(st.filter) > 0 {
				r := []rune(st.filter)
				st.filter = string(r[:len(r)-1])
				st.recompute()
			}
		case strings.HasPrefix(key, "char:"):
			st.filter += strings.TrimPrefix(key, "char:")
			st.recompute()
		}
		st.render(false)
	}
}

func (st *selectState) runMulti() ([]int, error) {
	if err := st.enter(); err != nil {
		return fallbackMultiSelect(st.sel, st.out, st.in)
	}
	defer st.close()
	st.recompute()
	st.render(true)
	for {
		key, err := st.readKey()
		if err != nil {
			return nil, err
		}
		switch {
		case key == "ctrl-c":
			fmt.Fprint(st.out, "\r\n")
			return nil, ErrAborted
		case key == "enter":
			out := []int{}
			for idx := range st.checked {
				out = append(out, idx)
			}
			sortInts(out)
			st.finish(labelsFor(st.sel.Options, out))
			return out, nil
		case key == "up":
			if st.cursor > 0 {
				st.cursor--
			}
		case key == "down":
			if st.cursor < len(st.rows)-1 {
				st.cursor++
			}
		case key == "space":
			if idx := st.indexInOptions(st.cursor); idx >= 0 {
				st.checked[idx] = !st.checked[idx]
			}
		case key == "backspace":
			if len(st.filter) > 0 {
				r := []rune(st.filter)
				st.filter = string(r[:len(r)-1])
				st.recompute()
			}
		case strings.HasPrefix(key, "char:"):
			st.filter += strings.TrimPrefix(key, "char:")
			st.recompute()
		}
		st.render(true)
	}
}

// finish redraws the prompt with just the answer, as gh does.
func (st *selectState) finish(answer string) {
	if st.lines > 0 {
		fmt.Fprint(st.out, moveUp(st.lines))
	}
	fmt.Fprintf(st.out, "%s%s?%s %s %s%s%s\r\n",
		cyan, bold, reset, st.sel.Message, green, answer, reset)
	fmt.Fprint(st.out, showCursor)
	st.lines = 0
}

func labelsFor(options []string, idxs []int) string {
	parts := []string{}
	for _, i := range idxs {
		if i >= 0 && i < len(options) {
			parts = append(parts, options[i])
		}
	}
	return strings.Join(parts, ", ")
}

func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}

// fallbackSelect is the numbered-input path used when raw mode is
// unavailable (dumb terminal, restricted environment).
func fallbackSelect(sel selector, out FileWriter, in FileReader, multi bool) (int, error) {
	if !multi {
		fmt.Fprintf(out, "%s%s?%s %s\n", cyan, bold, reset, sel.Message)
		for i, o := range sel.Options {
			fmt.Fprintf(out, "  %d) %s\n", i+1, o)
		}
		answer, err := readLine(in, out, "Enter a number", "", false)
		if err != nil {
			return 0, err
		}
		var n int
		if _, err := fmt.Sscanf(strings.TrimSpace(answer), "%d", &n); err != nil || n < 1 || n > len(sel.Options) {
			return 0, fmt.Errorf("invalid selection: %q", answer)
		}
		return n - 1, nil
	}
	idxs, err := fallbackMultiSelect(sel, out, in)
	if err != nil {
		return 0, err
	}
	if len(idxs) == 0 {
		return 0, nil
	}
	return idxs[0], nil
}

func fallbackMultiSelect(sel selector, out FileWriter, in FileReader) ([]int, error) {
	fmt.Fprintf(out, "%s%s?%s %s (comma-separated numbers)\n", cyan, bold, reset, sel.Message)
	for i, o := range sel.Options {
		fmt.Fprintf(out, "  %d) %s\n", i+1, o)
	}
	answer, err := readLine(in, out, "Select", "", false)
	if err != nil {
		return nil, err
	}
	out2 := []int{}
	for _, part := range strings.Split(answer, ",") {
		var n int
		if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &n); err == nil && n >= 1 && n <= len(sel.Options) {
			out2 = append(out2, n-1)
		}
	}
	sortInts(out2)
	return out2, nil
}
