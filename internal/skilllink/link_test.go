package skilllink

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLinkReadOnlyStorePreservesAgentFiles(t *testing.T) {
	home, store := t.TempDir(), t.TempDir()
	for _, name := range []string{"alpha", "beta"} {
		if err := os.Mkdir(filepath.Join(store, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(store, "alpha", "SKILL.md"), []byte("synthetic"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".agents", "skills", "beta"), 0o700); err != nil {
		t.Fatal(err)
	}
	result, err := Link(context.Background(), "codex", home, store)
	if err != nil || result.Linked != 1 || result.Conflicts != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".agents", "skills", "alpha", "SKILL.md"))
	if err != nil || string(data) != "synthetic" {
		t.Fatalf("linked skill unavailable: %q %v", data, err)
	}
	if info, err := os.Lstat(filepath.Join(home, ".agents", "skills", "beta")); err != nil || !info.IsDir() {
		t.Fatalf("existing agent directory changed: %v %v", info, err)
	}
	result, err = Link(context.Background(), "codex", home, store)
	if err != nil || result.Existing != 1 || result.Conflicts != 1 {
		t.Fatalf("repeat result=%+v err=%v", result, err)
	}
}

func TestLinkSkipsUnsafeStoreEntries(t *testing.T) {
	home, store, external := t.TempDir(), t.TempDir(), t.TempDir()
	if err := os.Symlink(external, filepath.Join(store, "escape")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if err := os.Mkdir(filepath.Join(store, ".hidden"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store, "file"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := Link(context.Background(), "claude", home, store)
	if err != nil || result.Linked != 0 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if _, err := os.Lstat(filepath.Join(home, ".claude", "skills", "escape")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("store symlink was exposed: %v", err)
	}
}

func TestLinkRejectsHomeEscape(t *testing.T) {
	home, store, external := t.TempDir(), t.TempDir(), t.TempDir()
	if err := os.Mkdir(filepath.Join(store, "alpha"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(home, ".agents")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := Link(context.Background(), "codex", home, store); err == nil {
		t.Fatal("expected home escape to fail")
	}
	if _, err := os.Lstat(filepath.Join(external, "skills")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("wrote outside home: %v", err)
	}
}

func TestLinkDoesNotReplaceExistingSymlink(t *testing.T) {
	home, store, external := t.TempDir(), t.TempDir(), t.TempDir()
	if err := os.Mkdir(filepath.Join(store, "alpha"), 0o700); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, ".agents", "skills")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(dir, "alpha")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	result, err := Link(context.Background(), "codex", home, store)
	if err != nil || result.Conflicts != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if target, err := os.Readlink(filepath.Join(dir, "alpha")); err != nil || target != external {
		t.Fatalf("existing link changed: %q, %v", target, err)
	}
}

func TestTargetMappingAndCancellation(t *testing.T) {
	for agent, want := range map[string]string{"codex": ".agents/skills", "devin": ".agents/skills", "claude": ".claude/skills", "opencode": ".config/opencode/skills", "copilot": ".copilot/skills", "cursor": ".cursor/skills", "droid": ".factory/skills"} {
		if got, ok := Target(agent); !ok || got != want {
			t.Fatalf("%s: got %s %t", agent, got, ok)
		}
	}
	if _, ok := Target("shell"); ok {
		t.Fatal("shell must not receive an agent store")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	home, store := t.TempDir(), t.TempDir()
	if err := os.Mkdir(filepath.Join(store, "alpha"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Link(ctx, "codex", home, store); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
