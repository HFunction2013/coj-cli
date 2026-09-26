package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/HFunction2013/coj-cli/internal/api"
	"github.com/HFunction2013/coj-cli/internal/cli"
	"github.com/HFunction2013/coj-cli/internal/prompter"
)

// resolveParams decides the parameters for one call.
//
// Three gh conventions are honoured here:
//  1. Explicitly provided fields always win — never nag the user.
//  2. Prompt only when the terminal can support it, and only for fields that
//     are genuinely required.
//  3. With --web, make no business call; just locate and open the page.
func resolveParams(ctx *cli.Context, ep *api.Endpoint) (*api.Params, error) {
	p, err := paramsFrom(ctx)
	if err != nil {
		return nil, err
	}
	if len(ctx.Args) > 0 {
		if len(ep.Params) > 0 {
			p.Fields[ep.Params[0].Name] = strings.Join(ctx.Args, ",")
		} else if len(ctx.Args) == 1 {
			p.Fields["F_ID"] = ctx.Args[0]
		}
	}
	if ctx.Flags.Bool("web") {
		return p, nil
	}

	missing := missingFields(ep, p)
	if len(missing) == 0 {
		return p, nil
	}

	io := ioFrom(ctx)
	if !io.CanPrompt() {
		// gh never hangs waiting on input it cannot receive.
		hints := make([]string, 0, len(missing))
		for _, m := range missing {
			hints = append(hints, fmt.Sprintf("--field %s=…", m))
		}
		return nil, fmt.Errorf("missing required field(s): %s\n\nProvide them, e.g.:\n  %s",
			strings.Join(missing, ", "),
			fmt.Sprintf("coj %s %s %s", ep.Resource, ep.Action, strings.Join(hints, " ")))
	}

	fmt.Fprintf(ctx.Stderr, "%s %s — %s\n", ep.Method, ep.Path, ep.Summary)
	for _, field := range missing {
		v, err := askField(ctx, ep, field)
		if err != nil {
			return nil, err
		}
		if v != "" {
			p.Fields[field] = v
			p.MarkPrompted(field)
		}
	}
	return p, nil
}

// missingFields lists fields the endpoint expects but the caller did not
// supply and that look required.
//
// "Required" means an ID reference (F_XxxID / F_ID). Pagination and optional
// filters stay silent — gh likewise asks only for what it truly needs.
//
// Endpoints such as doTest carry several optional context ids lifted from the
// route query (F_CourseID, F_Chapter, F_MatchID can all be empty when
// answering a standalone problem). Interrogating the user about each one
// would make every submit painful, so the rule is:
//
//   - if the caller supplied any id, assume they know what they are doing and
//     stay quiet;
//   - if they supplied none, ask about every missing id.
func missingFields(ep *api.Endpoint, p *api.Params) []string {
	suppliedAny := false
	missing := []string{}
	for _, ph := range ep.Params {
		if ph.Name == "page" || ph.Name == "limit" {
			continue
		}
		if !api.IsIDField(ph.Name) {
			continue
		}
		if _, ok := p.Fields[ph.Name]; ok {
			suppliedAny = true
			continue
		}
		missing = append(missing, ph.Name)
	}
	if suppliedAny || len(missing) == 0 {
		return nil
	}
	return missing
}

// askField prompts for one field, backed by real data where possible.
//
// This is where gh's MultiSelectWithSearch earns its keep: instead of
// materialising every row, the prompt opens on a short list with a Search
// sentinel; picking Search asks for a query and re-runs the search.
func askField(ctx *cli.Context, ep *api.Endpoint, field string) (string, error) {
	prompter_ := ioFrom(ctx).Prompter()
	c, err := clientFrom(ctx)
	if err != nil {
		return "", err
	}

	if src, ok := api.RefFor(field); ok {
		searchFunc := func(query string) prompter.MultiSelectSearchResult {
			res, err := searchRef(ctx, c, src, query)
			if err != nil {
				return prompter.MultiSelectSearchResult{Err: err}
			}
			return res
		}
		picked, err := prompter_.MultiSelectWithSearch(
			fmt.Sprintf("Select %s", humanField(field)),
			"Search",
			nil,
			nil,
			searchFunc,
		)
		if err == nil && len(picked) > 0 {
			return picked[0], nil
		}
		if err == prompter.ErrAborted {
			return "", err
		}
		if err != nil && err != prompter.ErrNotInteractive {
			// Search failed (network, permission) — fall back to free text.
			fmt.Fprintf(ctx.Stderr, "warning: could not load choices: %v\n", err)
		}
	}
	return prompter_.Input(humanField(field), "")
}

// searchRef runs one search against the backend and shapes it for the prompt.
func searchRef(ctx *cli.Context, c *api.Client, src api.RefSource, query string) (prompter.MultiSelectSearchResult, error) {
	ep := api.ByPath(src.ListPath)
	if ep == nil {
		return prompter.MultiSelectSearchResult{}, fmt.Errorf("unknown endpoint %s", src.ListPath)
	}
	p := api.NewParams()
	for k, v := range src.Extra {
		p.Fields[k] = v
	}
	p.Fields["page"] = "1"
	if query != "" {
		// Backends differ on the search key; send the common ones.
		p.Fields["F_Name"] = query
		p.Fields["keyword"] = query
	}
	resp, err := c.Call(ep, p)
	if err != nil {
		return prompter.MultiSelectSearchResult{}, err
	}
	if err := resp.Check(); err != nil {
		return prompter.MultiSelectSearchResult{}, err
	}
	return shapeOptions(resp.Data, src)
}

// shapeOptions pulls (label, key) pairs plus a remainder count out of a list
// response. The remainder is what drives the "Search (N more)" hint.
func shapeOptions(raw json.RawMessage, src api.RefSource) (prompter.MultiSelectSearchResult, error) {
	var payload struct {
		List  []map[string]interface{} `json:"list"`
		Count float64                  `json:"count"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil || len(payload.List) == 0 {
		var arr []map[string]interface{}
		if err2 := json.Unmarshal(raw, &arr); err2 != nil {
			return prompter.MultiSelectSearchResult{}, fmt.Errorf("unexpected response shape")
		}
		payload.List = arr
		payload.Count = float64(len(arr))
	}

	const initialPage = 10
	limit := initialPage
	if len(payload.List) < limit {
		limit = len(payload.List)
	}

	keys := make([]string, 0, limit)
	labels := make([]string, 0, limit)
	for _, item := range payload.List[:limit] {
		id := ""
		if v, ok := item[src.IDKey]; ok {
			id = trimFloat(fmt.Sprintf("%v", v))
		}
		if id == "" {
			continue
		}
		label := id
		for _, k := range src.LabelKeys {
			if v, ok := item[k]; ok {
				if s, ok := v.(string); ok && s != "" {
					label = s
					break
				}
			}
		}
		keys = append(keys, id)
		labels = append(labels, fmt.Sprintf("%s (%s)", label, id))
	}

	more := int(payload.Count) - len(keys)
	if more < 0 {
		more = 0
	}
	return prompter.MultiSelectSearchResult{Keys: keys, Labels: labels, MoreResults: more}, nil
}

func trimFloat(s string) string {
	return strings.TrimSuffix(s, ".000000")
}

// humanField turns F_MatchID into "Match ID" — splitting only at lower→upper
// boundaries so "ClassID" does not become "Class I D".
func humanField(f string) string {
	f = strings.TrimPrefix(f, "F_")
	rs := []rune(f)
	var b strings.Builder
	for i, r := range rs {
		if i > 0 && r >= 'A' && r <= 'Z' && rs[i-1] >= 'a' && rs[i-1] <= 'z' {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}
