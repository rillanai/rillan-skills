// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

// Package rillanskills exposes the bundled skill source tree as an embedded
// filesystem so the installer binary can ship without the source repo.
package rillanskills

import "embed"

// Skills is the embedded skill source tree. The installer reads from this FS
// rather than the disk so it can ship as a single binary. Each pack is a
// directory holding a router SKILL.md plus zero or more `<mode>.md` files.
//
//go:embed skills/*/*.md
var Skills embed.FS

// SkillsRoot is the directory inside the embedded FS where skill packs live.
const SkillsRoot = "skills"

// Packs returns the canonical list of skill pack directories embedded into
// the binary, in stable order.
func Packs() []string {
	return []string{
		"adr",
		"cicd",
		"docker",
		"go",
		"helm",
		"kubernetes",
		"operator",
		"planning",
		"python",
		"rfc",
		"rust",
		"security",
		"terraform",
	}
}
