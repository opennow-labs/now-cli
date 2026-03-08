package logging

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	logFileName = "nownow.log"
	maxLogSize  = 1 << 20 // 1 MB
)

// Init sets up structured logging to a file in configDir.
// Creates the directory if it doesn't exist.
// Truncates the log file if it exceeds 1 MB.
// Set NOWNOW_DEBUG=1 for debug-level output.
func Init(configDir string) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	logPath := filepath.Join(configDir, logFileName)

	// Simple rotation: keep the last half when file exceeds maxLogSize
	if info, err := os.Stat(logPath); err == nil && info.Size() > maxLogSize {
		if data, err := os.ReadFile(logPath); err == nil {
			half := len(data) / 2
			// Advance to the next newline so we don't start mid-line
			if i := bytes.IndexByte(data[half:], '\n'); i >= 0 {
				half += i + 1
			}
			if err := os.WriteFile(logPath, data[half:], 0600); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to rotate log: %v\n", err)
			}
		}
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	// Write to both file and stderr (stderr visible in --foreground mode)
	w := io.MultiWriter(f, os.Stderr)

	level := slog.LevelInfo
	if os.Getenv("NOWNOW_DEBUG") == "1" {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))

	return nil
}
