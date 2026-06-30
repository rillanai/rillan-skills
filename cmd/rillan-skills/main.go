// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

// Command rillan-skills is the project-scoped skill installer. It detects
// what languages and tools a target repository uses and writes only the
// relevant skill packs into that repo's tool-specific directories.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"time"

	rillanskills "github.com/rillanai/rillan-skills"
	"github.com/rillanai/rillan-skills/internal/detect"
	"github.com/rillanai/rillan-skills/internal/install"
	"github.com/rillanai/rillan-skills/internal/selfupdate"
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
  update      Refresh installed skills (nearest project, or --global) to this binary's version
  upgrade     Replace this binary with the latest release from GitHub
  version     Print the rillan-skills version

Common flags:
  --target string   Target repository directory (default ".")
  --packs string    Comma-separated pack list, overrides detection (e.g. "go,kubernetes")
  --tool string     Comma-separated tool list: claude,codex,opencode,grok or "all" (default "claude")
  --global          Install into each tool's user-level config dir instead of a repo
  --dry-run         Show what would happen without writing files
  --force           Overwrite installed files even when versions match

Examples:
  rillan-skills install --target . --tool claude
  rillan-skills install --packs go,security --dry-run
  rillan-skills install --global --tool claude
  rillan-skills install --global --tool all --dry-run
  rillan-skills detect --target ../some-repo
  rillan-skills update                  # refresh skills in the nearest install
  rillan-skills update --global --yes   # refresh global skills without prompting
  rillan-skills upgrade                 # replace this binary with the latest release
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
	case "update":
		err = runUpdate(args)
	case "upgrade":
		err = runUpgrade(args)
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
	global bool
}

func parseCommon(name string, args []string) (*commonFlags, []string, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	c := &commonFlags{}
	fs.StringVar(&c.target, "target", ".", "Target repository directory")
	fs.StringVar(&c.packs, "packs", "", "Comma-separated pack list (overrides detection)")
	fs.StringVar(&c.tools, "tool", "claude", "Comma-separated tool list: claude,codex,opencode,grok or \"all\"")
	fs.BoolVar(&c.dryRun, "dry-run", false, "Show what would happen without writing files")
	fs.BoolVar(&c.force, "force", false, "Overwrite installed files even when versions match")
	fs.BoolVar(&c.global, "global", false, "Install into each tool's user-level config dir instead of a repo")
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
		if t == "all" {
			return install.AllTools(), nil
		}
		switch install.Tool(t) {
		case install.ToolClaude, install.ToolCodex, install.ToolOpenCode, install.ToolGrok:
			out = append(out, install.Tool(t))
		default:
			return nil, fmt.Errorf("unknown tool %q (valid: claude, codex, opencode, grok, all)", t)
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
	// Global install targets the user's home, not a repo, so detection would be
	// meaningless. With no explicit --packs, default to every bundled pack;
	// an explicit --packs (handled above) still narrows the selection.
	if c.global {
		packs := rillanskills.Packs()
		reasons := map[string]string{}
		for _, p := range packs {
			reasons[p] = "global install (all packs)"
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
		Global: c.global,
		Logger: func(format string, a ...any) { fmt.Printf(format+"\n", a...) },
	})
	if err != nil {
		return err
	}
	verb := "installed"
	if c.dryRun {
		verb = "would install"
	}
	dest := c.target
	if c.global {
		dest = "user-level config directories"
	}
	fmt.Printf("\nrillan-skills: %s %d skill pack(s) into %s\n", verb, count, dest)
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
	removed, err := install.Uninstall(install.UninstallOptions{
		Target:     c.target,
		Tools:      tools,
		Global:     c.global,
		DryRun:     c.dryRun,
		KnownPacks: rillanskills.Packs(),
		Logger:     func(format string, a ...any) { fmt.Printf(format+"\n", a...) },
	})
	if err != nil {
		return err
	}
	verb := "removed"
	if c.dryRun {
		verb = "would remove"
	}
	fmt.Printf("\nrillan-skills: %s %d skill pack(s)\n", verb, removed)
	return nil
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

// runUpdate refreshes the skills recorded in an existing install — the nearest
// project (walking up from --target) or the user-level global install — to the
// versions bundled in this binary. It prompts before writing unless --yes.
func runUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	target := fs.String("target", ".", "Project directory to search upward from for an install")
	global := fs.Bool("global", false, "Refresh the user-level (global) install instead of the nearest project")
	dryRun := fs.Bool("dry-run", false, "Show what would change without writing files")
	force := fs.Bool("force", false, "Rewrite skills even when the installed version is unchanged")
	var yes bool
	fs.BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	fs.BoolVar(&yes, "y", false, "Alias for --yes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	absTarget, err := filepath.Abs(*target)
	if err != nil {
		return err
	}

	// Resolve which install to refresh: --global forces the user-level scope;
	// otherwise use the nearest project install, falling back to global.
	var root string
	isGlobal := *global
	switch {
	case isGlobal:
		home, herr := os.UserHomeDir()
		if herr != nil {
			return fmt.Errorf("resolve home directory: %w", herr)
		}
		root = home
	default:
		if pr, ok := install.NearestProjectRoot(absTarget); ok {
			root = pr
		} else {
			home, herr := os.UserHomeDir()
			if herr != nil {
				return fmt.Errorf("no project install found and cannot resolve home directory: %w", herr)
			}
			root, isGlobal = home, true
			fmt.Println("rillan-skills: no project install found nearby; targeting the global install")
		}
	}

	man, err := install.LoadManifest(root)
	if err != nil {
		return err
	}
	order, byTool := manifestByTool(man)

	scope := root
	if isGlobal {
		scope = "user-level config (global)"
	}
	if len(order) == 0 {
		fmt.Printf("rillan-skills: no installed skills recorded for %s; run `rillan-skills install` first\n", scope)
		return nil
	}

	fmt.Printf("rillan-skills %s — refresh installed skills at %s:\n", version, scope)
	for _, tool := range order {
		fmt.Printf("  %-9s %s\n", tool, strings.Join(byTool[tool], ", "))
	}
	ok, err := confirm("Update these skills now?", yes || *dryRun)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("aborted.")
		return nil
	}

	opts := install.Options{
		DryRun: *dryRun,
		Force:  *force,
		Logger: func(format string, a ...any) { fmt.Printf(format+"\n", a...) },
	}
	if isGlobal {
		opts.Global = true
		opts.Home = root
	} else {
		opts.Target = root
	}

	total := 0
	for _, tool := range order {
		opts.Tools = []install.Tool{install.Tool(tool)}
		opts.Packs = byTool[tool]
		n, rerr := install.Run(rillanskills.Skills, opts)
		if rerr != nil {
			return rerr
		}
		total += n
	}
	verb := "updated"
	if *dryRun {
		verb = "would update"
	}
	fmt.Printf("\nrillan-skills: %s %d skill pack(s) at %s\n", verb, total, scope)
	return nil
}

// manifestByTool groups a manifest's recorded packs by tool, preserving
// first-seen order so the refresh touches exactly what was installed.
func manifestByTool(m *install.Manifest) (order []string, byTool map[string][]string) {
	byTool = map[string][]string{}
	seen := map[string]map[string]bool{}
	for _, e := range m.Entries {
		if _, ok := byTool[e.Tool]; !ok {
			order = append(order, e.Tool)
			seen[e.Tool] = map[string]bool{}
		}
		if !seen[e.Tool][e.Pack] {
			seen[e.Tool][e.Pack] = true
			byTool[e.Tool] = append(byTool[e.Tool], e.Pack)
		}
	}
	return order, byTool
}

// runUpgrade replaces the running binary with the latest GitHub release for
// this platform. It verifies the release checksum before swapping and prompts
// before replacing unless --yes.
func runUpgrade(args []string) error {
	fs := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	check := fs.Bool("check", false, "Only report whether a newer release is available")
	force := fs.Bool("force", false, "Upgrade even when already on the latest version (or a dev build)")
	dryRun := fs.Bool("dry-run", false, "Show what would happen without replacing the binary")
	var yes bool
	fs.BoolVar(&yes, "yes", false, "Skip the confirmation prompt")
	fs.BoolVar(&yes, "y", false, "Alias for --yes")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if version == "dev" && !*force {
		fmt.Println(`rillan-skills: this is a local dev build (version "dev"); upgrade targets released binaries.`)
		fmt.Println("Install a release from https://github.com/rillanai/rillan-skills/releases, or pass --force to upgrade anyway.")
		return nil
	}

	opts := selfupdate.Options{
		CurrentVersion: version,
		GOOS:           runtime.GOOS,
		GOARCH:         runtime.GOARCH,
		Token:          os.Getenv("GITHUB_TOKEN"),
		Logger:         func(format string, a ...any) { fmt.Printf(format+"\n", a...) },
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	st, rel, err := selfupdate.Check(ctx, opts)
	if err != nil {
		return err
	}
	if !st.Newer && !*force {
		fmt.Printf("rillan-skills: already up to date (v%s)\n", st.Current)
		return nil
	}
	if *check {
		fmt.Printf("rillan-skills: update available %s -> %s (run `rillan-skills upgrade` to install)\n", st.Current, st.Latest)
		return nil
	}
	if *dryRun {
		fmt.Printf("rillan-skills: would upgrade %s -> %s\n", st.Current, st.Latest)
		return nil
	}

	ok, err := confirm(fmt.Sprintf("Upgrade rillan-skills %s -> %s?", st.Current, st.Latest), yes)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("aborted.")
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate current executable: %w", err)
	}
	if resolved, rerr := filepath.EvalSymlinks(exe); rerr == nil {
		exe = resolved
	}
	if err := selfupdate.Apply(ctx, opts, rel, exe); err != nil {
		fmt.Fprintln(os.Stderr, "rillan-skills: could not replace the binary in place (insufficient permissions or a read-only path).")
		fmt.Fprintln(os.Stderr, "Download the latest release manually: https://github.com/rillanai/rillan-skills/releases/latest")
		return err
	}
	fmt.Printf("rillan-skills: upgraded %s -> %s\n", st.Current, st.Latest)
	fmt.Println("Run `rillan-skills update` (or `update --global`) to refresh installed skills to the new version.")
	return nil
}

// confirm asks a yes/no question. It returns true on an affirmative answer, or
// immediately when assumeYes is set. On a non-interactive stdin (no TTY) without
// assumeYes it returns an error so automation must opt in explicitly via --yes.
func confirm(question string, assumeYes bool) (bool, error) {
	if assumeYes {
		return true, nil
	}
	if !stdinIsTerminal() {
		return false, fmt.Errorf("refusing to prompt on a non-interactive stdin; re-run with --yes to proceed")
	}
	fmt.Printf("%s [y/N] ", question)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func stdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
