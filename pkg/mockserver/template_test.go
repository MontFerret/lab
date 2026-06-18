package mockserver

import (
	"testing"
)

func TestCompileBodyTemplatesCompilesStrings(t *testing.T) {
	body := map[string]any{
		"greeting": "Hello {{ .Path.name }}",
		"static":   "no-template",
	}

	templates, err := compileBodyTemplates(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}

	if templates["Hello {{ .Path.name }}"] == nil {
		t.Fatal("expected template for greeting string")
	}

	if templates["no-template"] == nil {
		t.Fatal("expected template for static string")
	}
}

func TestCompileBodyTemplatesWalksNestedStructures(t *testing.T) {
	body := map[string]any{
		"user": map[string]any{
			"name": "{{ .Path.name }}",
		},
		"items": []any{
			"{{ .Path.id }}",
			"literal",
		},
	}

	templates, err := compileBodyTemplates(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if templates["{{ .Path.name }}"] == nil {
		t.Fatal("expected template for nested map string")
	}

	if templates["{{ .Path.id }}"] == nil {
		t.Fatal("expected template for array string")
	}

	if templates["literal"] == nil {
		t.Fatal("expected template for literal string")
	}
}

func TestCompileBodyTemplatesRejectsInvalidTemplate(t *testing.T) {
	body := map[string]any{
		"bad": "{{ .Unclosed",
	}

	_, err := compileBodyTemplates(body)
	if err == nil {
		t.Fatal("expected error for invalid template syntax")
	}
}

func TestRenderBodyUsesPreCompiledTemplates(t *testing.T) {
	body := map[string]any{
		"id": "{{ .Path.id }}",
	}

	templates, err := compileBodyTemplates(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx := TemplateContext{
		Path: map[string]string{"id": "42"},
	}

	result, err := renderBody(body, ctx, templates)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	if resultMap["id"] != "42" {
		t.Fatalf("expected id=42, got %v", resultMap["id"])
	}
}

func TestRenderBodyRendersNestedTemplates(t *testing.T) {
	body := map[string]any{
		"user": map[string]any{
			"name": "{{ .Path.name }}",
		},
		"tags": []any{
			"{{ .Path.tag }}",
			"fixed",
		},
	}

	templates, err := compileBodyTemplates(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx := TemplateContext{
		Path: map[string]string{"name": "Ada", "tag": "dev"},
	}

	result, err := renderBody(body, ctx, templates)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	resultMap := result.(map[string]any)
	userMap := resultMap["user"].(map[string]any)

	if userMap["name"] != "Ada" {
		t.Fatalf("expected name=Ada, got %v", userMap["name"])
	}

	tags := resultMap["tags"].([]any)
	if tags[0] != "dev" {
		t.Fatalf("expected first tag=dev, got %v", tags[0])
	}

	if tags[1] != "fixed" {
		t.Fatalf("expected second tag=fixed, got %v", tags[1])
	}
}

func TestRenderBodyPassesThroughNonStringValues(t *testing.T) {
	body := map[string]any{
		"count":  42,
		"active": true,
		"score":  3.14,
		"empty":  nil,
	}

	templates, err := compileBodyTemplates(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result, err := renderBody(body, TemplateContext{}, templates)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	resultMap := result.(map[string]any)

	if resultMap["count"] != 42 {
		t.Fatalf("expected count=42, got %v", resultMap["count"])
	}
	if resultMap["active"] != true {
		t.Fatalf("expected active=true, got %v", resultMap["active"])
	}
	if resultMap["score"] != 3.14 {
		t.Fatalf("expected score=3.14, got %v", resultMap["score"])
	}
	if resultMap["empty"] != nil {
		t.Fatalf("expected empty=nil, got %v", resultMap["empty"])
	}
}

func TestInvalidBodyTemplateFailsAtLoadTime(t *testing.T) {
	_, err := New(Options{SpecData: []byte(`
openapi: 3.1.0
info:
  title: Test
  version: 1.0.0
paths:
  /bad:
    get:
      x-lab-mock:
        body:
          broken: "{{ .Unclosed"
`)})
	if err == nil {
		t.Fatal("expected error for invalid body template at load time")
	}
}
