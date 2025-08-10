package filereader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"strings"
)

const maxFileSize = 5 * 1024 * 1024 // 5 MB limit per file

type secureFileReader struct {
	log domain.Logger
}

func NewSecureFileReader(log domain.Logger) domain.FileContentReader {
	return &secureFileReader{log: log}
}

func (r *secureFileReader) ReadContents(
	ctx context.Context,
	filePaths []string,
	rootDir string,
	progress func(current, total int64),
) (map[string]string, error) {
	contents := make(map[string]string)
	var totalSize int64

	// First pass: validate paths and calculate total size
	validatedPaths := make([]string, 0, len(filePaths))
	for _, relPath := range filePaths {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		absPath, err := r.sanitizeAndAbs(rootDir, relPath)
		if err != nil {
			r.log.Warning(fmt.Sprintf("Skipping invalid path %s: %v", relPath, err))
			continue
		}

		info, err := os.Stat(absPath)
		if err != nil {
			r.log.Warning(fmt.Sprintf("Skipping file %s: cannot stat - %v", relPath, err))
			continue
		}
		if info.IsDir() {
			r.log.Warning(fmt.Sprintf("Skipping directory: %s", relPath))
			continue
		}
		if info.Size() > maxFileSize {
			r.log.Warning(fmt.Sprintf("Skipping large file %s: size %d > %d", relPath, info.Size(), maxFileSize))
			continue
		}

		totalSize += info.Size()
		validatedPaths = append(validatedPaths, relPath)
	}

	var currentSize int64
	for _, relPath := range validatedPaths {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		absPath, _ := r.sanitizeAndAbs(rootDir, relPath) // Error already checked
		data, err := os.ReadFile(absPath)
		if err != nil {
			r.log.Warning(fmt.Sprintf("Skipping file %s: read error - %v", relPath, err))
			continue
		}

		contents[relPath] = string(data)
		currentSize += int64(len(data))
		progress(currentSize, totalSize)
	}

	return contents, nil
}

// sanitizeAndAbs ensures the path is within the rootDir to prevent traversal attacks.
func (r *secureFileReader) sanitizeAndAbs(rootDir, relPath string) (string, error) {
	cleanRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path for root: %w", err)
	}

	absPath := filepath.Join(cleanRootDir, relPath)

	if !strings.HasPrefix(absPath, cleanRootDir) {
		return "", fmt.Errorf("path traversal attempt detected: %s", relPath)
	}
	return absPath, nil
}
