// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.2.0", "1.1.0", 1},
		{"1.1.0", "1.2.0", -1},
		{"1.1.0", "1.1.0", 0},
		{"2.0.0", "1.9.9", 1},
		{"1.1.0", "dev", 1}, // dev sorts as 0.0.0
		{"0.0.0", "dev", 0},
	}
	for _, c := range cases {
		if got := CompareVersions(c.a, c.b); got != c.want {
			t.Errorf("CompareVersions(%q,%q)=%d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestNormalizeVersion(t *testing.T) {
	for in, want := range map[string]string{"v1.2.0": "1.2.0", "1.2.0": "1.2.0", " v0.1.0 ": "0.1.0"} {
		if got := NormalizeVersion(in); got != want {
			t.Errorf("NormalizeVersion(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestParseChecksums(t *testing.T) {
	data := []byte("abc123  file_one.tar.gz\nDEF456 *file_two.zip\n\nonlyonefield\n")
	got := ParseChecksums(data)
	if got["file_one.tar.gz"] != "abc123" {
		t.Errorf("file_one digest = %q", got["file_one.tar.gz"])
	}
	if got["file_two.zip"] != "def456" { // lowercased, "*" stripped
		t.Errorf("file_two digest = %q", got["file_two.zip"])
	}
	if len(got) != 2 {
		t.Errorf("want 2 entries, got %d: %v", len(got), got)
	}
}

func TestVerifySHA256(t *testing.T) {
	data := []byte("hello")
	sum := sha256.Sum256(data)
	if err := VerifySHA256(data, hex.EncodeToString(sum[:])); err != nil {
		t.Errorf("matching checksum should pass: %v", err)
	}
	if err := VerifySHA256(data, "deadbeef"); err == nil {
		t.Error("mismatched checksum should fail")
	}
}

func TestSelectAssets(t *testing.T) {
	rel := Release{Assets: []Asset{
		{Name: "rillan-skills_1.2.0_linux_amd64.tar.gz", URL: "u1"},
		{Name: "rillan-skills_1.2.0_windows_amd64.zip", URL: "u2"},
		{Name: "checksums.txt", URL: "u3"},
	}}
	a, s, err := SelectAssets(rel, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if a.URL != "u1" || s.URL != "u3" {
		t.Errorf("got archive=%q checksums=%q", a.URL, s.URL)
	}
	if _, _, err := SelectAssets(rel, "darwin", "arm64"); err == nil {
		t.Error("missing platform asset should error")
	}
}

func TestExtractBinaryTarGz(t *testing.T) {
	want := []byte("#!/fake binary\x00\x01")
	data := makeTarGz(t, "rillan-skills", want)
	got, err := ExtractBinary(data, "rillan-skills_1.2.0_linux_amd64.tar.gz", "rillan-skills")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("extracted %q, want %q", got, want)
	}
	if _, err := ExtractBinary(data, "x.tar.gz", "nonesuch"); err == nil {
		t.Error("missing binary should error")
	}
}

func TestExtractBinaryZip(t *testing.T) {
	want := []byte("MZ fake exe")
	data := makeZip(t, "rillan-skills.exe", want)
	got, err := ExtractBinary(data, "rillan-skills_1.2.0_windows_amd64.zip", "rillan-skills.exe")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("extracted %q, want %q", got, want)
	}
}

func TestReplaceExecutable(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "rillan-skills")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil { //nolint:gosec // test fixture
		t.Fatal(err)
	}
	newBin := []byte("new binary contents")
	if err := ReplaceExecutable(dest, newBin); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dest) //nolint:gosec // test path
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, newBin) {
		t.Errorf("dest = %q, want %q", got, newBin)
	}
	// no stray temp files left behind
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("expected only the replaced binary, got %d entries", len(entries))
	}
}

func TestCheckAndApply(t *testing.T) {
	binContent := []byte("the new rillan-skills binary")
	archiveName := "rillan-skills_1.2.0_linux_amd64.tar.gz"
	archive := makeTarGz(t, "rillan-skills", binContent)
	sum := sha256.Sum256(archive)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), archiveName)

	mux := http.NewServeMux()
	var base string
	mux.HandleFunc("/repos/rillanai/rillan-skills/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"tag_name":"v1.2.0","assets":[
			{"name":%q,"browser_download_url":%q},
			{"name":"checksums.txt","browser_download_url":%q}]}`,
			archiveName, base+"/dl/archive", base+"/dl/checksums")
	})
	mux.HandleFunc("/dl/archive", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive) })
	mux.HandleFunc("/dl/checksums", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(checksums)) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	base = srv.URL

	opts := Options{
		CurrentVersion: "1.1.0",
		GOOS:           "linux",
		GOARCH:         "amd64",
		APIBase:        srv.URL,
		HTTPClient:     srv.Client(),
	}

	st, rel, err := Check(context.Background(), opts)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Newer || st.Latest != "1.2.0" {
		t.Fatalf("Check = %+v, want newer 1.2.0", st)
	}

	dest := filepath.Join(t.TempDir(), "rillan-skills")
	if err := os.WriteFile(dest, []byte("old binary"), 0o755); err != nil { //nolint:gosec // test fixture
		t.Fatal(err)
	}
	if err := Apply(context.Background(), opts, rel, dest); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dest) //nolint:gosec // test path
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, binContent) {
		t.Errorf("after Apply, dest = %q, want %q", got, binContent)
	}
}

func TestApplyChecksumMismatch(t *testing.T) {
	archiveName := "rillan-skills_1.2.0_linux_amd64.tar.gz"
	archive := makeTarGz(t, "rillan-skills", []byte("real"))

	mux := http.NewServeMux()
	mux.HandleFunc("/dl/archive", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive) })
	mux.HandleFunc("/dl/checksums", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "%064x  %s\n", 0, archiveName) // wrong digest
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	rel := Release{Tag: "v1.2.0", Assets: []Asset{
		{Name: archiveName, URL: srv.URL + "/dl/archive"},
		{Name: "checksums.txt", URL: srv.URL + "/dl/checksums"},
	}}
	opts := Options{CurrentVersion: "1.1.0", GOOS: "linux", GOARCH: "amd64", HTTPClient: srv.Client()}
	dest := filepath.Join(t.TempDir(), "rillan-skills")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil { //nolint:gosec // test fixture
		t.Fatal(err)
	}
	if err := Apply(context.Background(), opts, rel, dest); err == nil {
		t.Fatal("Apply should fail on checksum mismatch")
	}
	got, _ := os.ReadFile(dest) //nolint:gosec // test path
	if string(got) != "old" {
		t.Errorf("dest should be untouched on failure, got %q", got)
	}
}

func makeTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func makeZip(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
