package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/MontFerret/api"
)

type (
	remoteVersion struct {
		Worker string `json:"worker"`
		Ferret string `json:"ferret"`
	}

	remoteInfo struct {
		IP      string        `json:"ip"`
		Version remoteVersion `json:"version"`
	}

	remoteQuery struct {
		Text   string         `json:"text"`
		Params map[string]any `json:"params"`
	}

	HTTPParams struct {
		Headers http.Header
		Path    string
		Cookies []http.Cookie
	}

	Remote struct {
		url    *url.URL
		client *http.Client
		params HTTPParams
	}
)

var _ api.Runtime = (*Remote)(nil)

// NewRemote borrows the default HTTP client and owns each response body.
func NewRemote(u string, params map[string]any) (*Remote, error) {
	p := HTTPParams{
		Headers: http.Header{
			"Content-Type":    []string{"application/json"},
			"Accept":          []string{"*/*"},
			"Accept-Charset":  []string{"utf-8"},
			"Accept-Encoding": []string{"gzip", "deflate"},
			"Cache-Control":   []string{"no-cache"},
		},
		Cookies: make([]http.Cookie, 0, 5),
	}

	if params != nil {
		headers, exists := params["headers"]

		if exists {
			headers, ok := headers.(map[string]any)

			if !ok {
				return nil, errors.New("invalid type of headers (expected map)")
			}

			for k, v := range headers {
				str, ok := v.(string)

				if !ok {
					return nil, fmt.Errorf("invalid value type of a header: %s (expected string)", k)
				}

				p.Headers.Add(k, str)
			}
		}

		urlPath, exists := params["path"]

		if exists {
			urlPath, ok := urlPath.(string)

			if !ok {
				return nil, errors.New("invalid type of path (expected string)")
			}

			p.Path = urlPath
		}

		cookies, exists := params["cookies"]

		if exists {
			cookies, ok := cookies.(map[string]any)

			if !ok {
				return nil, errors.New("invalid type of cookies (expected map)")
			}

			for k, v := range cookies {
				str, ok := v.(string)

				if !ok {
					return nil, fmt.Errorf("invalid value type of a header: %s (expected string)", k)
				}

				p.Cookies = append(p.Cookies, http.Cookie{
					Name:   k,
					Value:  str,
					Domain: u,
					Path:   p.Path,
				})
			}
		}
	}

	parsedURL, err := url.Parse(u)

	if err != nil {
		return nil, err
	}

	client := http.DefaultClient

	return &Remote{url: parsedURL, client: client, params: p}, nil
}

// Version reports the hosted Ferret version unchanged.
func (rt *Remote) Version(ctx context.Context) (api.Version, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	out, err := rt.makeRequest(ctx, "GET", rt.versionEndpoint(), nil)
	if out == nil {
		return "", err
	}

	info := remoteInfo{}
	if decodeErr := json.Unmarshal(out.Content, &info); decodeErr != nil {
		return "", errors.Join(err, fmt.Errorf("deserialize response data: %w", decodeErr))
	}

	return api.Version(info.Version.Ferret), err
}

// Run preserves encoded response bytes and any available output on read or close failure.
func (rt *Remote) Run(ctx context.Context, src api.Source, setters ...api.SessionOption) (*api.Output, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	opts, err := newRemoteSessionOptions(setters)
	if err != nil {
		return nil, fmt.Errorf("configure HTTP runtime execution: %w", err)
	}

	body, err := json.Marshal(remoteQuery{Text: src.Content, Params: opts.params})
	if err != nil {
		return nil, fmt.Errorf("serialize query: %w", err)
	}

	return rt.makeRequest(ctx, "POST", rt.runEndpoint(), body)
}

// Compile rejects reusable compilation without HTTP I/O.
func (rt *Remote) Compile(ctx context.Context, _ api.Source, _ ...api.PlanOption) (api.Plan, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("HTTP runtime does not support compilation: %w", errors.ErrUnsupported)
}

// CompileDebug rejects debug compilation without HTTP I/O.
func (rt *Remote) CompileDebug(ctx context.Context, _ api.Source, _ ...api.PlanOption) (api.Plan, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("HTTP runtime does not support debug compilation: %w", errors.ErrUnsupported)
}

// Close is repeatable and leaves the borrowed HTTP client usable.
func (rt *Remote) Close() error {
	return nil
}

func (rt *Remote) runEndpoint() string {
	if rt.params.Path != "" {
		return rt.params.Path
	}

	return rt.basePath()
}

func (rt *Remote) versionEndpoint() string {
	basePath := rt.basePath()

	if basePath == "/" {
		return "/info"
	}

	return path.Join(basePath, "info")
}

func (rt *Remote) basePath() string {
	if rt.url.Path == "" {
		return "/"
	}

	return rt.url.Path
}

func (rt *Remote) createRequest(ctx context.Context, method, endpoint string, body []byte) (*http.Request, error) {
	var reader io.Reader

	if body != nil {
		reader = bytes.NewReader(body)
	}

	u2, err := url.Parse(endpoint)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, rt.url.ResolveReference(u2).String(), reader)

	if err != nil {
		return nil, err
	}

	req.Header = rt.params.Headers.Clone()

	for _, c := range rt.params.Cookies {
		req.AddCookie(&c)
	}

	return req, nil
}

func (rt *Remote) makeRequest(ctx context.Context, method, endpoint string, body []byte) (output *api.Output, err error) {
	req, err := rt.createRequest(ctx, method, endpoint, body)

	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := rt.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("make HTTP request to remote runtime: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close response data: %w", closeErr))
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.New(resp.Status)
	}

	data, readErr := io.ReadAll(resp.Body)
	if readErr == nil || len(data) > 0 {
		output = &api.Output{Content: data, ContentType: resp.Header.Get("Content-Type")}
	}

	if readErr != nil {
		return output, fmt.Errorf("read response data: %w", readErr)
	}

	return output, nil
}
