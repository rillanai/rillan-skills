// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

// Package install writes embedded skill packs into a target repository using
// the layout each tool expects for project-scoped skills.
//
// Each pack is a directory holding a router SKILL.md plus zero or more
// `<mode>.md` files. Tools that support bundled, progressively-disclosed files
// (Claude, Codex) get the directory verbatim; OpenCode, which uses single-file
// agents, gets one self-contained file assembled from the router plus every
// mode body.
package install

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// skillsRoot is the directory inside the embedded FS where skill packs live.
const skillsRoot = "skills"

// routerFile is the per-pack entry point that carries the pack's frontmatter
// and version.
const routerFile = "SKILL.md"

// Tool identifies which assistant to install for.
type Tool string

// Supported tool identifiers.
const (
	ToolClaude   Tool = "claude"
	ToolCodex    Tool = "codex"
	ToolOpenCode Tool = "opencode"
	// ToolGrok is recognized so users can request it explicitly, but it has no
	// well-defined skills directory. The installer skips it with a clear
	// warning and a manual-copy hint rather than failing the whole run.
	ToolGrok Tool = "grok"
)

// AllTools returns every tool the installer knows about, in stable order.
// Tools without a resolvable skills directory (e.g. Grok) are included so
// "--tool all" can report them as skipped rather than silently omitting them.
func AllTools() []Tool {
	return []Tool{ToolClaude, ToolCodex, ToolOpenCode, ToolGrok}
}

// Options control an install run.
type Options struct {
	Target string // target repo dir (project install); ignored when Global is set
	Tools  []Tool
	Packs  []string
	DryRun bool
	Force  bool
	// Global installs into each tool's user-level config directory under Home
	// instead of into a per-repository target.
	Global bool
	// Home is the user's home directory used to resolve global paths. Empty
	// means resolve via os.UserHomeDir() (overridable mainly for tests).
	Home   string
	Logger func(format string, args ...any)
}

// SkillsDir returns the directory under which a tool's packs are written, and
// whether the tool is supported in the requested mode. For directory-based
// tools (Claude, Codex) this is the parent of the per-pack directories; for
// OpenCode it is the directory that holds the single-file agents.
//
// When global is true, target is ignored and paths are resolved under home;
// otherwise they are resolved under target. Tools with no resolvable directory
// (e.g. Grok) return ("", false).
func SkillsDir(t Tool, target, home string, global bool) (string, bool) {
	if global {
		switch t {
		case ToolClaude:
			return filepath.Join(home, ".claude", "skills"), true
		case ToolCodex:
			return filepath.Join(home, ".codex", "skills"), true
		case ToolOpenCode:
			// OpenCode loads user-level agents from the XDG config dir.
			return filepath.Join(home, ".config", "opencode", "agent"), true
		default:
			return "", false
		}
	}
	switch t {
	case ToolClaude:
		return filepath.Join(target, ".claude", "skills"), true
	case ToolCodex:
		return filepath.Join(target, ".codex", "skills"), true
	case ToolOpenCode:
		return filepath.Join(target, ".opencode", "agents"), true
	default:
		return "", false
	}
}

// ManualHint returns guidance for tools the installer cannot target directly,
// or "" when the tool is supported.
func ManualHint(t Tool) string {
	if t == ToolGrok {
		return "Grok has no standard skills directory — copy the desired packs from the rillan-skills 'skills/<pack>/' tree into your Grok prompt/config manually."
	}
	return ""
}

// ModeFile is one bundled `<mode>.md` reference file within a pack.
type ModeFile struct {
	Name     string // e.g. "policy.md"
	Contents []byte
}

// Pack is one resolved skill pack ready to be installed: a router plus its
// bundled mode files.
type Pack struct {
	Name    string // directory name, e.g. "go"
	Version string // from the router's version comment, "" if absent
	Router  []byte // contents of SKILL.md
	Modes   []ModeFile
}

// Run installs the requested packs into the target for each requested tool. It
// returns the number of packs that were written (or would be written, in
// dry-run mode).
func Run(efs fs.FS, opts Options) (int, error) {
	if opts.Logger == nil {
		opts.Logger = func(string, ...any) {}
	}
	if len(opts.Tools) == 0 {
		opts.Tools = []Tool{ToolClaude}
	}
	if opts.Global {
		if opts.Home == "" {
			h, err := os.UserHomeDir()
			if err != nil {
				return 0, fmt.Errorf("install: resolve home directory: %w", err)
			}
			opts.Home = h
		}
	} else {
		if opts.Target == "" {
			return 0, fmt.Errorf("install: target directory is required")
		}
		abs, err := filepath.Abs(opts.Target)
		if err != nil {
			return 0, err
		}
		if st, err := os.Stat(abs); err != nil || !st.IsDir() {
			return 0, fmt.Errorf("install: target %q is not a directory", abs)
		}
		opts.Target = abs
	}

	packs, err := loadPacks(efs, opts.Packs)
	if err != nil {
		return 0, err
	}

	root := opts.Target
	if opts.Global {
		root = opts.Home
	}
	man, err := LoadManifest(root)
	if err != nil {
		return 0, fmt.Errorf("install: read manifest: %w", err)
	}
	man.Global = opts.Global

	count := 0
	for _, t := range opts.Tools {
		base, ok := SkillsDir(t, opts.Target, opts.Home, opts.Global)
		if !ok {
			if hint := ManualHint(t); hint != "" {
				opts.Logger("[!] %s: skipped — %s", t, hint)
			} else {
				opts.Logger("[!] %s: no skills directory for this tool — skipped", t)
			}
			continue
		}
		for _, p := range packs {
			marker := routerDest(t, base, p.Name)
			// The manifest is the authoritative record; fall back to the
			// installed router's version marker for installs predating it.
			recorded := man.Recorded(string(t), p.Name)
			if recorded == "" {
				if existing, ok := readVersion(marker); ok {
					recorded = existing
				}
			}
			kind := classify(recorded, p.Version)
			if !opts.Force && kind == changeSame && p.Version != "" {
				opts.Logger("[=] %s/%s: up-to-date (v%s)", t, p.Name, p.Version)
				continue
			}
			if !opts.Force && kind == changeDown {
				opts.Logger("[!] %s/%s: installed v%s is newer than bundled v%s — skipping (use --force to downgrade)", t, p.Name, recorded, p.Version)
				continue
			}
			if opts.DryRun {
				opts.Logger("[dry-run] %s/%s %s -> %s", t, p.Name, changeVerb(kind, recorded, p.Version), marker)
				count++
				continue
			}
			if err := writePack(t, base, p); err != nil {
				return count, fmt.Errorf("install %s: %w", p.Name, err)
			}
			man.upsert(Entry{
				Tool:    string(t),
				Pack:    p.Name,
				Version: p.Version,
				Files:   relFiles(root, packFiles(t, base, p)),
			})
			opts.Logger("[+] %s/%s %s -> %s", t, p.Name, changeVerb(kind, recorded, p.Version), marker)
			count++
		}
	}
	if count > 0 && !opts.DryRun {
		if err := man.Save(root); err != nil {
			return count, fmt.Errorf("install: write manifest: %w", err)
		}
	}
	return count, nil
}

// writePack writes a pack under base using the layout the tool expects. base is
// the resolved skills/agents directory for the tool (see SkillsDir).
func writePack(t Tool, base string, p Pack) error {
	if t == ToolOpenCode {
		return writeFile(routerDest(t, base, p.Name), flattenForOpenCode(p))
	}
	// Claude, Codex: directory with SKILL.md + mode files.
	dir := InstalledPath(t, base, p.Name)
	if err := writeFile(filepath.Join(dir, routerFile), p.Router); err != nil {
		return err
	}
	for _, m := range p.Modes {
		if err := writeFile(filepath.Join(dir, m.Name), m.Contents); err != nil {
			return err
		}
	}
	return nil
}

// changeVerb renders a human status for an install action.
func changeVerb(kind changeKind, recorded, bundled string) string {
	switch kind {
	case changeUpgrade:
		return fmt.Sprintf("upgraded %s->%s", recorded, bundled)
	case changeDown:
		return fmt.Sprintf("downgraded %s->%s", recorded, bundled)
	case changeSame:
		return fmt.Sprintf("reinstalled v%s", bundled)
	default:
		return fmt.Sprintf("installed v%s", bundled)
	}
}

// packFiles lists the absolute files a pack writes under base, for the manifest.
func packFiles(t Tool, base string, p Pack) []string {
	if t == ToolOpenCode {
		return []string{routerDest(t, base, p.Name)}
	}
	dir := InstalledPath(t, base, p.Name)
	files := []string{filepath.Join(dir, routerFile)}
	for _, m := range p.Modes {
		files = append(files, filepath.Join(dir, m.Name))
	}
	return files
}

// relFiles converts absolute paths to slash-relative paths under root for
// stable, portable manifest storage.
func relFiles(root string, abs []string) []string {
	out := make([]string, 0, len(abs))
	for _, f := range abs {
		if rel, err := filepath.Rel(root, f); err == nil {
			out = append(out, filepath.ToSlash(rel))
		} else {
			out = append(out, filepath.ToSlash(f))
		}
	}
	return out
}

// InstalledPath returns the on-disk path a pack occupies under base: a per-pack
// directory for directory-based tools (Claude, Codex), or the single agent file
// for OpenCode. It is the unit removed on uninstall.
func InstalledPath(t Tool, base, pack string) string {
	if t == ToolOpenCode {
		return filepath.Join(base, pack+".md")
	}
	return filepath.Join(base, pack)
}

// routerDest returns the file whose version is checked for idempotent installs:
// the router SKILL.md for directory tools, or the single agent file for OpenCode.
func routerDest(t Tool, base, pack string) string {
	if t == ToolOpenCode {
		return InstalledPath(t, base, pack)
	}
	return filepath.Join(InstalledPath(t, base, pack), routerFile)
}

// flattenForOpenCode assembles a pack into a single self-contained agent file:
// the router (frontmatter + body) followed by each mode body under a heading.
// The router's version comment stays first so idempotency still works.
func flattenForOpenCode(p Pack) []byte {
	var b strings.Builder
	b.Write(p.Router)
	for _, m := range p.Modes {
		if b.Len() > 0 && !strings.HasSuffix(b.String(), "\n") {
			b.WriteString("\n")
		}
		b.WriteString("\n## ")
		b.WriteString(strings.TrimSuffix(m.Name, ".md"))
		b.WriteString("\n\n")
		b.Write(m.Contents)
	}
	return []byte(b.String())
}

func writeFile(dest string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil { //nolint:gosec // skills are intentionally world-readable in the user's home
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".rillan-skills-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := io.Copy(tmp, strings.NewReader(string(contents))); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, dest)
}

var versionCommentExpr = regexp.MustCompile(`(?m)^<!--\s*version:\s*([0-9]+\.[0-9]+\.[0-9]+)\s*-->`)

// loadPacks reads each requested pack directory from the embedded FS, taking
// SKILL.md as the router (and version source) and every other `*.md` as a mode.
func loadPacks(efs fs.FS, names []string) ([]Pack, error) {
	var out []Pack
	for _, name := range names {
		dir := path.Join(skillsRoot, name)
		entries, err := fs.ReadDir(efs, dir)
		if err != nil {
			// Pack not embedded — skip silently rather than fail; lets the
			// detector add forward-looking packs without breaking.
			continue
		}
		p := Pack{Name: name}
		var modes []ModeFile
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			body, err := fs.ReadFile(efs, path.Join(dir, e.Name()))
			if err != nil {
				return nil, fmt.Errorf("read %s/%s: %w", dir, e.Name(), err)
			}
			if e.Name() == routerFile {
				p.Router = body
				p.Version = parseVersion(body)
				continue
			}
			modes = append(modes, ModeFile{Name: e.Name(), Contents: body})
		}
		if p.Router == nil {
			return nil, fmt.Errorf("pack %q has no %s", name, routerFile)
		}
		sort.Slice(modes, func(i, j int) bool { return modes[i].Name < modes[j].Name })
		p.Modes = modes
		out = append(out, p)
	}
	return out, nil
}

func parseVersion(body []byte) string {
	if m := versionCommentExpr.FindSubmatch(body); len(m) == 2 {
		return string(m[1])
	}
	return ""
}

// readVersion returns the version comment of an installed skill file if it
// exists, used for idempotent installs.
func readVersion(dest string) (string, bool) {
	b, err := os.ReadFile(dest) //nolint:gosec // path is constructed from the configured target dir, not user-controlled at read time
	if err != nil {
		return "", false
	}
	v := parseVersion(b)
	return v, v != ""
}
