// Package skilllink prepares a read-only shared skill store for a guest agent.
package skilllink

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

const maxEntries = 10000

var skillName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// Target returns the agent-specific path relative to the agent's home.
func Target(agent string) (string, bool) {
	switch agent {
	case "codex", "devin", "kimi", "dsh":
		return ".agents/skills", true
	case "claude":
		return ".claude/skills", true
	case "opencode":
		return ".config/opencode/skills", true
	case "copilot":
		return ".copilot/skills", true
	case "cursor":
		return ".cursor/skills", true
	case "droid":
		return ".factory/skills", true
	default:
		return "", false
	}
}

// Result describes links added or already present and conflicts left untouched.
type Result struct {
	Linked    int
	Existing  int
	Conflicts int
}

// Link creates links to direct directories in store within the agent's home.
// Existing files and links are never overwritten or deleted. store must be the
// trusted, recursively read-only mount selected by the host; this function
// does not mount it or verify mount flags. It must run as the guest agent user.
func Link(ctx context.Context, agent, home, store string) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	target, ok := Target(agent)
	if !ok {
		return Result{}, errors.New("unsupported agent")
	}
	if !filepath.IsAbs(home) || !filepath.IsAbs(store) || filepath.Clean(home) != home || filepath.Clean(store) != store || home == "/" || store == "/" {
		return Result{}, errors.New("home and store must be clean absolute non-root paths")
	}
	storeRoot, err := os.OpenRoot(store)
	if err != nil {
		return Result{}, fmt.Errorf("open skills store: %w", err)
	}
	defer func() { _ = storeRoot.Close() }()
	entries, err := storeEntries(storeRoot)
	if err != nil {
		return Result{}, err
	}
	homeRoot, err := os.OpenRoot(home)
	if err != nil {
		return Result{}, fmt.Errorf("open agent home: %w", err)
	}
	defer func() { _ = homeRoot.Close() }()
	if err := homeRoot.MkdirAll(filepath.FromSlash(target), 0o700); err != nil {
		return Result{}, fmt.Errorf("create agent skills directory: %w", err)
	}
	destination, err := homeRoot.OpenRoot(filepath.FromSlash(target))
	if err != nil {
		return Result{}, fmt.Errorf("open agent skills directory: %w", err)
	}
	defer func() { _ = destination.Close() }()
	return linkStoreEntries(ctx, destination, store, entries)
}

func linkStoreEntries(ctx context.Context, destination *os.Root, store string, entries []os.DirEntry) (Result, error) {
	result, err := linkEntries(ctx, destination, store, entries)
	if err != nil {
		return result, err
	}
	return result, removeStaleLinks(destination, store, entries)
}

func storeEntries(root *os.Root) ([]os.DirEntry, error) {
	dir, err := root.Open(".")
	if err != nil {
		return nil, fmt.Errorf("open skills store directory: %w", err)
	}
	entries, readErr := dir.ReadDir(maxEntries + 1)
	closeErr := dir.Close()
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return nil, fmt.Errorf("read skills store: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close skills store: %w", closeErr)
	}
	if len(entries) > maxEntries {
		return nil, errors.New("skills store exceeds entry limit")
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	return entries, nil
}

func linkEntries(ctx context.Context, destination *os.Root, store string, entries []os.DirEntry) (Result, error) {
	var result Result
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		name := entry.Name()
		if !skillName.MatchString(name) || !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		if err := addLink(destination, name, filepath.Join(store, name), &result); err != nil {
			return result, err
		}
	}
	return result, nil
}

func addLink(root *os.Root, name, target string, result *Result) error {
	err := root.Symlink(target, name)
	if err == nil {
		result.Linked++
		return nil
	}
	if !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("create skill link: %w", err)
	}
	current, readErr := root.Readlink(name)
	if readErr == nil && current == target {
		result.Existing++
	} else {
		result.Conflicts++
	}
	return nil
}

// removeStaleLinks removes only links whose target is inside the store and
// whose entry disappeared. Agent-owned files and foreign links are untouched.
func removeStaleLinks(destination *os.Root, store string, entries []os.DirEntry) error {
	current := make(map[string]bool, len(entries))
	for _, entry := range entries {
		current[entry.Name()] = true
	}
	existing, err := destination.Open(".")
	if err != nil {
		return fmt.Errorf("open agent skills directory: %w", err)
	}
	children, readErr := existing.ReadDir(maxEntries + 1)
	closeErr := existing.Close()
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return fmt.Errorf("read agent skills directory: %w", readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close agent skills directory: %w", closeErr)
	}
	if len(children) > maxEntries {
		return errors.New("agent skills directory exceeds entry limit")
	}
	for _, child := range children {
		if err := removeStaleLink(destination, store, child, current); err != nil {
			return err
		}
	}
	return nil
}

func removeStaleLink(destination *os.Root, store string, child os.DirEntry, current map[string]bool) error {
	if current[child.Name()] || child.Type()&os.ModeSymlink == 0 {
		return nil
	}
	target, err := destination.Readlink(child.Name())
	if err != nil {
		return nil //nolint:nilerr // An unreadable entry is not ours to remove.
	}
	if target != filepath.Join(store, child.Name()) {
		return nil
	}
	if err := destination.Remove(child.Name()); err != nil {
		return fmt.Errorf("remove stale skill link: %w", err)
	}
	return nil
}
