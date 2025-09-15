package analysis

import (
	"fmt"
	"path/filepath"
	"regexp"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/policy"
	"strings"
	"sync"
	"time"
)

// TaskflowService interface for taskflow operations
type TaskflowService interface {
	// Add minimal interface methods needed
	ValidateTask(taskID string) error
}

// GuardrailsService implements policy enforcement within the analysis bounded context
type GuardrailsService struct {
	log           domain.Logger
	policies      []domain.GuardrailPolicy
	budgets       []domain.BudgetPolicy
	mu            sync.RWMutex
	taskflowSvc   TaskflowService
	config        domain.GuardrailConfig
	ephemeralMode bool
	ephemeralEnd  time.Time
	opaService    *policy.OPAService
}

// NewGuardrailsService creates a new guardrails service
func NewGuardrailsService(log domain.Logger, taskflowSvc TaskflowService) *GuardrailsService {
	service := &GuardrailsService{
		log:         log,
		taskflowSvc: taskflowSvc,
		policies:    make([]domain.GuardrailPolicy, 0),
		budgets:     make([]domain.BudgetPolicy, 0),
		config: domain.GuardrailConfig{
			FailClosed:           true,
			EnableEphemeralMode:  true,
			EphemeralTimeout:     5 * time.Minute,
			EnableTaskValidation: true,
			EnableBudgetTracking: true,
			EnablePathValidation: true,
		},
		opaService: policy.NewOPAService(log),
	}

	// Initialize default policies
	service.initializeDefaultPolicies()

	return service
}

// ValidatePath checks path against policies
func (s *GuardrailsService) ValidatePath(path string) ([]domain.GuardrailViolation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var violations []domain.GuardrailViolation

	// Check ephemeral mode for critical paths
	if s.ephemeralMode && s.isEphemeralExpired() {
		s.disableEphemeralMode()
	}

	// Check if path is allowed in ephemeral mode
	if s.ephemeralMode && s.isCriticalPath(path) {
		s.log.Info(fmt.Sprintf("Critical path %s allowed in ephemeral mode", path))
		return violations, nil
	}

	// Additional validation through OPA policies
	if s.opaService != nil {
		opaResult, err := s.opaService.ValidatePath(path)
		if err != nil {
			s.log.Warning(fmt.Sprintf("OPA validation failed: %v", err))
		} else if !opaResult.Valid {
			for _, violation := range opaResult.Violations {
				guardrailViolation := domain.GuardrailViolation{
					PolicyID:  "opa-policy",
					RuleID:    violation.Type,
					Severity:  domain.GuardrailSeverityBlock,
					Message:   violation.Message,
					FilePath:  path,
					Timestamp: time.Now(),
					Context: map[string]interface{}{
						"opa_violation":  true,
						"violation_type": violation.Type,
					},
				}
				violations = append(violations, guardrailViolation)
			}
		}
	}

	for _, policy := range s.policies {
		if !policy.Enabled {
			continue
		}

		if policy.Type == domain.GuardrailTypeForbiddenPath {
			for _, rule := range policy.Rules {
				if s.matchesPath(path, rule.Pattern) {
					violation := domain.GuardrailViolation{
						PolicyID:  policy.ID,
						RuleID:    rule.ID,
						Severity:  policy.Severity,
						Message:   rule.Message,
						FilePath:  path,
						Timestamp: time.Now(),
						Context: map[string]interface{}{
							"pattern":       rule.Pattern,
							"ephemeralMode": s.ephemeralMode,
							"failClosed":    s.config.FailClosed,
						},
					}
					violations = append(violations, violation)

					// With fail-closed behavior block immediately
					if s.config.FailClosed && policy.Severity == domain.GuardrailSeverityBlock {
						s.log.Error(fmt.Sprintf("Guardrail violation blocked: %s - %s", path, rule.Message))
						return violations, fmt.Errorf("guardrail violation: %s", rule.Message)
					}
				}
			}
		}
	}

	return violations, nil
}

// ValidateBudget checks budget constraints
func (s *GuardrailsService) ValidateBudget(budgetType domain.BudgetType, current int64) ([]domain.BudgetViolation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var violations []domain.BudgetViolation

	for _, budget := range s.budgets {
		if !budget.Enabled || budget.Type != budgetType {
			continue
		}

		if current > budget.Limit {
			violation := domain.BudgetViolation{
				PolicyID:  budget.ID,
				Type:      budget.Type,
				Current:   current,
				Limit:     budget.Limit,
				Unit:      budget.Unit,
				Message:   fmt.Sprintf("Budget exceeded: %d %s (limit: %d)", current, budget.Unit, budget.Limit),
				Timestamp: time.Now(),
				Context: map[string]interface{}{
					"timeWindow":   budget.TimeWindow,
					"failClosed":   s.config.FailClosed,
					"budgetType":   budgetType,
					"currentValue": current,
				},
			}
			violations = append(violations, violation)

			// With fail-closed behavior block budget exceeding
			if s.config.FailClosed {
				s.log.Error(fmt.Sprintf("Budget violation blocked: %s", violation.Message))
				return violations, fmt.Errorf("budget violation: %s", violation.Message)
			}
		}
	}

	return violations, nil
}

// ValidateTask checks task against policies
func (s *GuardrailsService) ValidateTask(taskID string, files []string, linesChanged int64) (*domain.TaskValidationResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := &domain.TaskValidationResult{
		TaskID:           taskID,
		Valid:            true,
		Violations:       make([]domain.GuardrailViolation, 0),
		BudgetViolations: make([]domain.BudgetViolation, 0),
		Timestamp:        time.Now(),
	}

	// Check file paths
	for _, file := range files {
		pathViolations, err := s.ValidatePath(file)
		if err != nil {
			result.Valid = false
			result.Error = err.Error()
			return result, err
		}
		result.Violations = append(result.Violations, pathViolations...)
	}

	// Check file budget
	if s.config.EnableBudgetTracking {
		fileBudgetViolations, err := s.ValidateBudget(domain.BudgetTypeFiles, int64(len(files)))
		if err != nil {
			result.Valid = false
			result.Error = err.Error()
			return result, err
		}
		result.BudgetViolations = append(result.BudgetViolations, fileBudgetViolations...)

		// Check lines changed budget
		linesBudgetViolations, err := s.ValidateBudget(domain.BudgetTypeLines, linesChanged)
		if err != nil {
			result.Valid = false
			result.Error = err.Error()
			return result, err
		}
		result.BudgetViolations = append(result.BudgetViolations, linesBudgetViolations...)
	}

	// Set validation result
	result.Valid = len(result.Violations) == 0 && len(result.BudgetViolations) == 0

	return result, nil
}

// EnableEphemeralMode enables ephemeral mode temporarily
func (s *GuardrailsService) EnableEphemeralMode(duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.config.EnableEphemeralMode {
		return fmt.Errorf("ephemeral mode is disabled in configuration")
	}

	s.ephemeralMode = true
	s.ephemeralEnd = time.Now().Add(duration)
	s.log.Info(fmt.Sprintf("Ephemeral mode enabled for %v", duration))

	return nil
}

// DisableEphemeralMode disables ephemeral mode
func (s *GuardrailsService) DisableEphemeralMode() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.disableEphemeralMode()
	return nil
}

// GetViolationReport gets violation report for a task
func (s *GuardrailsService) GetViolationReport(taskID string) (*domain.ViolationReport, error) {
	// Implementation would fetch violations from storage
	// For now, return empty report
	return &domain.ViolationReport{
		TaskID:           taskID,
		Violations:       make([]domain.GuardrailViolation, 0),
		BudgetViolations: make([]domain.BudgetViolation, 0),
		GeneratedAt:      time.Now(),
	}, nil
}

// AddPolicy adds a new guardrail policy
func (s *GuardrailsService) AddPolicy(policy domain.GuardrailPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate policy
	if err := s.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}

	s.policies = append(s.policies, policy)
	s.log.Info(fmt.Sprintf("Added guardrail policy: %s", policy.ID))

	return nil
}

// RemovePolicy removes a guardrail policy
func (s *GuardrailsService) RemovePolicy(policyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, policy := range s.policies {
		if policy.ID == policyID {
			s.policies = append(s.policies[:i], s.policies[i+1:]...)
			s.log.Info(fmt.Sprintf("Removed guardrail policy: %s", policyID))
			return nil
		}
	}

	return fmt.Errorf("policy not found: %s", policyID)
}

// AddBudgetPolicy adds a budget policy
func (s *GuardrailsService) AddBudgetPolicy(budget domain.BudgetPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate budget policy
	if err := s.validateBudgetPolicy(budget); err != nil {
		return fmt.Errorf("invalid budget policy: %w", err)
	}

	s.budgets = append(s.budgets, budget)
	s.log.Info(fmt.Sprintf("Added budget policy: %s", budget.ID))

	return nil
}

// UpdateConfig updates guardrail configuration
func (s *GuardrailsService) UpdateConfig(config domain.GuardrailConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	s.log.Info("Updated guardrail configuration")

	return nil
}

// GetConfig returns current configuration
func (s *GuardrailsService) GetConfig() domain.GuardrailConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config
}

// Private helper methods

func (s *GuardrailsService) initializeDefaultPolicies() {
	// Default forbidden paths policy
	defaultPolicy := domain.GuardrailPolicy{
		ID:          "default-forbidden-paths",
		Name:        "Default Forbidden Paths",
		Description: "Blocks access to sensitive system paths",
		Type:        domain.GuardrailTypeForbiddenPath,
		Severity:    domain.GuardrailSeverityBlock,
		Enabled:     true,
		Rules: []domain.GuardrailRule{
			{
				ID:      "no-system-files",
				Pattern: "^/(etc|sys|proc|dev)/.*",
				Message: "Access to system directories is forbidden",
			},
			{
				ID:      "no-hidden-config",
				Pattern: "^/.*\\.ssh/.*",
				Message: "Access to SSH configuration is forbidden",
			},
		},
	}

	// Default budget policies
	filesBudget := domain.BudgetPolicy{
		ID:         "default-files-budget",
		Name:       "Default Files Budget",
		Type:       domain.BudgetTypeFiles,
		Limit:      150,
		Unit:       "files",
		TimeWindow: time.Hour,
		Enabled:    true,
	}

	linesBudget := domain.BudgetPolicy{
		ID:         "default-lines-budget",
		Name:       "Default Lines Budget",
		Type:       domain.BudgetTypeLines,
		Limit:      1500,
		Unit:       "lines",
		TimeWindow: time.Hour,
		Enabled:    true,
	}

	s.policies = append(s.policies, defaultPolicy)
	s.budgets = append(s.budgets, filesBudget, linesBudget)
}

func (s *GuardrailsService) matchesPath(path, pattern string) bool {
	// Simple glob pattern matching
	matched, err := filepath.Match(pattern, path)
	if err == nil && matched {
		return true
	}

	// Regex pattern matching
	if strings.HasPrefix(pattern, "^") {
		regex, err := regexp.Compile(pattern)
		if err == nil {
			return regex.MatchString(path)
		}
	}

	return false
}

func (s *GuardrailsService) isCriticalPath(path string) bool {
	criticalPatterns := []string{
		"^/(etc|sys|proc|dev)/.*",
		".*\\.ssh/.*",
		".*\\.aws/.*",
		".*\\.kube/.*",
	}

	for _, pattern := range criticalPatterns {
		if s.matchesPath(path, pattern) {
			return true
		}
	}

	return false
}

func (s *GuardrailsService) isEphemeralExpired() bool {
	return time.Now().After(s.ephemeralEnd)
}

func (s *GuardrailsService) disableEphemeralMode() {
	s.ephemeralMode = false
	s.ephemeralEnd = time.Time{}
	s.log.Info("Ephemeral mode disabled")
}

func (s *GuardrailsService) validatePolicy(policy domain.GuardrailPolicy) error {
	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}

	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	if len(policy.Rules) == 0 {
		return fmt.Errorf("policy must have at least one rule")
	}

	for _, rule := range policy.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule ID is required")
		}
		if rule.Pattern == "" {
			return fmt.Errorf("rule pattern is required")
		}
	}

	return nil
}

func (s *GuardrailsService) validateBudgetPolicy(budget domain.BudgetPolicy) error {
	if budget.ID == "" {
		return fmt.Errorf("budget policy ID is required")
	}

	if budget.Name == "" {
		return fmt.Errorf("budget policy name is required")
	}

	if budget.Limit <= 0 {
		return fmt.Errorf("budget limit must be positive")
	}

	if budget.Unit == "" {
		return fmt.Errorf("budget unit is required")
	}

	return nil
}