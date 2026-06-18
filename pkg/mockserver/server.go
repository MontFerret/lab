package mockserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"text/template"
)

type (
	Options struct {
		SpecPath string
		SpecData []byte
	}

	Server struct {
		staticRoutes []*route
		paramRoutes  []*route
	}

	operation struct {
		status       int
		headers      map[string]string
		body         any
		hasBody      bool
		bodyTemplate *template.Template
	}
)

var supportedMethods = map[string]struct{}{
	http.MethodGet:    {},
	http.MethodPost:   {},
	http.MethodPut:    {},
	http.MethodPatch:  {},
	http.MethodDelete: {},
}

func New(opts Options) (*Server, error) {
	data, err := loadSpec(opts)
	if err != nil {
		return nil, err
	}

	root, err := parseSpec(data)
	if err != nil {
		return nil, err
	}

	return buildServer(root)
}

func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(s.ServeHTTP)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := strings.ToUpper(r.Method)
	pathMatched := false
	allowed := make(map[string]struct{})

	for _, rt := range s.staticRoutes {
		if rt.path != r.URL.Path {
			continue
		}

		pathMatched = true
		op := rt.ops[method]
		if op == nil {
			addAllowedMethods(allowed, rt)
			writeMethodNotAllowed(w, allowed)
			return
		}

		s.serveOperation(w, r, op, nil)
		return
	}

	for _, rt := range s.paramRoutes {
		params, ok := rt.match(r.URL.Path)
		if !ok {
			continue
		}

		pathMatched = true
		op := rt.ops[method]
		if op != nil {
			s.serveOperation(w, r, op, params)
			return
		}

		addAllowedMethods(allowed, rt)
	}

	if pathMatched {
		writeMethodNotAllowed(w, allowed)
		return
	}

	http.NotFound(w, r)
}

func (s *Server) serveOperation(w http.ResponseWriter, r *http.Request, op *operation, params map[string]string) {
	body, err := requestBody(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx := TemplateContext{
		Method:  r.Method,
		Path:    params,
		Query:   r.URL.Query(),
		Headers: r.Header,
		Body:    body,
	}

	for name, value := range op.headers {
		w.Header().Set(name, value)
	}

	switch {
	case op.hasBody:
		if !hasHeader(w.Header(), "Content-Type") {
			w.Header().Set("Content-Type", "application/json")
		}

		rendered, err := renderBody(op.body, ctx)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		payload, err := json.Marshal(rendered)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(op.status)
		_, _ = w.Write(payload)
	case op.bodyTemplate != nil:
		if !hasHeader(w.Header(), "Content-Type") {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}

		var buf bytes.Buffer
		if err := op.bodyTemplate.Execute(&buf, ctx); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(op.status)
		_, _ = w.Write(buf.Bytes())
	default:
		w.WriteHeader(op.status)
	}
}

func sortRoutes(routes []*route) {
	sort.Slice(routes, func(i, j int) bool {
		return routeLess(routes[i], routes[j])
	})
}
