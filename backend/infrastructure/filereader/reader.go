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
const expandDirectories = true // Flag to enable directory expansion

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
	// Guard against nil context and progress to avoid panics
	if ctx == nil {
		ctx = context.Background()
	}
	if progress == nil {
		progress = func(current, total int64) {}
	}

	contents := make(map[string]string)
	var totalSize int64

	// First pass: validate paths and calculate total size
	// Map to track inputPath -> absPath for later use
	pathMapping := make(map[string]string)
	validatedPaths := make([]string, 0, len(filePaths))
	
	for _, inputPath := range filePaths {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Get absolute path - if already absolute, use as is; otherwise join with rootDir
		var absPath string
		var err error
		if filepath.IsAbs(inputPath) {
			absPath, err = filepath.Abs(inputPath)
			if err != nil {
				r.log.Warning(fmt.Sprintf("Skipping invalid path %s: %v", inputPath, err))
				continue
			}
		} else {
			absPath, err = r.sanitizeAndAbs(rootDir, inputPath)
			if err != nil {
				r.log.Warning(fmt.Sprintf("Skipping invalid path %s: %v (rootDir: %s)", inputPath, err, rootDir))
				continue
			}
		}
		
		pathMapping[inputPath] = absPath

		info, err := os.Stat(absPath)
		if err != nil {
			r.log.Warning(fmt.Sprintf("Skipping path %s: cannot stat - %v", inputPath, err))
			continue
		}
		
		if info.IsDir() {
			if expandDirectories {
				expandedPaths, err := r.expandDirectory(absPath, rootDir)
				if err != nil {
					r.log.Warning(fmt.Sprintf("Error expanding directory %s: %v", inputPath, err))
					continue
				}
				
				// Add expanded file paths to validatedPaths
				for _, expandedPath := range expandedPaths {
					// Stat the file to check its size
					fileInfo, err := os.Stat(expandedPath)
					if err != nil {
						r.log.Warning(fmt.Sprintf("Skipping file %s: cannot stat - %v", expandedPath, err))
						continue
					}
					
					if fileInfo.Size() > maxFileSize {
						r.log.Warning(fmt.Sprintf("Skipping large file %s: size %d > %d", expandedPath, fileInfo.Size(), maxFileSize))
						continue
					}
					
					// Convert absolute path to relative path for consistent keys
					relPath, err := filepath.Rel(rootDir, expandedPath)
					if err != nil {
						r.log.Warning(fmt.Sprintf("Cannot get relative path for %s: %v", expandedPath, err))
						relPath = expandedPath // fallback to absolute
					}
					// Normalize to forward slashes for cross-platform consistency
					relPath = filepath.ToSlash(relPath)
					
					totalSize += fileInfo.Size()
					validatedPaths = append(validatedPaths, relPath)
					pathMapping[relPath] = expandedPath
				}
				
				r.log.Info(fmt.Sprintf("Expanded directory %s: found %d files", inputPath, len(expandedPaths)))
			} else {
				r.log.Warning(fmt.Sprintf("Skipping directory (directory expansion disabled): %s", inputPath))
				continue
			}
		} else {
			if info.Size() > maxFileSize {
				r.log.Warning(fmt.Sprintf("Skipping large file %s: size %d > %d", inputPath, info.Size(), maxFileSize))
				continue
			}

			totalSize += info.Size()
			validatedPaths = append(validatedPaths, inputPath)
		}
	}

	var currentSize int64
	for _, inputPath := range validatedPaths {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Get absolute path from mapping, fallback to inputPath if it's already absolute
		absPath := pathMapping[inputPath]
		if absPath == "" {
			// inputPath might already be an absolute path (e.g., from directory expansion)
			if filepath.IsAbs(inputPath) {
				absPath = inputPath
			} else {
				r.log.Warning(fmt.Sprintf("Skipping file %s: no path mapping found", inputPath))
				continue
			}
		}
		
		// Read file using absolute path
		data, err := os.ReadFile(absPath)
		if err != nil {
			r.log.Warning(fmt.Sprintf("Skipping file %s: read error - %v", inputPath, err))
			continue
		}

		// Use original inputPath as key in contents map (maintains contract)
		contents[inputPath] = string(data)
		currentSize += int64(len(data))
		progress(currentSize, totalSize)
	}

	return contents, nil
}

// sanitizeAndAbs converts path to absolute, allowing files from any location
func (r *secureFileReader) sanitizeAndAbs(rootDir, relPath string) (string, error) {
	// If path is already absolute, use it directly
	if filepath.IsAbs(relPath) {
		absPath, err := filepath.Abs(relPath)
		if err != nil {
			return "", fmt.Errorf("could not get absolute path: %w", err)
		}
		return absPath, nil
	}

	// Otherwise, join with rootDir
	cleanRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path for root: %w", err)
	}
	
	absPath := filepath.Join(cleanRootDir, relPath)
	absPath, err = filepath.Abs(absPath)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path: %w", err)
	}
	
	return absPath, nil
}

// expandDirectory recursively walks through a directory and returns all file paths
func (r *secureFileReader) expandDirectory(dirPath, rootDir string) ([]string, error) {
	var files []string

// Limit recursion depth to prevent DoS
const maxDepth = 20
const maxFiles = 10000 // Limit total files to prevent DoS

err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
	if err != nil {
		r.log.Warning(fmt.Sprintf("Error accessing path %s: %v", path, err))
		return nil // Continue walking despite errors
	}
	
	// Check if we've reached the file limit
	if len(files) >= maxFiles {
		r.log.Warning(fmt.Sprintf("Stopping directory expansion: max files limit (%d) reached", maxFiles))
		return filepath.SkipAll // Stop walking completely
	}
	
	// Skip directories, only add files
	if d.IsDir() {
		// Check recursion depth
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return nil // Continue walking
		}
		
		// Count separators to determine depth
		depth := strings.Count(relPath, string(filepath.Separator)) + 1
		if depth > maxDepth {
			r.log.Warning(fmt.Sprintf("Skipping directory %s: max depth exceeded", path))
			return filepath.SkipDir
		}
		
		return nil
	}
	
	// Add the file path to the list
	files = append(files, path)
	
	return nil
})

return files, err
}
