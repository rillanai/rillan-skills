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

// fixture builds an in-memory skill tree: one multi-mode pack ("go" with a
// router + two modes) and one single-mode pack ("security" with only a router).
func fixture() fstest.MapFS {
	router := func(name string) []byte {
		return []byte("---\nname: " + name + "\ndescription: test\n---\n\n<!-- version: 3.0.0 -->\n# " + name + "\n")
	}
	mode := func(title string) []byte {
		return []byte("<!-- version: 3.0.0 -->\n# " + title + "\nbody\n")
	}
	return fstest.MapFS{
		"skills/go/SKILL.md":       {Data: router("go")},
		"skills/go/policy.md":      {Data: mode("policy")},
		"skills/go/dev.md":         {Data: mode("dev")},
		"skills/security/SKILL.md": {Data: router("security")},
	}
}

func TestLoadPacks(t *testing.T) {
	packs, err := loadPacks(fixture(), []string{"go", "security", "absent"})
	if err != nil {
		t.Fatalf("loadPacks: %v", err)
	}
	if len(packs) != 2 {
		t.Fatalf("want 2 packs (absent skipped), got %d", len(packs))
	}
	gp := packs[0]
	if gp.Name != "go" || gp.Version != "3.0.0" {
		t.Errorf("go pack: name=%q version=%q", gp.Name, gp.Version)
	}
	if len(gp.Modes) != 2 {
		t.Fatalf("go pack: want 2 modes, got %d", len(gp.Modes))
	}
	// Modes must be sorted for stable output.
	if gp.Modes[0].Name != "dev.md" || gp.Modes[1].Name != "policy.md" {
		t.Errorf("modes not sorted: %q, %q", gp.Modes[0].Name, gp.Modes[1].Name)
	}
	if len(packs[1].Modes) != 0 {
		t.Errorf("security pack: want 0 modes, got %d", len(packs[1].Modes))
	}
}

func TestLoadPacksMissingRouter(t *testing.T) {
	mfs := fstest.MapFS{"skills/broken/dev.md": {Data: []byte("# dev")}}
	if _, err := loadPacks(mfs, []string{"broken"}); err == nil {
		t.Fatal("want error for pack with no SKILL.md, got nil")
	}
}

func TestRouterDest(t *testing.T) {
	tgt := filepath.FromSlash("/repo")
	cases := map[Tool]string{
		ToolClaude:   filepath.FromSlash("/repo/.claude/skills/go/SKILL.md"),
		ToolCodex:    filepath.FromSlash("/repo/.codex/skills/go/SKILL.md"),
		ToolOpenCode: filepath.FromSlash("/repo/.opencode/agents/go.md"),
	}
	for tool, want := range cases {
		if got := routerDest(tool, tgt, "go"); got != want {
			t.Errorf("%s: got %q want %q", tool, got, want)
		}
	}
}

func TestRunClaudeWritesTree(t *testing.T) {
	dir := t.TempDir()
	n, err := Run(fixture(), Options{Target: dir, Tools: []Tool{ToolClaude}, Packs: []string{"go", "security"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 2 {
		t.Errorf("want 2 packs written, got %d", n)
	}
	for _, rel := range []string{
		".claude/skills/go/SKILL.md",
		".claude/skills/go/policy.md",
		".claude/skills/go/dev.md",
		".claude/skills/security/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
}

func TestRunOpenCodeFlattens(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(fixture(), Options{Target: dir, Tools: []Tool{ToolOpenCode}, Packs: []string{"go"}}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	agent := filepath.Join(dir, filepath.FromSlash(".opencode/agents/go.md"))
	b, err := os.ReadFile(agent) //nolint:gosec // test-controlled path
	if err != nil {
		t.Fatalf("read agent: %v", err)
	}
	s := string(b)
	// Single self-contained file: router frontmatter first, then each mode body.
	if !strings.HasPrefix(s, "---\nname: go\n") {
		t.Error("flattened agent should start with router frontmatter")
	}
	for _, want := range []string{"## dev", "## policy"} {
		if !strings.Contains(s, want) {
			t.Errorf("flattened agent missing mode heading %q", want)
		}
	}
	// No separate mode files for OpenCode.
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(".opencode/agents/go"))); err == nil {
		t.Error("OpenCode should not create a pack directory")
	}
}

func TestRunIdempotentAndForce(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Target: dir, Tools: []Tool{ToolClaude}, Packs: []string{"go"}}
	if _, err := Run(fixture(), opts); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	// Same version → skip.
	n, err := Run(fixture(), opts)
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if n != 0 {
		t.Errorf("want 0 packs rewritten at same version, got %d", n)
	}
	// Force → rewrite.
	forced := opts
	forced.Force = true
	n, err = Run(fixture(), forced)
	if err != nil {
		t.Fatalf("forced Run: %v", err)
	}
	if n != 1 {
		t.Errorf("want 1 pack rewritten with --force, got %d", n)
	}
}

func TestRunRejectsBadTarget(t *testing.T) {
	if _, err := Run(fixture(), Options{Target: ""}); err == nil {
		t.Error("want error for empty target")
	}
	f := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(fixture(), Options{Target: f, Packs: []string{"go"}}); err == nil {
		t.Error("want error when target is not a directory")
	}
}
