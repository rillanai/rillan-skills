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
)

// Options control an install run.
type Options struct {
	Target string
	Tools  []Tool
	Packs  []string
	DryRun bool
	Force  bool
	Logger func(format string, args ...any)
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

	packs, err := loadPacks(efs, opts.Packs)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, t := range opts.Tools {
		for _, p := range packs {
			marker := routerDest(t, opts.Target, p.Name)
			if !opts.Force {
				if existing, ok := readVersion(marker); ok && existing == p.Version && p.Version != "" {
					opts.Logger("[=] %s/%s: already at v%s", t, p.Name, p.Version)
					continue
				}
			}
			if opts.DryRun {
				opts.Logger("[dry-run] %s -> %s", p.Name, marker)
				count++
				continue
			}
			if err := writePack(t, opts.Target, p); err != nil {
				return count, fmt.Errorf("install %s: %w", p.Name, err)
			}
			opts.Logger("[+] %s installed -> %s", p.Name, marker)
			count++
		}
	}
	return count, nil
}

// writePack writes a pack into the target using the layout the tool expects.
func writePack(t Tool, target string, p Pack) error {
	if t == ToolOpenCode {
		return writeFile(routerDest(t, target, p.Name), flattenForOpenCode(p))
	}
	// Claude, Codex: directory with SKILL.md + mode files.
	dir := packDir(t, target, p.Name)
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

// packDir returns the per-pack directory for tools that install a directory.
func packDir(t Tool, target, pack string) string {
	switch t {
	case ToolCodex:
		return filepath.Join(target, ".codex", "skills", pack)
	case ToolClaude:
		return filepath.Join(target, ".claude", "skills", pack)
	default:
		return filepath.Join(target, ".claude", "skills", pack)
	}
}

// routerDest returns the file whose version is checked for idempotent installs:
// the router SKILL.md for directory tools, or the single agent file for OpenCode.
func routerDest(t Tool, target, pack string) string {
	if t == ToolOpenCode {
		return filepath.Join(target, ".opencode", "agents", pack+".md")
	}
	return filepath.Join(packDir(t, target, pack), routerFile)
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
