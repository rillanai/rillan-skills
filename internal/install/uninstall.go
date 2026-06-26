// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"fmt"
	"os"
	"path/filepath"
)

// UninstallOptions control an uninstall run.
type UninstallOptions struct {
	Target string // repo dir (project mode); ignored when Global is set
	Home   string // home dir for --global; empty resolves via os.UserHomeDir
	Tools  []Tool
	Global bool
	DryRun bool
	// KnownPacks is the fallback pack set used for tools with no manifest
	// entries (e.g. installs predating the manifest).
	KnownPacks []string
	Logger     func(format string, args ...any)
}

// Uninstall removes the packs this installer recorded, per tool. It uses the
// manifest as the source of truth for which files to delete so it never
// removes unrelated files sharing a skills directory; for tools with no
// manifest entry it falls back to removing known pack paths that exist on disk.
func Uninstall(opts UninstallOptions) (int, error) {
	if opts.Logger == nil {
		opts.Logger = func(string, ...any) {}
	}
	if opts.Global {
		if opts.Home == "" {
			h, err := os.UserHomeDir()
			if err != nil {
				return 0, fmt.Errorf("uninstall: resolve home directory: %w", err)
			}
			opts.Home = h
		}
	} else {
		abs, err := filepath.Abs(opts.Target)
		if err != nil {
			return 0, err
		}
		opts.Target = abs
	}

	root := opts.Target
	if opts.Global {
		root = opts.Home
	}
	man, err := LoadManifest(root)
	if err != nil {
		return 0, fmt.Errorf("uninstall: read manifest: %w", err)
	}

	removed := 0
	changed := false
	for _, t := range opts.Tools {
		base, ok := SkillsDir(t, opts.Target, opts.Home, opts.Global)
		if !ok {
			if hint := ManualHint(t); hint != "" {
				opts.Logger("[!] %s: skipped — %s", t, hint)
			}
			continue
		}

		// Packs recorded for this tool, from the manifest.
		var recorded []string
		for _, e := range man.Entries {
			if e.Tool == string(t) {
				recorded = append(recorded, e.Pack)
			}
		}

		if len(recorded) > 0 {
			for _, pack := range recorded {
				idx := man.find(string(t), pack)
				if idx < 0 {
					continue
				}
				files := man.Entries[idx].Files
				didRemove := false
				for _, rel := range files {
					p := filepath.Join(root, filepath.FromSlash(rel))
					if removePath(p, opts) {
						didRemove = true
					}
				}
				// Drop the now-empty pack directory for directory-based tools.
				if t != ToolOpenCode {
					_ = os.Remove(InstalledPath(t, base, pack))
				}
				if didRemove {
					removed++
				}
				if !opts.DryRun {
					man.remove(string(t), pack)
					changed = true
				}
			}
			continue
		}

		// Fallback: no manifest record for this tool — remove known pack paths
		// that are present on disk.
		for _, pack := range opts.KnownPacks {
			p := InstalledPath(t, base, pack)
			if _, err := os.Stat(p); err != nil {
				continue
			}
			if removePath(p, opts) {
				removed++
			}
		}
	}

	if changed && !opts.DryRun {
		if err := man.Save(root); err != nil {
			return removed, fmt.Errorf("uninstall: write manifest: %w", err)
		}
	}
	return removed, nil
}

// removePath removes a file or directory, honoring dry-run, and reports whether
// it acted (or would act). Missing paths are a no-op.
func removePath(p string, opts UninstallOptions) bool {
	if _, err := os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	if opts.DryRun {
		opts.Logger("[dry-run] would remove %s", p)
		return true
	}
	if err := os.RemoveAll(p); err != nil {
		opts.Logger("[!] could not remove %s: %v", p, err)
		return false
	}
	opts.Logger("[-] removed %s", p)
	return true
}
