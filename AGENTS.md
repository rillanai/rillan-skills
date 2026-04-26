# Repository Guidelines

## Project Structure & Module Organization

This repository is a skills bundle plus one installer script. Language-specific skills live in `go/`, `rust/`, `python/`, `terraform/`, `helm/`, `kubernetes/`, and `operator/`, with one `*.skill.md` file per mode such as `dev`, `audit`, `docs`, `test`, and `migrate`. Go, Rust, and Python also include `policy` and `workflow` layers. The installer is `install.sh`, the main user documentation is `README.md`, and licensing metadata is in `LICENSE`.

## Build, Test, and Development Commands

There is no build step. Use the installer and its preview modes as the primary verification path.

- `./install.sh --list` shows detected tools and available skills.
- `./install.sh --dry-run` previews installs without changing user directories.
- `./install.sh --tool=codex --lang=go` tests a targeted install path.
- `./install.sh --uninstall --tool=codex` verifies uninstall behavior for one tool.

Run commands from the repository root.

## Coding Style & Naming Conventions

Keep Markdown direct and tool-oriented. Use short sections, imperative guidance, and fenced examples only where they add value. Skill files should remain named `<mode>.skill.md` inside the appropriate language directory. Preserve the top-of-file metadata format:

- `<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->`
- `<!-- SPDX-License-Identifier: Apache-2.0 -->`
- `<!-- version: X.Y.Z -->`

For `install.sh`, follow existing Bash style: `set -euo pipefail`, quoted variables, small functions, and 4-space indentation inside blocks.

## Testing Guidelines

This repo does not use a formal test framework. Validate changes by exercising installer flows with `--list`, `--dry-run`, and a narrow real install target when safe. When editing skill files, confirm version headers are preserved and that referenced paths or commands still match the repo layout.

## Commit & Pull Request Guidelines

Current history is minimal (`Initial commit`, `Import`), so prefer short, imperative commit subjects such as `Add SPDX headers to skill files` or `Clarify install examples`. Keep each commit scoped to one logical change.

Pull requests should include a concise summary, affected directories, verification commands run, and any README updates needed for user-facing behavior changes.
