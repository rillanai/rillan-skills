// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

// Package rillanskills exposes the bundled skill source tree as an embedded
// filesystem so the installer binary can ship without the source repo.
package rillanskills

import "embed"

//go:embed adr/*.skill.md cicd/*.skill.md docker/*.skill.md go/*.skill.md helm/*.skill.md kubernetes/*.skill.md operator/*.skill.md planning/*.skill.md python/*.skill.md rfc/*.skill.md rust/*.skill.md security/*.skill.md terraform/*.skill.md
var Skills embed.FS

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
