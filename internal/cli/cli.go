// Package cli is a dependency-free command framework shaped like gh:
//
//	coj <subject> <verb> [flags]
//
// For example:
//
//	coj auth login --username alice
//	coj match list --page 2 --json
//	coj api POST /Subject/getSubjectListNew --field page=1
package cli

import (
	"fmt"
	"io"

	"github.com/HFunction2013/coj-cli/internal/iostreams"
	"sort"
	"strings"
	"text/tabwriter"
)

// Command is a node in the command tree.
// A node with Subcommands only routes; one without Run prints help.
type Command struct {
	Name        string
	Aliases     []string
	Summary     string // one-line description
	Description string // long description
	Example     string
	Group       string // section heading in help

	Flags *FlagSet
	Run   func(*Context) error

	Subcommands []*Command
	parent      *Command
}

// FullName returns the command name from the root down.
func (c *Command) FullName() string {
	parts := []string{}
	for p := c; p != nil; p = p.parent {
		parts = append([]string{p.Name}, parts...)
	}
	return strings.Join(parts, " ")
}

func (c *Command) add(sub *Command) {
	sub.parent = c
	c.Subcommands = append(c.Subcommands, sub)
}

// Add registers subcommands.
func (c *Command) Add(subs ...*Command) {
	for _, s := range subs {
		c.add(s)
	}
}

// Find looks up a direct subcommand by name or alias.
func (c *Command) Find(name string) *Command {
	for _, s := range c.Subcommands {
		if s.Name == name {
			return s
		}
		for _, a := range s.Aliases {
			if a == name {
				return s
			}
		}
	}
	return nil
}

// Context carries all state needed for one execution.
type Context struct {
	Cmd     *Command
	Flags   *FlagSet
	Args    []string
	Stdout  io.Writer
	Stderr  io.Writer
	Version string
	// IO carries terminal capabilities; commands consult it instead of
	// probing the terminal directly (gh's pkg/iostreams convention).
	IO *iostreams.IOStreams
}

// Flag describes a command-line flag.
type Flag struct {
	Name     string
	Short    string
	Usage    string
	Default  string
	Bool     bool
	Repeated bool // may appear more than once, e.g. --field
	Hidden   bool
}

// FlagSet is a parsed set of flags.
type FlagSet struct {
	defs   []*Flag
	values map[string][]string
	seen   map[string]bool
}

// NewFlagSet creates a flag set.
func NewFlagSet(defs ...*Flag) *FlagSet {
	return &FlagSet{defs: defs, values: map[string][]string{}, seen: map[string]bool{}}
}

// Add appends flag definitions.
func (fs *FlagSet) Add(defs ...*Flag) {
	fs.defs = append(fs.defs, defs...)
}

// def finds a definition by name or short name.
func (fs *FlagSet) def(name string) *Flag {
	for _, d := range fs.defs {
		if d.Name == name || (d.Short != "" && d.Short == name) {
			return d
		}
	}
	return nil
}

// Get returns a flag's string value (or its default when unset).
func (fs *FlagSet) Get(name string) string {
	if v, ok := fs.values[name]; ok && len(v) > 0 {
		return v[len(v)-1]
	}
	if d := fs.def(name); d != nil {
		return d.Default
	}
	return ""
}

// GetAll returns every value of a repeatable flag.
func (fs *FlagSet) GetAll(name string) []string {
	return fs.values[name]
}

// Bool returns the value of a boolean flag.
func (fs *FlagSet) Bool(name string) bool {
	if v, ok := fs.values[name]; ok && len(v) > 0 {
		return v[len(v)-1] != "false"
	}
	if d := fs.def(name); d != nil {
		return d.Default == "true"
	}
	return false
}

// Set is called by the parser.
func (fs *FlagSet) Set(name, value string) {
	fs.values[name] = append(fs.values[name], value)
	fs.seen[name] = true
}

// parse parses argv.
// With stopAtPositional=true it stops at the first non-flag argument
// the leaf command is unknown here, so its own flags cannot be parsed yet）；
// (routing phase, where the leaf command is still unknown); with false it
// collects positional args and keeps parsing flags after them (leaf phase,
// so `coj match del-match 42 --host x` works).
func (fs *FlagSet) parse(argv []string, stopAtPositional bool) (rest []string, err error) {
	i := 0
	for i < len(argv) {
		arg := argv[i]
		if !strings.HasPrefix(arg, "-") || arg == "-" || arg == "--" {
			if arg == "--" {
				return append(rest, argv[i+1:]...), nil
			}
			rest = append(rest, arg)
			i++
			if stopAtPositional {
				return append(rest, argv[i:]...), nil
			}
			continue
		}

		name := strings.TrimPrefix(strings.TrimPrefix(arg, "-"), "-")
		inline := ""
		if eq := strings.Index(name, "="); eq >= 0 {
			inline, name = name[eq+1:], name[:eq]
		}

		d := fs.def(name)
		if d == nil {
			// Unknown flag: still accept the name=value form, as in --page=1
			return nil, fmt.Errorf("unknown flag: --%s", name)
		}
		if d.Bool {
			if inline == "" {
				inline = "true"
			}
			fs.Set(d.Name, inline)
			i++
			continue
		}
		if inline != "" {
			fs.Set(d.Name, inline)
			i++
			continue
		}
		if i+1 >= len(argv) {
			return nil, fmt.Errorf("flag needs an argument: --%s", d.Name)
		}
		fs.Set(d.Name, argv[i+1])
		i += 2
	}
	return rest, nil
}

// Visible returns the non-hidden flag definitions.
func (fs *FlagSet) Visible() []*Flag {
	out := []*Flag{}
	for _, d := range fs.defs {
		if !d.Hidden {
			out = append(out, d)
		}
	}
	return out
}

// Router holds the root command.
type Router struct {
	Root    *Command
	Version string
	Global  []*Flag // flags accepted anywhere on the command line
}

// Execute parses and runs.
func (r *Router) Execute(argv []string, stdout, stderr io.Writer) error {
	cmd, fs, args, err := r.resolve(argv)
	if err != nil {
		return err
	}
	if cmd == nil {
		r.usage(r.Root, stdout)
		return nil
	}
	ctx := &Context{Cmd: cmd, Flags: fs, Args: args, Stdout: stdout, Stderr: stderr, Version: r.Version}
	if cmd.Run == nil || fs.Bool("help") {
		helpCmd(cmd, ctx)
		return nil
	}
	if fs.Bool("version") {
		fmt.Fprintf(stdout, "%s version %s\n", r.Root.Name, r.Version)
		return nil
	}
	return cmd.Run(ctx)
}

// resolve walks down the command tree until nothing else matches.
func (r *Router) resolve(argv []string) (*Command, *FlagSet, []string, error) {
	fs := NewFlagSet(r.Global...)
	rest, err := fs.parse(argv, true)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(rest) == 0 {
		return nil, nil, nil, nil
	}
	head := rest[0]
	if head == "help" || head == "--help" || head == "-h" {
		return r.Root, fs, rest[1:], nil
	}

	cmd := r.Root.Find(head)
	if cmd == nil {
		return nil, nil, nil, fmt.Errorf("unknown command %q for %q\n\nRun 'coj --help' for usage.", head, r.Root.Name)
	}
	cur := cmd
	remaining := rest[1:]
	// merge the parent's and the current command's flag definitions
	all := append([]*Flag{}, fs.defs...)
	if r.Root.Flags != nil {
		all = append(all, r.Root.Flags.defs...)
	}
	for len(remaining) > 0 {
		sub := cur.Find(remaining[0])
		if sub == nil {
			break
		}
		cur = sub
		remaining = remaining[1:]
	}
	if cur.Flags != nil {
		all = append(all, cur.Flags.defs...)
	}
	cfs := NewFlagSet(all...)
	// keep already-parsed values (global flags may precede the subcommand)
	for k, v := range fs.values {
		cfs.values[k] = v
	}
	rest2, err := cfs.parse(remaining, false)
	if err != nil {
		return nil, nil, nil, err
	}
	return cur, cfs, rest2, nil
}

func (r *Router) usage(cmd *Command, w io.Writer) {
	if cmd.Description != "" {
		fmt.Fprintln(w, cmd.Description)
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "USAGE\n  %s <command> [flags]\n\n", cmd.FullName())
	groups := map[string][]*Command{}
	order := []string{}
	for _, s := range cmd.Subcommands {
		g := s.Group
		if g == "" {
			g = "COMMANDS"
		}
		if _, ok := groups[g]; !ok {
			order = append(order, g)
		}
		groups[g] = append(groups[g], s)
	}
	for _, g := range order {
		fmt.Fprintf(w, "%s\n", g)
		tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
		for _, s := range groups[g] {
			fmt.Fprintf(tw, "  %s\t%s\n", s.Name, s.Summary)
		}
		tw.Flush()
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "FLAGS\n")
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	for _, f := range r.Global {
		fmt.Fprintf(tw, "  --%s\t%s\n", f.Name, f.Usage)
	}
	tw.Flush()
	fmt.Fprintln(w)
	fmt.Fprintf(w, "EXAMPLES\n  %s auth login --username alice\n  %s match list --json\n", cmd.Name, cmd.Name)
}

func helpCmd(cmd *Command, ctx *Context) {
	fs := cmd.Flags
	if fs == nil {
		fs = NewFlagSet()
	}
	if cmd.Description != "" {
		fmt.Fprintf(ctx.Stdout, "%s\n\n", cmd.Description)
	} else if cmd.Summary != "" {
		fmt.Fprintf(ctx.Stdout, "%s\n\n", cmd.Summary)
	}
	fmt.Fprintf(ctx.Stdout, "USAGE\n  %s [flags]\n\n", cmd.FullName())
	if flags := fs.Visible(); len(flags) > 0 {
		fmt.Fprintf(ctx.Stdout, "FLAGS\n")
		tw := tabwriter.NewWriter(ctx.Stdout, 0, 4, 2, ' ', 0)
		for _, f := range flags {
			label := "--" + f.Name
			if f.Short != "" {
				label = "-" + f.Short + ", " + label
			}
			if !f.Bool {
				label += " <value>"
			}
			fmt.Fprintf(tw, "  %s\t%s\n", label, f.Usage)
		}
		tw.Flush()
		fmt.Fprintln(ctx.Stdout)
	}
	if len(cmd.Subcommands) > 0 {
		fmt.Fprintf(ctx.Stdout, "SUBCOMMANDS\n")
		tw := tabwriter.NewWriter(ctx.Stdout, 0, 4, 2, ' ', 0)
		sorted := append([]*Command{}, cmd.Subcommands...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
		for _, s := range sorted {
			fmt.Fprintf(tw, "  %s\t%s\n", s.Name, s.Summary)
		}
		tw.Flush()
		fmt.Fprintln(ctx.Stdout)
	}
	if cmd.Example != "" {
		fmt.Fprintf(ctx.Stdout, "EXAMPLES\n%s\n", cmd.Example)
	}
}

func globalFlags() []*Flag {
	return []*Flag{
		{Name: "help", Short: "h", Usage: "Show help", Bool: true},
		{Name: "version", Usage: "Show version", Bool: true},
	}
}
