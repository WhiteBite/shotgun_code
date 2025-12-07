package application

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"shotgun_code/domain"
	"time"
)

// TaskProtocolConfigService manages Task Protocol configuration
type TaskProtocolConfigService struct {
	log        domain.Logger
	fileSystem domain.FileSystemProvider
}

// TaskProtocolConfiguration represents the full configuration structure
type TaskProtocolConfiguration struct {
	TaskProtocol TaskProtocolConfig `yaml:"task_protocol"`
}

// TaskProtocolConfig represents the main protocol configuration
type TaskProtocolConfig struct {
	MaxRetries        int                       `yaml:"max_retries"`
	FailFast          bool                      `yaml:"fail_fast"`
	TimeoutSeconds    int                       `yaml:"timeout_seconds"`
	SelfCorrection    SelfCorrectionConf        `yaml:"self_correction"`
	Stages            StagesConfig              `yaml:"stages"`
	LanguageSpecific  map[string]LanguageConfig `yaml:"language_specific"`
	ErrorCorrection   ErrorCorrectionConfig     `yaml:"error_correction"`
	GuardrailPolicies GuardrailPolicyConfig     `yaml:"guardrail_policies"`
}

// SelfCorrectionConf represents self-correction configuration
type SelfCorrectionConf struct {
	Enabled      bool `yaml:"enabled"`
	MaxAttempts  int  `yaml:"max_attempts"`
	AIAssistance bool `yaml:"ai_assistance"`
}

// StagesConfig represents configuration for all stages
type StagesConfig struct {
	Linting    StageConfig `yaml:"linting"`
	Building   StageConfig `yaml:"building"`
	Testing    StageConfig `yaml:"testing"`
	Guardrails StageConfig `yaml:"guardrails"`
}

// StageConfig represents configuration for a single stage
type StageConfig struct {
	Enabled           bool     `yaml:"enabled"`
	TimeoutSeconds    int      `yaml:"timeout_seconds"`
	Tools             []string `yaml:"tools,omitempty"`
	FailOnWarning     bool     `yaml:"fail_on_warning,omitempty"`
	StrictMode        bool     `yaml:"strict_mode,omitempty"`
	Parallel          bool     `yaml:"parallel_builds,omitempty"`
	Scope             string   `yaml:"scope,omitempty"`
	CoverageThreshold int      `yaml:"coverage_threshold,omitempty"`
	EnforcePolicies   bool     `yaml:"enforce_policies,omitempty"`
	FailOnViolation   bool     `yaml:"fail_on_violation,omitempty"`
}

// LanguageConfig represents language-specific configuration
type LanguageConfig struct {
	Linting  LanguageStageConfig `yaml:"linting"`
	Building LanguageStageConfig `yaml:"building"`
	Testing  LanguageStageConfig `yaml:"testing"`
}

// LanguageStageConfig represents language-specific stage configuration
type LanguageStageConfig struct {
	Tools             []string `yaml:"tools,omitempty"`
	Rules             string   `yaml:"rules,omitempty"`
	FailOnWarning     bool     `yaml:"fail_on_warning,omitempty"`
	RaceDetection     bool     `yaml:"race_detection,omitempty"`
	StrictMode        bool     `yaml:"strict_mode,omitempty"`
	Tags              []string `yaml:"tags,omitempty"`
	Parallel          bool     `yaml:"parallel,omitempty"`
	CoverageThreshold int      `yaml:"coverage_threshold,omitempty"`
	Extends           []string `yaml:"extends,omitempty"`
	NoImplicitAny     bool     `yaml:"no_implicit_any,omitempty"`
	TypeChecking      bool     `yaml:"type_checking,omitempty"`
	Framework         string   `yaml:"framework,omitempty"`
	BabelTranspile    bool     `yaml:"babel_transpile,omitempty"`
	Minify            bool     `yaml:"minify,omitempty"`
}

// ErrorCorrectionConfig represents error correction configuration
type ErrorCorrectionConfig struct {
	ImportErrors  CorrectionRuleConfig `yaml:"import_errors"`
	SyntaxErrors  CorrectionRuleConfig `yaml:"syntax_errors"`
	TypeErrors    CorrectionRuleConfig `yaml:"type_errors"`
	LintingErrors CorrectionRuleConfig `yaml:"linting_errors"`
	TestErrors    CorrectionRuleConfig `yaml:"test_errors"`
}

// CorrectionRuleConfig represents configuration for a correction rule
type CorrectionRuleConfig struct {
	AutoFix    bool     `yaml:"auto_fix"`
	Priority   int      `yaml:"priority"`
	Strategies []string `yaml:"strategies"`
}

// GuardrailPolicyConfig represents guardrail policy configuration
type GuardrailPolicyConfig struct {
	Security       SecurityConfig       `yaml:"security"`
	ResourceLimits ResourceLimitsConfig `yaml:"resource_limits"`
	Quality        QualityConfig        `yaml:"quality"`
}

// SecurityConfig represents security-related configuration
type SecurityConfig struct {
	Enabled             bool `yaml:"enabled"`
	ScanVulnerabilities bool `yaml:"scan_vulnerabilities"`
	CheckLicenses       bool `yaml:"check_licenses"`
}

// ResourceLimitsConfig represents resource limit configuration
type ResourceLimitsConfig struct {
	MaxFilesChanged int    `yaml:"max_files_changed"`
	MaxLinesChanged int    `yaml:"max_lines_changed"`
	MaxMemoryUsage  string `yaml:"max_memory_usage"`
}

// QualityConfig represents quality-related configuration
type QualityConfig struct {
	MinCodeCoverage      int  `yaml:"min_code_coverage"`
	MaxComplexity        int  `yaml:"max_complexity"`
	RequireDocumentation bool `yaml:"require_documentation"`
}

// NewTaskProtocolConfigService creates a new configuration service
func NewTaskProtocolConfigService(log domain.Logger, fileSystem domain.FileSystemProvider) *TaskProtocolConfigService {
	return &TaskProtocolConfigService{
		log:        log,
		fileSystem: fileSystem,
	}
}

// LoadConfiguration loads the task protocol configuration from file
func (s *TaskProtocolConfigService) LoadConfiguration(configPath string) (*domain.TaskProtocolConfig, error) {
	s.log.Info(fmt.Sprintf("Loading task protocol configuration from: %s", configPath))

	// Check if config file exists, if not create default
	if !s.fileExists(configPath) {
		s.log.Info("Configuration file not found, creating default configuration")
		if err := s.createDefaultConfiguration(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default configuration: %w", err)
		}
	}

	// Load configuration from file
	// In a real implementation, this would use a YAML parser like gopkg.in/yaml.v3
	// For now, return a default configuration
	return s.createDefaultDomainConfig(), nil
}

// SaveConfiguration saves the task protocol configuration to file
func (s *TaskProtocolConfigService) SaveConfiguration(config *domain.TaskProtocolConfig, configPath string) error {
	s.log.Info(fmt.Sprintf("Saving task protocol configuration to: %s", configPath))

	// Convert domain config to file format and save
	// In a real implementation, this would serialize to YAML
	return s.createDefaultConfiguration(configPath)
}

// GetConfigurationForProject gets project-specific configuration
func (s *TaskProtocolConfigService) GetConfigurationForProject(ctx context.Context, projectPath string, languages []string) (*domain.TaskProtocolConfig, error) {
	// Try to load project-specific config first
	projectConfigPath := filepath.Join(projectPath, ".ark", "protocol.yaml")

	var config *domain.TaskProtocolConfig
	var err error

	if s.fileExists(projectConfigPath) {
		s.log.Info("Loading project-specific task protocol configuration")
		config, err = s.LoadConfiguration(projectConfigPath)
	} else {
		// Fall back to global config
		globalConfigPath := "config/task_protocol.yaml"
		s.log.Info("Loading global task protocol configuration")
		config, err = s.LoadConfiguration(globalConfigPath)
	}

	if err != nil {
		return nil, err
	}

	// Customize config for the specific project and languages
	config.ProjectPath = projectPath
	config.Languages = languages

	return config, nil
}

// ValidateConfiguration validates the configuration
func (s *TaskProtocolConfigService) ValidateConfiguration(config *domain.TaskProtocolConfig) error {
	if config.ProjectPath == "" {
		return fmt.Errorf("project path is required")
	}

	if len(config.Languages) == 0 {
		return fmt.Errorf("at least one language must be specified")
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if config.SelfCorrection.MaxAttempts < 0 {
		return fmt.Errorf("self correction max attempts cannot be negative")
	}

	// Validate enabled stages
	if len(config.EnabledStages) == 0 {
		return fmt.Errorf("at least one stage must be enabled")
	}

	return nil
}

// Helper methods

func (s *TaskProtocolConfigService) createDefaultDomainConfig() *domain.TaskProtocolConfig {
	return &domain.TaskProtocolConfig{
		Languages: []string{"go", "typescript"},
		EnabledStages: []domain.ProtocolStage{
			domain.StageLinting,
			domain.StageBuilding,
			domain.StageTesting,
			domain.StageGuardrails,
		},
		MaxRetries: 3,
		FailFast:   false,
		SelfCorrection: domain.SelfCorrectionConfig{
			Enabled:      true,
			MaxAttempts:  5,
			AIAssistance: true,
		},
		Timeouts: map[string]time.Duration{
			"linting":    5 * time.Minute,
			"building":   10 * time.Minute,
			"testing":    15 * time.Minute,
			"guardrails": 2 * time.Minute,
		},
	}
}

func (s *TaskProtocolConfigService) createDefaultConfiguration(configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create default YAML content (simplified)
	defaultContent := `# Task Protocol Configuration
task_protocol:
  max_retries: 3
  fail_fast: false
  timeout_seconds: 1800
  
  self_correction:
    enabled: true
    max_attempts: 5
    ai_assistance: true
    
  stages:
    linting:
      enabled: true
      timeout_seconds: 300
    building:
      enabled: true
      timeout_seconds: 600
    testing:
      enabled: true
      timeout_seconds: 900
    guardrails:
      enabled: true
      timeout_seconds: 120
`

	// Write to file
	return s.fileSystem.WriteFile(configPath, []byte(defaultContent), 0o644)
}

func (s *TaskProtocolConfigService) fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// GetStageConfiguration returns configuration for a specific stage
func (s *TaskProtocolConfigService) GetStageConfiguration(stage domain.ProtocolStage, language string) (map[string]interface{}, error) {
	// Return stage-specific configuration
	// In a real implementation, this would extract from the loaded config
	config := make(map[string]interface{})

	switch stage {
	case domain.StageLinting:
		config["tools"] = s.getLintingToolsForLanguage(language)
		config["fail_on_warning"] = false
	case domain.StageBuilding:
		config["strict_mode"] = true
		config["parallel"] = true
	case domain.StageTesting:
		config["coverage_threshold"] = 80
		config["parallel"] = true
	case domain.StageGuardrails:
		config["enforce_policies"] = true
		config["fail_on_violation"] = true
	}

	return config, nil
}

func (s *TaskProtocolConfigService) getLintingToolsForLanguage(language string) []string {
	switch language {
	case "go":
		return []string{"staticcheck", "go vet"}
	case "typescript":
		return []string{"eslint", "@typescript-eslint"}
	case "javascript":
		return []string{"eslint"}
	default:
		return []string{}
	}
}
