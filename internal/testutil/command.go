package testutil

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

type (
	CommandResult struct {
		Dir    string
		Name   string
		Args   []string
		Stdout string
		Stderr string
	}

	StartedCommand struct {
		dir    string
		name   string
		args   []string
		cmd    *exec.Cmd
		stdout *lockedBuffer
		stderr *lockedBuffer
		done   chan error
	}

	lockedBuffer struct {
		mu     sync.Mutex
		buffer bytes.Buffer
	}
)

func RunCommand(t testing.TB, ctx context.Context, dir string, env []string, name string, args ...string) CommandResult {
	t.Helper()

	result, err := RunCommandResult(ctx, dir, env, name, args...)
	if err != nil {
		t.Fatalf("%s", result.FailureMessage(err))
	}
	return result
}

func RunCommandResult(ctx context.Context, dir string, env []string, name string, args ...string) (CommandResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return CommandResult{
		Dir:    dir,
		Name:   name,
		Args:   append([]string(nil), args...),
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}, err
}

func StartCommand(ctx context.Context, dir string, env []string, name string, args ...string) (*StartedCommand, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	started := &StartedCommand{
		dir:    dir,
		name:   name,
		args:   append([]string(nil), args...),
		cmd:    cmd,
		stdout: &lockedBuffer{},
		stderr: &lockedBuffer{},
		done:   make(chan error, 1),
	}
	cmd.Stdout = started.stdout
	cmd.Stderr = started.stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		started.done <- cmd.Wait()
		close(started.done)
	}()

	return started, nil
}

func (c *StartedCommand) Done() <-chan error {
	if c == nil {
		done := make(chan error)
		close(done)
		return done
	}
	return c.done
}

func (c *StartedCommand) Stop(gracePeriod time.Duration) error {
	if c == nil || c.cmd == nil || c.cmd.Process == nil {
		return nil
	}

	select {
	case err := <-c.done:
		return err
	default:
	}

	_ = c.cmd.Process.Signal(os.Interrupt)

	timer := time.NewTimer(gracePeriod)
	defer timer.Stop()

	select {
	case err := <-c.done:
		return err
	case <-timer.C:
		_ = c.cmd.Process.Kill()
		err := <-c.done
		return fmt.Errorf("command did not stop within %s: %w", gracePeriod, err)
	}
}

func (c *StartedCommand) Output() string {
	if c == nil {
		return ""
	}

	result := CommandResult{
		Dir:    c.dir,
		Name:   c.name,
		Args:   c.args,
		Stdout: c.stdout.String(),
		Stderr: c.stderr.String(),
	}
	return result.String()
}

func (r CommandResult) FailureMessage(err error) string {
	return fmt.Sprintf("command failed: %s\nerror: %v\n%s", r.CommandLine(), err, r.String())
}

func (r CommandResult) CommandLine() string {
	parts := append([]string{r.Name}, r.Args...)
	return strings.Join(parts, " ")
}

func (r CommandResult) String() string {
	return fmt.Sprintf("dir: %s\nstdout:\n%s\nstderr:\n%s", r.Dir, r.Stdout, r.Stderr)
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
}
