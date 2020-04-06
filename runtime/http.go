package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/pkg/errors"
)

type (
	HTTPParams struct {
		Headers http.Header
		Path    string
		Cookies []http.Cookie
	}

	HTTP struct {
		url    string
		params HTTPParams
	}
)

func NewHTTP(url string, params map[string]interface{}) (*HTTP, error) {
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
					Domain: url,
					Path:   p.Path,
				})
			}
		}
	}

	return &HTTP{url: url, params: p}, nil
}

func (r HTTP) Run(ctx context.Context, query string, params map[string]interface{}) ([]byte, error) {
	j, err := json.Marshal(map[string]interface{}{
		"text":   query,
		"params": params,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare a request body")
	}

	b := &bytes.Buffer{}
	b.Write(j)

	u := r.url

	if r.params.Path != "" {
		u = path.Join(u, r.params.Path)
	}

	req, err := http.NewRequest("POST", u, b)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create a request")
	}

	req.Header = r.params.Headers

	for _, c := range r.params.Cookies {
		req.AddCookie(&c)
	}

	fmt.Println("headers:", req.Header)

	c := http.Client{}
	resp, err := c.Do(req.WithContext(ctx))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.Wrap(err, "failed to read a response body")
	}

	return data, nil
}
