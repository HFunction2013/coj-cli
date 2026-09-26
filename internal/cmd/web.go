package cmd

import (
	"fmt"
	"strings"

	"github.com/HFunction2013/coj-cli/internal/api"
	"github.com/HFunction2013/coj-cli/internal/browser"
	"github.com/HFunction2013/coj-cli/internal/cli"
)

// openWeb handles --web.
//
// gh's model: --web is not "abandon the CLI", it is "do the lookup, then hand
// the result to the browser". pr view --web still resolves the PR (asking only
// for the url field) and prints "Opening … in your browser." We follow that:
// resolve what we can from flags, prompt only if an id is genuinely needed,
// then open.
func openWeb(ctx *cli.Context, ep *api.Endpoint, p *api.Params) error {
	io := ioFrom(ctx)
	page, needID := api.WebURL(ep.Path, p.Fields)

	if needID {
		if !io.CanPrompt() {
			return fmt.Errorf("this page requires an id; pass it with --field <ID field>=<value>")
		}
		field := idFieldFor(ep, ctx)
		v, err := askField(ctx, ep, field)
		if err != nil {
			return err
		}
		if v == "" {
			return fmt.Errorf("no id provided for %s", page)
		}
		p.Fields[field] = v
		page, needID = api.WebURL(ep.Path, p.Fields)
		if needID {
			return fmt.Errorf("still missing an id for %s", page)
		}
	}
	if page == "" {
		return fmt.Errorf("no web page mapped for %s; try: coj browse", ep.Path)
	}

	url := browser.PageURL(hostOf(ctx), page)
	if io.IsStdoutTTY() {
		fmt.Fprintf(io.ErrOut, "Opening %s in your browser.\n", url)
	}
	return io.Browser().Browse(url)
}

// idFieldFor picks which field supplies the page id.
func idFieldFor(ep *api.Endpoint, ctx *cli.Context) string {
	candidates := []string{}
	for _, ph := range ep.Params {
		if api.IsIDField(ph.Name) {
			candidates = append(candidates, ph.Name)
		}
	}
	if len(candidates) == 0 {
		return "F_ID"
	}
	if len(candidates) == 1 || !ioFrom(ctx).CanPrompt() {
		return candidates[0]
	}
	idx, err := ioFrom(ctx).Prompter().Select(
		"Which id does this page need?", "", readable(candidates))
	if err != nil || idx < 0 || idx >= len(candidates) {
		return candidates[0]
	}
	return candidates[idx]
}

func readable(fields []string) []string {
	out := make([]string, len(fields))
	for i, f := range fields {
		out[i] = humanField(f)
	}
	return out
}

// browseCmd is the gh browse equivalent.
func browseCmd() *cli.Command {
	return &cli.Command{
		Name:    "browse",
		Aliases: []string{"open", "web"},
		Summary: "Open the CandyOJ web UI in a browser",
		Description: `Open a page of the CandyOJ web UI.

With no argument it opens the home page. A bare resource name resolves to that
resource's management page, exactly like gh browse resolving a repo.`,
		Example: `  coj browse                 # home
  coj browse /MatchManage    # matches admin page
  coj browse match           # shorthand: resolve by resource name
  coj browse --no-browser    # print the URL instead
`,
		Flags: cli.NewFlagSet(
			&cli.Flag{Name: "no-browser", Usage: "Print the URL instead of opening it", Bool: true},
		),
		Run: func(ctx *cli.Context) error {
			io := ioFrom(ctx)
			target := ""
			if len(ctx.Args) > 0 {
				target = ctx.Args[0]
			}
			page := resolveBrowsePage(target)
			url := browser.PageURL(hostOf(ctx), page)
			if ctx.Flags.Bool("no-browser") {
				fmt.Fprintln(ctx.Stdout, url)
				return nil
			}
			if io.IsStdoutTTY() {
				fmt.Fprintf(io.ErrOut, "Opening %s in your browser.\n", url)
			}
			return io.Browser().Browse(url)
		},
	}
}

func resolveBrowsePage(target string) string {
	t := strings.Trim(target, "/")
	if t == "" {
		return "/Home"
	}
	if strings.HasPrefix(target, "/") {
		return "/" + t
	}
	for _, r := range api.Resources {
		if r.Name == t {
			return managePageOf(r.Name)
		}
		for _, a := range r.Aliases {
			if a == t {
				return managePageOf(r.Name)
			}
		}
	}
	return "/" + t
}

func managePageOf(res string) string {
	for _, ep := range api.Endpoints {
		if ep.Resource != res {
			continue
		}
		if t, ok := api.WebTargets[ep.Path]; ok && !strings.Contains(t.Path, ":id") {
			return t.Path
		}
	}
	return "/Home"
}

// ---- wiring ----
