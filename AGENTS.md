# Repository Guidelines

## Project Structure & Module Organization

This repository is a skills bundle plus a Go-based installer (`rillan-skills`). All skill content lives under `skills/`, organized by pack: `skills/go/`, `skills/rust/`, `skills/python/`, `skills/terraform/`, `skills/helm/`, `skills/kubernetes/`, `skills/operator/`, plus the cross-cutting packs `skills/adr/`, `skills/cicd/`, `skills/docker/`, `skills/planning/`, `skills/rfc/`, `skills/security/`. Each pack is a **router** (`SKILL.md`, the only file with frontmatter) plus zero or more frontmatter-free `<mode>.md` files that the router points to for on-demand reading. Multi-mode packs carry `dev.md`, `audit.md`, `docs.md`, `test.md`, `migrate.md` (Go/Rust/Python add `policy.md`, `workflow.md`, `ci.md`; cicd uses `core.md`, `github-actions.md`, `azure-devops.md`, `gitops.md`, `supply-chain.md`). Single-mode packs (`adr`, `rfc`, `planning`, `docker`, `security`) are just `SKILL.md`.

The installer source is in `cmd/rillan-skills/` with shared logic in `internal/detect/` and `internal/install/`. The bundled skill tree is exposed to the binary through `embed.go`. User documentation is `README.md`, licensing metadata is in `LICENSE` and `NOTICE`.

## Build, Test, and Development Commands

The build runner is [Task](https://taskfile.dev). Run from the repository root:

- `task build` — build the `rillan-skills` binary into `bin/`.
- `task list` — list every skill bundled in the binary.
- `task detect` — run pack detection against the repo itself.
- `task vet` — `go vet ./...`.
- `task test` — `go test ./...`.
- `task install` — install the built binary into `~/.local/bin`.
- `task clean` — remove `bin/`.

For a one-off dry-run install against a sample target:

```bash
./bin/rillan-skills install --target /path/to/repo --tool claude --dry-run
```

## Coding Style & Naming Conventions

Keep Markdown direct and tool-oriented. Use short sections, imperative guidance, and fenced examples only where they add value. Each pack's router is `skills/<pack>/SKILL.md` (with YAML frontmatter); mode files are `skills/<pack>/<mode>.md` (no frontmatter — only `SKILL.md` is registered as a skill). Preserve the top-of-file metadata format. `SKILL.md` carries frontmatter then the comment headers; mode files carry just the comment headers:

- `name:` + `description:` YAML frontmatter (router `SKILL.md` only)
- `<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->`
- `<!-- SPDX-License-Identifier: Apache-2.0 -->`
- `<!-- version: X.Y.Z -->` (the router's version drives install idempotency for the whole pack)

For Go code, follow standard `gofmt` + `go vet` conventions. Keep `cmd/rillan-skills/main.go` thin; put real logic in `internal/`.

## Testing Guidelines

Run `task ci` (or at minimum `task vet` and `task test`) before sending changes. The installer and detector are covered by unit tests in `internal/install/` and `internal/detect/`; keep them green and add cases for new behavior. When editing skill files, confirm the `<!-- version: X.Y.Z -->` header is preserved, that mode files stay frontmatter-free, and that cross-references use the path form (same-pack `policy.md`, cross-pack `cicd/core.md`). For installer changes, exercise both `task detect` (against this repo) and a `--dry-run install --target <repo>` against a representative sample.

## Commit & Pull Request Guidelines

Use [Conventional Commit](https://www.conventionalcommits.org) subjects — `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `ci:`, etc., with `feat!:` or a `BREAKING CHANGE:` footer for breaking changes. release-please derives the next version and CHANGELOG from these, so a non-conventional subject simply won't appear in a release. When squash-merging, make the PR title conventional. Keep each commit scoped to one logical change. Sign commits (DCO sign-off plus SSH/GPG signature) when the project requires it; the bot-authored release-please PR is exempt from the DCO check.

Pull requests should include a concise summary, the affected directories, the verification commands you ran, and any README updates needed for user-facing behavior changes. See the README "Releasing" section for how merges become releases.
