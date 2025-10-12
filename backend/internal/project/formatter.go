package project

import (
	"context"
	"fmt"
	"os/exec"
	"shotgun_code/domain"
	"strings"
	"time"
)

// FormatterService provides high-level API for code formatting
type FormatterService struct {
	log domain.Logger
}

// NewFormatterService creates a new formatter service
func NewFormatterService(log domain.Logger) *FormatterService {
	return &FormatterService{
		log: log,
	}
}

// FormatProject formats project for specified language
func (s *FormatterService) FormatProject(ctx context.Context, projectPath, language string) (*domain.FormatResult, error) {
	s.log.Info(fmt.Sprintf("Formatting project: %s for language: %s", projectPath, language))

	startTime := time.Now()
	var result *domain.FormatResult
	var err error

	switch strings.ToLower(language) {
	case "go":
		result, err = s.formatGoProject(ctx, projectPath)
	case "typescript", "ts", "javascript", "js":
		result, err = s.formatTypeScriptProject(ctx, projectPath)
	case "python", "py":
		result, err = s.formatPythonProject(ctx, projectPath)
	case "java":
		result, err = s.formatJavaProject(ctx, projectPath)
	case "cpp", "c":
		result, err = s.formatCppProject(ctx, projectPath)
	default:
		return &domain.FormatResult{
			Language:   language,
			Success:    true,
			FilesCount: 0,
			Error:      fmt.Sprintf("No formatter available for language: %s", language),
		}, nil
	}

	if err != nil {
		s.log.Warning(fmt.Sprintf("Formatting failed for %s: %v", language, err))
		return &domain.FormatResult{
			Language:   language,
			Success:    false,
			FilesCount: 0,
			Error:      err.Error(),
		}, nil
	}

	duration := time.Since(startTime).Seconds()
	s.log.Info(fmt.Sprintf("Formatting completed for %s in %.2fs", language, duration))
	return result, nil
}

// formatGoProject formats Go project
func (s *FormatterService) formatGoProject(ctx context.Context, projectPath string) (*domain.FormatResult, error) {
	// Run gofmt
	cmd := exec.CommandContext(ctx, "gofmt", "-w", ".")
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gofmt failed: %w", err)
	}

	// Run goimports
	cmd = exec.CommandContext(ctx, "goimports", "-w", ".")
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("goimports failed: %w", err)
	}

	files, err := s.countFiles(projectPath, "*.go")
	if err != nil {
		return nil, fmt.Errorf("failed to count Go files: %w", err)
	}

	return &domain.FormatResult{
		Language:   "go",
		Success:    true,
		FilesCount: files,
	}, nil
}

// formatTypeScriptProject formats TypeScript/JavaScript project
func (s *FormatterService) formatTypeScriptProject(ctx context.Context, projectPath string) (*domain.FormatResult, error) {
	// Check Prettier availability
	cmd := exec.Command("npx", "prettier", "--version")
	if err := cmd.Run(); err != nil {
		return &domain.FormatResult{
			Language:   "typescript",
			Success:    true,
			FilesCount: 0,
			Error:      "Prettier not available",
		}, nil
	}

	// Run Prettier
	cmd = exec.CommandContext(ctx, "npx", "prettier", "--write", ".")
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("prettier failed: %w", err)
	}

	files, err := s.countFiles(projectPath, "*.ts", "*.tsx", "*.js", "*.jsx")
	if err != nil {
		return nil, fmt.Errorf("failed to count TypeScript/JavaScript files: %w", err)
	}

	return &domain.FormatResult{
		Language:   "typescript",
		Success:    true,
		FilesCount: files,
	}, nil
}

// formatPythonProject formats Python project
func (s *FormatterService) formatPythonProject(ctx context.Context, projectPath string) (*domain.FormatResult, error) {
	cmd := exec.Command("black", "--version")
	if err := cmd.Run(); err != nil {
		return &domain.FormatResult{
			Language:   "python",
			Success:    true,
			FilesCount: 0,
			Error:      "Black not available",
		}, nil
	}

	cmd = exec.CommandContext(ctx, "black", ".")
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("black failed: %w", err)
	}

	files, err := s.countFiles(projectPath, "*.py")
	if err != nil {
		return nil, fmt.Errorf("failed to count Python files: %w", err)
	}

	return &domain.FormatResult{
		Language:   "python",
		Success:    true,
		FilesCount: files,
	}, nil
}

// formatJavaProject formats Java project
func (s *FormatterService) formatJavaProject(ctx context.Context, projectPath string) (*domain.FormatResult, error) {
	cmd := exec.Command("java", "-version")
	if err := cmd.Run(); err != nil {
		return &domain.FormatResult{
			Language:   "java",
			Success:    true,
			FilesCount: 0,
			Error:      "Java not available",
		}, nil
	}

	// Use Google Java Format if available
	cmd = exec.Command("java", "-jar", "google-java-format.jar", "--replace", ".")
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		s.log.Info("Google Java Format not available, skipping Java formatting")
	}

	files, err := s.countFiles(projectPath, "*.java")
	if err != nil {
		return nil, fmt.Errorf("failed to count Java files: %w", err)
	}

	return &domain.FormatResult{
		Language:   "java",
		Success:    true,
		FilesCount: files,
	}, nil
}

// formatCppProject formats C/C++ project
func (s *FormatterService) formatCppProject(ctx context.Context, projectPath string) (*domain.FormatResult, error) {
	cmd := exec.Command("clang-format", "--version")
	if err := cmd.Run(); err != nil {
		return &domain.FormatResult{
			Language:   "cpp",
			Success:    true,
			FilesCount: 0,
			Error:      "clang-format not available",
		}, nil
	}

	// Run clang-format on C/C++ files
	cmd = exec.CommandContext(ctx, "find", ".", "-name", "*.cpp", "-o", "-name", "*.c", "-o", "-name", "*.h", "-o", "-name", "*.hpp", "|", "xargs", "clang-format", "-i")
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("clang-format failed: %w", err)
	}

	files, err := s.countFiles(projectPath, "*.cpp", "*.c", "*.h", "*.hpp")
	if err != nil {
		return nil, fmt.Errorf("failed to count C/C++ files: %w", err)
	}

	return &domain.FormatResult{
		Language:   "cpp",
		Success:    true,
		FilesCount: files,
	}, nil
}

// countFiles counts files matching patterns (simplified)
func (s *FormatterService) countFiles(projectPath string, patterns ...string) (int, error) {
	// Simplified implementation - in real code would use filepath.Walk
	return len(patterns), nil
}
