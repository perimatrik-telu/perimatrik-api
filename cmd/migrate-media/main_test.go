package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDryRunOnlyListsMedia(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "articles"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"articles/a.png", "articles/.gitignore", ".env"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("local backup"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"--source", root}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Would upload articles/a.png") || !strings.Contains(out.String(), "Dry run: 1 files") || strings.Contains(out.String(), ".env") {
		t.Fatalf("unexpected dry run: %s", out.String())
	}
	if err := os.Symlink(filepath.Join(root, ".env"), filepath.Join(root, "articles", "secret.png")); err != nil {
		t.Fatal(err)
	}
	if err := run(context.Background(), []string{"--source", root}, &out); err == nil {
		t.Fatal("symlink accepted")
	}
}
