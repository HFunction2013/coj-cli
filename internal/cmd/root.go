// Package cmd defines the CLI command tree.
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/HFunction2013/coj-cli/internal/api"
	"github.com/HFunction2013/coj-cli/internal/cli"
	"github.com/HFunction2013/coj-cli/internal/config"
	"github.com/HFunction2013/coj-cli/internal/iostreams"
	"github.com/HFunction2013/coj-cli/internal/output"
)

// Version is injected by main.
var Version = "dev"

// GlobalFlags are the flags accepted anywhere on the command line.
func GlobalFlags() []*cli.Flag {
	return []*cli.Flag{
		{Name: "host", Usage: "Site URL, defaults to https://candyoj.com"},
		{Name: "session", Usage: "Use this PHPSESSID directly (skip stored credentials)"},
		{Name: "json", Usage: "Print raw data as JSON", Bool: true},
		{Name: "jq", Usage: "Extract fields by path, e.g. '.list[].F_Name'"},
		{Name: "template", Usage: "Render output with a Go template, e.g. '{{.count}}'"},
		{Name: "field", Short: "F", Usage: "Request field as k=v; repeatable (query vs body decided per endpoint)", Repeated: true},
		{Name: "raw-field", Usage: "Field forced into the URL query string as k=v; repeatable", Repeated: true},
		{Name: "file", Usage: "File field to upload as k=/path/to/file; repeatable", Repeated: true},
		{Name: "page", Usage: "Page number, same as --field page=N"},
		{Name: "limit", Usage: "Page size; ignored by some endpoints (frontend truncates)"},
		{Name: "force", Usage: "Force past a quota limit (461) by sending force=1", Bool: true},
		{Name: "paginate", Usage: "Fetch every page automatically (backend paginates by page)", Bool: true},
		{Name: "web", Short: "w", Usage: "Open the matching page in a browser instead of calling the API", Bool: true},
		{Name: "no-prompt", Usage: "Disable interactive prompts (error on missing fields)", Bool: true},
		{Name: "verbose", Usage: "Print a request/response summary for debugging", Bool: true},
		{Name: "timeout", Usage: "Request timeout in seconds, default 30"},
		{Name: "no-header", Usage: "Omit the table header", Bool: true},
		{Name: "help", Short: "h", Usage: "Show help", Bool: true},
	}
}

// New builds the whole command tree.
func New() *cli.Router {
	root := &cli.Command{
		Name:    "coj",
		Summary: "Work with CandyOJ from the command line",
		Description: `coj is the command-line client for CandyOJ, an online judge
and programming-teaching platform.

Commands follow a REST resource model:

    coj <resource> <action> [flags]

A resource is a backend controller (user / match / homework / subject …);
an action is one endpoint.
For example /Match/getMatchList is reached as:

    coj match list
    coj match get-match-list     # the full action name works too

Any endpoint is also reachable through the escape hatch:

    coj api POST /Subject/getSubjectListNew --field page=1

Authentication: most endpoints need a login session. Start with

    coj auth login --username <name> --password <secret>

The session lives under ` + "`COJ_CONFIG_DIR`" + ` (default ~/.config/coj/session).

Note: most POST endpoints of this backend take parameters in the URL query
string rather than a request body. coj picks the right encoding per endpoint,
so you never have to think about it.`,
		Flags: cli.NewFlagSet(GlobalFlags()...),
	}

	root.Add(
		authCmd(),
		apiCmd(),
		browseCmd(),
		statusCmd(),
		versionCmd(),
	)
	// 190 endpoints, generated per resource
	root.Add(resourceCommands()...)

	return &cli.Router{Root: root, Version: Version, Global: GlobalFlags()}
}

// ---- shared helpers ----

// clientFrom builds a client from the global flags.
func clientFrom(ctx *cli.Context) (*api.Client, error) {
	host := ctx.Flags.Get("host")
	if host == "" {
		if cfg, err := config.Load(); err == nil && cfg.Host != "" {
			host = cfg.Host
		} else {
			host = "https://candyoj.com"
		}
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	c, err := api.NewClient(host)
	if err != nil {
		return nil, err
	}
	c.Verbose = ctx.Flags.Bool("verbose")
	c.Log = ctx.Stderr
	if secs := ctx.Flags.Get("timeout"); secs != "" {
		var n int
		if _, err := fmt.Sscanf(secs, "%d", &n); err == nil && n > 0 {
			c.HTTP.Timeout = time.Duration(n) * time.Second
		}
	}
	if sid := ctx.Flags.Get("session"); sid != "" {
		c.SetSession(sid)
	} else if sid := config.LoadSession(); sid != "" {
		c.SetSession(sid)
	}
	return c, nil
}

// paramsFrom builds request parameters from the flags.
func paramsFrom(ctx *cli.Context) (*api.Params, error) {
	p := api.NewParams()
	for _, kv := range ctx.Flags.GetAll("field") {
		k, v, err := splitKV(kv)
		if err != nil {
			return nil, err
		}
		p.Fields[k] = v
	}
	for _, kv := range ctx.Flags.GetAll("raw-field") {
		k, v, err := splitKV(kv)
		if err != nil {
			return nil, err
		}
		p.Raw[k] = append(p.Raw[k], v)
	}
	for _, kv := range ctx.Flags.GetAll("file") {
		k, v, err := splitKV(kv)
		if err != nil {
			return nil, err
		}
		p.Files[k] = v
	}
	if page := ctx.Flags.Get("page"); page != "" {
		p.Fields["page"] = page
	}
	if limit := ctx.Flags.Get("limit"); limit != "" {
		p.Fields["limit"] = limit
	}
	p.Force = ctx.Flags.Bool("force")
	return p, nil
}

func splitKV(s string) (string, string, error) {
	eq := strings.Index(s, "=")
	if eq <= 0 {
		return "", "", fmt.Errorf("field must look like k=v, got %q", s)
	}
	return s[:eq], s[eq+1:], nil
}

// outputOpts decides how to render output.
func outputOpts(ctx *cli.Context) output.Options {
	if expr := ctx.Flags.Get("jq"); expr != "" {
		return output.Options{Mode: output.ModeJQ, Expr: expr}
	}
	if expr := ctx.Flags.Get("template"); expr != "" {
		return output.Options{Mode: output.ModeTemplate, Expr: expr}
	}
	if ctx.Flags.Bool("json") {
		return output.Options{Mode: output.ModeJSON}
	}
	return output.Options{Mode: output.ModeTable, NoHeader: ctx.Flags.Bool("no-header")}
}

// emit renders data (falls back to the whole response when data is empty).
func emit(ctx *cli.Context, data interface{}) error {
	return output.Render(ctx.Stdout, data, outputOpts(ctx))
}

func versionCmd() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Summary: "Show the coj version",
		Run: func(ctx *cli.Context) error {
			fmt.Fprintf(ctx.Stdout, "coj version %s\n", ctx.Version)
			return nil
		},
	}
}

func statusCmd() *cli.Command {
	return &cli.Command{
		Name:    "status",
		Summary: "Show current host, login state and config paths",
		Example: "  coj status\n",
		Run: func(ctx *cli.Context) error {
			cfg, _ := config.Load()
			c, err := clientFrom(ctx)
			if err != nil {
				return err
			}
			rows := []map[string]interface{}{}
			add := func(k string, v interface{}) {
				rows = append(rows, map[string]interface{}{"key": k, "value": v})
			}
			add("host", c.BaseURL)
			add("config", config.Path())
			add("session_file", config.SessionPath())
			add("logged_in", c.SessionID() != "")
			if c.SessionID() != "" {
				add("session", c.SessionID())
				if cfg != nil && cfg.Username != "" {
					add("username", cfg.Username)
					add("role", config.RoleName(cfg.UserType))
				}
			}
			return emit(ctx, rows)
		},
	}
}

// requireSession fails with an actionable message when there is no session.
func requireSession(ctx *cli.Context, c *api.Client) error {
	if c.SessionID() != "" {
		return nil
	}
	return fmt.Errorf("not logged in; run: coj auth login --username <name> --password <secret>")
}

func stderrf(ctx *cli.Context, format string, args ...interface{}) {
	fmt.Fprintf(ctx.Stderr, format+"\n", args...)
}

// ioFrom builds the IOStreams for this invocation, cached on the context.
// --no-prompt maps to gh's neverPrompt switch.
func ioFrom(ctx *cli.Context) *iostreams.IOStreams {
	if ctx.IO != nil {
		return ctx.IO
	}
	io := iostreams.System()
	if ctx.Flags.Bool("no-prompt") {
		io.SetNeverPrompt(true)
	}
	ctx.IO = io
	return io
}

// hostOf returns the site URL in use, for building web URLs.
func hostOf(ctx *cli.Context) string {
	host := ctx.Flags.Get("host")
	if host == "" {
		if cfg, err := config.Load(); err == nil && cfg.Host != "" {
			host = cfg.Host
		}
	}
	if host == "" {
		host = "https://candyoj.com"
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	return strings.TrimRight(host, "/")
}

// rememberedUserID / rememberedClassID：coj 没有 git 上下文，
// 可用的是登录后记下来的身份。对应 gh "无参从上下文推断"。
func rememberedUserID() string {
	if cfg, err := config.Load(); err == nil && cfg.UserID != "" {
		return cfg.UserID
	}
	return ""
}

func rememberedClassID() string {
	if cfg, err := config.Load(); err == nil && cfg.ClassID != "" {
		return cfg.ClassID
	}
	return ""
}
