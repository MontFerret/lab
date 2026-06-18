package mockserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func loadSpec(opts Options) ([]byte, error) {
	hasPath := strings.TrimSpace(opts.SpecPath) != ""
	hasData := len(opts.SpecData) > 0

	switch {
	case hasPath && hasData:
		return nil, errors.New("mock API spec path and data are mutually exclusive")
	case hasPath:
		data, err := os.ReadFile(opts.SpecPath)
		if err != nil {
			return nil, errors.Wrapf(err, "read mock API spec %q", opts.SpecPath)
		}

		return data, nil
	case hasData:
		return opts.SpecData, nil
	default:
		return nil, errors.New("mock API spec path or data is required")
	}
}

func parseSpec(data []byte) (map[string]any, error) {
	var raw any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, errors.Wrap(err, "parse mock API spec")
	}

	root, ok := normalizeValue(raw).(map[string]any)
	if !ok {
		return nil, errors.New("mock API spec must be an object")
	}

	return root, nil
}

func buildServer(root map[string]any) (*Server, error) {
	pathsRaw, ok := root["paths"].(map[string]any)
	if !ok {
		return nil, errors.New("mock API spec paths must be an object")
	}

	server := &Server{}
	routes := make(map[string]*route)
	matchKeys := make(map[string]string)

	for path, pathItemRaw := range pathsRaw {
		pathItem, ok := pathItemRaw.(map[string]any)
		if !ok {
			return nil, errors.Errorf("mock API path %q must be an object", path)
		}

		for methodRaw, operationRaw := range pathItem {
			method := strings.ToUpper(methodRaw)
			if _, ok := supportedMethods[method]; !ok {
				continue
			}

			operationMap, ok := operationRaw.(map[string]any)
			if !ok {
				return nil, errors.Errorf("mock API operation %s %s must be an object", methodRaw, path)
			}

			mockRaw, ok := operationMap["x-lab-mock"]
			if !ok {
				continue
			}

			op, err := parseOperation(path, methodRaw, mockRaw)
			if err != nil {
				return nil, err
			}

			rt, err := routeForPath(routes, matchKeys, path)
			if err != nil {
				return nil, err
			}

			if rt.ops[method] != nil {
				return nil, errors.Errorf("duplicate mock API operation %s %s", methodRaw, path)
			}

			rt.ops[method] = op
		}
	}

	for _, rt := range routes {
		if len(rt.ops) == 0 {
			continue
		}

		if rt.static {
			server.staticRoutes = append(server.staticRoutes, rt)
		} else {
			server.paramRoutes = append(server.paramRoutes, rt)
		}
	}

	sortRoutes(server.staticRoutes)
	sortRoutes(server.paramRoutes)

	return server, nil
}

func routeForPath(routes map[string]*route, matchKeys map[string]string, path string) (*route, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, errors.Errorf("mock API path %q must start with /", path)
	}

	if rt, ok := routes[path]; ok {
		return rt, nil
	}

	segments := parseRouteSegments(path)
	matchKey := routeMatchKey(segments)
	if existing, ok := matchKeys[matchKey]; ok && existing != path {
		return nil, errors.Errorf("ambiguous mock API routes %q and %q", existing, path)
	}

	rt := &route{
		path:     path,
		segments: segments,
		static:   routeIsStatic(segments),
		ops:      make(map[string]*operation),
	}

	routes[path] = rt
	matchKeys[matchKey] = path

	return rt, nil
}

func parseOperation(path string, method string, raw any) (*operation, error) {
	mock, ok := raw.(map[string]any)
	if !ok {
		return nil, errors.Errorf("x-lab-mock for %s %s must be an object", method, path)
	}

	op := &operation{
		status:  http.StatusOK,
		headers: make(map[string]string),
	}

	if rawStatus, ok := mock["status"]; ok {
		status, err := parseStatus(rawStatus)
		if err != nil {
			return nil, errors.Wrapf(err, "x-lab-mock status for %s %s", method, path)
		}

		op.status = status
	}

	if rawHeaders, ok := mock["headers"]; ok {
		headers, err := parseHeaders(rawHeaders)
		if err != nil {
			return nil, errors.Wrapf(err, "x-lab-mock headers for %s %s", method, path)
		}

		op.headers = headers
	}

	rawBody, hasBody := mock["body"]
	rawBodyTemplate, hasBodyTemplate := mock["bodyTemplate"]
	if hasBody && hasBodyTemplate {
		return nil, errors.Errorf("x-lab-mock body and bodyTemplate for %s %s are mutually exclusive", method, path)
	}

	if hasBody {
		op.body = rawBody
		op.hasBody = true

		if err := validateBodyTemplates(rawBody); err != nil {
			return nil, errors.Wrapf(err, "x-lab-mock body for %s %s", method, path)
		}
	}

	if hasBodyTemplate {
		bodyTemplate, ok := rawBodyTemplate.(string)
		if !ok {
			return nil, errors.Errorf("x-lab-mock bodyTemplate for %s %s must be a string", method, path)
		}

		tmpl, err := parseTemplate(bodyTemplate)
		if err != nil {
			return nil, errors.Wrapf(err, "x-lab-mock bodyTemplate for %s %s", method, path)
		}

		op.bodyTemplate = tmpl
	}

	return op, nil
}

func parseStatus(raw any) (int, error) {
	switch value := raw.(type) {
	case int:
		if value < 100 || value > 599 {
			return 0, errors.Errorf("must be between 100 and 599")
		}

		return value, nil
	case int64:
		if value < 100 || value > 599 {
			return 0, errors.Errorf("must be between 100 and 599")
		}

		return int(value), nil
	case float64:
		status := int(value)
		if value != float64(status) || status < 100 || status > 599 {
			return 0, errors.Errorf("must be an integer between 100 and 599")
		}

		return status, nil
	default:
		return 0, errors.Errorf("must be a number")
	}
}

func parseHeaders(raw any) (map[string]string, error) {
	source, ok := raw.(map[string]any)
	if !ok {
		return nil, errors.New("must be an object")
	}

	headers := make(map[string]string, len(source))
	for name, rawValue := range source {
		value, ok := rawValue.(string)
		if !ok {
			return nil, errors.Errorf("header %q must be a string", name)
		}

		headers[name] = value
	}

	return headers, nil
}

func requestBody(r *http.Request) (any, error) {
	if r.Body == nil {
		return nil, nil
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}

	var body any
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, err
	}

	return body, nil
}

func writeMethodNotAllowed(w http.ResponseWriter, allowed map[string]struct{}) {
	methods := make([]string, 0, len(allowed))
	for method := range allowed {
		methods = append(methods, method)
	}

	sort.Strings(methods)
	if len(methods) > 0 {
		w.Header().Set("Allow", strings.Join(methods, ", "))
	}

	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func addAllowedMethods(allowed map[string]struct{}, rt *route) {
	for method := range rt.ops {
		allowed[method] = struct{}{}
	}
}

func normalizeValue(value any) any {
	switch typed := value.(type) {
	case map[interface{}]interface{}:
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			out[fmt.Sprint(key)] = normalizeValue(child)
		}

		return out
	case map[string]interface{}:
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			out[key] = normalizeValue(child)
		}

		return out
	case []interface{}:
		out := make([]any, len(typed))
		for idx, child := range typed {
			out[idx] = normalizeValue(child)
		}

		return out
	default:
		return value
	}
}

func hasHeader(headers http.Header, name string) bool {
	for headerName := range headers {
		if strings.EqualFold(headerName, name) {
			return true
		}
	}

	return false
}

func parseRouteSegments(path string) []routeSegment {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}

	parts := strings.Split(trimmed, "/")
	segments := make([]routeSegment, 0, len(parts))

	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") && len(part) > 2 {
			segments = append(segments, routeSegment{
				value: strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}"),
				param: true,
			})
			continue
		}

		segments = append(segments, routeSegment{value: part})
	}

	return segments
}

func routeIsStatic(segments []routeSegment) bool {
	for _, segment := range segments {
		if segment.param {
			return false
		}
	}

	return true
}

func routeMatchKey(segments []routeSegment) string {
	if len(segments) == 0 {
		return "/"
	}

	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment.param {
			parts = append(parts, "{}")
		} else {
			parts = append(parts, segment.value)
		}
	}

	return "/" + strings.Join(parts, "/")
}

func routeLess(left, right *route) bool {
	leftStatic := routeStaticSegmentCount(left)
	rightStatic := routeStaticSegmentCount(right)
	if leftStatic != rightStatic {
		return leftStatic > rightStatic
	}

	if len(left.segments) != len(right.segments) {
		return len(left.segments) > len(right.segments)
	}

	return left.path < right.path
}

func routeStaticSegmentCount(rt *route) int {
	var count int
	for _, segment := range rt.segments {
		if !segment.param {
			count++
		}
	}

	return count
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}

	return strings.Split(trimmed, "/")
}
