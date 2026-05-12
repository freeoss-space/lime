package managers

import (
	"context"
	"os"
	"os/exec"
)

// Commander abstracts shell command execution, enabling test doubles.
type Commander interface {
	// Run executes the command, streaming stdout/stderr to the terminal.
	Run(ctx context.Context, name string, args ...string) error
	// Output executes the command and captures its combined output.
	Output(ctx context.Context, name string, args ...string) ([]byte, error)
	// LookPath reports whether name is available on PATH.
	LookPath(name string) (string, error)
}

// RealCommander executes commands for real.
type RealCommander struct{}

func (RealCommander) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (RealCommander) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

func (RealCommander) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// FakeCommander is a test double that records calls and returns preset results.
type FakeCommander struct {
	// RunErr is returned by Run if non-nil.
	RunErr error
	// OutputData is returned by Output.
	OutputData []byte
	// OutputErr is returned by Output.
	OutputErr error
	// LookPathResult maps binary names to their expected paths.
	LookPathResult map[string]string
	// LookPathErr is returned when a name is not in LookPathResult.
	LookPathErr error
	// Calls records all (name, args) pairs passed to Run.
	Calls [][]string
}

func (f *FakeCommander) Run(_ context.Context, name string, args ...string) error {
	f.Calls = append(f.Calls, append([]string{name}, args...))
	return f.RunErr
}

func (f *FakeCommander) Output(_ context.Context, name string, args ...string) ([]byte, error) {
	f.Calls = append(f.Calls, append([]string{name}, args...))
	return f.OutputData, f.OutputErr
}

func (f *FakeCommander) LookPath(name string) (string, error) {
	if f.LookPathResult != nil {
		if p, ok := f.LookPathResult[name]; ok {
			return p, nil
		}
	}
	if f.LookPathErr != nil {
		return "", f.LookPathErr
	}
	return "", &exec.Error{Name: name, Err: exec.ErrNotFound}
}
