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
