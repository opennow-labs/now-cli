//go:build !windows

package cmd

import "os"

// removeSelf deletes the running binary. On Unix this is safe because the
// kernel keeps the inode alive until the process exits.
// The second return value is always empty on Unix (no pending cleanup).
func removeSelf(exe string) (pendingCleanup string, err error) {
	return "", os.Remove(exe)
}
