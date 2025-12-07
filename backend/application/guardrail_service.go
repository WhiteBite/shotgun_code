package application

import (
	"fmt"
	"path/filepath"
	"regexp"
	"shotgun_code/domain"
	"strings"
	"sync"
	"time"
)

// Task type constants for ephemeral mode
const (
	taskTypeScaffold = "scaffold"
	taskTypeDepsFix  = "deps_fix"
)

// GuardrailServiceImpl реализует GuardrailService
type GuardrailServiceImpl struct {
	log              domain.Logger
	policies         []domain.GuardrailPolicy
	budgets          []domain.BudgetPolicy
	mu               sync.RWMutex
	config           domain.GuardrailConfig
	ephemeralMode    bool
	ephemeralEnd     time.Time
	opaService       domain.OPAService
	fileStatProvider domain.FileStatProvider
	taskTypeProvider domain.TaskTypeProvider // Replace taskflowService field with taskTypeProvider
}

// NewGuardrailService создает новый сервис guardrails
func NewGuardrailService(log domain.Logger, opaService domain.OPAService, fileStatProvider domain.FileStatProvider) domain.GuardrailService {
	service := &GuardrailServiceImpl{
		log:      log,
		policies: make([]domain.GuardrailPolicy, 0),
		budgets:  make([]domain.BudgetPolicy, 0),
		config: domain.GuardrailConfig{
			FailClosed:           true,
			EnableEphemeralMode:  true,
			EphemeralTimeout:     5 * time.Minute,
			EnableTaskValidation: true,
			EnableBudgetTracking: true,
			EnablePathValidation: true,
		},
		opaService:       opaService,
		fileStatProvider: fileStatProvider,
	}

	// Инициализируем политики по умолчанию
	service.initializeDefaultPolicies()

	return service
}

// SetTaskTypeProvider устанавливает провайдер типов задач для разрешения циклической зависимости
func (s *GuardrailServiceImpl) SetTaskTypeProvider(taskTypeProvider domain.TaskTypeProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.taskTypeProvider = taskTypeProvider
}

// ValidatePath проверяет путь на соответствие политикам
func (s *GuardrailServiceImpl) ValidatePath(path string) ([]domain.GuardrailViolation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.checkEphemeralExpiry()

	if s.ephemeralMode && s.isCriticalPath(path) {
		s.log.Info(fmt.Sprintf("Critical path %s allowed in ephemeral mode", path))
		return nil, nil
	}

	var violations []domain.GuardrailViolation
	violations = append(violations, s.validateWithOPA(path)...)
	policyViolations, err := s.validateWithPolicies(path)
	if err != nil {
		return append(violations, policyViolations...), err
	}
	return append(violations, policyViolations...), nil
}

// checkEphemeralExpiry checks and disables expired ephemeral mode
func (s *GuardrailServiceImpl) checkEphemeralExpiry() {
	if s.ephemeralMode && s.isEphemeralExpired() {
		s.disableEphemeralMode()
	}
}

// validateWithOPA validates path using OPA service
func (s *GuardrailServiceImpl) validateWithOPA(path string) []domain.GuardrailViolation {
	if s.opaService == nil {
		return nil
	}
	opaResult, err := s.opaService.ValidatePath(path)
	if err != nil {
		s.log.Warning(fmt.Sprintf("OPA validation failed: %v", err))
		return nil
	}
	if opaResult.Valid {
		return nil
	}

	violations := make([]domain.GuardrailViolation, 0, len(opaResult.Violations))
	for _, v := range opaResult.Violations {
		violations = append(violations, domain.GuardrailViolation{
			PolicyID: "opa-policy", RuleID: v.Type, Severity: domain.GuardrailSeverityBlock,
			Message: v.Message, FilePath: path, Timestamp: time.Now(),
			Context: map[string]interface{}{"opa_violation": true, "violation_type": v.Type},
		})
	}
	return violations
}

// validateWithPolicies validates path against configured policies
func (s *GuardrailServiceImpl) validateWithPolicies(path string) ([]domain.GuardrailViolation, error) {
	violations := make([]domain.GuardrailViolation, 0, len(s.policies))
	for _, policy := range s.policies {
		if !policy.Enabled || policy.Type != domain.GuardrailTypeForbiddenPath {
			continue
		}
		v, err := s.checkPolicyRules(path, policy)
		violations = append(violations, v...)
		if err != nil {
			return violations, err
		}
	}
	return violations, nil
}

// checkPolicyRules checks path against policy rules
func (s *GuardrailServiceImpl) checkPolicyRules(path string, policy domain.GuardrailPolicy) ([]domain.GuardrailViolation, error) {
	violations := make([]domain.GuardrailViolation, 0, len(policy.Rules))
	for _, rule := range policy.Rules {
		if !s.matchesPath(path, rule.Pattern) {
			continue
		}
		violation := domain.GuardrailViolation{
			PolicyID: policy.ID, RuleID: rule.ID, Severity: policy.Severity,
			Message: rule.Message, FilePath: path, Timestamp: time.Now(),
			Context: map[string]interface{}{"pattern": rule.Pattern, "ephemeralMode": s.ephemeralMode, "failClosed": s.config.FailClosed},
		}
		violations = append(violations, violation)

		if s.config.FailClosed && policy.Severity == domain.GuardrailSeverityBlock {
			s.log.Error(fmt.Sprintf("Guardrail violation blocked: %s - %s", path, rule.Message))
			return violations, fmt.Errorf("guardrail violation: %s", rule.Message)
		}
	}
	return violations, nil
}

// ValidateBudget проверяет бюджетные ограничения
func (s *GuardrailServiceImpl) ValidateBudget(budgetType domain.BudgetType, current int64) ([]domain.BudgetViolation, error) {
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

			// При fail-closed поведении блокируем превышение бюджета
			if s.config.FailClosed {
				s.log.Error(fmt.Sprintf("Budget violation blocked: %s", violation.Message))
				return violations, fmt.Errorf("budget violation: %s", violation.Message)
			}
		}
	}

	return violations, nil
}

// ValidateTask проверяет задачу на соответствие политикам
func (s *GuardrailServiceImpl) ValidateTask(taskID string, files []string, linesChanged int64) (*domain.TaskValidationResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := &domain.TaskValidationResult{
		TaskID:           taskID,
		Valid:            true,
		Violations:       make([]domain.GuardrailViolation, 0),
		BudgetViolations: make([]domain.BudgetViolation, 0),
		Timestamp:        time.Now(),
	}

	// Проверяем пути файлов
	for _, file := range files {
		pathViolations, err := s.ValidatePath(file)
		if err != nil {
			result.Valid = false
			result.Error = err.Error()
			return result, err
		}
		result.Violations = append(result.Violations, pathViolations...)
	}

	// Проверяем бюджет файлов
	if s.config.EnableBudgetTracking {
		fileBudgetViolations, err := s.ValidateBudget(domain.BudgetTypeFiles, int64(len(files)))
		if err != nil {
			result.Valid = false
			result.Error = err.Error()
			return result, err
		}
		result.BudgetViolations = append(result.BudgetViolations, fileBudgetViolations...)
	}

	// Проверяем бюджет строк
	if s.config.EnableBudgetTracking {
		lineBudgetViolations, err := s.ValidateBudget(domain.BudgetTypeLines, linesChanged)
		if err != nil {
			result.Valid = false
			result.Error = err.Error()
			return result, err
		}
		result.BudgetViolations = append(result.BudgetViolations, lineBudgetViolations...)
	}

	// Определяем общую валидность
	if len(result.Violations) > 0 || len(result.BudgetViolations) > 0 {
		result.Valid = false
		if s.config.FailClosed {
			result.Error = "Task validation failed due to guardrail violations"
		}
	}

	return result, nil
}

// EnableEphemeralMode включает ephemeral mode для критических путей
// This method now accepts taskType parameter directly instead of checking with TaskflowService
func (s *GuardrailServiceImpl) EnableEphemeralMode(taskID string, taskType string, duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Safety check: verify taskTypeProvider is initialized
	if s.taskTypeProvider == nil {
		s.log.Warning("TaskTypeProvider not initialized - ephemeral mode validation skipped")
		// Fallback: allow only if explicitly scaffold/deps_fix
		if taskType != taskTypeScaffold && taskType != taskTypeDepsFix {
			return fmt.Errorf("ephemeral mode only allowed for scaffold/deps_fix tasks")
		}
	}

	// Проверяем, что это задача scaffold или deps_fix
	if taskType != taskTypeScaffold && taskType != taskTypeDepsFix {
		return fmt.Errorf("ephemeral mode only allowed for scaffold/deps_fix tasks")
	}

	s.ephemeralMode = true
	s.ephemeralEnd = time.Now().Add(duration)
	s.log.Info(fmt.Sprintf("Ephemeral mode enabled for task %s until %s", taskID, s.ephemeralEnd.Format(time.RFC3339)))

	return nil
}

// DisableEphemeralMode отключает ephemeral mode
func (s *GuardrailServiceImpl) DisableEphemeralMode() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ephemeralMode = false
	s.ephemeralEnd = time.Time{}
	s.log.Info("Ephemeral mode disabled")
}

// GetPolicies возвращает все политики
func (s *GuardrailServiceImpl) GetPolicies() ([]domain.GuardrailPolicy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]domain.GuardrailPolicy, len(s.policies))
	copy(policies, s.policies)
	return policies, nil
}

// GetBudgetPolicies возвращает бюджетные политики
func (s *GuardrailServiceImpl) GetBudgetPolicies() ([]domain.BudgetPolicy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	budgets := make([]domain.BudgetPolicy, len(s.budgets))
	copy(budgets, s.budgets)
	return budgets, nil
}

// AddPolicy добавляет новую политику
func (s *GuardrailServiceImpl) AddPolicy(policy domain.GuardrailPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем уникальность ID
	for _, existing := range s.policies {
		if existing.ID == policy.ID {
			return fmt.Errorf("policy with ID %s already exists", policy.ID)
		}
	}

	s.policies = append(s.policies, policy)
	s.log.Info(fmt.Sprintf("Added guardrail policy: %s", policy.Name))
	return nil
}

// RemovePolicy удаляет политику
func (s *GuardrailServiceImpl) RemovePolicy(policyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, policy := range s.policies {
		if policy.ID == policyID {
			s.policies = append(s.policies[:i], s.policies[i+1:]...)
			s.log.Info(fmt.Sprintf("Removed guardrail policy: %s", policy.Name))
			return nil
		}
	}

	return fmt.Errorf("policy with ID %s not found", policyID)
}

// UpdatePolicy обновляет политику
func (s *GuardrailServiceImpl) UpdatePolicy(policy domain.GuardrailPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.policies {
		if existing.ID == policy.ID {
			s.policies[i] = policy
			s.log.Info(fmt.Sprintf("Updated guardrail policy: %s", policy.Name))
			return nil
		}
	}

	return fmt.Errorf("policy with ID %s not found", policy.ID)
}

// AddBudgetPolicy добавляет бюджетную политику
func (s *GuardrailServiceImpl) AddBudgetPolicy(policy domain.BudgetPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем уникальность ID
	for _, existing := range s.budgets {
		if existing.ID == policy.ID {
			return fmt.Errorf("budget policy with ID %s already exists", policy.ID)
		}
	}

	s.budgets = append(s.budgets, policy)
	s.log.Info(fmt.Sprintf("Added budget policy: %s", policy.Name))
	return nil
}

// RemoveBudgetPolicy удаляет бюджетную политику
func (s *GuardrailServiceImpl) RemoveBudgetPolicy(policyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, policy := range s.budgets {
		if policy.ID == policyID {
			s.budgets = append(s.budgets[:i], s.budgets[i+1:]...)
			s.log.Info(fmt.Sprintf("Removed budget policy: %s", policy.Name))
			return nil
		}
	}

	return fmt.Errorf("budget policy with ID %s not found", policyID)
}

// UpdateBudgetPolicy обновляет бюджетную политику
func (s *GuardrailServiceImpl) UpdateBudgetPolicy(policy domain.BudgetPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.budgets {
		if existing.ID == policy.ID {
			s.budgets[i] = policy
			s.log.Info(fmt.Sprintf("Updated budget policy: %s", policy.Name))
			return nil
		}
	}

	return fmt.Errorf("budget policy with ID %s not found", policy.ID)
}

// GetConfig возвращает конфигурацию guardrails
func (s *GuardrailServiceImpl) GetConfig() domain.GuardrailConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config
}

// UpdateConfig обновляет конфигурацию guardrails
func (s *GuardrailServiceImpl) UpdateConfig(config domain.GuardrailConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	s.log.Info("Updated guardrail configuration")
	return nil
}

// matchesPath проверяет, соответствует ли путь паттерну
func (s *GuardrailServiceImpl) matchesPath(path, pattern string) bool {
	// Если паттерн содержит glob символы, используем filepath.Match
	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			s.log.Warning(fmt.Sprintf("Invalid glob pattern: %s", pattern))
			return false
		}
		return matched
	}

	// Иначе используем regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		s.log.Warning(fmt.Sprintf("Invalid regex pattern: %s", pattern))
		return false
	}
	return re.MatchString(path)
}

// isCriticalPath проверяет, является ли путь критическим
func (s *GuardrailServiceImpl) isCriticalPath(path string) bool {
	criticalPatterns := []string{
		"go\\.mod",
		"package\\.json",
		"package-lock\\.json",
		"yarn\\.lock",
		"pnpm-lock\\.yaml",
	}

	for _, pattern := range criticalPatterns {
		if s.matchesPath(path, pattern) {
			return true
		}
	}

	return false
}

// isEphemeralExpired проверяет, истек ли ephemeral mode
func (s *GuardrailServiceImpl) isEphemeralExpired() bool {
	return time.Now().After(s.ephemeralEnd)
}

// disableEphemeralMode отключает ephemeral mode
func (s *GuardrailServiceImpl) disableEphemeralMode() {
	s.ephemeralMode = false
	s.ephemeralEnd = time.Time{}
	s.log.Info("Ephemeral mode expired and disabled")
}

// initializeDefaultPolicies инициализирует политики по умолчанию
func (s *GuardrailServiceImpl) initializeDefaultPolicies() {
	// Политика запрещенных путей
	forbiddenPathsPolicy := domain.GuardrailPolicy{
		ID:          "forbidden-paths",
		Name:        "Forbidden Paths",
		Description: "Запрещенные пути и файлы",
		Type:        domain.GuardrailTypeForbiddenPath,
		Severity:    domain.GuardrailSeverityBlock,
		Enabled:     true,
		Rules: []domain.GuardrailRule{
			{
				ID:          "go-mod",
				Pattern:     "go\\.mod",
				Description: "Запрещено изменять go.mod",
				Action:      domain.GuardrailActionBlock,
				Message:     "Изменение go.mod запрещено",
			},
			{
				ID:          "package-json",
				Pattern:     "package\\.json|package-lock\\.json",
				Description: "Запрещено изменять package.json и package-lock.json",
				Action:      domain.GuardrailActionBlock,
				Message:     "Изменение package.json запрещено",
			},
			{
				ID:          "node-modules",
				Pattern:     "node_modules/.*",
				Description: "Запрещено изменять node_modules",
				Action:      domain.GuardrailActionBlock,
				Message:     "Изменение node_modules запрещено",
			},
			{
				ID:          "secrets",
				Pattern:     ".*\\.(key|pem|p12|pfx|env|secret)",
				Description: "Запрещено изменять файлы с секретами",
				Action:      domain.GuardrailActionBlock,
				Message:     "Изменение файлов с секретами запрещено",
			},
			{
				ID:          "binary-files",
				Pattern:     ".*\\.(exe|dll|so|dylib|bin|jar|war|ear)",
				Description: "Запрещено изменять бинарные файлы",
				Action:      domain.GuardrailActionBlock,
				Message:     "Изменение бинарных файлов запрещено",
			},
			{
				ID:          "temp-dirs",
				Pattern:     "(tmp|temp|cache)/.*",
				Description: "Запрещено изменять временные каталоги",
				Action:      domain.GuardrailActionBlock,
				Message:     "Изменение временных каталогов запрещено",
			},
			{
				ID:          "cursor-rules",
				Pattern:     "\\.cursor/rules/cutc\\.mdc",
				Description: "Запрещено изменять внутренний протокол",
				Action:      domain.GuardrailActionBlock,
				Message:     "Изменение .cursor/rules/cutc.mdc запрещено",
			},
		},
	}

	// Бюджетные политики
	fileBudgetPolicy := domain.BudgetPolicy{
		ID:          "max-files",
		Name:        "Maximum Files",
		Description: "Максимальное количество изменяемых файлов",
		Type:        domain.BudgetTypeFiles,
		Limit:       150,
		Unit:        domain.BudgetUnitCount,
		TimeWindow:  time.Hour,
		Enabled:     true,
	}

	lineBudgetPolicy := domain.BudgetPolicy{
		ID:          "max-lines",
		Name:        "Maximum Lines",
		Description: "Максимальное количество изменяемых строк",
		Type:        domain.BudgetTypeLines,
		Limit:       1500,
		Unit:        domain.BudgetUnitCount,
		TimeWindow:  time.Hour,
		Enabled:     true,
	}

	tokenBudgetPolicy := domain.BudgetPolicy{
		ID:          "max-tokens",
		Name:        "Maximum Tokens",
		Description: "Максимальное количество токенов в запросе",
		Type:        domain.BudgetTypeTokens,
		Limit:       10000,
		Unit:        domain.BudgetUnitCount,
		TimeWindow:  time.Hour,
		Enabled:     true,
	}

	// Добавляем политики
	s.policies = append(s.policies, forbiddenPathsPolicy)
	s.budgets = append(s.budgets, fileBudgetPolicy, lineBudgetPolicy, tokenBudgetPolicy)

	s.log.Info("Initialized default guardrail policies")
}
