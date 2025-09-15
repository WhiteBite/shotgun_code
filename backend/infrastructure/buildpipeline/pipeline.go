package buildpipeline

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"shotgun_code/domain"
	"shotgun_code/infrastructure/sandbox"
	"strings"
	"time"
)

// BuildPipelineImpl реализует BuildPipeline
type BuildPipelineImpl struct {
	log           domain.Logger
	sandboxRunner domain.SandboxRunner
}

// NewBuildPipeline создает новый build pipeline
func NewBuildPipeline(log domain.Logger) *BuildPipelineImpl {
	return &BuildPipelineImpl{
		log:           log,
		sandboxRunner: sandbox.NewSandboxRunner(log),
	}
}

// Build выполняет сборку проекта
func (p *BuildPipelineImpl) Build(ctx context.Context, projectPath, language string) (*domain.BuildResult, error) {
	p.log.Info(fmt.Sprintf("Building %s project at %s", language, projectPath))

	startTime := time.Now()

	var result *domain.BuildResult
	var err error

	switch language {
	case "go":
		result, err = p.buildGo(ctx, projectPath)
	case "typescript", "ts":
		result, err = p.buildTypeScript(ctx, projectPath)
	case "java":
		result, err = p.buildJava(ctx, projectPath)
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	if err != nil {
		return nil, err
	}

	result.Duration = time.Since(startTime).Seconds()
	return result, nil
}

// TypeCheck выполняет проверку типов
func (p *BuildPipelineImpl) TypeCheck(ctx context.Context, projectPath, language string) (*domain.TypeCheckResult, error) {
	p.log.Info(fmt.Sprintf("Type checking %s project at %s", language, projectPath))

	startTime := time.Now()

	var result *domain.TypeCheckResult
	var err error

	switch language {
	case "go":
		result, err = p.typeCheckGo(ctx, projectPath)
	case "typescript", "ts":
		result, err = p.typeCheckTypeScript(ctx, projectPath)
	case "java":
		result, err = p.typeCheckJava(ctx, projectPath)
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	if err != nil {
		return nil, err
	}

	result.Duration = time.Since(startTime).Seconds()
	return result, nil
}

// BuildAndTypeCheck выполняет сборку и проверку типов
func (p *BuildPipelineImpl) BuildAndTypeCheck(ctx context.Context, projectPath, language string) (*domain.BuildResult, *domain.TypeCheckResult, error) {
	p.log.Info(fmt.Sprintf("Building and type checking %s project at %s", language, projectPath))

	// Сначала выполняем проверку типов
	typeCheckResult, err := p.TypeCheck(ctx, projectPath, language)
	if err != nil {
		return nil, nil, fmt.Errorf("type check failed: %w", err)
	}

	// Затем выполняем сборку
	buildResult, err := p.Build(ctx, projectPath, language)
	if err != nil {
		return nil, typeCheckResult, fmt.Errorf("build failed: %w", err)
	}

	return buildResult, typeCheckResult, nil
}

// BuildInSandbox выполняет сборку в песочнице
func (p *BuildPipelineImpl) BuildInSandbox(ctx context.Context, projectPath, language string, sandboxConfig domain.SandboxConfig) (*domain.BuildResult, error) {
	p.log.Info(fmt.Sprintf("Building %s project in sandbox at %s", language, projectPath))

	// Проверяем доступность песочницы
	if !p.sandboxRunner.IsAvailable(ctx) {
		p.log.Warning("Sandbox not available, falling back to local build")
		return p.Build(ctx, projectPath, language)
	}

	// Определяем команду для сборки
	var command []string
	switch language {
	case "go":
		command = []string{"go", "build", "./..."}
	case "java":
		if _, err := os.Stat(filepath.Join(projectPath, "pom.xml")); err == nil {
			command = []string{"mvn", "compile"}
		} else if _, err := os.Stat(filepath.Join(projectPath, "build.gradle")); err == nil {
			command = []string{"gradle", "build"}
		} else {
			return nil, fmt.Errorf("no build configuration found for Java project")
		}
	case "typescript", "ts":
		command = []string{"npm", "run", "build"}
	default:
		return nil, fmt.Errorf("unsupported language for sandbox build: %s", language)
	}

	// Настраиваем монтирование проекта
	if len(sandboxConfig.Mounts) == 0 {
		sandboxConfig.Mounts = []domain.SandboxMount{
			{
				Source:   projectPath,
				Target:   "/workspace",
				ReadOnly: false,
				Type:     "bind",
			},
		}
	}

	// Настраиваем рабочую директорию
	if sandboxConfig.WorkingDir == "" {
		sandboxConfig.WorkingDir = "/workspace"
	}

	// Настраиваем таймаут
	if sandboxConfig.Timeout == 0 {
		sandboxConfig.Timeout = 300 // 5 минут по умолчанию
	}

	// Запускаем команду в песочнице
	result, err := p.sandboxRunner.Run(ctx, sandboxConfig, command)
	if err != nil {
		return nil, fmt.Errorf("sandbox execution failed: %w", err)
	}

	// Конвертируем результат
	buildResult := &domain.BuildResult{
		Language:    language,
		ProjectPath: projectPath,
		Success:     result.Success,
		Output:      result.Output,
		Error:       result.Error,
		Duration:    result.Duration,
		Metadata: map[string]interface{}{
			"sandbox": true,
			"engine":  sandboxConfig.Engine,
			"image":   sandboxConfig.Image,
		},
	}

	if !result.Success {
		buildResult.Warnings = append(buildResult.Warnings, "Build executed in sandbox but failed")
	}

	return buildResult, nil
}

// GetSupportedLanguages возвращает поддерживаемые языки
func (p *BuildPipelineImpl) GetSupportedLanguages() []string {
	return []string{"go", "typescript", "ts", "java"}
}

// buildGo выполняет сборку Go проекта
func (p *BuildPipelineImpl) buildGo(ctx context.Context, projectPath string) (*domain.BuildResult, error) {
	result := &domain.BuildResult{
		Language:    "go",
		ProjectPath: projectPath,
	}

	// Проверяем наличие go.mod
	if _, err := os.Stat(filepath.Join(projectPath, "go.mod")); os.IsNotExist(err) {
		result.Success = false
		result.Error = "go.mod not found"
		return result, nil
	}

	// Выполняем go build
	cmd := exec.CommandContext(ctx, "go", "build", "-o", "shotgun.exe", ".")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, nil
	}

	result.Success = true

	// Ищем артефакты
	if _, err := os.Stat(filepath.Join(projectPath, "shotgun.exe")); err == nil {
		result.Artifacts = append(result.Artifacts, "shotgun.exe")
	}

	return result, nil
}

// typeCheckGo выполняет проверку типов Go проекта
func (p *BuildPipelineImpl) typeCheckGo(ctx context.Context, projectPath string) (*domain.TypeCheckResult, error) {
	result := &domain.TypeCheckResult{
		Language:    "go",
		ProjectPath: projectPath,
	}

	// Выполняем go vet
	cmd := exec.CommandContext(ctx, "go", "vet", "./...")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	if err != nil {
		result.Success = false
		result.Error = err.Error()

		// Парсим ошибки
		result.Issues = p.parseGoVetIssues(string(output))
		return result, nil
	}

	result.Success = true
	return result, nil
}

// buildTypeScript выполняет сборку TypeScript проекта
func (p *BuildPipelineImpl) buildTypeScript(ctx context.Context, projectPath string) (*domain.BuildResult, error) {
	result := &domain.BuildResult{
		Language:    "typescript",
		ProjectPath: projectPath,
	}

	// Проверяем наличие package.json
	if _, err := os.Stat(filepath.Join(projectPath, "package.json")); os.IsNotExist(err) {
		result.Success = false
		result.Error = "package.json not found"
		return result, nil
	}

	// Выполняем npm run build или tsc
	cmd := exec.CommandContext(ctx, "npm", "run", "build")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	if err != nil {
		// Пробуем tsc напрямую
		cmd = exec.CommandContext(ctx, "npx", "tsc")
		cmd.Dir = projectPath

		output, err = cmd.CombinedOutput()
		result.Output = string(output)

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			return result, nil
		}
	}

	result.Success = true

	// Ищем артефакты в dist/
	if distPath := filepath.Join(projectPath, "dist"); distPath != "" {
		if _, err := os.Stat(distPath); err == nil {
			result.Artifacts = append(result.Artifacts, "dist/")
		}
	}

	return result, nil
}

// typeCheckTypeScript выполняет проверку типов TypeScript проекта
func (p *BuildPipelineImpl) typeCheckTypeScript(ctx context.Context, projectPath string) (*domain.TypeCheckResult, error) {
	result := &domain.TypeCheckResult{
		Language:    "typescript",
		ProjectPath: projectPath,
	}

	// Выполняем tsc --noEmit
	cmd := exec.CommandContext(ctx, "npx", "tsc", "--noEmit")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	if err != nil {
		result.Success = false
		result.Error = err.Error()

		// Парсим ошибки TypeScript
		result.Issues = p.parseTypeScriptIssues(string(output))
		return result, nil
	}

	result.Success = true
	return result, nil
}

// buildJava выполняет сборку Java проекта
func (p *BuildPipelineImpl) buildJava(ctx context.Context, projectPath string) (*domain.BuildResult, error) {
	result := &domain.BuildResult{
		Language:    "java",
		ProjectPath: projectPath,
	}

	// Проверяем наличие Java среды
	if !p.checkJavaEnvironment() {
		result.Success = false
		result.Error = "Java environment not available (java, javac, mvn, or gradle not found)"
		result.Warnings = append(result.Warnings, "Java build tools not available, skipping build")
		return result, nil
	}

	// Проверяем наличие pom.xml или build.gradle
	if _, err := os.Stat(filepath.Join(projectPath, "pom.xml")); err == nil {
		// Maven проект
		result, err = p.buildMavenProject(ctx, projectPath)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Warnings = append(result.Warnings, "Maven build failed, but continuing")
		}
	} else if _, err := os.Stat(filepath.Join(projectPath, "build.gradle")); err == nil {
		// Gradle проект
		result, err = p.buildGradleProject(ctx, projectPath)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Warnings = append(result.Warnings, "Gradle build failed, but continuing")
		}
	} else {
		result.Success = false
		result.Error = "neither pom.xml nor build.gradle found"
		result.Warnings = append(result.Warnings, "No Java build configuration found")
	}

	return result, nil
}

// typeCheckJava выполняет проверку типов Java проекта
func (p *BuildPipelineImpl) typeCheckJava(ctx context.Context, projectPath string) (*domain.TypeCheckResult, error) {
	result := &domain.TypeCheckResult{
		Language:    "java",
		ProjectPath: projectPath,
	}

	// Проверяем наличие Java среды
	if !p.checkJavaEnvironment() {
		result.Success = false
		result.Error = "Java environment not available"
		return result, nil
	}

	// Проверяем наличие pom.xml или build.gradle
	if _, err := os.Stat(filepath.Join(projectPath, "pom.xml")); err == nil {
		// Maven проект - компиляция включает проверку типов
		cmd := exec.CommandContext(ctx, "mvn", "compile", "-q")
		cmd.Dir = projectPath

		output, err := cmd.CombinedOutput()
		result.Output = string(output)

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Issues = p.parseMavenIssues(string(output))
			return result, nil
		}

		result.Success = true

		// Дополнительно запускаем ErrorProne, если доступен
		if p.hasErrorPronePlugin(projectPath) {
			errorProneResult := p.runErrorProne(ctx, projectPath, "mvn")
			if !errorProneResult.Success {
				result.Issues = append(result.Issues, p.parseErrorProneIssues(errorProneResult.Output)...)
			}
		}
	} else if _, err := os.Stat(filepath.Join(projectPath, "build.gradle")); err == nil {
		// Gradle проект - компиляция включает проверку типов
		cmd := exec.CommandContext(ctx, "gradle", "compileJava", "--quiet")
		cmd.Dir = projectPath

		output, err := cmd.CombinedOutput()
		result.Output = string(output)

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Issues = p.parseGradleIssues(string(output))
			return result, nil
		}

		result.Success = true
	} else {
		result.Success = false
		result.Error = "neither pom.xml nor build.gradle found"
	}

	return result, nil
}

// parseGoVetIssues парсит ошибки go vet
func (p *BuildPipelineImpl) parseGoVetIssues(output string) []*domain.TypeIssue {
	var issues []*domain.TypeIssue

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 3)
			if len(parts) >= 3 {
				issue := &domain.TypeIssue{
					File:     parts[0],
					Severity: "warning",
					Message:  strings.TrimSpace(parts[2]),
				}
				issues = append(issues, issue)
			}
		}
	}

	return issues
}

// parseTypeScriptIssues парсит ошибки TypeScript
func (p *BuildPipelineImpl) parseTypeScriptIssues(output string) []*domain.TypeIssue {
	var issues []*domain.TypeIssue

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ".ts") && strings.Contains(line, "error TS") {
			// Пример: src/file.ts(10,15): error TS2307: Cannot find module './module'
			re := regexp.MustCompile(`([^(]+)\((\d+),(\d+)\):\s+error\s+(TS\d+):\s+(.+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 6 {
				issue := &domain.TypeIssue{
					File:     matches[1],
					Line:     p.parseInt(matches[2]),
					Column:   p.parseInt(matches[3]),
					Code:     matches[4],
					Message:  matches[5],
					Severity: "error",
				}
				issues = append(issues, issue)
			}
		}
	}

	return issues
}

// parseMavenIssues парсит ошибки Maven
func (p *BuildPipelineImpl) parseMavenIssues(output string) []*domain.TypeIssue {
	var issues []*domain.TypeIssue

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "[ERROR]") && strings.Contains(line, ".java") {
			// Пример: [ERROR] /path/to/file.java:[10,15] cannot find symbol
			re := regexp.MustCompile(`\[ERROR\]\s+([^:]+):\[(\d+),(\d+)\]\s+(.+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 5 {
				issue := &domain.TypeIssue{
					File:     matches[1],
					Line:     p.parseInt(matches[2]),
					Column:   p.parseInt(matches[3]),
					Message:  matches[4],
					Severity: "error",
				}
				issues = append(issues, issue)
			}
		}
	}

	return issues
}

// parseGradleIssues парсит ошибки Gradle
func (p *BuildPipelineImpl) parseGradleIssues(output string) []*domain.TypeIssue {
	var issues []*domain.TypeIssue

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "error:") && strings.Contains(line, ".java") {
			// Пример: /path/to/file.java:10:15: error: cannot find symbol
			re := regexp.MustCompile(`([^:]+):(\d+):(\d+):\s+error:\s+(.+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 5 {
				issue := &domain.TypeIssue{
					File:     matches[1],
					Line:     p.parseInt(matches[2]),
					Column:   p.parseInt(matches[3]),
					Message:  matches[4],
					Severity: "error",
				}
				issues = append(issues, issue)
			}
		}
	}

	return issues
}

// parseInt парсит строку в int
func (p *BuildPipelineImpl) parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

// checkJavaEnvironment проверяет наличие Java среды
func (p *BuildPipelineImpl) checkJavaEnvironment() bool {
	// Проверяем наличие java
	if _, err := exec.LookPath("java"); err != nil {
		p.log.Warning("Java runtime not found")
		return false
	}

	// Проверяем наличие javac
	if _, err := exec.LookPath("javac"); err != nil {
		p.log.Warning("Java compiler not found")
		return false
	}

	// Проверяем наличие mvn или gradle
	if _, err := exec.LookPath("mvn"); err != nil {
		if _, err := exec.LookPath("gradle"); err != nil {
			p.log.Warning("Neither Maven nor Gradle found")
			return false
		}
	}

	return true
}

// buildMavenProject выполняет сборку Maven проекта
func (p *BuildPipelineImpl) buildMavenProject(ctx context.Context, projectPath string) (*domain.BuildResult, error) {
	result := &domain.BuildResult{
		Language:    "java",
		ProjectPath: projectPath,
	}

	// Проверяем наличие Maven
	if _, err := exec.LookPath("mvn"); err != nil {
		result.Success = false
		result.Error = "Maven not found in PATH"
		result.Warnings = append(result.Warnings, "Maven build tool not available")
		return result, nil
	}

	// Выполняем Maven compile
	cmd := exec.CommandContext(ctx, "mvn", "compile", "-q")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Warnings = append(result.Warnings, "Maven compilation failed")
		return result, nil
	}

	result.Success = true
	result.Artifacts = append(result.Artifacts, "target/")

	// Пытаемся запустить тесты, если они есть
	if p.hasJUnitTests(projectPath) {
		testResult := p.runJUnitTests(ctx, projectPath, "mvn")
		if !testResult.Success {
			result.Warnings = append(result.Warnings, "JUnit tests failed: "+testResult.Error)
		}
	}

	return result, nil
}

// buildGradleProject выполняет сборку Gradle проекта
func (p *BuildPipelineImpl) buildGradleProject(ctx context.Context, projectPath string) (*domain.BuildResult, error) {
	result := &domain.BuildResult{
		Language:    "java",
		ProjectPath: projectPath,
	}

	// Проверяем наличие Gradle
	if _, err := exec.LookPath("gradle"); err != nil {
		result.Success = false
		result.Error = "Gradle not found in PATH"
		result.Warnings = append(result.Warnings, "Gradle build tool not available")
		return result, nil
	}

	// Выполняем Gradle build
	cmd := exec.CommandContext(ctx, "gradle", "build", "--quiet")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Warnings = append(result.Warnings, "Gradle build failed")
		return result, nil
	}

	result.Success = true
	result.Artifacts = append(result.Artifacts, "build/")

	// Пытаемся запустить тесты, если они есть
	if p.hasJUnitTests(projectPath) {
		testResult := p.runJUnitTests(ctx, projectPath, "gradle")
		if !testResult.Success {
			result.Warnings = append(result.Warnings, "JUnit tests failed: "+testResult.Error)
		}
	}

	return result, nil
}

// hasJUnitTests проверяет наличие JUnit тестов
func (p *BuildPipelineImpl) hasJUnitTests(projectPath string) bool {
	// Проверяем наличие тестовых директорий
	testDirs := []string{
		filepath.Join(projectPath, "src", "test", "java"),
		filepath.Join(projectPath, "test"),
		filepath.Join(projectPath, "tests"),
	}

	for _, dir := range testDirs {
		if _, err := os.Stat(dir); err == nil {
			return true
		}
	}

	return false
}

// runJUnitTests запускает JUnit тесты
func (p *BuildPipelineImpl) runJUnitTests(ctx context.Context, projectPath, buildTool string) *domain.BuildResult {
	result := &domain.BuildResult{
		Language:    "java",
		ProjectPath: projectPath,
	}

	var cmd *exec.Cmd
	if buildTool == "mvn" {
		cmd = exec.CommandContext(ctx, "mvn", "test", "-q")
	} else {
		cmd = exec.CommandContext(ctx, "gradle", "test", "--quiet")
	}
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Warnings = append(result.Warnings, "JUnit test execution failed")
		return result
	}

	result.Success = true
	result.Artifacts = append(result.Artifacts, "test-results/")
	return result
}

// runErrorProne запускает ErrorProne анализ
func (p *BuildPipelineImpl) runErrorProne(ctx context.Context, projectPath, buildTool string) *domain.BuildResult {
	result := &domain.BuildResult{
		Language:    "java",
		ProjectPath: projectPath,
	}

	// ErrorProne доступен только для Maven
	if buildTool != "mvn" {
		result.Success = false
		result.Error = "ErrorProne only supported with Maven"
		result.Warnings = append(result.Warnings, "ErrorProne analysis skipped (Gradle not supported)")
		return result
	}

	// Проверяем наличие ErrorProne plugin в pom.xml
	if !p.hasErrorPronePlugin(projectPath) {
		result.Success = false
		result.Error = "ErrorProne plugin not configured"
		result.Warnings = append(result.Warnings, "ErrorProne analysis skipped (plugin not configured)")
		return result
	}

	// Запускаем ErrorProne анализ
	cmd := exec.CommandContext(ctx, "mvn", "compile", "-Perror-prone")
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Warnings = append(result.Warnings, "ErrorProne analysis failed")
		return result
	}

	result.Success = true
	result.Artifacts = append(result.Artifacts, "error-prone-reports/")
	return result
}

// hasErrorPronePlugin проверяет наличие ErrorProne plugin в pom.xml
func (p *BuildPipelineImpl) hasErrorPronePlugin(projectPath string) bool {
	pomPath := filepath.Join(projectPath, "pom.xml")
	content, err := os.ReadFile(pomPath)
	if err != nil {
		return false
	}

	return strings.Contains(string(content), "error-prone") ||
		strings.Contains(string(content), "errorprone")
}

// parseErrorProneIssues парсит ошибки ErrorProne
func (p *BuildPipelineImpl) parseErrorProneIssues(output string) []*domain.TypeIssue {
	var issues []*domain.TypeIssue

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "[ERROR]") && strings.Contains(line, "error-prone") {
			// Пример: [ERROR] /path/to/file.java:[10,15] error: [SomeError] description
			re := regexp.MustCompile(`\[ERROR\]\s+([^:]+):\[(\d+),(\d+)\]\s+error:\s+\[([^\]]+)\]\s+(.+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 6 {
				issue := &domain.TypeIssue{
					File:     matches[1],
					Line:     p.parseInt(matches[2]),
					Column:   p.parseInt(matches[3]),
					Code:     matches[4],
					Message:  matches[5],
					Severity: "error",
				}
				issues = append(issues, issue)
			}
		}
	}

	return issues
}
