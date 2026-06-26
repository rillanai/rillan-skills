// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

// Package detect inspects a target repository's filesystem and returns the
// set of skill packs that are relevant to that repo. Detection is
// deliberately filesystem-only — no parsing of project files beyond cheap
// content checks — so it stays fast and stable across ecosystems.
package detect

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// cloudBlock matches a HashiCorp `cloud {` block (HCP Terraform / Terraform
// Cloud execution backend), anchored at line start to avoid matching
// identifiers like `cloudfront` or `icloud`.
var cloudBlock = regexp.MustCompile(`(?m)^\s*cloud\s*\{`)

// tfeProvider matches use of the `tfe` provider — the provider block, its
// required_providers source (`hashicorp/tfe`), or any quoted `tfe_*` resource
// or data type (e.g. `data "tfe_outputs"`, `resource "tfe_workspace"`). This
// catches self-hosted TFE and workspace-management repos that never reference
// app.terraform.io.
var tfeProvider = regexp.MustCompile(`provider\s+"tfe"|hashicorp/tfe|"tfe_\w`)

// Result is the outcome of a detection pass over a target directory.
type Result struct {
	Packs   []string // skill packs (e.g. "go", "rust", "kubernetes")
	Reasons map[string]string
}

// Cross-cutting packs are always installed regardless of detection.
var crossCutting = []string{"adr", "rfc", "planning", "security"}

// Run walks the target repo and returns the relevant pack set.
func Run(target string) (Result, error) {
	r := Result{Reasons: map[string]string{}}
	for _, p := range crossCutting {
		r.Reasons[p] = "cross-cutting baseline"
	}

	hits, err := scan(target)
	if err != nil {
		return r, err
	}

	add := func(pack, reason string) {
		if _, ok := r.Reasons[pack]; ok {
			return
		}
		r.Reasons[pack] = reason
	}

	if hits.has("go.mod") {
		add("go", "go.mod present")
	}
	if hits.has("Cargo.toml") {
		add("rust", "Cargo.toml present")
	}
	if hits.has("pyproject.toml") || hits.hasGlob("requirements*.txt") || hits.has("setup.py") {
		add("python", "Python project marker present")
	}
	if hits.hasGlob("*.tf") || hits.has("terraform.tfvars") {
		add("terraform", "Terraform sources present")
		// HCP Terraform / Terraform Cloud is an overlay on top of terraform:
		// a `cloud {}` block, an app.terraform.io reference, or tfe_outputs.
		if hits.hasHCPTerraform() {
			add("hcp-terraform", "HCP Terraform / Terraform Cloud usage detected")
		}
	}
	if hits.has("Chart.yaml") {
		add("helm", "Chart.yaml present")
	}
	// Kubernetes: kustomization.yaml, or any k8s-shaped manifest under deploy/manifests.
	if hits.has("kustomization.yaml") || hits.has("kustomization.yml") || hits.hasK8sManifest() {
		add("kubernetes", "Kubernetes manifests present")
	}
	// Operator: api/*/*_types.go is the kubebuilder convention.
	if hits.hasOperatorMarkers() {
		add("operator", "kubebuilder-style API types present")
	}
	if hits.has("Dockerfile") || hits.hasGlob("Dockerfile.*") {
		add("docker", "Dockerfile present")
	}

	// CI platform packs are nested under "cicd" — emit cicd if any CI present.
	if hits.hasDir(".github/workflows") || hits.has("azure-pipelines.yml") || hits.has(".gitlab-ci.yml") {
		add("cicd", "CI workflows present")
	}

	// Stable order
	r.Packs = append(r.Packs, crossCutting...)
	// hcp-terraform follows terraform: it is an overlay loaded after the base.
	for _, p := range []string{"cicd", "docker", "go", "helm", "kubernetes", "operator", "python", "rust", "terraform", "hcp-terraform"} {
		if _, ok := r.Reasons[p]; ok {
			r.Packs = append(r.Packs, p)
		}
	}
	return r, nil
}

type hits struct {
	root  string // target root; relative file keys are resolved against it for content reads
	files map[string]bool
	dirs  map[string]bool
}

// abs resolves a recorded (slash-relative) file key to an absolute path under
// the scanned root, so content sniffs work regardless of the process CWD.
func (h *hits) abs(rel string) string {
	return filepath.Join(h.root, filepath.FromSlash(rel))
}

func (h *hits) has(name string) bool    { return h.files[name] }
func (h *hits) hasDir(name string) bool { return h.dirs[name] }
func (h *hits) hasGlob(pattern string) bool {
	for f := range h.files {
		ok, _ := filepath.Match(pattern, filepath.Base(f))
		if ok {
			return true
		}
	}
	return false
}

// hasK8sManifest looks for files that smell like k8s manifests under common
// deploy paths. Cheap content sniff: top-level apiVersion + kind keys.
func (h *hits) hasK8sManifest() bool {
	for f := range h.files {
		base := filepath.Base(f)
		if !strings.HasSuffix(base, ".yaml") && !strings.HasSuffix(base, ".yml") {
			continue
		}
		if !underAny(f, "deploy", "manifests", "k8s", "kustomize") {
			continue
		}
		b, err := os.ReadFile(h.abs(f)) //nolint:gosec // f comes from the user-supplied target tree we are deliberately scanning
		if err != nil || len(b) > 64*1024 {
			continue
		}
		s := string(b)
		if strings.Contains(s, "\napiVersion:") && strings.Contains(s, "\nkind:") {
			return true
		}
		if strings.HasPrefix(s, "apiVersion:") {
			return true
		}
	}
	return false
}

// hasOperatorMarkers detects kubebuilder-style operators: api/<group>/*_types.go
// plus a controller-runtime import somewhere.
func (h *hits) hasOperatorMarkers() bool {
	hasTypes := false
	for f := range h.files {
		if strings.HasSuffix(f, "_types.go") && strings.Contains(f, "/api/") {
			hasTypes = true
			break
		}
	}
	if !hasTypes {
		return false
	}
	for f := range h.files {
		if filepath.Base(f) != "go.mod" {
			continue
		}
		b, err := os.ReadFile(h.abs(f)) //nolint:gosec // f is a go.mod inside the user-supplied target tree
		if err != nil {
			continue
		}
		if strings.Contains(string(b), "sigs.k8s.io/controller-runtime") {
			return true
		}
	}
	return false
}

// hasHCPTerraform detects HCP Terraform / Terraform Cloud usage in any .tf
// file: a `cloud {}` execution-backend block, an app.terraform.io registry or
// state reference, or the tfe provider / tfe_outputs data source.
func (h *hits) hasHCPTerraform() bool {
	for f := range h.files {
		if !strings.HasSuffix(f, ".tf") {
			continue
		}
		b, err := os.ReadFile(h.abs(f)) //nolint:gosec // f is a .tf file inside the user-supplied target tree we are deliberately scanning
		if err != nil || len(b) > 256*1024 {
			continue
		}
		s := string(b)
		if strings.Contains(s, "app.terraform.io") ||
			cloudBlock.MatchString(s) ||
			tfeProvider.MatchString(s) {
			return true
		}
	}
	return false
}

func underAny(path string, dirs ...string) bool {
	for p := range strings.SplitSeq(filepath.ToSlash(path), "/") {
		if slices.Contains(dirs, p) {
			return true
		}
	}
	return false
}

// scan walks the target up to a reasonable depth and records files and dirs.
// Skips vendored/build/.git directories and anything beyond depth 6.
func scan(root string) (*hits, error) {
	root = filepath.Clean(root)
	h := &hits{root: root, files: map[string]bool{}, dirs: map[string]bool{}}
	rootDepth := strings.Count(root, string(os.PathSeparator))

	skip := map[string]bool{
		".git": true, "node_modules": true, "vendor": true, "target": true,
		"dist": true, "build": true, ".venv": true, "venv": true,
		"__pycache__": true, ".tox": true, ".terraform": true, ".idea": true,
		".vscode": true, ".next": true, "out": true,
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		depth := strings.Count(path, string(os.PathSeparator)) - rootDepth
		if d.IsDir() {
			if depth > 6 {
				return fs.SkipDir
			}
			if skip[d.Name()] {
				return fs.SkipDir
			}
			h.dirs[filepath.ToSlash(rel)] = true
			return nil
		}
		h.files[filepath.ToSlash(rel)] = true
		return nil
	})
	return h, err
}
