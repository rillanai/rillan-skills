// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"fmt"
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
		base, ok := SkillsDir(tool, tgt, "", false)
		if !ok {
			t.Fatalf("%s: SkillsDir returned not-ok", tool)
		}
		if got := routerDest(tool, base, "go"); got != want {
			t.Errorf("%s: got %q want %q", tool, got, want)
		}
	}
}

func TestSkillsDirProject(t *testing.T) {
	tgt := filepath.FromSlash("/repo")
	cases := map[Tool]string{
		ToolClaude:   filepath.FromSlash("/repo/.claude/skills"),
		ToolCodex:    filepath.FromSlash("/repo/.codex/skills"),
		ToolOpenCode: filepath.FromSlash("/repo/.opencode/agents"),
	}
	for tool, want := range cases {
		got, ok := SkillsDir(tool, tgt, "", false)
		if !ok || got != want {
			t.Errorf("%s: got %q ok=%v, want %q", tool, got, ok, want)
		}
	}
}

func TestSkillsDirGlobal(t *testing.T) {
	home := filepath.FromSlash("/home/u")
	cases := map[Tool]string{
		ToolClaude:   filepath.FromSlash("/home/u/.claude/skills"),
		ToolCodex:    filepath.FromSlash("/home/u/.codex/skills"),
		ToolOpenCode: filepath.FromSlash("/home/u/.config/opencode/agent"),
	}
	for tool, want := range cases {
		got, ok := SkillsDir(tool, "", home, true)
		if !ok || got != want {
			t.Errorf("%s: got %q ok=%v, want %q", tool, got, ok, want)
		}
	}
}

func TestSkillsDirUnsupportedTool(t *testing.T) {
	for _, global := range []bool{false, true} {
		if got, ok := SkillsDir(ToolGrok, "/repo", "/home/u", global); ok {
			t.Errorf("grok (global=%v): want unsupported, got %q ok=%v", global, got, ok)
		}
		if got, ok := SkillsDir(Tool("nope"), "/repo", "/home/u", global); ok {
			t.Errorf("unknown tool (global=%v): want unsupported, got %q ok=%v", global, got, ok)
		}
	}
	if hint := ManualHint(ToolGrok); hint == "" {
		t.Error("ManualHint(grok): want non-empty guidance")
	}
	if hint := ManualHint(ToolClaude); hint != "" {
		t.Errorf("ManualHint(claude): want empty, got %q", hint)
	}
}

func TestRunGlobalWritesUnderHome(t *testing.T) {
	home := t.TempDir()
	n, err := Run(fixture(), Options{
		Global: true, Home: home,
		Tools: []Tool{ToolClaude, ToolCodex, ToolOpenCode},
		Packs: []string{"go", "security"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 2 packs x 3 tools.
	if n != 6 {
		t.Errorf("want 6 packs written, got %d", n)
	}
	for _, rel := range []string{
		".claude/skills/go/SKILL.md",
		".claude/skills/go/policy.md",
		".codex/skills/security/SKILL.md",
		".config/opencode/agent/go.md",
	} {
		if _, err := os.Stat(filepath.Join(home, filepath.FromSlash(rel))); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
}

func TestRunGlobalSkipsUnsupportedTool(t *testing.T) {
	home := t.TempDir()
	var logs []string
	n, err := Run(fixture(), Options{
		Global: true, Home: home,
		Tools:  []Tool{ToolGrok},
		Packs:  []string{"go"},
		Logger: func(format string, _ ...any) { logs = append(logs, format) },
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 0 {
		t.Errorf("grok: want 0 written (skipped), got %d", n)
	}
	// Nothing should have been written under home.
	if entries, _ := os.ReadDir(home); len(entries) != 0 {
		t.Errorf("grok: want empty home, got %d entries", len(entries))
	}
	var warned bool
	for _, l := range logs {
		if strings.Contains(l, "skipped") {
			warned = true
		}
	}
	if !warned {
		t.Error("grok: want a skip warning logged")
	}
}

func TestRunGlobalDryRun(t *testing.T) {
	home := t.TempDir()
	var logs []string
	n, err := Run(fixture(), Options{
		Global: true, Home: home, DryRun: true,
		Tools:  []Tool{ToolClaude},
		Packs:  []string{"go"},
		Logger: func(format string, a ...any) { logs = append(logs, fmtLine(format, a...)) },
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 1 {
		t.Errorf("want 1 pack in dry-run count, got %d", n)
	}
	// Dry-run must not write anything.
	if _, err := os.Stat(filepath.Join(home, ".claude")); !os.IsNotExist(err) {
		t.Errorf("dry-run wrote files: %v", err)
	}
	var sawTarget bool
	for _, l := range logs {
		if strings.Contains(l, "dry-run") && strings.Contains(l, filepath.FromSlash(".claude/skills/go/SKILL.md")) {
			sawTarget = true
		}
	}
	if !sawTarget {
		t.Errorf("dry-run: want a log line naming the global target, got %v", logs)
	}
}

func fmtLine(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
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
