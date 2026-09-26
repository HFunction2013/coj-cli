// Package output provides gh-style rendering:
//   - human-readable tables by default
//   - --json prints raw JSON
//   - --jq '<expr>' extracts fields by path (.a.b / .list[].name / .list[0].x)
//   - --template '{{...}}' renders a Go template
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"
)

// Mode is the output mode.
type Mode int

const (
	ModeTable Mode = iota
	ModeJSON
	ModeJQ
	ModeTemplate
)

// Options are the output options.
type Options struct {
	Mode     Mode
	Expr     string // jq expression or Go template
	NoHeader bool
}

// Render renders data according to the mode.
func Render(w io.Writer, data interface{}, opts Options) error {
	switch opts.Mode {
	case ModeJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case ModeJQ:
		v, err := Extract(data, opts.Expr)
		if err != nil {
			return err
		}
		return writeValue(w, v)
	case ModeTemplate:
		tmpl, err := template.New("out").Funcs(funcs()).Parse(opts.Expr)
		if err != nil {
			return fmt.Errorf("cannot parse template: %w", err)
		}
		return tmpl.Execute(w, data)
	default:
		return writeTable(w, data, opts)
	}
}

func writeValue(w io.Writer, v interface{}) error {
	switch t := v.(type) {
	case string:
		fmt.Fprintln(w, t)
		return nil
	case nil:
		return nil
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ---- tables ----

func writeTable(w io.Writer, data interface{}, opts Options) error {
	rows, cols, ok := flatten(data)
	if !ok {
		// not a list or object: pretty-print it
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if !opts.NoHeader && len(cols) > 0 {
		for i, c := range cols {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, strings.ToUpper(c))
		}
		fmt.Fprintln(tw)
	}
	for _, r := range rows {
		for i, c := range cols {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, cell(r[c]))
		}
		fmt.Fprintln(tw)
	}
	return tw.Flush()
}

func cell(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

// flatten normalises common response shapes into rows.
// Handles: []any, a map carrying a list field, and a top-level object.
func flatten(data interface{}) ([]map[string]interface{}, []string, bool) {
	// any slice (e.g. []map[string]interface{}) is normalised to []interface{}
	v := reflect.ValueOf(data)
	if v.IsValid() && (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) {
		if _, ok := data.([]interface{}); !ok {
			items := make([]interface{}, 0, v.Len())
			for i := 0; i < v.Len(); i++ {
				items = append(items, v.Index(i).Interface())
			}
			return flatten(items)
		}
	}
	switch t := data.(type) {
	case []interface{}:
		rows := []map[string]interface{}{}
		seen := map[string]bool{}
		cols := []string{}
		for _, item := range t {
			m, ok := toStringMap(item)
			if !ok {
				rows = append(rows, map[string]interface{}{"value": item})
				if !seen["value"] {
					seen["value"] = true
					cols = append(cols, "value")
				}
				continue
			}
			rows = append(rows, m)
			for k := range m {
				if !seen[k] {
					seen[k] = true
					cols = append(cols, k)
				}
			}
		}
		sort.Strings(cols)
		return rows, cols, true

	case map[string]interface{}:
		// the common shape: { list: [...], count: N }
		for _, key := range []string{"list", "data", "rows", "items"} {
			if inner, ok := t[key]; ok {
				if arr, ok := inner.([]interface{}); ok {
					return flatten(arr)
				}
			}
		}
		// plain object: one row, keys sorted
		cols := []string{}
		for k := range t {
			cols = append(cols, k)
		}
		sort.Strings(cols)
		return []map[string]interface{}{t}, cols, true
	}
	return nil, nil, false
}

func toStringMap(v interface{}) (map[string]interface{}, bool) {
	switch t := v.(type) {
	case map[string]interface{}:
		return t, true
	case string:
		// the backend sometimes returns a JSON string
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(t), &m); err == nil {
			return m, true
		}
	}
	return nil, false
}

// ---- jq subset ----

// Extract evaluates a path expression such as .data.list[].F_Name
func Extract(data interface{}, expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" || expr == "." {
		return data, nil
	}
	if !strings.HasPrefix(expr, ".") {
		return nil, fmt.Errorf("jq expression must start with a dot: %q", expr)
	}
	return walk(data, expr[1:])
}

func walk(v interface{}, rest string) (interface{}, error) {
	if rest == "" {
		return v, nil
	}
	// parse one segment
	i := 0
	for i < len(rest) && rest[i] != '.' && rest[i] != '[' {
		i++
	}
	seg := rest[:i]
	remain := rest[i:]

	if seg != "" {
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected an object before field %q", seg)
		}
		val, ok := m[seg]
		if !ok {
			return nil, fmt.Errorf("no such field: %s", seg)
		}
		v = val
	}

	for strings.HasPrefix(remain, "[") {
		end := strings.Index(remain, "]")
		if end < 0 {
			return nil, fmt.Errorf("missing closing bracket: %s", remain)
		}
		inner := remain[1:end]
		remain = remain[end+1:]
		arr, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected an array before an index, got %T", v)
		}
		if inner == "" {
			// [] expands: evaluate the rest of the path per item
			if remain == "" {
				return arr, nil
			}
			out := make([]interface{}, 0, len(arr))
			for _, item := range arr {
				got, err := walk(item, strings.TrimPrefix(remain, "."))
				if err != nil {
					continue
				}
				out = append(out, got)
			}
			return out, nil
		}
		idx, err := strconv.Atoi(inner)
		if err != nil {
			return nil, fmt.Errorf("index must be an integer or empty: [%s]", inner)
		}
		if idx < 0 || idx >= len(arr) {
			return nil, fmt.Errorf("index out of range: [%d]", idx)
		}
		v = arr[idx]
	}
	return walk(v, strings.TrimPrefix(remain, "."))
}

func funcs() template.FuncMap {
	return template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"color": func(s string) string { return s },
	}
}

// Stdout is the shared output target, handy for tests.
var Stdout io.Writer = os.Stdout
