package project

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/testengine"
)

// TestService provides high-level API for testing operations
type TestService struct {
	log         domain.Logger
	testEngine  domain.TestEngine
	symbolGraph domain.SymbolGraphBuilder
}

// NewTestService creates a new test service
func NewTestService(log domain.Logger, symbolGraph domain.SymbolGraphBuilder) *TestService {
	engine := testengine.NewTestEngine(log, symbolGraph)

	// Register test runners for supported languages
	engine.RegisterTestRunner("go", testengine.NewGoTestRunner(log))

	// Register test analyzers for supported languages
	engine.RegisterTestAnalyzer("go", testengine.NewGoTestAnalyzer(log))

	return &TestService{
		log:         log,
		testEngine:  engine,
		symbolGraph: symbolGraph,
	}
}

// RunTests executes tests according to configuration
func (s *TestService) RunTests(ctx context.Context, config *domain.TestConfig) ([]*domain.TestResult, error) {
	s.log.Info(fmt.Sprintf("Running tests with scope: %s", config.Scope))

	// If scope includes affected, build affected graph
	if config.Scope == domain.TestScopeAffected || config.Scope == domain.TestScopeAffectedSmoke {
		changedFiles := []string{} // In real usage, this comes from version control

		affectedGraph, err := s.testEngine.BuildAffectedGraph(ctx, changedFiles, config.ProjectPath)
		if err != nil {
			s.log.Warning(fmt.Sprintf("Failed to build affected graph: %v", err))
			return s.testEngine.RunTests(ctx, config)
		}

		return s.testEngine.RunTargetedTests(ctx, config, affectedGraph)
	}

	return s.testEngine.RunTests(ctx, config)
}

// RunSmokeTests executes only smoke tests
func (s *TestService) RunSmokeTests(ctx context.Context, projectPath, language string) ([]*domain.TestResult, error) {
	config := &domain.TestConfig{
		Language:    language,
		ProjectPath: projectPath,
		Scope:       domain.TestScopeSmoke,
		Timeout:     60,
		Verbose:     true,
	}
	return s.RunTests(ctx, config)
}

// ValidateTestResults validates test results
func (s *TestService) ValidateTestResults(results []*domain.TestResult) *domain.TestValidationResult {
	validation := &domain.TestValidationResult{
		TotalTests:      len(results),
		PassedTests:     0,
		FailedTests:     0,
		SkippedTests:    0,
		TotalDuration:   0.0,
		FailedTestPaths: []string{},
	}

	for _, result := range results {
		validation.TotalDuration += result.Duration

		if result.Success {
			validation.PassedTests++
		} else {
			validation.FailedTests++
			validation.FailedTestPaths = append(validation.FailedTestPaths, result.TestPath)
		}
	}

	validation.Success = validation.FailedTests == 0
	if validation.TotalTests > 0 {
		validation.SuccessRate = float64(validation.PassedTests) / float64(validation.TotalTests) * 100
	}

	return validation
}
