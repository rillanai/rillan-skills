# Changelog

## [1.1.0](https://github.com/rillanai/rillan-skills/compare/v1.0.0...v1.1.0) (2026-06-26)


### Features

* add --global install mode with per-tool paths and install manifest ([755be36](https://github.com/rillanai/rillan-skills/commit/755be36161f6371714823c4a04f58e2abd1d4382))
* **go:** add fuzzing + parallel-by-default testing and modernize Go skills ([a4d2d11](https://github.com/rillanai/rillan-skills/commit/a4d2d1106ee1a9117bcf25b77565350e5f5b0e83))
* make code/infra skills graphify-aware when a graph is present ([#6](https://github.com/rillanai/rillan-skills/issues/6)) ([39274a5](https://github.com/rillanai/rillan-skills/commit/39274a5c7d0dd0a99b32ec16194e3d95a80043c9))
* modernize Go/Operator skills and split Terraform into core + hcp-terraform overlay ([55084da](https://github.com/rillanai/rillan-skills/commit/55084da5884801697d7f03dd913d28880b5f7766))
* **operator:** prefer ValidatingAdmissionPolicy and modernize controller-runtime guidance ([f3d5850](https://github.com/rillanai/rillan-skills/commit/f3d585084c0326693dc3cc38e4af641ca1e9d477))
* **terraform:** split HCP/TFC specifics into a hcp-terraform overlay ([fbff610](https://github.com/rillanai/rillan-skills/commit/fbff610f918188f6a0c9c5c0599633b139683e36))


### Bug Fixes

* address PR review — manifest path safety, downgrade guard, schema check ([558a697](https://github.com/rillanai/rillan-skills/commit/558a6975018f61c90df4a98d4221d76e9907b5d9))
* **terraform:** address Copilot review on hcp-terraform overlay ([cdb93b2](https://github.com/rillanai/rillan-skills/commit/cdb93b272f13e97549205a39078158b74088980a))

## 1.0.0 (2026-06-25)


### Features

* fan-out skill architecture, dep refresh, and release pipeline ([4fc2da7](https://github.com/rillanai/rillan-skills/commit/4fc2da7cd182ed4acddebba1a6c2b9d7f9c3d547))
* fan-out skill architecture, dep refresh, and release pipeline ([c1171ee](https://github.com/rillanai/rillan-skills/commit/c1171eea9c1e88933228a3150f2547a45e2f42d6))
