package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/HFunction2013/coj-cli/internal/api"
	"github.com/HFunction2013/coj-cli/internal/cli"
)

// apiCmd is the gh-style escape hatch: call any endpoint directly.
func apiCmd() *cli.Command {
	return &cli.Command{
		Name:    "api",
		Summary: "Call any endpoint directly (escape hatch)",
		Description: `Call backend endpoints directly, covering anything the
resource commands do not wrap yet.

    coj api <method> <path> [flags]

The path may include or omit the /api prefix. coj inspects the extracted
endpoint definitions to decide whether parameters belong in the URL query
string or a JSON body; unknown endpoints default to the query string, which
is what most endpoints of this backend expect.

Use --body to force a JSON request body.`,
		Example: `  coj api POST /Match/getMatchList --field page=1
  coj api GET '/User/login?user_name=alice&user_pwd=secret'
  coj api POST /TestLog/doTest --body --field F_CodeContent=@solution.cpp
  coj api --list                       # list known endpoints
`,
		Flags: cli.NewFlagSet(
			&cli.Flag{Name: "body", Usage: "Force a JSON body instead of a query string", Bool: true},
			&cli.Flag{Name: "list", Usage: "List every known endpoint and exit", Bool: true},
			&cli.Flag{Name: "filter", Usage: "Filter the endpoint list by keyword (with --list)"},
			&cli.Flag{Name: "raw", Usage: "Print the full response including code and msg", Bool: true},
		),
		Run: func(ctx *cli.Context) error {
			if ctx.Flags.Bool("list") {
				return listEndpoints(ctx)
			}
			if len(ctx.Args) < 2 {
				return fmt.Errorf("usage: coj api <method> <path> [flags]")
			}
			method := strings.ToUpper(ctx.Args[0])
			path := ctx.Args[1]

			// support an inline query string: /User/login?a=b
			p, err := paramsFrom(ctx)
			if err != nil {
				return err
			}
			if eq := strings.Index(path, "?"); eq >= 0 {
				q, err := url.ParseQuery(path[eq+1:])
				if err != nil {
					return err
				}
				for k, vs := range q {
					for _, v := range vs {
						p.Raw[k] = append(p.Raw[k], v)
					}
				}
				path = path[:eq]
			}
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			path = strings.TrimPrefix(path, "/api")

			ep := api.ByPath(path)
			if ep == nil {
				ep = &api.Endpoint{Path: path, Method: method, Style: api.StyleQuery}
			} else if method != "" && method != "DEFAULT" {
				ep = &api.Endpoint{Path: ep.Path, Method: method, Style: ep.Style}
			}
			if ctx.Flags.Bool("body") {
				ep = &api.Endpoint{Path: ep.Path, Method: ep.Method, Style: api.StyleBody}
			}
			if ep.Method == "" {
				ep.Method = "POST"
			}

			// support --field k=@file to read a file
			for k, v := range p.Fields {
				if strings.HasPrefix(v, "@") {
					data, err := readFileMaybe(v[1:])
					if err != nil {
						return err
					}
					p.Fields[k] = string(data)
				}
			}

			c, err := clientFrom(ctx)
			if err != nil {
				return err
			}
			resp, err := c.Call(ep, p)
			if err != nil {
				return err
			}
			if ctx.Flags.Bool("raw") {
				return emit(ctx, map[string]interface{}{
					"code": resp.Code,
					"msg":  resp.Msg,
					"data": jsonValue(resp.Data),
				})
			}
			if err := resp.Check(); err != nil {
				return err
			}
			return emit(ctx, jsonValue(resp.Data))
		},
	}
}

func jsonValue(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	return v
}

func listEndpoints(ctx *cli.Context) error {
	filter := strings.ToLower(ctx.Flags.Get("filter"))
	rows := []map[string]interface{}{}
	for i := range api.Endpoints {
		ep := &api.Endpoints[i]
		label := ep.Resource + " " + ep.Action
		if filter != "" &&
			!strings.Contains(strings.ToLower(ep.Path), filter) &&
			!strings.Contains(strings.ToLower(label), filter) &&
			!strings.Contains(strings.ToLower(ep.Summary), filter) {
			continue
		}
		aliases := ""
		if len(ep.Aliases) > 0 {
			aliases = strings.Join(ep.Aliases, ",")
		}
		rows = append(rows, map[string]interface{}{
			"path":    ep.Path,
			"method":  ep.Method,
			"style":   ep.Style,
			"command": label,
			"aliases": aliases,
			"summary": ep.Summary,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i]["path"].(string) < rows[j]["path"].(string)
	})
	return emit(ctx, rows)
}

func readFileMaybe(path string) ([]byte, error) {
	return readFile(path)
}
