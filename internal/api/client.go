// Package api wraps the CandyOJ backend REST calls.
//
// Backend contract (recovered from the production frontend bundle):
//   - base URL: same origin, /api/
//   - response: { "code": int, "msg": string, "data": any }
//     code == 200 means success, anything else is a business error
//   - auth: PHP session, cookie PHPSESSID
//   - parameter placement: see Endpoint.Style
//
// An easy trap: most POST endpoints of this project take parameters in the
// URL query string (that is how the frontend $post is written); only a few
// endpoints ($post2) use a JSON body. The client therefore encodes strictly
// according to Endpoint.Style.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Response is the uniform backend response shape.
type Response struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`

	// these help debugging and are not returned by the backend
	Status int    `json:"-"`
	Path   string `json:"-"`
}

// OK reports whether the call succeeded.
func (r *Response) OK() bool { return r.Code == 200 }

// ErrNotLogin means the session is gone (backend code 440).
var ErrNotLogin = fmt.Errorf("not logged in or the session expired (440)")

// ErrQuota means a quota was exceeded (461); --force retries with force=1.
type ErrQuota struct {
	Msg string
}

func (e *ErrQuota) Error() string {
	return "quota exceeded (461): " + e.Msg + "; re-run with --force to continue"
}

// APIError is any non-200 business error.
type APIError struct {
	Code int
	Msg  string
	Path string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("endpoint %s returned code=%d: %s", e.Path, e.Code, e.Msg)
}

// Client holds the session and configuration.
type Client struct {
	BaseURL  string
	HTTP     *http.Client
	Timeout  time.Duration
	Insecure bool
	Verbose  bool
	Log      io.Writer
}

// NewClient creates a client.
func NewClient(baseURL string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid host URL: %w", err)
	}
	return &Client{
		BaseURL: strings.TrimRight(u.String(), "/"),
		HTTP:    &http.Client{Jar: jar, Timeout: 30 * time.Second},
		Timeout: 30 * time.Second,
		Log:     os.Stderr,
	}, nil
}

// SessionID returns the session cookie, empty when not logged in.
func (c *Client) SessionID() string {
	u, _ := url.Parse(c.BaseURL)
	for _, ck := range c.HTTP.Jar.Cookies(u) {
		if ck.Name == "PHPSESSID" {
			return ck.Value
		}
	}
	return ""
}

// SetSession injects a session id (used by --session or config restore).
func (c *Client) SetSession(id string) {
	if id == "" {
		return
	}
	u, _ := url.Parse(c.BaseURL)
	c.HTTP.Jar.SetCookies(u, []*http.Cookie{{Name: "PHPSESSID", Value: id, Path: "/"}})
}

// Params is the parameter set for one call.
type Params struct {
	Fields map[string]string   // plain fields
	Files  map[string]string   // file fields (name -> local path)
	Force  bool                // retry after a quota confirmation
	Raw    map[string][]string // extra query parameters, already encoded

	// prompted records fields filled by asking the user, so callers can tell
	// an interactive run from a scripted one (see resolvedByPrompting).
	prompted map[string]bool
}

// PromptedFields lists fields that were filled interactively.
func (p *Params) PromptedFields() []string {
	out := make([]string, 0, len(p.prompted))
	for k := range p.prompted {
		out = append(out, k)
	}
	return out
}

// MarkPrompted records that a field was supplied interactively.
func (p *Params) MarkPrompted(field string) {
	if p.prompted == nil {
		p.prompted = map[string]bool{}
	}
	p.prompted[field] = true
}

// NewParams creates a parameter set.
func NewParams() *Params {
	return &Params{Fields: map[string]string{}, Files: map[string]string{}, Raw: map[string][]string{}}
}

// Call invokes one endpoint.
func (c *Client) Call(ep *Endpoint, p *Params) (*Response, error) {
	if ep == nil {
		return nil, fmt.Errorf("no endpoint given")
	}
	target := c.BaseURL + "/api" + ep.Path

	var body io.Reader
	ctype := ""
	query := url.Values{}

	switch ep.Style {
	case StyleBody:
		payload := map[string]string{}
		for k, v := range p.Fields {
			payload[k] = v
		}
		if p.Force {
			payload["force"] = "1"
		}
		buf, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(buf)
		ctype = "application/json"

	case StyleUpload:
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		for k, v := range p.Fields {
			_ = mw.WriteField(k, v)
		}
		for k, path := range p.Files {
			if err := addFile(mw, k, path); err != nil {
				return nil, err
			}
		}
		if p.Force {
			_ = mw.WriteField("force", "1")
		}
		if err := mw.Close(); err != nil {
			return nil, err
		}
		body = &buf
		ctype = mw.FormDataContentType()

	default: // StyleQuery — note that POST also uses the query string
		for k, v := range p.Fields {
			query.Set(k, v)
		}
		if p.Force {
			query.Set("force", "1")
		}
	}

	for k, vs := range p.Raw {
		for _, v := range vs {
			query.Add(k, v)
		}
	}
	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	method := ep.Method
	if method == "" {
		method = "POST"
	}
	req, err := http.NewRequest(method, target, body)
	if err != nil {
		return nil, err
	}
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	if c.Verbose && c.Log != nil {
		fmt.Fprintf(c.Log, "> %s %s\n", method, target)
		if len(p.Fields) > 0 {
			fmt.Fprintf(c.Log, "> params: %v\n", p.Fields)
		}
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if c.Verbose && c.Log != nil {
		snippet := string(raw)
		if len(snippet) > 800 {
			snippet = snippet[:800] + "…"
		}
		fmt.Fprintf(c.Log, "< %d %s\n", resp.StatusCode, snippet)
	}

	out := &Response{Status: resp.StatusCode, Path: ep.Path}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			// the backend may return HTML (e.g. a WAF block or a 502 page)
			preview := string(raw)
			if len(preview) > 200 {
				preview = preview[:200]
			}
			return out, fmt.Errorf("response is not valid JSON (HTTP %d): %s", resp.StatusCode, preview)
		}
	}
	if resp.StatusCode >= 400 {
		return out, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return out, nil
}

// Check validates the business code, turning 440/461 into typed errors.
func (r *Response) Check() error {
	if r == nil {
		return fmt.Errorf("empty response")
	}
	switch r.Code {
	case 200:
		return nil
	case 440:
		return ErrNotLogin
	case 461:
		return &ErrQuota{Msg: r.Msg}
	default:
		return &APIError{Code: r.Code, Msg: r.Msg, Path: r.Path}
	}
}

func addFile(mw *multipart.Writer, field, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()
	part, err := mw.CreateFormFile(field, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f)
	return err
}

// DataString decodes data as a string (the backend sometimes returns one).
func (r *Response) DataString() string {
	var s string
	if err := json.Unmarshal(r.Data, &s); err == nil {
		return s
	}
	return string(r.Data)
}

// DataInt decodes data as an integer.
func (r *Response) DataInt() (int, bool) {
	var n int
	if err := json.Unmarshal(r.Data, &n); err == nil {
		return n, true
	}
	var s string
	if err := json.Unmarshal(r.Data, &s); err == nil {
		if v, err := strconv.Atoi(s); err == nil {
			return v, true
		}
	}
	return 0, false
}
