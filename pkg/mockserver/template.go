package mockserver

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
)

type (
	TemplateContext struct {
		Method  string
		Path    map[string]string
		Query   map[string][]string
		Headers map[string][]string
		Body    any
	}
)

func renderBody(value any, ctx TemplateContext) (any, error) {
	switch typed := value.(type) {
	case string:
		return renderStringTemplate(typed, ctx)
	case []any:
		out := make([]any, len(typed))
		for idx, child := range typed {
			rendered, err := renderBody(child, ctx)
			if err != nil {
				return nil, err
			}

			out[idx] = rendered
		}

		return out, nil
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			rendered, err := renderBody(child, ctx)
			if err != nil {
				return nil, err
			}

			out[key] = rendered
		}
		return out, nil
	default:
		return value, nil
	}
}

func renderStringTemplate(source string, ctx TemplateContext) (string, error) {
	tmpl, err := parseTemplate(source)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", errors.Wrap(err, "render template")
	}

	return buf.String(), nil
}

func validateBodyTemplates(value any) error {
	switch typed := value.(type) {
	case string:
		_, err := parseTemplate(typed)
		return err
	case []any:
		for _, child := range typed {
			if err := validateBodyTemplates(child); err != nil {
				return err
			}
		}
	case map[string]any:
		for _, child := range typed {
			if err := validateBodyTemplates(child); err != nil {
				return err
			}
		}
	}

	return nil
}

func templateFuncs() template.FuncMap {
	funcs := sprig.TxtFuncMap()

	delete(funcs, "env")
	delete(funcs, "expandenv")
	delete(funcs, "getHostByName")

	return funcs
}

func parseTemplate(source string) (*template.Template, error) {
	return template.New("mock").Funcs(templateFuncs()).Parse(source)
}
