package templates

import (
	"bytes"
	"encoding/json"
	"html/template"
	"testing"
)

// This test ensures our template helper used by submissions page
// correctly handles json.RawMessage (the type of Submission.SubmittedData).
func TestUnmarshalJSON_AcceptsRawMessage(t *testing.T) {
	tm := &TemplateManager{}
	fm := tm.templateFuncMap()

	fn, ok := fm["unmarshalJSON"].(func(interface{}) (map[string]interface{}, error))
	if !ok {
		t.Fatalf("unmarshalJSON func not found or wrong signature")
	}

	payload := json.RawMessage(`{"name":"Alice","email":"alice@example.com"}`)
	m, err := fn(payload)
	if err != nil {
		t.Fatalf("unmarshalJSON returned error: %v", err)
	}

	if got, want := m["name"], "Alice"; got != want {
		t.Fatalf("unexpected parsed value for name: got=%v want=%v", got, want)
	}
}

// This test executes a tiny template that uses the helper in a pipeline,
// mirroring how the submissions template uses it: {{$data := .SubmittedData | unmarshalJSON}}
func TestUnmarshalJSON_InTemplatePipeline(t *testing.T) {
	tm := &TemplateManager{}
	fm := tm.templateFuncMap()

	tmpl := template.Must(template.New("t").Funcs(fm).Parse(`{{$m := . | unmarshalJSON}}{{index $m "name"}}`))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, json.RawMessage(`{"name":"Bob"}`)); err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	if got, want := buf.String(), "Bob"; got != want {
		t.Fatalf("unexpected template output: got=%q want=%q", got, want)
	}
}

// Sanity check: strings are also accepted by the helper.
func TestUnmarshalJSON_AcceptsString(t *testing.T) {
	tm := &TemplateManager{}
	fm := tm.templateFuncMap()

	fn := fm["unmarshalJSON"].(func(interface{}) (map[string]interface{}, error))
	m, err := fn(`{"ok":true}`)
	if err != nil {
		t.Fatalf("unmarshalJSON returned error for string input: %v", err)
	}
	if got, ok := m["ok"].(bool); !ok || !got {
		t.Fatalf("unexpected parsed value for ok: got=%v", m["ok"])
	}
}
