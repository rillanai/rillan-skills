# Repository Guidelines

## Project Structure & Module Organization

This repository is a skills bundle plus a Go-based installer (`rillan-skills`). All skill content lives under `skills/`, organized by pack: `skills/go/`, `skills/rust/`, `skills/python/`, `skills/terraform/`, `skills/helm/`, `skills/kubernetes/`, `skills/operator/`, plus the cross-cutting packs `skills/adr/`, `skills/cicd/`, `skills/docker/`, `skills/planning/`, `skills/rfc/`, `skills/security/`. Each pack contains one `*.skill.md` file per mode (`dev`, `audit`, `docs`, `test`, `migrate`); Go, Rust, and Python additionally have `policy`, `workflow`, and `ci` files.

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

Keep Markdown direct and tool-oriented. Use short sections, imperative guidance, and fenced examples only where they add value. Skill files must remain named `<mode>.skill.md` inside the appropriate `skills/<pack>/` directory. Preserve the top-of-file metadata format:

- `<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->`
- `<!-- SPDX-License-Identifier: Apache-2.0 -->`
- `<!-- version: X.Y.Z -->`

For Go code, follow standard `gofmt` + `go vet` conventions. Keep `cmd/rillan-skills/main.go` thin; put real logic in `internal/`.

## Testing Guidelines

Run `task vet` and `task test` before sending changes. When editing skill files, confirm the `<!-- version: X.Y.Z -->` header is preserved and that any referenced paths or commands still match the repo layout. For installer changes, exercise both `task detect` (against this repo) and a `--dry-run install --target <repo>` against a representative sample.

## Commit & Pull Request Guidelines

Use short, imperative commit subjects such as `Move skill packs into skills/` or `Fix wide language matrix in README`. Keep each commit scoped to one logical change. Sign commits (DCO sign-off plus SSH/GPG signature) when the project requires it.

Pull requests should include a concise summary, the affected directories, the verification commands you ran, and any README updates needed for user-facing behavior changes.
