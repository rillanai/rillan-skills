// SPDX-FileCopyrightText: 2026 Rillan AI LLC
// SPDX-License-Identifier: Apache-2.0

package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ReplaceExecutable atomically replaces destPath with newBin (mode 0755). It
// writes a temp file in the same directory — so the final swap is a rename on
// the same filesystem — then renames it into place. On Windows, where a running
// image cannot be overwritten directly, the current file is moved aside first
// and the new one renamed in.
func ReplaceExecutable(destPath string, newBin []byte) error {
	dir := filepath.Dir(destPath)
	tmp, err := os.CreateTemp(dir, ".rillan-skills-update-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s (need write access to upgrade in place): %w", dir, err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }

	if _, err := tmp.Write(newBin); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Chmod(tmpName, 0o755); err != nil { //nolint:gosec // a CLI binary must be executable
		cleanup()
		return err
	}

	if runtime.GOOS == "windows" {
		old := destPath + ".old"
		_ = os.Remove(old)
		if err := os.Rename(destPath, old); err != nil {
			cleanup()
			return err
		}
		if err := os.Rename(tmpName, destPath); err != nil {
			_ = os.Rename(old, destPath) // roll back the move-aside
			cleanup()
			return err
		}
		_ = os.Remove(old) // best effort; may stay locked until the old process exits
		return nil
	}

	if err := os.Rename(tmpName, destPath); err != nil {
		cleanup()
		return err
	}
	return nil
}
