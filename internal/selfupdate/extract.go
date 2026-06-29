// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"path"
	"strings"
)

// ExtractBinary returns the bytes of binName from a release archive. assetName
// selects the format: a ".zip" suffix means a zip archive, otherwise tar.gz.
// Reads are bounded by maxDownload to cap decompression.
func ExtractBinary(data []byte, assetName, binName string) ([]byte, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractZip(data, binName)
	}
	return extractTarGz(data, binName)
}

func extractTarGz(data []byte, binName string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg || path.Base(hdr.Name) != binName {
			continue
		}
		return io.ReadAll(io.LimitReader(tr, maxDownload))
	}
	return nil, fmt.Errorf("binary %q not found in archive", binName)
}

func extractZip(data []byte, binName string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for _, f := range zr.File {
		if path.Base(f.Name) != binName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(io.LimitReader(rc, maxDownload))
		_ = rc.Close()
		return b, err
	}
	return nil, fmt.Errorf("binary %q not found in archive", binName)
}
