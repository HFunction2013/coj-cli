package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/HFunction2013/coj-cli/internal/api"
	"github.com/HFunction2013/coj-cli/internal/cli"
)

// resourceCommands expands the 190 endpoints into a command tree:
//
//	coj <resource> <action> [flags]
//
// every action gets a kebab name (get-match-list) and a gh-style alias (list).
func resourceCommands() []*cli.Command {
	byRes := map[string][]*api.Endpoint{}
	order := []string{}
	for i := range api.Endpoints {
		ep := &api.Endpoints[i]
		if _, ok := byRes[ep.Resource]; !ok {
			order = append(order, ep.Resource)
		}
		byRes[ep.Resource] = append(byRes[ep.Resource], ep)
	}
	sort.Strings(order)

	summaryOf := map[string]string{}
	aliasOf := map[string][]string{}
	for _, r := range api.Resources {
		summaryOf[r.Name] = r.Summary
		aliasOf[r.Name] = r.Aliases
	}

	out := []*cli.Command{}
	for _, res := range order {
		eps := byRes[res]
		parent := &cli.Command{
			Name:        res,
			Aliases:     aliasOf[res],
			Summary:     summaryOf[res],
			Description: fmt.Sprintf("%s\n\n%d endpoints, all served by the backend /%s controller.", summaryOf[res], len(eps), controllerOf(eps[0])),
			Group:       "RESOURCE COMMANDS",
		}
		acts := []*cli.Command{}
		for _, ep := range eps {
			acts = append(acts, endpointCommand(res, ep))
		}
		sort.Slice(acts, func(i, j int) bool { return acts[i].Name < acts[j].Name })
		parent.Add(acts...)
		out = append(out, parent)
	}
	return out
}

func controllerOf(ep *api.Endpoint) string {
	parts := strings.Split(strings.Trim(ep.Path, "/"), "/")
	return parts[0]
}

// endpointCommand builds the command for one endpoint.
func endpointCommand(res string, ep *api.Endpoint) *cli.Command {
	names := []string{ep.Action}
	names = append(names, ep.Aliases...)

	desc := &strings.Builder{}
	fmt.Fprintf(desc, "%s\n\nEndpoint: %s %s\nParameters: %s\n", ep.Summary, ep.Method, ep.Path, styleDesc(ep.Style))
	if len(ep.Params) > 0 {
		fmt.Fprintf(desc, "\nKnown fields\n")
		for _, p := range ep.Params {
			if p.Hint != "" && p.Hint != `""` {
				fmt.Fprintf(desc, "  %-22s frontend source: %s\n", p.Name, p.Hint)
			} else {
				fmt.Fprintf(desc, "  %s\n", p.Name)
			}
		}
	}

	ex := &strings.Builder{}
	fmt.Fprintf(ex, "  coj %s %s", res, names[0])
	for _, p := range ep.Params {
		fmt.Fprintf(ex, " --field %s=…", p.Name)
	}
	fmt.Fprintln(ex)
	if ep.Style == api.StyleUpload {
		fmt.Fprintf(ex, "  coj %s %s --file file=./data.xlsx\n", res, names[0])
	}

	return &cli.Command{
		Name:        names[0],
		Aliases:     names[1:],
		Summary:     ep.Summary,
		Description: desc.String(),
		Example:     ex.String(),
		Run: func(ctx *cli.Context) error {
			c, err := clientFrom(ctx)
			if err != nil {
				return err
			}

			// gh: --web is an alternative to interactivity, not an extra mode.
			// It skips the API call entirely and opens the matching page.
			if ctx.Flags.Bool("web") {
				p, err := paramsFrom(ctx)
				if err != nil {
					return err
				}
				if len(ctx.Args) > 0 {
					if len(ep.Params) > 0 {
						p.Fields[ep.Params[0].Name] = strings.Join(ctx.Args, ",")
					} else if len(ctx.Args) == 1 {
						p.Fields["F_ID"] = ctx.Args[0]
					}
				}
				return openWeb(ctx, ep, p)
			}

			// gh: most commands require auth up front, unless --web can carry it.
			if err := requireSession(ctx, c); err != nil {
				return err
			}

			io := ioFrom(ctx)
			pr := io.Prompter()

			p, err := resolveParams(ctx, ep)
			if err != nil {
				if errors.Is(err, errCancelled) {
					fmt.Fprintln(ctx.Stderr, "Cancelled")
					return nil
				}
				return err
			}
			hadToPrompt := resolvedByPrompting(ep, p)

			// gh ConfirmDeletion: destructive verbs make you type the target.
			if err := confirmDestructive(io, pr, ep, p); err != nil {
				return err
			}

			// gh confirmSubmission: "What's next?" with "Continue in browser".
			// Only for write actions, and only when we actually prompted —
			// scripted runs must stay quiet and deterministic.
			if io.CanPrompt() && isWriteAction(ep) && hadToPrompt {
				action, err := confirmSubmission(io, pr, api.WebTargets[ep.Path].Path != "")
				if err != nil {
					return err
				}
				switch action {
				case CancelAction:
					fmt.Fprintln(ctx.Stderr, "Cancelled")
					return nil
				case PreviewAction:
					return openWeb(ctx, ep, p)
				}
			}

			if ctx.Flags.Bool("paginate") {
				merged, err := callPaginated(ctx, c, ep, p)
				if err != nil {
					return err
				}
				return emit(ctx, merged)
			}

			resp, err := c.Call(ep, p)
			if err != nil {
				return err
			}
			if err := resp.Check(); err != nil {
				return err
			}
			if len(resp.Data) == 0 {
				fmt.Fprintln(ctx.Stdout, "ok")
				return nil
			}
			return emit(ctx, jsonValue(resp.Data))
		},
	}
}

func styleDesc(s string) string {
	switch s {
	case api.StyleBody:
		return "JSON request body ($post2)"
	case api.StyleUpload:
		return "multipart/form-data ($upfile)"
	default:
		return "URL query string (note: even POST sends params here)"
	}
}

// readFile reads a file (used by --field k=@file).
func readFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}
	return b, nil
}

// callPaginated walks pages: list endpoints return { count, list } keyed by page.
func callPaginated(ctx *cli.Context, c *api.Client, ep *api.Endpoint, p *api.Params) (interface{}, error) {
	merged := []interface{}{}
	total := -1
	for page := 1; page <= 1000; page++ {
		p.Fields["page"] = strconv.Itoa(page)
		resp, err := c.Call(ep, p)
		if err != nil {
			return nil, err
		}
		if err := resp.Check(); err != nil {
			return nil, err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(resp.Data, &m); err != nil {
			return jsonValue(resp.Data), nil
		}
		arr, ok := m["list"].([]interface{})
		if !ok {
			return jsonValue(resp.Data), nil
		}
		merged = append(merged, arr...)
		if n, ok := m["count"].(float64); ok {
			total = int(n)
		}
		if len(arr) == 0 {
			break
		}
		if total >= 0 && len(merged) >= total {
			break
		}
		if ctx.Flags.Bool("verbose") {
			stderrf(ctx, "fetched %d/%d", len(merged), total)
		}
	}
	return map[string]interface{}{"count": len(merged), "list": merged}, nil
}
