package modification

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"shotgun_code/domain"
	"sort"
	"strings"
	"time"
)

// RepairService implements domain.RepairService for code repair operations
type RepairService struct {
	log domain.Logger
}

// NewRepairService creates a new repair service
func NewRepairService(log domain.Logger) *RepairService {
	return &RepairService{
		log: log,
	}
}

// ExecuteRepair performs a repair cycle to fix code issues
func (s *RepairService) ExecuteRepair(ctx context.Context, req domain.RepairRequest) (*domain.RepairResult, error) {
	startTime := time.Now()
	s.log.Info(fmt.Sprintf("Starting repair cycle for project: %s", req.ProjectPath))

	// Check if project exists
	if _, err := os.Stat(req.ProjectPath); os.IsNotExist(err) {
		return &domain.RepairResult{
			Success: false,
			Error:   fmt.Sprintf("project path does not exist: %s", req.ProjectPath),
		}, nil
	}

	// Get available rules if none specified
	if len(req.Rules) == 0 {
		rules, err := s.GetAvailableRules(req.Language)
		if err != nil {
			return &domain.RepairResult{
				Success: false,
				Error:   fmt.Sprintf("failed to get rules: %v", err),
			}, err
		}
		req.Rules = rules
	}

	// Sort rules by priority
	sort.Slice(req.Rules, func(i, j int) bool {
		return req.Rules[i].Priority > req.Rules[j].Priority
	})

	// Execute repair cycle
	for attempt := 1; attempt <= req.MaxAttempts; attempt++ {
		s.log.Info(fmt.Sprintf("Repair attempt %d/%d", attempt, req.MaxAttempts))

		// Analyze errors and apply rules
		appliedRules := s.applyRepairRules(ctx, req.ProjectPath, req.ErrorOutput, req.Rules)

		if len(appliedRules) == 0 {
			s.log.Info("No applicable repair rules found")
			break
		}

		// Check if errors are fixed
		success, newErrors := s.verifyRepair(ctx, req.ProjectPath, req.Language)
		if success {
			duration := time.Since(startTime)
			return &domain.RepairResult{
				Success:    true,
				FixedFiles: appliedRules,
				Duration:   duration,
				Attempts:   attempt,
			}, nil
		}

		// Update errors for next attempt
		req.ErrorOutput = newErrors
	}

	duration := time.Since(startTime)
	return &domain.RepairResult{
		Success:  false,
		Error:    "repair cycle completed but errors remain",
		Duration: duration,
		Attempts: req.MaxAttempts,
	}, nil
}

// GetAvailableRules returns available rules for the language
func (s *RepairService) GetAvailableRules(language string) ([]domain.RepairRule, error) {
	rules := s.getDefaultRules(language)
	return rules, nil
}

// AddRule adds a new repair rule
func (s *RepairService) AddRule(rule domain.RepairRule) error {
	s.log.Info(fmt.Sprintf("Adding repair rule: %s", rule.Name))
	return nil
}

// RemoveRule removes a repair rule
func (s *RepairService) RemoveRule(ruleID string) error {
	s.log.Info(fmt.Sprintf("Removing repair rule: %s", ruleID))
	return nil
}

// ValidateRule validates a repair rule
func (s *RepairService) ValidateRule(rule domain.RepairRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if rule.Pattern == "" {
		return fmt.Errorf("rule pattern is required")
	}

	// Validate regex pattern
	if _, err := regexp.Compile(rule.Pattern); err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	return nil
}

// applyRepairRules applies rules to the project
func (s *RepairService) applyRepairRules(ctx context.Context, projectPath, errorOutput string, rules []domain.RepairRule) []string {
	var fixedFiles []string

	for _, rule := range rules {
		// Check if rule matches errors
		if !s.matchesError(errorOutput, rule) {
			continue
		}

		// Apply rule
		files, err := s.applyRule(ctx, projectPath, rule)
		if err != nil {
			s.log.Warning(fmt.Sprintf("Failed to apply rule %s: %v", rule.Name, err))
			continue
		}

		fixedFiles = append(fixedFiles, files...)
		s.log.Info(fmt.Sprintf("Applied rule %s to %d files", rule.Name, len(files)))
	}

	return fixedFiles
}

// matchesError checks if rule matches the errors
func (s *RepairService) matchesError(errorOutput string, rule domain.RepairRule) bool {
	re, err := regexp.Compile(rule.Pattern)
	if err != nil {
		s.log.Warning(fmt.Sprintf("Invalid regex in rule %s: %v", rule.Name, err))
		return false
	}
	return re.MatchString(errorOutput)
}

// applyRule applies a specific rule
func (s *RepairService) applyRule(ctx context.Context, projectPath string, rule domain.RepairRule) ([]string, error) {
	var fixedFiles []string

	switch rule.Category {
	case "format":
		fixedFiles = s.applyFormatRule(ctx, projectPath, rule)
	case "import":
		fixedFiles = s.applyImportRule(ctx, projectPath, rule)
	case "syntax":
		fixedFiles = s.applySyntaxRule(ctx, projectPath, rule)
	default:
		s.log.Warning(fmt.Sprintf("Unknown rule category: %s", rule.Category))
	}

	return fixedFiles, nil
}

// applyFormatRule applies formatting rules
func (s *RepairService) applyFormatRule(ctx context.Context, projectPath string, rule domain.RepairRule) []string {
	var fixedFiles []string

	// Apply language-specific formatters
	if strings.Contains(rule.Language, "go") {
		// gofmt
		cmd := exec.CommandContext(ctx, "gofmt", "-w", ".")
		cmd.Dir = projectPath
		if err := cmd.Run(); err == nil {
			fixedFiles = append(fixedFiles, "*.go")
		}

		// goimports
		cmd = exec.CommandContext(ctx, "goimports", "-w", ".")
		cmd.Dir = projectPath
		if err := cmd.Run(); err == nil {
			fixedFiles = append(fixedFiles, "*.go")
		}
	} else if strings.Contains(rule.Language, "typescript") || strings.Contains(rule.Language, "javascript") {
		// prettier
		cmd := exec.CommandContext(ctx, "npx", "prettier", "--write", ".")
		cmd.Dir = projectPath
		if err := cmd.Run(); err == nil {
			fixedFiles = append(fixedFiles, "*.ts", "*.js", "*.vue")
		}
	}

	return fixedFiles
}

// applyImportRule applies import-related rules
func (s *RepairService) applyImportRule(ctx context.Context, projectPath string, rule domain.RepairRule) []string {
	var fixedFiles []string

	if strings.Contains(rule.Language, "go") {
		// go mod tidy
		cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
		cmd.Dir = projectPath
		if err := cmd.Run(); err == nil {
			fixedFiles = append(fixedFiles, "go.mod", "go.sum")
		}
	} else if strings.Contains(rule.Language, "typescript") || strings.Contains(rule.Language, "javascript") {
		// npm install
		cmd := exec.CommandContext(ctx, "npm", "install")
		cmd.Dir = projectPath
		if err := cmd.Run(); err == nil {
			fixedFiles = append(fixedFiles, "package.json", "package-lock.json")
		}
	}

	return fixedFiles
}

// applySyntaxRule applies syntax-related rules
func (s *RepairService) applySyntaxRule(ctx context.Context, projectPath string, rule domain.RepairRule) []string {
	// Simple implementation returns empty list
	// Real implementation would contain logic for fixing syntax errors
	return []string{}
}

// verifyRepair checks if errors are fixed
func (s *RepairService) verifyRepair(ctx context.Context, projectPath, language string) (bool, string) {
	var cmd *exec.Cmd

	switch language {
	case "go":
		cmd = exec.CommandContext(ctx, "go", "build", "./...")
	case "typescript", "javascript":
		cmd = exec.CommandContext(ctx, "npx", "tsc", "--noEmit")
	default:
		return false, "unsupported language for verification"
	}

	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()

	if err == nil {
		return true, ""
	}

	return false, string(output)
}

// getDefaultRules returns default rules for the language
func (s *RepairService) getDefaultRules(language string) []domain.RepairRule {
	switch language {
	case "go":
		return []domain.RepairRule{
			{
				ID:          "go-format",
				Name:        "Go Format",
				Description: "Format Go code with gofmt and goimports",
				Pattern:     `gofmt|goimports`,
				Fix:         "format",
				Priority:    100,
				Language:    "go",
				Category:    "format",
			},
			{
				ID:          "go-imports",
				Name:        "Go Imports",
				Description: "Fix Go imports with go mod tidy",
				Pattern:     `undefined|imported and not used`,
				Fix:         "imports",
				Priority:    90,
				Language:    "go",
				Category:    "import",
			},
		}
	case "typescript", "javascript":
		return []domain.RepairRule{
			{
				ID:          "ts-format",
				Name:        "TypeScript Format",
				Description: "Format TypeScript code with prettier",
				Pattern:     `prettier|format`,
				Fix:         "format",
				Priority:    100,
				Language:    "typescript",
				Category:    "format",
			},
			{
				ID:          "ts-imports",
				Name:        "TypeScript Imports",
				Description: "Fix TypeScript imports with npm install",
				Pattern:     `Cannot find module|Module not found`,
				Fix:         "imports",
				Priority:    90,
				Language:    "typescript",
				Category:    "import",
			},
		}
	default:
		return []domain.RepairRule{}
	}
}