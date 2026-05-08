package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConvertFile(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.jsonc")
	src := `{
  // a comment
  "a": 1,
  "b": [1, 2, 3,],  // trailing comma + line comment
  /* block comment */
  "c": "hi",
}
`
	if err := os.WriteFile(in, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := convert(in, strings.NewReader(""), &buf); err != nil {
		t.Fatal(err)
	}

	var v map[string]any
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput:\n%s", err, buf.String())
	}
	if v["a"] != float64(1) {
		t.Errorf("a = %v, want 1", v["a"])
	}
	if v["c"] != "hi" {
		t.Errorf("c = %v, want hi", v["c"])
	}
	b, ok := v["b"].([]any)
	if !ok || len(b) != 3 {
		t.Errorf("b = %v, want [1 2 3]", v["b"])
	}
}

func TestConvertStdin(t *testing.T) {
	src := `{ "x": 1, /* hi */ "y": [1, 2,], }`
	var buf bytes.Buffer
	if err := convert("-", strings.NewReader(src), &buf); err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if v["x"] != float64(1) {
		t.Errorf("x = %v, want 1", v["x"])
	}
}

func TestConvertParseError(t *testing.T) {
	var buf bytes.Buffer
	err := convert("-", strings.NewReader("not json"), &buf)
	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
}

func TestConvertAppendsTrailingNewline(t *testing.T) {
	// Source with no trailing newline — convert must still add one so
	// successive documents in multi-file mode aren't run together.
	var buf bytes.Buffer
	if err := convert("-", strings.NewReader(`{"a":1}`), &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.Bytes()
	if len(out) == 0 || out[len(out)-1] != '\n' {
		t.Fatalf("expected trailing newline, got %q", out)
	}
}

func TestConvertMultipleFilesConcatenated(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.jsonc")
	b := filepath.Join(dir, "b.jsonc")
	if err := os.WriteFile(a, []byte(`{"a":1,}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte(`{"b":2,}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	for _, p := range []string{a, b} {
		if err := convert(p, strings.NewReader(""), &buf); err != nil {
			t.Fatal(err)
		}
	}

	// Output must be two JSON values separated by a newline (a stream,
	// not a single document — see usage docs).
	dec := json.NewDecoder(&buf)
	var first, second map[string]any
	if err := dec.Decode(&first); err != nil {
		t.Fatalf("decode first doc: %v", err)
	}
	if err := dec.Decode(&second); err != nil {
		t.Fatalf("decode second doc: %v", err)
	}
	if first["a"] != float64(1) || second["b"] != float64(2) {
		t.Errorf("got first=%v second=%v", first, second)
	}
}
