package wrap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Options configures the wrap command behavior.
type Options struct {
	Args      []string // Command and arguments to execute
	Name      string   // Human-readable name (defaults to command basename)
	OnSuccess string   // Custom success message template
	OnFailure string   // Custom failure message template
	Quiet     bool     // Suppress nownow's own output
	PushFn    func(msg string) error // Function to push status
}

// Run executes the wrapped command, pushes status, and returns the exit code.
func Run(opts Options) int {
	if len(opts.Args) == 0 {
		fmt.Fprintln(os.Stderr, "nownow wrap: no command specified")
		return 1
	}

	name := opts.Name
	if name == "" {
		name = filepath.Base(opts.Args[0])
	}

	start := time.Now()
	cmd := exec.Command(opts.Args[0], opts.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Command not found or other exec error
			fmt.Fprintf(os.Stderr, "nownow wrap: %v\n", err)
			exitCode = 1
		}
	}

	msg := buildMessage(opts, name, exitCode, duration)

	if opts.PushFn != nil {
		if pushErr := opts.PushFn(msg); pushErr != nil {
			if !opts.Quiet {
				fmt.Fprintf(os.Stderr, "nownow wrap: push failed: %v\n", pushErr)
			}
		} else if !opts.Quiet {
			fmt.Fprintf(os.Stderr, "pushed: %s\n", msg)
		}
	}

	return exitCode
}

func buildMessage(opts Options, name string, exitCode int, duration time.Duration) string {
	var tmpl string
	if exitCode == 0 {
		tmpl = opts.OnSuccess
		if tmpl == "" {
			tmpl = "{name} completed"
		}
	} else {
		tmpl = opts.OnFailure
		if tmpl == "" {
			tmpl = "{name} failed (exit {exit_code})"
		}
	}

	return expandVars(tmpl, opts.Args, name, exitCode, duration)
}

func expandVars(tmpl string, args []string, name string, exitCode int, duration time.Duration) string {
	r := strings.NewReplacer(
		"{cmd}", strings.Join(args, " "),
		"{name}", name,
		"{exit_code}", fmt.Sprintf("%d", exitCode),
		"{duration}", formatDuration(duration),
	)
	return r.Replace(tmpl)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if s == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dm%ds", m, s)
}
