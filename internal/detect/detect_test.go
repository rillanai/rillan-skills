// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

package detect

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// writeTree creates the given files (relative path → contents) under a fresh
// temp dir and returns the dir.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestCrossCuttingAlwaysPresent(t *testing.T) {
	r, err := Run(writeTree(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"adr", "rfc", "planning", "security", "socratic"} {
		if !slices.Contains(r.Packs, p) {
			t.Errorf("cross-cutting pack %q missing from %v", p, r.Packs)
		}
	}
}

func TestDetectByMarker(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"go", map[string]string{"go.mod": "module x\n"}, "go"},
		{"rust", map[string]string{"Cargo.toml": "[package]\n"}, "rust"},
		{"python", map[string]string{"pyproject.toml": "[project]\n"}, "python"},
		{"terraform", map[string]string{"main.tf": "resource {}\n"}, "terraform"},
		{"helm", map[string]string{"Chart.yaml": "name: c\n"}, "helm"},
		{"docker", map[string]string{"Dockerfile": "FROM scratch\n"}, "docker"},
		{"kustomize", map[string]string{"kustomization.yaml": "resources: []\n"}, "kubernetes"},
		{"ci", map[string]string{".github/workflows/ci.yml": "on: push\n"}, "cicd"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, err := Run(writeTree(t, c.files))
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Contains(r.Packs, c.want) {
				t.Errorf("want pack %q for %v, got %v", c.want, c.files, r.Packs)
			}
		})
	}
}

func TestDetectHCPTerraform(t *testing.T) {
	// A cloud {} block alongside plain Terraform sources triggers both packs.
	cloud := map[string]string{
		"main.tf":    "resource {}\n",
		"backend.tf": "terraform {\n  cloud {\n    organization = \"acme\"\n  }\n}\n",
	}
	r, err := Run(writeTree(t, cloud))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"terraform", "hcp-terraform"} {
		if !slices.Contains(r.Packs, p) {
			t.Errorf("want pack %q for cloud block, got %v", p, r.Packs)
		}
	}

	// An app.terraform.io registry source also triggers the overlay.
	reg := map[string]string{
		"main.tf": "module \"net\" {\n  source = \"app.terraform.io/acme/net/aws\"\n}\n",
	}
	r2, err := Run(writeTree(t, reg))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(r2.Packs, "hcp-terraform") {
		t.Errorf("want hcp-terraform for app.terraform.io source, got %v", r2.Packs)
	}

	// The tfe provider / tfe_* resources trigger the overlay too — covers
	// self-hosted TFE and workspace-management repos with no app.terraform.io.
	tfe := map[string]string{
		"workspaces.tf": "resource \"tfe_workspace\" \"app\" {\n  organization = \"acme\"\n}\n",
	}
	r3, err := Run(writeTree(t, tfe))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(r3.Packs, "hcp-terraform") {
		t.Errorf("want hcp-terraform for tfe_* resource, got %v", r3.Packs)
	}

	// Plain Terraform with no TFC markers must NOT trigger the overlay.
	r4, err := Run(writeTree(t, map[string]string{"main.tf": "resource \"aws_s3_bucket\" \"b\" {}\n"}))
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(r4.Packs, "hcp-terraform") {
		t.Errorf("hcp-terraform should not trigger without TFC markers, got %v", r4.Packs)
	}

	// Overlay is emitted after its base in the stable pack order.
	ti, hi := slices.Index(r.Packs, "terraform"), slices.Index(r.Packs, "hcp-terraform")
	if ti < 0 || hi < 0 || hi < ti {
		t.Errorf("want terraform before hcp-terraform in pack order, got %v", r.Packs)
	}
}

func TestDetectOperator(t *testing.T) {
	// Detection keys on a path containing "/api/" (i.e. api nested under another
	// dir) plus a controller-runtime dependency in go.mod.
	files := map[string]string{
		"go.mod":                     "module x\nrequire sigs.k8s.io/controller-runtime v0.19.0\n",
		"pkg/api/v1/widget_types.go": "package v1\n",
	}
	r, err := Run(writeTree(t, files))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(r.Packs, "operator") {
		t.Errorf("want operator pack, got %v", r.Packs)
	}
	// A go.mod without controller-runtime must NOT trigger operator.
	r2, err := Run(writeTree(t, map[string]string{
		"go.mod":                     "module x\n",
		"pkg/api/v1/widget_types.go": "package v1\n",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(r2.Packs, "operator") {
		t.Errorf("operator should not trigger without controller-runtime, got %v", r2.Packs)
	}
}
