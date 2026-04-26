// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

// Package install writes embedded skill files into a target repository
// using the layout each tool expects for project-scoped skills.
package install

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// skillsRoot is the directory inside the embedded FS where skill packs live.
const skillsRoot = "skills"

// Tool identifies which assistant to install for.
type Tool string

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

// Skill is one resolved skill file ready to be installed.
type Skill struct {
	Pack     string // e.g. "go"
	File     string // e.g. "ci.skill.md"
	Name     string // e.g. "go-ci" (from frontmatter or derived)
	Version  string // e.g. "0.1.0" or "" if absent
	Source   string // path inside the embed.FS
	Contents []byte
}

// Run installs the requested packs into the target for each requested tool.
// It returns the number of skills that were written (or would be written, in
// dry-run mode).
func Run(efs embed.FS, opts Options) (int, error) {
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

	skills, err := loadSkills(efs, opts.Packs)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, t := range opts.Tools {
		for _, s := range skills {
			dest := destPath(t, opts.Target, s)
			if !opts.Force {
				if existing, ok := readVersion(dest); ok && existing == s.Version && s.Version != "" {
					opts.Logger("[=] %s/%s: already at v%s", t, s.Name, s.Version)
					continue
				}
			}
			if opts.DryRun {
				opts.Logger("[dry-run] %s -> %s", s.Name, dest)
				count++
				continue
			}
			if err := writeFile(dest, s.Contents); err != nil {
				return count, fmt.Errorf("install %s: %w", s.Name, err)
			}
			opts.Logger("[+] %s installed -> %s", s.Name, dest)
			count++
		}
	}
	return count, nil
}

// destPath returns the project-scoped destination for a skill.
func destPath(t Tool, target string, s Skill) string {
	switch t {
	case ToolClaude:
		return filepath.Join(target, ".claude", "skills", s.Name, "SKILL.md")
	case ToolCodex:
		return filepath.Join(target, ".codex", "skills", s.Name, "SKILL.md")
	case ToolOpenCode:
		return filepath.Join(target, ".opencode", "agents", s.Name+".md")
	default:
		return filepath.Join(target, ".claude", "skills", s.Name, "SKILL.md")
	}
}

func writeFile(path string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".skaphos-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := io.Copy(tmp, strings.NewReader(string(contents))); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

var (
	frontmatterName    = regexp.MustCompile(`(?m)^name:\s*(.+?)\s*$`)
	versionCommentExpr = regexp.MustCompile(`(?m)^<!--\s*version:\s*([0-9]+\.[0-9]+\.[0-9]+)\s*-->`)
)

func loadSkills(efs embed.FS, packs []string) ([]Skill, error) {
	var out []Skill
	for _, pack := range packs {
		dir := path.Join(skillsRoot, pack)
		entries, err := fs.ReadDir(efs, dir)
		if err != nil {
			// Pack not embedded — skip silently rather than fail; lets the
			// detector add forward-looking packs without breaking.
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".skill.md") {
				continue
			}
			src := path.Join(dir, e.Name())
			body, err := fs.ReadFile(efs, src)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", src, err)
			}
			s := Skill{
				Pack:     pack,
				File:     e.Name(),
				Source:   src,
				Contents: body,
				Version:  parseVersion(body),
				Name:     parseName(body, pack, e.Name()),
			}
			out = append(out, s)
		}
	}
	return out, nil
}

func parseName(body []byte, pack, file string) string {
	if fm := extractFrontmatter(body); fm != "" {
		if m := frontmatterName.FindStringSubmatch(fm); len(m) == 2 {
			return strings.TrimSpace(m[1])
		}
	}
	base := strings.TrimSuffix(file, ".skill.md")
	return pack + "-" + base
}

func parseVersion(body []byte) string {
	if m := versionCommentExpr.FindSubmatch(body); len(m) == 2 {
		return string(m[1])
	}
	return ""
}

// extractFrontmatter returns the YAML frontmatter block (without the ---
// fences) if the file starts with one, else "".
func extractFrontmatter(body []byte) string {
	s := string(body)
	if !strings.HasPrefix(s, "---\n") {
		return ""
	}
	rest := s[4:]
	head, _, ok := strings.Cut(rest, "\n---")
	if !ok {
		return ""
	}
	return head
}

// readVersion returns the version comment of an installed skill file if it
// exists, used for idempotent installs.
func readVersion(path string) (string, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	v := parseVersion(b)
	return v, v != ""
}
