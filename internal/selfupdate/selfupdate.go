// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

// Package selfupdate replaces the running rillan-skills binary with the latest
// release published on GitHub. It is deliberately standard-library only: the
// installer binary ships with no non-stdlib runtime dependencies, and a
// self-updater for an install/security-adjacent tool is the last place to add
// one. The flow is: query the GitHub releases API, pick the archive for this
// GOOS/GOARCH, verify its SHA-256 against the release's checksums.txt, extract
// the binary, and atomically swap it over the current executable.
package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPIBase = "https://api.github.com"
	repoOwner      = "rillanai"
	repoName       = "rillan-skills"
	// binaryBase is the binary name goreleaser packs into each archive (with a
	// ".exe" suffix added on Windows).
	binaryBase = "rillan-skills"
	userAgent  = "rillan-skills-selfupdate"
	// maxDownload caps any single response body to a sane ceiling so a bad or
	// hostile redirect cannot stream unbounded data into memory.
	maxDownload = 256 << 20 // 256 MiB
)

// Asset is one downloadable file attached to a release.
type Asset struct {
	Name string
	URL  string
}

// Release is the subset of a GitHub release this package needs.
type Release struct {
	Tag    string
	Assets []Asset
}

// Options configure an upgrade. Zero values fall back to working defaults;
// APIBase and HTTPClient exist mainly so tests can avoid the real network.
type Options struct {
	CurrentVersion string // built-in version, e.g. "1.1.0" or "dev"
	GOOS           string // runtime.GOOS
	GOARCH         string // runtime.GOARCH
	HTTPClient     *http.Client
	APIBase        string // "" -> https://api.github.com
	Token          string // optional; sent only to the API host (rate limits)
	Logger         func(format string, args ...any)
}

func (o Options) client() *http.Client {
	if o.HTTPClient != nil {
		return o.HTTPClient
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func (o Options) apiBase() string {
	if o.APIBase != "" {
		return strings.TrimRight(o.APIBase, "/")
	}
	return defaultAPIBase
}

// Status is the result of comparing the built-in version to the latest release.
type Status struct {
	Current string // built-in version as-is
	Latest  string // latest release version, "v" stripped
	Newer   bool   // a strictly newer release is available
}

// Check fetches the latest release and reports whether it is newer than the
// running binary. It performs no writes.
func Check(ctx context.Context, o Options) (Status, Release, error) {
	rel, err := LatestRelease(ctx, o)
	if err != nil {
		return Status{}, Release{}, err
	}
	latest := NormalizeVersion(rel.Tag)
	st := Status{
		Current: o.CurrentVersion,
		Latest:  latest,
		Newer:   CompareVersions(latest, NormalizeVersion(o.CurrentVersion)) > 0,
	}
	return st, rel, nil
}

// LatestRelease returns metadata for the repository's latest published release.
func LatestRelease(ctx context.Context, o Options) (Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", o.apiBase(), repoOwner, repoName)
	body, err := o.get(ctx, url, "application/vnd.github+json", true)
	if err != nil {
		return Release{}, fmt.Errorf("fetch latest release: %w", err)
	}
	var raw struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return Release{}, fmt.Errorf("parse release metadata: %w", err)
	}
	if raw.TagName == "" {
		return Release{}, fmt.Errorf("latest release has no tag")
	}
	rel := Release{Tag: raw.TagName}
	for _, a := range raw.Assets {
		rel.Assets = append(rel.Assets, Asset{Name: a.Name, URL: a.URL})
	}
	return rel, nil
}

// Apply downloads the release archive for this platform, verifies its checksum,
// extracts the binary, and atomically replaces destPath (the running
// executable). destPath should already be resolved through any symlinks.
func Apply(ctx context.Context, o Options, rel Release, destPath string) error {
	archive, sums, err := SelectAssets(rel, o.GOOS, o.GOARCH)
	if err != nil {
		return err
	}
	archiveBytes, err := o.get(ctx, archive.URL, "application/octet-stream", false)
	if err != nil {
		return fmt.Errorf("download %s: %w", archive.Name, err)
	}
	sumsBytes, err := o.get(ctx, sums.URL, "text/plain", false)
	if err != nil {
		return fmt.Errorf("download %s: %w", sums.Name, err)
	}
	want, ok := ParseChecksums(sumsBytes)[archive.Name]
	if !ok {
		return fmt.Errorf("%s has no checksum for %s", sums.Name, archive.Name)
	}
	if err := VerifySHA256(archiveBytes, want); err != nil {
		return fmt.Errorf("verify %s: %w", archive.Name, err)
	}
	bin, err := ExtractBinary(archiveBytes, archive.Name, binaryFileName(o.GOOS))
	if err != nil {
		return fmt.Errorf("extract %s: %w", archive.Name, err)
	}
	if err := ReplaceExecutable(destPath, bin); err != nil {
		return fmt.Errorf("replace %s: %w", destPath, err)
	}
	return nil
}

// SelectAssets picks the release archive matching goos/goarch and the
// checksums.txt asset.
func SelectAssets(rel Release, goos, goarch string) (archive, checksums Asset, err error) {
	suffix := fmt.Sprintf("_%s_%s%s", goos, goarch, archiveExt(goos))
	for _, a := range rel.Assets {
		switch {
		case strings.HasSuffix(a.Name, suffix):
			archive = a
		case a.Name == "checksums.txt":
			checksums = a
		}
	}
	if archive.URL == "" {
		return Asset{}, Asset{}, fmt.Errorf("no release asset for %s/%s (looked for *%s)", goos, goarch, suffix)
	}
	if checksums.URL == "" {
		return Asset{}, Asset{}, fmt.Errorf("release has no checksums.txt asset")
	}
	return archive, checksums, nil
}

func archiveExt(goos string) string {
	if goos == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

func binaryFileName(goos string) string {
	if goos == "windows" {
		return binaryBase + ".exe"
	}
	return binaryBase
}

// get performs a GET and returns the body, bounded by maxDownload. token is sent
// only when withToken is true (the API host), never to asset/CDN redirects.
func (o Options) get(ctx context.Context, url, accept string, withToken bool) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil) //nolint:gosec // url is built from our own repo's GitHub release metadata
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", accept)
	if withToken && o.Token != "" {
		req.Header.Set("Authorization", "Bearer "+o.Token)
	}
	resp, err := o.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDownload))
	if err != nil {
		return nil, err
	}
	return body, nil
}

// NormalizeVersion strips a single leading "v" (release tags are "vX.Y.Z").
func NormalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// CompareVersions compares dotted numeric versions ("X.Y.Z"). Non-numeric or
// unparsable fields (e.g. "dev") sort as 0. Returns -1, 0, or 1.
func CompareVersions(a, b string) int {
	as, bs := splitVersion(a), splitVersion(b)
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

func splitVersion(v string) [3]int {
	var out [3]int
	parts := strings.SplitN(v, ".", 3)
	for i := 0; i < len(parts) && i < 3; i++ {
		n, _ := strconv.Atoi(strings.TrimSpace(parts[i]))
		out[i] = n
	}
	return out
}
