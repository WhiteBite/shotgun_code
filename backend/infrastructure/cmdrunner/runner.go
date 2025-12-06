package cmdrunner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// Runner implements command execution with safety limits
type Runner struct {
	// Default timeout for commands
	defaultTimeout time.Duration

	// Semaphore for limiting concurrent commands
	sem chan struct{}

	// Buffer pool for output
	bufferPool sync.Pool
}

const (
	maxConcurrentCommands = 8
	defaultCommandTimeout = 5 * time.Minute
	maxOutputSize         = 10 * 1024 * 1024 // 10MB
)

// NewRunner creates a new command runner
func NewRunner() *Runner {
	return &Runner{
		defaultTimeout: defaultCommandTimeout,
		sem:            make(chan struct{}, maxConcurrentCommands),
		bufferPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// RunCommand executes a command with context
func (r *Runner) RunCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return r.RunCommandInDir(ctx, "", name, args...)
}

// RunCommandInDir executes a command in a specific directory
func (r *Runner) RunCommandInDir(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	// Acquire semaphore
	select {
	case r.sem <- struct{}{}:
		defer func() { <-r.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Add timeout if not already set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.defaultTimeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	// Get buffers from pool
	stdout := r.getBuffer()
	stderr := r.getBuffer()
	defer r.putBuffer(stdout)
	defer r.putBuffer(stderr)

	cmd.Stdout = &limitedWriter{buf: stdout, limit: maxOutputSize}
	cmd.Stderr = &limitedWriter{buf: stderr, limit: maxOutputSize}

	err := cmd.Run()
	if err != nil {
		// Include stderr in error message
		if stderr.Len() > 0 {
			return stdout.Bytes(), fmt.Errorf("%w: %s", err, stderr.String())
		}
		return stdout.Bytes(), err
	}

	return stdout.Bytes(), nil
}

// RunCommandWithInput executes a command with stdin input
func (r *Runner) RunCommandWithInput(ctx context.Context, input []byte, name string, args ...string) ([]byte, error) {
	// Acquire semaphore
	select {
	case r.sem <- struct{}{}:
		defer func() { <-r.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Add timeout if not already set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.defaultTimeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = bytes.NewReader(input)

	stdout := r.getBuffer()
	stderr := r.getBuffer()
	defer r.putBuffer(stdout)
	defer r.putBuffer(stderr)

	cmd.Stdout = &limitedWriter{buf: stdout, limit: maxOutputSize}
	cmd.Stderr = &limitedWriter{buf: stderr, limit: maxOutputSize}

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return stdout.Bytes(), fmt.Errorf("%w: %s", err, stderr.String())
		}
		return stdout.Bytes(), err
	}

	return stdout.Bytes(), nil
}

// IsCommandAvailable checks if a command is available
func (r *Runner) IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func (r *Runner) getBuffer() *bytes.Buffer {
	buf := r.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func (r *Runner) putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	r.bufferPool.Put(buf)
}

// limitedWriter limits the amount of data written to prevent OOM
type limitedWriter struct {
	buf     *bytes.Buffer
	limit   int
	written int
}

func (w *limitedWriter) Write(p []byte) (n int, err error) {
	if w.written >= w.limit {
		return len(p), nil // Silently discard
	}

	remaining := w.limit - w.written
	if len(p) > remaining {
		p = p[:remaining]
	}

	n, err = w.buf.Write(p)
	w.written += n
	return len(p), err // Return original length to avoid errors
}
