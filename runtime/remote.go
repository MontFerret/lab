package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
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
		Text   string                 `json:"text"`
		Params map[string]interface{} `json:"params"`
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

func NewRemote(u string, params map[string]interface{}) (*Remote, error) {
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
			headers, ok := headers.(map[string]interface{})

			if !ok {
				return nil, errors.New("invalid type of headers (expected map)")
			}

			for k, v := range headers {
				str, ok := v.(string)

				if !ok {
					return nil, errors.Errorf("invalid value type of a header: %s (expected string)", k)
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
			cookies, ok := cookies.(map[string]interface{})

			if !ok {
				return nil, errors.New("invalid type of cookies (expected map)")
			}

			for k, v := range cookies {
				str, ok := v.(string)

				if !ok {
					return nil, errors.Errorf("invalid value type of a header: %s (expected string)", k)
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

func (rt *Remote) Version(ctx context.Context) (string, error) {
	data, err := rt.makeRequest(ctx, "GET", "/info", nil)

	if err != nil {
		return "", err
	}

	info := remoteInfo{}

	if err := json.Unmarshal(data, &info); err != nil {
		return "", errors.Wrap(err, "deserialize response data")
	}

	return info.Version.Ferret, nil
}

func (rt *Remote) Run(ctx context.Context, query string, params map[string]interface{}) ([]byte, error) {
	body, err := json.Marshal(remoteQuery{
		Text:   query,
		Params: params,
	})

	if err != nil {
		return nil, errors.Wrap(err, "serialize query")
	}

	return rt.makeRequest(ctx, "POST", "/", body)
}

func (rt *Remote) createRequest(ctx context.Context, method, endpoint string, body []byte) (*http.Request, error) {
	var reader io.Reader = nil

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

	req.Header = rt.params.Headers

	for _, c := range rt.params.Cookies {
		req.AddCookie(&c)
	}

	return req, nil
}

func (rt *Remote) makeRequest(ctx context.Context, method, endpoint string, body []byte) ([]byte, error) {
	req, err := rt.createRequest(ctx, method, endpoint, body)

	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}

	resp, err := rt.client.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "make HTTP request to remote runtime")
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.New(resp.Status)
	}

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.Wrap(err, "read response data")
	}

	return data, nil
}
