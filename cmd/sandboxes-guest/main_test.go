package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinkSkillsCommand(t *testing.T) {
	home, store := t.TempDir(), t.TempDir()
	if err := os.Mkdir(filepath.Join(store, "alpha"), 0o700); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	code := run(context.Background(), []string{"link-skills", "--agent", "codex", "--home", home, "--store", store}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), "linked=1") {
		t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
	}
	if _, err := os.Lstat(filepath.Join(home, ".agents", "skills", "alpha")); err != nil {
		t.Fatal(err)
	}
}
