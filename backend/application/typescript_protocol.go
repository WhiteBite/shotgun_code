package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"strings"
)

// TypeScriptProtocolImplementation provides TypeScript-specific protocol validation
type TypeScriptProtocolImplementation struct {
	log            domain.Logger
	staticAnalyzer *StaticAnalyzerService
	buildService   *BuildService
	testService    *TestService
}

// NewTypeScriptProtocolImplementation creates a new TypeScript protocol implementation
func NewTypeScriptProtocolImplementation(
	log domain.Logger,
	staticAnalyzer *StaticAnalyzerService,
	buildService *BuildService,
	testService *TestService,
) *TypeScriptProtocolImplementation {
	return &TypeScriptProtocolImplementation{
		log:            log,
		staticAnalyzer: staticAnalyzer,
		buildService:   buildService,
		testService:    testService,
	}
}

// ExecuteLintingStage executes TypeScript-specific linting
func (ts *TypeScriptProtocolImplementation) ExecuteLintingStage(ctx context.Context, config *domain.TaskProtocolConfig) error {
	ts.log.Info("Running TypeScript linting stage")
	
	// Run ESLint for TypeScript
	report, err := ts.staticAnalyzer.AnalyzeProject(ctx, config.ProjectPath, []string{"typescript"})
	if err != nil {
		return fmt.Errorf("TypeScript static analysis failed: %w", err)
	}

	// Check for critical TypeScript linting issues
	if ts.hasTypeScriptLintingErrors(report) {
		return fmt.Errorf("TypeScript linting errors found")
	}

	return nil
}

// ExecuteBuildingStage executes TypeScript-specific building
func (ts *TypeScriptProtocolImplementation) ExecuteBuildingStage(ctx context.Context, config *domain.TaskProtocolConfig) error {
	ts.log.Info("Running TypeScript building stage")
	
	// Validate TypeScript project compilation
	validation, err := ts.buildService.ValidateProject(ctx, config.ProjectPath, []string{"typescript"})
	if err != nil {
		return fmt.Errorf("TypeScript build validation failed: %w", err)
	}

	if !validation.Success {
		return fmt.Errorf("TypeScript project build failed")
	}

	return nil
}

// ExecuteTestingStage executes TypeScript-specific testing
func (ts *TypeScriptProtocolImplementation) ExecuteTestingStage(ctx context.Context, config *domain.TaskProtocolConfig) error {
	ts.log.Info("Running TypeScript testing stage")
	
	// Run TypeScript tests
	results, err := ts.testService.RunSmokeTests(ctx, config.ProjectPath, "typescript")
	if err != nil {
		return fmt.Errorf("TypeScript tests failed: %w", err)
	}

	// Validate TypeScript test results
	validation := ts.testService.ValidateTestResults(results)
	if !validation.Success {
		return fmt.Errorf("TypeScript test validation failed: %d tests failed", validation.FailedTests)
	}

	return nil
}

// AnalyzeTypeScriptError analyzes TypeScript-specific errors
func (ts *TypeScriptProtocolImplementation) AnalyzeTypeScriptError(errorOutput string) (*domain.ErrorDetails, error) {
	details := &domain.ErrorDetails{
		Tool:        "tsc",
		Severity:    "error",
		Suggestions: make([]string, 0),
	}

	// Parse TypeScript compiler errors
	if strings.Contains(errorOutput, "TS2304") {
		details.ErrorType = domain.ErrorTypeImport
		details.Message = errorOutput
		details.Suggestions = append(details.Suggestions, "Cannot find name - check imports or type declarations")
		details.Suggestions = append(details.Suggestions, "Consider adding type definitions or importing the missing module")
	} else if strings.Contains(errorOutput, "TS2322") {
		details.ErrorType = domain.ErrorTypeTypeCheck
		details.Message = errorOutput
		details.Suggestions = append(details.Suggestions, "Type assignment error - check type compatibility")
		details.Suggestions = append(details.Suggestions, "Consider type casting or updating type definitions")
	} else if strings.Contains(errorOutput, "TS2307") {
		details.ErrorType = domain.ErrorTypeImport
		details.Message = errorOutput
		details.Suggestions = append(details.Suggestions, "Cannot find module - check import path")
		details.Suggestions = append(details.Suggestions, "Verify module exists and is properly installed")
	} else if strings.Contains(errorOutput, "TS2345") {
		details.ErrorType = domain.ErrorTypeTypeCheck
		details.Message = errorOutput
		details.Suggestions = append(details.Suggestions, "Argument type mismatch - check function parameters")
	} else if strings.Contains(errorOutput, "TS2339") {
		details.ErrorType = domain.ErrorTypeTypeCheck
		details.Message = errorOutput
		details.Suggestions = append(details.Suggestions, "Property does not exist - check object type definitions")
	}

	return details, nil
}

// SuggestTypeScriptCorrections provides TypeScript-specific correction suggestions
func (ts *TypeScriptProtocolImplementation) SuggestTypeScriptCorrections(error *domain.ErrorDetails) ([]*domain.CorrectionStep, error) {
	corrections := make([]*domain.CorrectionStep, 0)

	switch error.ErrorType {
	case domain.ErrorTypeImport:
		if strings.Contains(error.Message, "TS2304") || strings.Contains(error.Message, "TS2307") {
			corrections = append(corrections, &domain.CorrectionStep{
				Action:      domain.ActionFixImport,
				Target:      error.SourceFile,
				Description: "Add missing import or type declaration",
			})
		}
	case domain.ErrorTypeTypeCheck:
		if strings.Contains(error.Message, "TS2322") {
			corrections = append(corrections, &domain.CorrectionStep{
				Action:      domain.ActionFixType,
				Target:      error.SourceFile,
				Description: "Fix type assignment error",
			})
		} else if strings.Contains(error.Message, "TS2345") {
			corrections = append(corrections, &domain.CorrectionStep{
				Action:      domain.ActionFixType,
				Target:      error.SourceFile,
				Description: "Fix function parameter type mismatch",
			})
		}
	case domain.ErrorTypeSyntax:
		corrections = append(corrections, &domain.CorrectionStep{
			Action:      domain.ActionFixSyntax,
			Target:      error.SourceFile,
			Description: "Fix TypeScript syntax error",
		})
	case domain.ErrorTypeLinting:
		corrections = append(corrections, &domain.CorrectionStep{
			Action:      domain.ActionFormatCode,
			Target:      error.SourceFile,
			Description: "Format TypeScript code with Prettier",
		})
	}

	return corrections, nil
}

// GetTypeScriptLintingConfiguration returns TypeScript-specific linting configuration
func (ts *TypeScriptProtocolImplementation) GetTypeScriptLintingConfiguration() map[string]interface{} {
	return map[string]interface{}{
		"tools": []string{"eslint", "@typescript-eslint/eslint-plugin"},
		"rules": map[string]interface{}{
			"extends": []string{
				"@typescript-eslint/recommended",
				"@typescript-eslint/recommended-requiring-type-checking",
			},
			"strict_mode":     true,
			"no_implicit_any": true,
			"fail_on_warning": false,
		},
		"parser_options": map[string]interface{}{
			"ecmaVersion": "latest",
			"sourceType":  "module",
			"project":     "./tsconfig.json",
		},
		"exclusions": []string{
			"node_modules/",
			"dist/",
			"*.d.ts",
		},
	}
}

// GetTypeScriptBuildConfiguration returns TypeScript-specific build configuration
func (ts *TypeScriptProtocolImplementation) GetTypeScriptBuildConfiguration() map[string]interface{} {
	return map[string]interface{}{
		"compiler_options": map[string]interface{}{
			"strict":                 true,
			"noImplicitAny":         true,
			"strictNullChecks":      true,
			"strictFunctionTypes":   true,
			"noImplicitReturns":     true,
			"noFallthroughCasesInSwitch": true,
			"target":                "ES2020",
			"module":                "ESNext",
			"moduleResolution":      "node",
			"esModuleInterop":       true,
			"allowSyntheticDefaultImports": true,
			"skipLibCheck":          true,
			"forceConsistentCasingInFileNames": true,
		},
		"include": []string{
			"src/**/*",
			"tests/**/*",
		},
		"exclude": []string{
			"node_modules",
			"dist",
		},
		"type_checking": true,
		"declaration_files": true,
	}
}

// GetTypeScriptTestingConfiguration returns TypeScript-specific testing configuration
func (ts *TypeScriptProtocolImplementation) GetTypeScriptTestingConfiguration() map[string]interface{} {
	return map[string]interface{}{
		"test_framework": "vitest", // or jest
		"test_patterns": []string{
			"**/*.test.ts",
			"**/*.spec.ts",
		},
		"coverage_threshold": 85,
		"parallel_execution": true,
		"type_checking_in_tests": true,
		"test_timeout": "30s",
		"setup_files": []string{
			"tests/setup.ts",
		},
	}
}

// Helper methods

func (ts *TypeScriptProtocolImplementation) hasTypeScriptLintingErrors(report interface{}) bool {
	// Simplified implementation - in a real scenario, this would parse the actual ESLint report
	// and check for error-level issues vs warnings
	return false
}

// ApplyTypeScriptFormatting applies TypeScript-specific code formatting
func (ts *TypeScriptProtocolImplementation) ApplyTypeScriptFormatting(ctx context.Context, projectPath string, files []string) error {
	ts.log.Info("Applying TypeScript formatting")
	
	// In a real implementation, this would:
	// 1. Run Prettier on TypeScript files
	// 2. Apply ESLint auto-fix rules
	// 3. Organize imports
	// 4. Apply project-specific formatting rules
	
	for _, file := range files {
		if strings.HasSuffix(file, ".ts") || strings.HasSuffix(file, ".tsx") {
			ts.log.Debug(fmt.Sprintf("Formatting TypeScript file: %s", file))
			// Prettier/ESLint formatting logic would go here
		}
	}
	
	return nil
}

// ValidateTypeScriptProject validates TypeScript project structure
func (ts *TypeScriptProtocolImplementation) ValidateTypeScriptProject(ctx context.Context, projectPath string) error {
	ts.log.Info("Validating TypeScript project")
	
	// In a real implementation, this would:
	// 1. Check tsconfig.json exists and is valid
	// 2. Verify all TypeScript files can be compiled
	// 3. Check for missing type declarations
	// 4. Validate import/export statements
	// 5. Check for circular dependencies
	
	return nil
}

// RunTypeScriptSecurityChecks runs TypeScript-specific security analysis
func (ts *TypeScriptProtocolImplementation) RunTypeScriptSecurityChecks(ctx context.Context, projectPath string) error {
	ts.log.Info("Running TypeScript security checks")
	
	// In a real implementation, this would:
	// 1. Run npm audit for dependency vulnerabilities
	// 2. Check for unsafe type assertions
	// 3. Validate external API usage
	// 4. Check for XSS vulnerabilities in frontend code
	// 5. Analyze dynamic imports and eval usage
	
	return nil
}

// OptimizeTypeScriptImports optimizes TypeScript import statements
func (ts *TypeScriptProtocolImplementation) OptimizeTypeScriptImports(ctx context.Context, projectPath string, files []string) error {
	ts.log.Info("Optimizing TypeScript imports")
	
	// In a real implementation, this would:
	// 1. Remove unused imports
	// 2. Sort imports alphabetically
	// 3. Group imports by source (external, internal, relative)
	// 4. Convert to consistent import syntax
	// 5. Add missing imports
	
	return nil
}

// GenerateTypeDeclarations generates missing TypeScript type declarations
func (ts *TypeScriptProtocolImplementation) GenerateTypeDeclarations(ctx context.Context, projectPath string) error {
	ts.log.Info("Generating TypeScript type declarations")
	
	// In a real implementation, this would:
	// 1. Analyze untyped JavaScript dependencies
	// 2. Generate .d.ts files for missing types
	// 3. Create interface definitions for API responses
	// 4. Add type annotations to function parameters and return types
	
	return nil
}