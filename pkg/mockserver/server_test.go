package mockserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestStaticRouteResponse(t *testing.T) {
	server := newTestHTTPServer(t, `
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /users:
    get:
      x-lab-mock:
        status: 200
        body:
          users: []
`)
	defer server.Close()

	resp, body := doRequest(t, http.MethodGet, server.URL+"/users", "", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected JSON content type, got %q", got)
	}

	assertJSONBody(t, body, map[string]any{"users": []any{}})
}

func TestPathParameterResponse(t *testing.T) {
	server := newTestHTTPServer(t, `
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /users/{id}:
    get:
      x-lab-mock:
        body:
          id: "{{ .Path.id }}"
`)
	defer server.Close()

	resp, body := doRequest(t, http.MethodGet, server.URL+"/users/123", "", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	assertJSONBody(t, body, map[string]any{"id": "123"})
}

func TestQueryAndHeaderAccess(t *testing.T) {
	server := newTestHTTPServer(t, `
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /search:
    get:
      x-lab-mock:
        body:
          q: "{{ index (index .Query \"q\") 0 }}"
          token: "{{ index (index .Headers \"X-Demo\") 0 }}"
`)
	defer server.Close()

	resp, body := doRequest(t, http.MethodGet, server.URL+"/search?q=ferret", "", map[string]string{
		"X-Demo": "lab",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	assertJSONBody(t, body, map[string]any{
		"q":     "ferret",
		"token": "lab",
	})
}

func TestJSONRequestBodyAccess(t *testing.T) {
	server := newTestHTTPServer(t, `
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /users:
    post:
      x-lab-mock:
        status: 201
        body:
          name: "{{ .Body.name }}"
`)
	defer server.Close()

	resp, body := doRequest(t, http.MethodPost, server.URL+"/users", `{"name":"Ada"}`, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	assertJSONBody(t, body, map[string]any{"name": "Ada"})
}

func TestConfiguredHeaders(t *testing.T) {
	server := newTestHTTPServer(t, `
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /headers:
    get:
      x-lab-mock:
        headers:
          x-demo: lab
        body:
          ok: true
`)
	defer server.Close()

	resp, _ := doRequest(t, http.MethodGet, server.URL+"/headers", "", nil)
	defer resp.Body.Close()

	if got := resp.Header.Get("X-Demo"); got != "lab" {
		t.Fatalf("expected x-demo header, got %q", got)
	}
}

func TestBodyTemplate(t *testing.T) {
	server := newTestHTTPServer(t, `
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /users/{id}:
    post:
      x-lab-mock:
        bodyTemplate: |
          {"id": {{ .Path.id }}, "name": {{ toJson .Body.name }}}
`)
	defer server.Close()

	resp, body := doRequest(t, http.MethodPost, server.URL+"/users/123", `{"name":"Ada"}`, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	assertJSONBody(t, body, map[string]any{
		"id":   float64(123),
		"name": "Ada",
	})
}

func TestParameterizedRoutesPreferMoreSpecificMatch(t *testing.T) {
	server := newTestHTTPServer(t, `
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /{collection}/{id}:
    get:
      x-lab-mock:
        body:
          route: generic
          collection: "{{ .Path.collection }}"
          id: "{{ .Path.id }}"
  /users/{id}:
    get:
      x-lab-mock:
        body:
          route: users
          id: "{{ .Path.id }}"
`)
	defer server.Close()

	resp, body := doRequest(t, http.MethodGet, server.URL+"/users/123", "", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	assertJSONBody(t, body, map[string]any{
		"route": "users",
		"id":    "123",
	})
}

func TestMethodNotAllowedReportsAllowedMethods(t *testing.T) {
	server := newTestHTTPServer(t, `
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /users/{id}:
    get:
      x-lab-mock:
        body:
          ok: true
    delete:
      x-lab-mock:
        status: 204
`)
	defer server.Close()

	resp, _ := doRequest(t, http.MethodPost, server.URL+"/users/123", "", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}

	if got := resp.Header.Get("Allow"); got != "DELETE, GET" {
		t.Fatalf("expected Allow header, got %q", got)
	}
}

func TestValidationRejectsBodyAndBodyTemplate(t *testing.T) {
	_, err := New(Options{SpecData: []byte(`
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /users:
    get:
      x-lab-mock:
        body:
          ok: true
        bodyTemplate: "ok"
`)})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("expected mutually exclusive error, got %v", err)
	}
}

func TestNotFoundAndMethodNotAllowed(t *testing.T) {
	server := newTestHTTPServer(t, `
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /users:
    get:
      x-lab-mock:
        body:
          ok: true
`)
	defer server.Close()

	resp, _ := doRequest(t, http.MethodGet, server.URL+"/missing", "", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	resp, _ = doRequest(t, http.MethodPost, server.URL+"/users", "", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestSprigHelpersAreAvailable(t *testing.T) {
	server := newTestHTTPServer(t, `
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /ids:
    get:
      x-lab-mock:
        body:
          id: "{{ uuidv4 }}"
          code: "{{ randAlphaNum 8 }}"
`)
	defer server.Close()

	resp, body := doRequest(t, http.MethodGet, server.URL+"/ids", "", nil)
	defer resp.Body.Close()

	var payload map[string]string
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !regexp.MustCompile(`^[0-9a-f-]{36}$`).MatchString(payload["id"]) {
		t.Fatalf("expected UUID, got %q", payload["id"])
	}

	if len(payload["code"]) != 8 {
		t.Fatalf("expected 8-character code, got %q", payload["code"])
	}
}

func TestDangerousSprigHelpersAreNotAvailable(t *testing.T) {
	tests := []string{
		`{{ env "HOME" }}`,
		`{{ expandenv "$HOME" }}`,
		`{{ getHostByName "localhost" }}`,
	}

	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			_, err := New(Options{SpecData: []byte(`
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /unsafe:
    get:
      x-lab-mock:
        bodyTemplate: '` + source + `'
`)})
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestDangerousSprigHelpersAreNotAvailableInBody(t *testing.T) {
	_, err := New(Options{SpecData: []byte(`
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /unsafe:
    get:
      x-lab-mock:
        body:
          home: '{{ env "HOME" }}'
`)})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestInvalidJSONRequestBodyReturnsBadRequest(t *testing.T) {
	server := newTestHTTPServer(t, `
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /users:
    post:
      x-lab-mock:
        body:
          ok: true
`)
	defer server.Close()

	resp, _ := doRequest(t, http.MethodPost, server.URL+"/users", `{`, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func newTestHTTPServer(t *testing.T, spec string) *httptest.Server {
	t.Helper()

	server, err := New(Options{SpecData: []byte(spec)})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	return httptest.NewServer(server.Handler())
}

func doRequest(t *testing.T, method string, target string, body string, headers map[string]string) (*http.Response, []byte) {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, target, reader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	for name, value := range headers {
		req.Header.Set(name, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		t.Fatalf("failed to read response: %v", err)
	}

	resp.Body = io.NopCloser(bytes.NewReader(data))

	return resp, data
}

func assertJSONBody(t *testing.T, data []byte, expected any) {
	t.Helper()

	var actual any
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("failed to decode response body %q: %v", string(data), err)
	}

	expectedData, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("failed to marshal expected body: %v", err)
	}

	var normalizedExpected any
	if err := json.Unmarshal(expectedData, &normalizedExpected); err != nil {
		t.Fatalf("failed to normalize expected body: %v", err)
	}

	if fmtJSON(actual) != fmtJSON(normalizedExpected) {
		t.Fatalf("expected body %s, got %s", fmtJSON(normalizedExpected), fmtJSON(actual))
	}
}

func fmtJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}
