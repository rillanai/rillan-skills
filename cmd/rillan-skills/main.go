// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

// Command rillan-skills is the project-scoped skill installer. It detects
// what languages and tools a target repository uses and writes only the
// relevant skill packs into that repo's tool-specific directories.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	rillanskills "github.com/rillanai/rillan-skills"
	"github.com/rillanai/rillan-skills/internal/detect"
	"github.com/rillanai/rillan-skills/internal/install"
)

// version is the build version, overridden at release time via
// -ldflags "-X main.version=...". Defaults to "dev" for local builds.
var version = "dev"

const usage = `rillan-skills — project-scoped skill installer

Usage:
  rillan-skills <command> [flags]

Commands:
  install     Install relevant skills into a target repository
  detect      Print which skill packs would be installed (filesystem scan only)
  list        List all skills bundled in this binary
  uninstall   Remove installed skill files from a target repository
  version     Print the rillan-skills version

Common flags:
  --target string   Target repository directory (default ".")
  --packs string    Comma-separated pack list, overrides detection (e.g. "go,kubernetes")
  --tool string     Comma-separated tool list: claude,codex,opencode (default "claude")
  --dry-run         Show what would happen without writing files
  --force           Overwrite installed files even when versions match

Examples:
  rillan-skills install --target . --tool claude
  rillan-skills install --packs go,security --dry-run
  rillan-skills detect --target ../some-repo
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	var err error
	switch cmd {
	case "install":
		err = runInstall(args)
	case "detect":
		err = runDetect(args)
	case "list":
		err = runList(args)
	case "uninstall":
		err = runUninstall(args)
	case "version", "--version", "-v":
		fmt.Printf("rillan-skills %s\n", version)
		return
	case "-h", "--help", "help":
		fmt.Print(usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "rillan-skills: unknown command %q\n\n%s", cmd, usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "rillan-skills:", err)
		os.Exit(1)
	}
}

type commonFlags struct {
	target string
	packs  string
	tools  string
	dryRun bool
	force  bool
}

func parseCommon(name string, args []string) (*commonFlags, []string, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	c := &commonFlags{}
	fs.StringVar(&c.target, "target", ".", "Target repository directory")
	fs.StringVar(&c.packs, "packs", "", "Comma-separated pack list (overrides detection)")
	fs.StringVar(&c.tools, "tool", "claude", "Comma-separated tool list: claude,codex,opencode")
	fs.BoolVar(&c.dryRun, "dry-run", false, "Show what would happen without writing files")
	fs.BoolVar(&c.force, "force", false, "Overwrite installed files even when versions match")
	if err := fs.Parse(args); err != nil {
		return nil, nil, err
	}
	abs, err := filepath.Abs(c.target)
	if err != nil {
		return nil, nil, err
	}
	c.target = abs
	return c, fs.Args(), nil
}

func resolveTools(s string) ([]install.Tool, error) {
	var out []install.Tool
	for raw := range strings.SplitSeq(s, ",") {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		switch install.Tool(t) {
		case install.ToolClaude, install.ToolCodex, install.ToolOpenCode:
			out = append(out, install.Tool(t))
		default:
			return nil, fmt.Errorf("unknown tool %q (valid: claude, codex, opencode)", t)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no tools selected")
	}
	return out, nil
}

func resolvePacks(c *commonFlags) ([]string, map[string]string, error) {
	if c.packs != "" {
		var packs []string
		reasons := map[string]string{}
		for raw := range strings.SplitSeq(c.packs, ",") {
			p := strings.TrimSpace(raw)
			if p == "" {
				continue
			}
			if !slices.Contains(rillanskills.Packs(), p) {
				return nil, nil, fmt.Errorf("unknown pack %q (valid: %s)", p, strings.Join(rillanskills.Packs(), ", "))
			}
			packs = append(packs, p)
			reasons[p] = "explicit --packs"
		}
		return packs, reasons, nil
	}
	r, err := detect.Run(c.target)
	if err != nil {
		return nil, nil, fmt.Errorf("detect: %w", err)
	}
	return r.Packs, r.Reasons, nil
}

func runInstall(args []string) error {
	c, _, err := parseCommon("install", args)
	if err != nil {
		return err
	}
	tools, err := resolveTools(c.tools)
	if err != nil {
		return err
	}
	packs, reasons, err := resolvePacks(c)
	if err != nil {
		return err
	}
	if len(packs) == 0 {
		fmt.Println("rillan-skills: no skill packs detected; nothing to install")
		return nil
	}
	printPackList(packs, reasons)
	count, err := install.Run(rillanskills.Skills, install.Options{
		Target: c.target,
		Tools:  tools,
		Packs:  packs,
		DryRun: c.dryRun,
		Force:  c.force,
		Logger: func(format string, a ...any) { fmt.Printf(format+"\n", a...) },
	})
	if err != nil {
		return err
	}
	verb := "installed"
	if c.dryRun {
		verb = "would install"
	}
	fmt.Printf("\nrillan-skills: %s %d skill pack(s) into %s\n", verb, count, c.target)
	return nil
}

func runDetect(args []string) error {
	c, _, err := parseCommon("detect", args)
	if err != nil {
		return err
	}
	r, err := detect.Run(c.target)
	if err != nil {
		return err
	}
	if len(r.Packs) == 0 {
		fmt.Println("rillan-skills: no skill packs detected")
		return nil
	}
	printPackList(r.Packs, r.Reasons)
	return nil
}

func runList(args []string) error {
	_, _, err := parseCommon("list", args)
	if err != nil {
		return err
	}
	fmt.Println("Each pack installs as one root skill (SKILL.md) that routes to its mode files:")
	fmt.Println()
	for _, pack := range rillanskills.Packs() {
		entries, err := fs.ReadDir(rillanskills.Skills, rillanskills.SkillsRoot+"/"+pack)
		if err != nil {
			continue
		}
		var modes []string
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "SKILL.md" {
				continue
			}
			modes = append(modes, strings.TrimSuffix(e.Name(), ".md"))
		}
		sort.Strings(modes)
		desc := "single skill"
		if len(modes) > 0 {
			desc = "modes: " + strings.Join(modes, ", ")
		}
		fmt.Printf("%-12s %s\n", pack, desc)
	}
	return nil
}

func runUninstall(args []string) error {
	c, _, err := parseCommon("uninstall", args)
	if err != nil {
		return err
	}
	tools, err := resolveTools(c.tools)
	if err != nil {
		return err
	}
	removed := 0
	for _, t := range tools {
		root := uninstallRoot(t, c.target)
		entries, err := os.ReadDir(root)
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			continue
		}
		for _, e := range entries {
			path := filepath.Join(root, e.Name())
			if c.dryRun {
				fmt.Printf("[dry-run] would remove %s\n", path)
				removed++
				continue
			}
			if err := os.RemoveAll(path); err != nil {
				return err
			}
			fmt.Printf("[-] removed %s\n", path)
			removed++
		}
	}
	verb := "removed"
	if c.dryRun {
		verb = "would remove"
	}
	fmt.Printf("\nrillan-skills: %s %d skill entries\n", verb, removed)
	return nil
}

func uninstallRoot(t install.Tool, target string) string {
	switch t {
	case install.ToolClaude:
		return filepath.Join(target, ".claude", "skills")
	case install.ToolCodex:
		return filepath.Join(target, ".codex", "skills")
	case install.ToolOpenCode:
		return filepath.Join(target, ".opencode", "agents")
	default:
		return filepath.Join(target, ".claude", "skills")
	}
}

func printPackList(packs []string, reasons map[string]string) {
	fmt.Println("Skill packs:")
	for _, p := range packs {
		why := reasons[p]
		if why == "" {
			why = "selected"
		}
		fmt.Printf("  %-12s — %s\n", p, why)
	}
	fmt.Println()
}
