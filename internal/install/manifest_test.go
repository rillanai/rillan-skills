// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// versionedFixture builds an in-memory "go" pack at an arbitrary version so
// tests can drive the upgrade-decision matrix without hard-coding numbers.
func versionedFixture(version string) fstest.MapFS {
	router := []byte("---\nname: go\ndescription: test\n---\n\n<!-- version: " + version + " -->\n# go\n")
	mode := []byte("<!-- version: " + version + " -->\n# dev\nbody\n")
	return fstest.MapFS{
		"skills/go/SKILL.md": {Data: router},
		"skills/go/dev.md":   {Data: mode},
	}
}

func TestManifestRoundTrip(t *testing.T) {
	root := t.TempDir()
	m := &Manifest{Global: true}
	m.upsert(Entry{Tool: "claude", Pack: "go", Version: "3.0.0", Files: []string{".claude/skills/go/SKILL.md"}})
	m.upsert(Entry{Tool: "codex", Pack: "security", Version: "1.2.3", Files: []string{".codex/skills/security/SKILL.md"}})
	if err := m.Save(root); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(ManifestPath(root)); err != nil {
		t.Fatalf("manifest not written: %v", err)
	}

	got, err := LoadManifest(root)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if got.Schema != manifestSchema || !got.Global {
		t.Errorf("schema/global not round-tripped: %+v", got)
	}
	if v := got.Recorded("claude", "go"); v != "3.0.0" {
		t.Errorf("recorded claude/go = %q, want 3.0.0", v)
	}
	if v := got.Recorded("codex", "security"); v != "1.2.3" {
		t.Errorf("recorded codex/security = %q, want 1.2.3", v)
	}
	if v := got.Recorded("claude", "absent"); v != "" {
		t.Errorf("recorded missing entry = %q, want empty", v)
	}
}

func TestLoadManifestMissingIsEmpty(t *testing.T) {
	m, err := LoadManifest(t.TempDir())
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(m.Entries) != 0 {
		t.Errorf("want empty manifest, got %d entries", len(m.Entries))
	}
}

func TestClassifyMatrix(t *testing.T) {
	cases := []struct {
		recorded, bundled string
		want              changeKind
	}{
		{"", "3.0.0", changeFresh},
		{"3.0.0", "3.0.0", changeSame},
		{"3.0.0", "3.1.0", changeUpgrade},
		{"3.0.0", "2.9.0", changeDown},
		{"3.0.0", "3.0.1", changeUpgrade},
		{"3.10.0", "3.9.0", changeDown}, // numeric, not lexical
	}
	for _, c := range cases {
		if got := classify(c.recorded, c.bundled); got != c.want {
			t.Errorf("classify(%q,%q)=%v want %v", c.recorded, c.bundled, got, c.want)
		}
	}
}

func TestInstallWritesManifest(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(versionedFixture("3.0.0"), Options{
		Target: dir, Tools: []Tool{ToolClaude}, Packs: []string{"go"},
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if v := m.Recorded("claude", "go"); v != "3.0.0" {
		t.Fatalf("manifest version = %q, want 3.0.0", v)
	}
	idx := m.find("claude", "go")
	files := m.Entries[idx].Files
	if len(files) != 2 {
		t.Errorf("want 2 recorded files, got %v", files)
	}
}

func TestInstallUpgradeDecision(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Target: dir, Tools: []Tool{ToolClaude}, Packs: []string{"go"}}

	// Fresh install.
	if n, err := Run(versionedFixture("3.0.0"), opts); err != nil || n != 1 {
		t.Fatalf("fresh: n=%d err=%v", n, err)
	}
	// Equal version → skip.
	if n, err := Run(versionedFixture("3.0.0"), opts); err != nil || n != 0 {
		t.Fatalf("equal: n=%d err=%v want 0", n, err)
	}
	// Newer bundled → upgrade, with an "upgraded" status line.
	var logs []string
	up := opts
	up.Logger = func(format string, a ...any) { logs = append(logs, fmtLine(format, a...)) }
	if n, err := Run(versionedFixture("3.1.0"), up); err != nil || n != 1 {
		t.Fatalf("upgrade: n=%d err=%v want 1", n, err)
	}
	var sawUpgrade bool
	for _, l := range logs {
		if strings.Contains(l, "upgraded 3.0.0->3.1.0") {
			sawUpgrade = true
		}
	}
	if !sawUpgrade {
		t.Errorf("want an 'upgraded 3.0.0->3.1.0' status, got %v", logs)
	}
	m, _ := LoadManifest(dir)
	if v := m.Recorded("claude", "go"); v != "3.1.0" {
		t.Errorf("manifest after upgrade = %q, want 3.1.0", v)
	}
}

func TestUninstallUsesManifest(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Target: dir, Tools: []Tool{ToolClaude}, Packs: []string{"go"}}
	if _, err := Run(versionedFixture("3.0.0"), opts); err != nil {
		t.Fatalf("install: %v", err)
	}
	// A user file that the installer did NOT record must survive uninstall.
	userFile := filepath.Join(dir, ".claude", "skills", "mine", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(userFile), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(userFile, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}

	n, err := Uninstall(UninstallOptions{
		Target: dir, Tools: []Tool{ToolClaude}, KnownPacks: []string{"go"},
	})
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if n != 1 {
		t.Errorf("want 1 pack removed, got %d", n)
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "skills", "go")); !os.IsNotExist(err) {
		t.Errorf("go pack should be gone: %v", err)
	}
	if _, err := os.Stat(userFile); err != nil {
		t.Errorf("unrelated user skill must be preserved: %v", err)
	}
	// Manifest entry should be cleared.
	m, _ := LoadManifest(dir)
	if v := m.Recorded("claude", "go"); v != "" {
		t.Errorf("manifest entry not cleared: %q", v)
	}
}

func TestUninstallDryRunAndUnsupported(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(versionedFixture("3.0.0"), Options{
		Target: dir, Tools: []Tool{ToolClaude}, Packs: []string{"go"},
	}); err != nil {
		t.Fatalf("install: %v", err)
	}
	var logs []string
	n, err := Uninstall(UninstallOptions{
		Target: dir, Tools: []Tool{ToolClaude, ToolGrok}, KnownPacks: []string{"go"},
		DryRun: true,
		Logger: func(format string, a ...any) { logs = append(logs, fmtLine(format, a...)) },
	})
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if n != 1 {
		t.Errorf("dry-run: want 1 reported, got %d", n)
	}
	// Dry-run must not delete.
	if _, err := os.Stat(filepath.Join(dir, ".claude", "skills", "go")); err != nil {
		t.Errorf("dry-run removed files: %v", err)
	}
	var sawWould, sawGrokSkip bool
	for _, l := range logs {
		if strings.Contains(l, "would remove") {
			sawWould = true
		}
		if strings.Contains(strings.ToLower(l), "grok") {
			sawGrokSkip = true
		}
	}
	if !sawWould {
		t.Errorf("want a 'would remove' line, got %v", logs)
	}
	if !sawGrokSkip {
		t.Errorf("want a grok skip line, got %v", logs)
	}
}

func TestUninstallGlobal(t *testing.T) {
	home := t.TempDir()
	if _, err := Run(versionedFixture("3.0.0"), Options{
		Global: true, Home: home, Tools: []Tool{ToolClaude, ToolOpenCode}, Packs: []string{"go"},
	}); err != nil {
		t.Fatalf("install: %v", err)
	}
	n, err := Uninstall(UninstallOptions{
		Global: true, Home: home, Tools: []Tool{ToolClaude, ToolOpenCode}, KnownPacks: []string{"go"},
	})
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if n != 2 {
		t.Errorf("want 2 packs removed (2 tools), got %d", n)
	}
	for _, rel := range []string{".claude/skills/go", ".config/opencode/agent/go.md"} {
		if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Errorf("%s should be removed: %v", rel, err)
		}
	}
}

func TestLoadManifestRejectsUnknownSchema(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(ManifestPath(root)), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ManifestPath(root), []byte(`{"schema":99,"entries":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(root); err == nil {
		t.Fatal("want error for unsupported schema, got nil")
	}
}

func TestInstallDowngradeSkippedWithoutForce(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Target: dir, Tools: []Tool{ToolClaude}, Packs: []string{"go"}}

	// Install a newer version first.
	if _, err := Run(versionedFixture("3.5.0"), opts); err != nil {
		t.Fatalf("install newer: %v", err)
	}
	// A bundle with an older version must be skipped by default.
	var logs []string
	down := opts
	down.Logger = func(format string, a ...any) { logs = append(logs, fmtLine(format, a...)) }
	n, err := Run(versionedFixture("3.0.0"), down)
	if err != nil {
		t.Fatalf("downgrade run: %v", err)
	}
	if n != 0 {
		t.Errorf("downgrade without --force: want 0 written, got %d", n)
	}
	var warned bool
	for _, l := range logs {
		if strings.Contains(l, "newer than bundled") {
			warned = true
		}
	}
	if !warned {
		t.Errorf("want a downgrade-skip warning, got %v", logs)
	}
	if v := mustManifest(t, dir).Recorded("claude", "go"); v != "3.5.0" {
		t.Errorf("manifest must stay at 3.5.0, got %q", v)
	}

	// With --force the downgrade is applied.
	forced := opts
	forced.Force = true
	if n, err := Run(versionedFixture("3.0.0"), forced); err != nil || n != 1 {
		t.Fatalf("forced downgrade: n=%d err=%v want 1", n, err)
	}
	if v := mustManifest(t, dir).Recorded("claude", "go"); v != "3.0.0" {
		t.Errorf("manifest after forced downgrade: want 3.0.0, got %q", v)
	}
}

func TestUninstallEmptyTargetErrors(t *testing.T) {
	if _, err := Uninstall(UninstallOptions{Target: "", Tools: []Tool{ToolClaude}}); err == nil {
		t.Fatal("want error for empty target in project mode, got nil")
	}
}

func TestUninstallRejectsOutOfScopePaths(t *testing.T) {
	root := t.TempDir()
	// A sentinel file outside the scope root that a tampered manifest tries to
	// delete via traversal must survive.
	outside := filepath.Join(filepath.Dir(root), "victim-"+filepath.Base(root)+".txt")
	if err := os.WriteFile(outside, []byte("do not delete"), 0o600); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(outside) }()

	// Hand-craft a manifest with malicious file entries.
	m := &Manifest{}
	m.upsert(Entry{Tool: "claude", Pack: "go", Files: []string{
		"../" + filepath.Base(outside), // relative traversal escape
		outside,                        // absolute path
	}})
	if err := m.Save(root); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var logs []string
	if _, err := Uninstall(UninstallOptions{
		Target: root, Tools: []Tool{ToolClaude}, KnownPacks: []string{"go"},
		Logger: func(format string, a ...any) { logs = append(logs, fmtLine(format, a...)) },
	}); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if _, err := os.Stat(outside); err != nil {
		t.Errorf("out-of-scope file must be preserved: %v", err)
	}
	var refused bool
	for _, l := range logs {
		if strings.Contains(l, "out-of-scope") {
			refused = true
		}
	}
	if !refused {
		t.Errorf("want an out-of-scope refusal log, got %v", logs)
	}
}

func mustManifest(t *testing.T, root string) *Manifest {
	t.Helper()
	m, err := LoadManifest(root)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	return m
}
