// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// manifestSchema is the on-disk manifest format version, bumped only on
// incompatible structural changes.
const manifestSchema = 1

// manifestRelPath is the manifest location relative to the scope root (the
// repo target for project installs, or the home dir for --global installs).
var manifestRelPath = filepath.Join(".rillan-skills", "manifest.json")

// Entry records one installed pack for one tool so a later run can decide
// whether to upgrade it and an uninstall can remove exactly what was written.
type Entry struct {
	Tool    string `json:"tool"`
	Pack    string `json:"pack"`
	Version string `json:"version"`
	// Files are slash-relative to the scope root (target or home).
	Files []string `json:"files"`
}

// Manifest is the install ledger persisted at <root>/.rillan-skills/manifest.json.
type Manifest struct {
	Schema  int     `json:"schema"`
	Global  bool    `json:"global"`
	Updated string  `json:"updated"`
	Entries []Entry `json:"entries"`
}

// ManifestPath returns the manifest path for a given scope root.
func ManifestPath(root string) string {
	return filepath.Join(root, manifestRelPath)
}

// NearestProjectRoot walks up from start looking for a project install (a
// .rillan-skills/manifest.json). It returns the first directory that has one
// and true, or ("", false) when none is found before the filesystem root.
func NearestProjectRoot(start string) (string, bool) {
	dir := start
	for {
		if fi, err := os.Stat(ManifestPath(dir)); err == nil && !fi.IsDir() {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// LoadManifest reads the manifest at root, returning an empty (non-nil)
// manifest when none exists yet.
func LoadManifest(root string) (*Manifest, error) {
	b, err := os.ReadFile(ManifestPath(root)) //nolint:gosec // path derived from the configured scope root
	if err != nil {
		if os.IsNotExist(err) {
			return &Manifest{Schema: manifestSchema}, nil
		}
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	// schema 0 means "written before the field existed" — treat as current.
	// Any other unrecognized value is incompatible: refuse rather than risk an
	// older binary misreading a newer manifest's layout.
	if m.Schema == 0 {
		m.Schema = manifestSchema
	} else if m.Schema != manifestSchema {
		return nil, fmt.Errorf("manifest at %s has unsupported schema %d (this binary supports %d); upgrade rillan-skills", ManifestPath(root), m.Schema, manifestSchema)
	}
	return &m, nil
}

// Save writes the manifest atomically to root.
func (m *Manifest) Save(root string) error {
	m.Schema = manifestSchema
	m.Updated = time.Now().UTC().Format(time.RFC3339)
	sort.Slice(m.Entries, func(i, j int) bool {
		if m.Entries[i].Tool != m.Entries[j].Tool {
			return m.Entries[i].Tool < m.Entries[j].Tool
		}
		return m.Entries[i].Pack < m.Entries[j].Pack
	})
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(ManifestPath(root), append(b, '\n'))
}

// find returns the index of the entry for (tool, pack), or -1.
func (m *Manifest) find(tool, pack string) int {
	for i := range m.Entries {
		if m.Entries[i].Tool == tool && m.Entries[i].Pack == pack {
			return i
		}
	}
	return -1
}

// Recorded returns the version recorded for (tool, pack), or "" if absent.
func (m *Manifest) Recorded(tool, pack string) string {
	if i := m.find(tool, pack); i >= 0 {
		return m.Entries[i].Version
	}
	return ""
}

// upsert inserts or replaces the entry for (tool, pack).
func (m *Manifest) upsert(e Entry) {
	if i := m.find(e.Tool, e.Pack); i >= 0 {
		m.Entries[i] = e
		return
	}
	m.Entries = append(m.Entries, e)
}

// remove drops the entry for (tool, pack) and reports whether one existed.
func (m *Manifest) remove(tool, pack string) bool {
	if i := m.find(tool, pack); i >= 0 {
		m.Entries = append(m.Entries[:i], m.Entries[i+1:]...)
		return true
	}
	return false
}

// changeKind classifies an install relative to a previously recorded version.
type changeKind int

const (
	changeFresh   changeKind = iota // nothing recorded before
	changeSame                      // recorded == bundled
	changeUpgrade                   // recorded < bundled
	changeDown                      // recorded > bundled
)

// classify compares a recorded version against the bundled one.
func classify(recorded, bundled string) changeKind {
	if recorded == "" {
		return changeFresh
	}
	switch compareSemver(bundled, recorded) {
	case 0:
		return changeSame
	case 1:
		return changeUpgrade
	default:
		return changeDown
	}
}

// compareSemver compares dotted numeric versions ("X.Y.Z"). Non-numeric or
// unparsable fields sort as 0. Returns -1, 0, or 1.
func compareSemver(a, b string) int {
	as, bs := splitSemver(a), splitSemver(b)
	for i := 0; i < 3; i++ {
		if as[i] != bs[i] {
			if as[i] < bs[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func splitSemver(v string) [3]int {
	var out [3]int
	parts := strings.SplitN(v, ".", 3)
	for i := 0; i < len(parts) && i < 3; i++ {
		n, _ := strconv.Atoi(strings.TrimSpace(parts[i]))
		out[i] = n
	}
	return out
}
