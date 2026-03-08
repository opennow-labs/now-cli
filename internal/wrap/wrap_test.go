package wrap

import (
	"strings"
	"testing"
	"time"
)

func TestRunSuccess(t *testing.T) {
	var pushed string
	exitCode := Run(Options{
		Args:  []string{"echo", "hello"},
		Quiet: true,
		PushFn: func(msg string) error {
			pushed = msg
			return nil
		},
	})

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if pushed != "echo completed" {
		t.Errorf("expected 'echo completed', got %q", pushed)
	}
}

func TestRunFailure(t *testing.T) {
	var pushed string
	exitCode := Run(Options{
		Args:  []string{"false"},
		Quiet: true,
		PushFn: func(msg string) error {
			pushed = msg
			return nil
		},
	})

	if exitCode == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(pushed, "failed") {
		t.Errorf("expected failure message, got %q", pushed)
	}
}

func TestRunCustomMessages(t *testing.T) {
	var pushed string
	Run(Options{
		Args:      []string{"echo", "hi"},
		Name:      "My Task",
		OnSuccess: "Done: {name} in {duration}",
		Quiet:     true,
		PushFn: func(msg string) error {
			pushed = msg
			return nil
		},
	})

	if !strings.HasPrefix(pushed, "Done: My Task in ") {
		t.Errorf("unexpected message: %q", pushed)
	}
}

func TestRunCustomFailureMessage(t *testing.T) {
	var pushed string
	Run(Options{
		Args:      []string{"false"},
		Name:      "backup",
		OnFailure: "{name} broke with code {exit_code}",
		Quiet:     true,
		PushFn: func(msg string) error {
			pushed = msg
			return nil
		},
	})

	if pushed != "backup broke with code 1" {
		t.Errorf("unexpected message: %q", pushed)
	}
}

func TestRunNoCommand(t *testing.T) {
	exitCode := Run(Options{
		Quiet: true,
	})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
}

func TestRunWithCmdVariable(t *testing.T) {
	var pushed string
	Run(Options{
		Args:      []string{"echo", "hello", "world"},
		OnSuccess: "ran: {cmd}",
		Quiet:     true,
		PushFn: func(msg string) error {
			pushed = msg
			return nil
		},
	})

	if pushed != "ran: echo hello world" {
		t.Errorf("unexpected message: %q", pushed)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{3 * time.Second, "3s"},
		{59 * time.Second, "59s"},
		{60 * time.Second, "1m"},
		{135 * time.Second, "2m15s"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
