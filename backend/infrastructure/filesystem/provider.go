package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Provider implements file system operations with safety checks and pooling
type Provider struct {
	// Buffer pool for I/O operations
	bufferPool sync.Pool
}

// NewProvider creates a new filesystem provider
func NewProvider() *Provider {
	return &Provider{
		bufferPool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 32*1024) // 32KB buffers
				return &buf
			},
		},
	}
}

// ReadFile reads file content with buffered I/O
func (p *Provider) ReadFile(filename string) ([]byte, error) {
	// Validate path
	if err := p.validatePath(filename); err != nil {
		return nil, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size for pre-allocation
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Limit file size to prevent OOM
	const maxFileSize = 100 * 1024 * 1024 // 100MB
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d)", info.Size(), maxFileSize)
	}

	// Pre-allocate buffer
	data := make([]byte, 0, info.Size())
	buf := p.getBuffer()
	defer p.putBuffer(buf)

	for {
		n, err := file.Read(*buf)
		if n > 0 {
			data = append(data, (*buf)[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
	}

	return data, nil
}

// WriteFile writes data to file atomically
func (p *Provider) WriteFile(filename string, data []byte, perm int) error {
	// Validate path
	if err := p.validatePath(filename); err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to temp file first for atomicity
	tempFile := filename + ".tmp"
	if err := os.WriteFile(tempFile, data, os.FileMode(perm)); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Rename for atomic update
	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile) // Cleanup on failure
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// MkdirAll creates directory with all parents
func (p *Provider) MkdirAll(path string, perm int) error {
	if err := p.validatePath(path); err != nil {
		return err
	}
	return os.MkdirAll(path, os.FileMode(perm))
}

// Remove removes a file
func (p *Provider) Remove(name string) error {
	if err := p.validatePath(name); err != nil {
		return err
	}
	return os.Remove(name)
}

// RemoveAll removes a directory and all contents
func (p *Provider) RemoveAll(path string) error {
	if err := p.validatePath(path); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

// Stat returns file info
func (p *Provider) Stat(name string) (os.FileInfo, error) {
	if err := p.validatePath(name); err != nil {
		return nil, err
	}
	return os.Stat(name)
}

// Exists checks if file exists
func (p *Provider) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// validatePath validates file path for security
func (p *Provider) validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}

	// Clean path to prevent traversal attacks
	cleaned := filepath.Clean(path)

	// Check for suspicious patterns
	if filepath.IsAbs(cleaned) {
		// Allow absolute paths but check for dangerous locations
		dangerous := []string{"/etc", "/usr", "/bin", "/sbin", "/var", "/root"}
		for _, d := range dangerous {
			if len(cleaned) >= len(d) && cleaned[:len(d)] == d {
				return fmt.Errorf("access to system directory not allowed: %s", d)
			}
		}
	}

	return nil
}

// getBuffer gets a buffer from pool
func (p *Provider) getBuffer() *[]byte {
	return p.bufferPool.Get().(*[]byte)
}

// putBuffer returns buffer to pool
func (p *Provider) putBuffer(buf *[]byte) {
	p.bufferPool.Put(buf)
}

// CopyFile copies a file
func (p *Provider) CopyFile(src, dst string) error {
	if err := p.validatePath(src); err != nil {
		return err
	}
	if err := p.validatePath(dst); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	buf := p.getBuffer()
	defer p.putBuffer(buf)

	_, err = io.CopyBuffer(dstFile, srcFile, *buf)
	if err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	return dstFile.Sync()
}
