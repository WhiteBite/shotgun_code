package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"sort"
	"strings"
	"time"
)

// RouterPlannerService предоставляет эвристический планировщик TPL (DAG) без LLM
type RouterPlannerService struct {
	log            domain.Logger
	buildService   domain.IBuildService
	testService    domain.ITestService
	staticAnalyzer domain.IStaticAnalyzerService
	repairService  domain.RepairService
	// Добавьте другие необходимые сервисы
}

// NewRouterPlannerService создает новый сервис планировщика
func NewRouterPlannerService(
	log domain.Logger,
	buildService domain.IBuildService,
	testService domain.ITestService,
	staticAnalyzer domain.IStaticAnalyzerService,
	repairService domain.RepairService,
) *RouterPlannerService {
	return &RouterPlannerService{
		log:            log,
		buildService:   buildService,
		testService:    testService,
		staticAnalyzer: staticAnalyzer,
		repairService:  repairService,
	}
}

// TaskPipelineStep представляет шаг в пайплайне задачи
type TaskPipelineStep struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Type        TaskPipelineStepType    `json:"type"`
	Status      TaskPipelineStepStatus  `json:"status"`
	Priority    int                     `json:"priority"`
	DependsOn   []string                `json:"depends_on"`
	Config      map[string]interface{}  `json:"config"`
	StartedAt   *time.Time              `json:"started_at,omitempty"`
	CompletedAt *time.Time              `json:"completed_at,omitempty"`
	Duration    time.Duration           `json:"duration,omitempty"`
	Error       string                  `json:"error,omitempty"`
	Result      *TaskPipelineStepResult `json:"result,omitempty"`
}

// TaskPipelineStepType определяет тип шага пайплайна
type TaskPipelineStepType string

const (
	StepTypeRetrieve TaskPipelineStepType = "retrieve"
	StepTypeASTSynth TaskPipelineStepType = "ast_synth"
	StepTypeCompile  TaskPipelineStepType = "compile"
	StepTypeTest     TaskPipelineStepType = "test"
	StepTypeStatic   TaskPipelineStepType = "static"
	StepTypeRepair   TaskPipelineStepType = "repair"
	StepTypeFormat   TaskPipelineStepType = "format"
	StepTypeValidate TaskPipelineStepType = "validate"
)

// TaskPipelineStepStatus определяет статус шага пайплайна
type TaskPipelineStepStatus string

const (
	StepStatusPending   TaskPipelineStepStatus = "pending"
	StepStatusRunning   TaskPipelineStepStatus = "running"
	StepStatusCompleted TaskPipelineStepStatus = "completed"
	StepStatusFailed    TaskPipelineStepStatus = "failed"
)

// TaskPipelineStepResult содержит результат выполнения шага
type TaskPipelineStepResult struct {
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Metrics     map[string]float64     `json:"metrics,omitempty"`
	Artifacts   []string               `json:"artifacts,omitempty"`
	Warnings    []string               `json:"warnings,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

// TaskPipeline представляет полный пайплайн для задачи
type TaskPipeline struct {
	TaskID      string              `json:"task_id"`
	Steps       []*TaskPipelineStep `json:"steps"`
	Status      TaskPipelineStatus  `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	StartedAt   *time.Time          `json:"started_at,omitempty"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
	Duration    time.Duration       `json:"duration,omitempty"`
	Error       string              `json:"error,omitempty"`
	Policy      *PipelinePolicy     `json:"policy"`
}

// TaskPipelineStatus определяет статус пайплайна
type TaskPipelineStatus string

const (
	PipelineStatusPending   TaskPipelineStatus = "pending"
	PipelineStatusRunning   TaskPipelineStatus = "running"
	PipelineStatusCompleted TaskPipelineStatus = "completed"
	PipelineStatusFailed    TaskPipelineStatus = "failed"
)

// PipelinePolicy определяет политику выполнения пайплайна
type PipelinePolicy struct {
	EnableRetrieve bool          `json:"enable_retrieve"`
	EnableASTSynth bool          `json:"enable_ast_synth"`
	EnableCompile  bool          `json:"enable_compile"`
	EnableTest     bool          `json:"enable_test"`
	EnableStatic   bool          `json:"enable_static"`
	EnableRepair   bool          `json:"enable_repair"`
	EnableFormat   bool          `json:"enable_format"`
	EnableValidate bool          `json:"enable_validate"`
	FailFast       bool          `json:"fail_fast"`
	RetryFailed    bool          `json:"retry_failed"`
	MaxRetries     int           `json:"max_retries"`
	ParallelSteps  bool          `json:"parallel_steps"`
	Timeout        time.Duration `json:"timeout"`
}

// getDependsOn determines the dependency for a step based on enabled policies
func getDependsOn(taskID string, stepID int, policy *PipelinePolicy, checkOrder []bool) []string {
	for _, enabled := range checkOrder {
		if enabled {
			return []string{fmt.Sprintf("%s-step-%d", taskID, stepID-1)}
		}
	}
	return []string{fmt.Sprintf("%s-step-%d", taskID, 1)}
}

// CreatePipeline создает пайплайн для задачи
func (r *RouterPlannerService) CreatePipeline(ctx context.Context, task domain.Task, policy *PipelinePolicy) (*TaskPipeline, error) {
	r.log.Info(fmt.Sprintf("Creating pipeline for task: %s", task.ID))

	// Если политика не предоставлена, определяем ее на основе типа задачи
	if policy == nil {
		policy = r.determinePolicy(task)
	}

	// Создаем шаги пайплайна
	steps := r.createPipelineSteps(task, policy)

	// Сортируем шаги по приоритету и зависимостям
	r.sortPipelineSteps(steps)

	pipeline := &TaskPipeline{
		TaskID:    task.ID,
		Steps:     steps,
		Status:    PipelineStatusPending,
		CreatedAt: time.Now(),
		Policy:    policy,
	}

	r.log.Info(fmt.Sprintf("Created pipeline with %d steps for task: %s", len(steps), task.ID))
	return pipeline, nil
}

// ExecutePipeline выполняет пайплайн
func (r *RouterPlannerService) ExecutePipeline(ctx context.Context, pipeline *TaskPipeline) error {
	r.log.Info(fmt.Sprintf("Executing pipeline for task: %s", pipeline.TaskID))

	now := time.Now()
	pipeline.StartedAt = &now
	pipeline.Status = PipelineStatusRunning

	// Выполняем шаги последовательно или параллельно
	if pipeline.Policy.ParallelSteps {
		return r.executePipelineParallel(ctx, pipeline)
	}
	return r.executePipelineSequential(ctx, pipeline)
}

// determinePolicy определяет политику выполнения на основе задачи
func (r *RouterPlannerService) determinePolicy(task domain.Task) *PipelinePolicy {
	policy := &PipelinePolicy{
		EnableRetrieve: true,
		EnableASTSynth: true,
		EnableCompile:  true,
		EnableTest:     true,
		EnableStatic:   true,
		EnableRepair:   true,
		EnableFormat:   true,
		EnableValidate: true,
		FailFast:       true,
		RetryFailed:    true,
		MaxRetries:     3,
		ParallelSteps:  false,
		Timeout:        30 * time.Minute,
	}

	// Настройки на основе типа задачи
	switch {
	case strings.Contains(task.ID, "ark-100"):
		// Правила и конфигурация - только валидация
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableStatic = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-110"):
		// Индексатор - retrieve + ast_synth
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableStatic = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-120"):
		// Context Pack - retrieve + ast_synth + validate
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableStatic = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-130"):
		// Apply Engine - полный пайплайн
		policy.EnableRepair = true
	case strings.Contains(task.ID, "ark-140"):
		// Derived Diff - retrieve + ast_synth + validate
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableStatic = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-150"):
		// Verification Pipeline - полный пайплайн
		policy.EnableRepair = true
	case strings.Contains(task.ID, "ark-160"):
		// Targeted Tests - retrieve + test
		policy.EnableASTSynth = false
		policy.EnableCompile = false
		policy.EnableStatic = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-170"):
		// Static Analyzers - retrieve + static
		policy.EnableASTSynth = false
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableRepair = false
	case strings.Contains(task.ID, "ark-180"):
		// Router Planner - только планирование
		policy.EnableCompile = false
		policy.EnableTest = false
		policy.EnableStatic = false
		policy.EnableRepair = false
		policy.EnableFormat = false
	case strings.Contains(task.ID, "ark-999"):
		// Final Gate - полный пайплайн с валидацией
		policy.EnableRepair = true
		policy.FailFast = true
	}

	return policy
}

// stepBuilder helps build pipeline steps
type stepBuilder struct {
	taskID  string
	stepID  int
	steps   []*TaskPipelineStep
	prevDep string
}

func newStepBuilder(taskID string) *stepBuilder {
	return &stepBuilder{taskID: taskID, stepID: 1, steps: make([]*TaskPipelineStep, 0)}
}

func (sb *stepBuilder) addStep(name string, stepType TaskPipelineStepType, priority int, config map[string]interface{}) {
	dependsOn := []string{}
	if sb.prevDep != "" {
		dependsOn = []string{sb.prevDep}
	}
	step := &TaskPipelineStep{
		ID: fmt.Sprintf("%s-step-%d", sb.taskID, sb.stepID), Name: name, Type: stepType,
		Status: StepStatusPending, Priority: priority, DependsOn: dependsOn, Config: config,
	}
	sb.steps = append(sb.steps, step)
	sb.prevDep = step.ID
	sb.stepID++
}

// createPipelineSteps создает шаги пайплайна
func (r *RouterPlannerService) createPipelineSteps(task domain.Task, policy *PipelinePolicy) []*TaskPipelineStep {
	sb := newStepBuilder(task.ID)

	if policy.EnableRetrieve {
		sb.addStep("Retrieve Context", StepTypeRetrieve, 1, map[string]interface{}{
			"task_id": task.ID, "step_file": task.StepFile, "budgets": task.Budgets, "dependencies": task.DependsOn,
		})
	}
	if policy.EnableASTSynth {
		sb.addStep("AST Synthesis", StepTypeASTSynth, 2, map[string]interface{}{
			"task_id": task.ID, "symbol_graph": true, "context_pack": true,
		})
	}
	if policy.EnableCompile {
		sb.addStep("Compile", StepTypeCompile, 3, map[string]interface{}{
			"task_id": task.ID, "project_path": task.Metadata["project_path"], "language": "go", "build_mode": "debug",
		})
	}
	if policy.EnableTest {
		sb.addStep("Test", StepTypeTest, 4, map[string]interface{}{
			"task_id": task.ID, "project_path": task.Metadata["project_path"], "test_mode": "targeted", "coverage": true,
		})
	}
	if policy.EnableStatic {
		sb.addStep("Static Analysis", StepTypeStatic, 5, map[string]interface{}{
			"task_id": task.ID, "project_path": task.Metadata["project_path"],
			"analyzers": []string{"staticcheck", "go vet", "eslint"}, "fail_on_error": false,
		})
	}
	if policy.EnableFormat {
		sb.addStep("Format", StepTypeFormat, 6, map[string]interface{}{
			"task_id": task.ID, "formatters": []string{"gofmt", "goimports", "prettier"},
		})
	}
	if policy.EnableValidate {
		sb.addStep("Validate", StepTypeValidate, 7, map[string]interface{}{
			"task_id": task.ID, "validation_rules": []string{"syntax", "semantics", "policy"},
		})
	}
	if policy.EnableRepair {
		sb.addStep("Repair", StepTypeRepair, 8, map[string]interface{}{
			"task_id": task.ID, "project_path": task.Metadata["project_path"],
			"repair_strategies": []string{"auto_fix", "suggestions", "rollback"}, "max_attempts": policy.MaxRetries,
		})
	}

	return sb.steps
}

// sortPipelineSteps сортирует шаги по приоритету и зависимостям
func (r *RouterPlannerService) sortPipelineSteps(steps []*TaskPipelineStep) {
	sort.Slice(steps, func(i, j int) bool {
		// Сначала по приоритету
		if steps[i].Priority != steps[j].Priority {
			return steps[i].Priority < steps[j].Priority
		}
		// Затем по количеству зависимостей
		return len(steps[i].DependsOn) < len(steps[j].DependsOn)
	})
}

// executePipelineSequential выполняет пайплайн последовательно
func (r *RouterPlannerService) executePipelineSequential(ctx context.Context, pipeline *TaskPipeline) error {
	for _, step := range pipeline.Steps {
		if err := r.executeStep(ctx, step); err != nil {
			if pipeline.Policy.FailFast {
				pipeline.Status = PipelineStatusFailed
				pipeline.Error = err.Error()
				now := time.Now()
				pipeline.CompletedAt = &now
				pipeline.Duration = now.Sub(*pipeline.StartedAt)
				return fmt.Errorf("step %s failed: %w", step.ID, err)
			}
			// Продолжаем выполнение других шагов
		}
	}

	pipeline.Status = PipelineStatusCompleted
	now := time.Now()
	pipeline.CompletedAt = &now
	pipeline.Duration = now.Sub(*pipeline.StartedAt)

	r.log.Info(fmt.Sprintf("Pipeline completed for task: %s", pipeline.TaskID))
	return nil
}

// executePipelineParallel выполняет пайплайн параллельно
func (r *RouterPlannerService) executePipelineParallel(ctx context.Context, pipeline *TaskPipeline) error {
	// Упрощенная реализация параллельного выполнения
	// В реальной реализации нужно учитывать зависимости между шагами
	return r.executePipelineSequential(ctx, pipeline)
}

// executeStep выполняет один шаг пайплайна
func (r *RouterPlannerService) executeStep(ctx context.Context, step *TaskPipelineStep) error {
	r.log.Info(fmt.Sprintf("Executing step: %s (%s)", step.Name, step.ID))

	now := time.Now()
	step.StartedAt = &now
	step.Status = StepStatusRunning

	// Выполняем шаг в зависимости от типа
	var err error
	switch step.Type {
	case StepTypeRetrieve:
		err = r.executeRetrieveStep(ctx, step)
	case StepTypeASTSynth:
		err = r.executeASTSynthStep(ctx, step)
	case StepTypeCompile:
		err = r.executeCompileStep(ctx, step)
	case StepTypeTest:
		err = r.executeTestStep(ctx, step)
	case StepTypeStatic:
		err = r.executeStaticStep(ctx, step)
	case StepTypeFormat:
		err = r.executeFormatStep(ctx, step)
	case StepTypeValidate:
		err = r.executeValidateStep(ctx, step)
	case StepTypeRepair:
		err = r.executeRepairStep(ctx, step)
	default:
		err = fmt.Errorf("unknown step type: %s", step.Type)
	}

	// Обновляем статус шага
	completedAt := time.Now()
	step.CompletedAt = &completedAt
	step.Duration = completedAt.Sub(*step.StartedAt)

	if err != nil {
		step.Status = StepStatusFailed
		step.Error = err.Error()
		r.log.Error(fmt.Sprintf("Step %s failed: %v", step.ID, err))
	} else {
		step.Status = StepStatusCompleted
		step.Result = &TaskPipelineStepResult{
			Success: true,
			Message: fmt.Sprintf("Step %s completed successfully", step.Name),
		}
		r.log.Info(fmt.Sprintf("Step %s completed successfully", step.ID))
	}

	return err
}

// executeRetrieveStep выполняет шаг извлечения контекста
func (r *RouterPlannerService) executeRetrieveStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага retrieve
	time.Sleep(100 * time.Millisecond)
	return nil
}

// executeASTSynthStep выполняет шаг синтеза AST
func (r *RouterPlannerService) executeASTSynthStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага ast_synth
	time.Sleep(200 * time.Millisecond)
	return nil
}

// executeCompileStep выполняет шаг компиляции
func (r *RouterPlannerService) executeCompileStep(ctx context.Context, step *TaskPipelineStep) error {
	projectPath, ok := step.Config["project_path"].(string)
	if !ok {
		return fmt.Errorf("project_path not found in step config")
	}
	language, ok := step.Config["language"].(string)
	if !ok {
		language = "go" // default to go
	}

	result, err := r.buildService.Build(ctx, projectPath, language)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	step.Result = &TaskPipelineStepResult{
		Success: result.Success,
		Message: result.Output,
		Data: map[string]interface{}{
			"duration": result.Duration,
		},
		Artifacts: result.Artifacts,
	}

	if !result.Success {
		return fmt.Errorf("compilation failed: %s", result.Output)
	}

	return nil
}

// executeTestStep выполняет шаг тестирования
func (r *RouterPlannerService) executeTestStep(ctx context.Context, step *TaskPipelineStep) error {
	projectPath, ok := step.Config["project_path"].(string)
	if !ok {
		// Пытаемся получить из зависимостей, если не нашли в конфиге
		// Это временное решение, в идеале project_path должен быть везде
		if pipeline, ok := ctx.Value("pipeline").(*TaskPipeline); ok {
			if task, ok := pipeline.Steps[0].Config["task"].(domain.Task); ok {
				projectPath = task.Metadata["project_path"].(string)
			}
		}
		if projectPath == "" {
			return fmt.Errorf("project_path not found for test step")
		}
	}

	testConfig := &domain.TestConfig{
		ProjectPath: projectPath,
		Language:    "go", // Пока что хардкод
		Scope:       domain.TestScopeAll,
		Coverage:    true,
		Timeout:     int((5 * time.Minute).Seconds()),
	}

	results, err := r.testService.RunTests(ctx, testConfig)
	if err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// Агрегируем результаты
	success := true
	messages := make([]string, 0, len(results))
	totalDuration := 0.0
	for _, res := range results {
		if !res.Success {
			success = false
		}
		messages = append(messages, fmt.Sprintf("Test: %s, Success: %v, Output: %s", res.TestName, res.Success, res.Output))
		totalDuration += res.Duration
	}

	step.Result = &TaskPipelineStepResult{
		Success: success,
		Message: strings.Join(messages, "\n"),
		Data: map[string]interface{}{
			"duration":  totalDuration,
			"tests_run": len(results),
		},
	}

	if !success {
		return fmt.Errorf("one or more tests failed")
	}

	return nil
}

// executeStaticStep выполняет шаг статического анализа
func (r *RouterPlannerService) executeStaticStep(ctx context.Context, step *TaskPipelineStep) error {
	projectPath, ok := step.Config["project_path"].(string)
	if !ok {
		// Временное решение для получения project_path
		if pipeline, ok := ctx.Value("pipeline").(*TaskPipeline); ok {
			if task, ok := pipeline.Steps[0].Config["task"].(domain.Task); ok {
				projectPath = task.Metadata["project_path"].(string)
			}
		}
		if projectPath == "" {
			return fmt.Errorf("project_path not found for static analysis step")
		}
	}

	languages, ok := step.Config["languages"].([]string)
	if !ok {
		languages = []string{"go"} // default
	}

	failOnError, ok := step.Config["fail_on_error"].(bool)
	if !ok {
		failOnError = false // default
	}

	results, err := r.staticAnalyzer.AnalyzeProject(ctx, projectPath, languages)
	if err != nil {
		return fmt.Errorf("static analysis failed: %w", err)
	}

	// Агрегируем результаты
	success := true
	totalIssues := 0
	totalDuration := 0.0
	for _, result := range results.Results {
		totalIssues += len(result.Issues)
		totalDuration += result.Duration
	}

	if failOnError && totalIssues > 0 {
		success = false
	}

	step.Result = &TaskPipelineStepResult{
		Success: success,
		Message: fmt.Sprintf("Static analysis completed with %d issues.", totalIssues),
		Data: map[string]interface{}{
			"total_issues": totalIssues,
			"duration":     totalDuration,
		},
	}

	if !success {
		return fmt.Errorf("static analysis found critical issues")
	}

	return nil
}

// executeFormatStep выполняет шаг форматирования
func (r *RouterPlannerService) executeFormatStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага format
	time.Sleep(150 * time.Millisecond)
	return nil
}

// executeValidateStep выполняет шаг валидации
func (r *RouterPlannerService) executeValidateStep(ctx context.Context, step *TaskPipelineStep) error {
	// Симуляция выполнения шага validate
	time.Sleep(100 * time.Millisecond)
	return nil
}

// executeRepairStep выполняет шаг исправления
func (r *RouterPlannerService) executeRepairStep(ctx context.Context, step *TaskPipelineStep) error {
	projectPath, ok := step.Config["project_path"].(string)
	if !ok {
		// Временное решение
		if pipeline, ok := ctx.Value("pipeline").(*TaskPipeline); ok {
			if task, ok := pipeline.Steps[0].Config["task"].(domain.Task); ok {
				projectPath = task.Metadata["project_path"].(string)
			}
		}
		if projectPath == "" {
			return fmt.Errorf("project_path not found for repair step")
		}
	}

	errorOutput, ok := step.Config["error_output"].(string)
	if !ok {
		// Если ошибки нет в конфиге, значит предыдущие шаги прошли успешно
		step.Result = &TaskPipelineStepResult{
			Success: true,
			Message: "No errors to repair.",
		}
		return nil
	}

	language, ok := step.Config["language"].(string)
	if !ok {
		language = "go" // default
	}

	maxAttempts, ok := step.Config["max_attempts"].(int)
	if !ok {
		maxAttempts = 3 // default
	}

	repairRequest := domain.RepairRequest{
		ProjectPath: projectPath,
		ErrorOutput: errorOutput,
		Language:    language,
		MaxAttempts: maxAttempts,
	}

	result, err := r.repairService.ExecuteRepair(ctx, repairRequest)
	if err != nil {
		return fmt.Errorf("repair execution failed: %w", err)
	}

	step.Result = &TaskPipelineStepResult{
		Success: result.Success,
		Message: fmt.Sprintf("Repair finished. Success: %v. Attempts: %d.", result.Success, result.Attempts),
		Data: map[string]interface{}{
			"attempts": result.Attempts,
		},
	}

	if !result.Success {
		return fmt.Errorf("automated repair failed after %d attempts", result.Attempts)
	}

	return nil
}

// GetPipelineStatus возвращает статус пайплайна
func (r *RouterPlannerService) GetPipelineStatus(pipeline *TaskPipeline) map[string]interface{} {
	completed := 0
	failed := 0
	pending := 0
	running := 0

	for _, step := range pipeline.Steps {
		switch step.Status {
		case StepStatusCompleted:
			completed++
		case StepStatusFailed:
			failed++
		case StepStatusPending:
			pending++
		case StepStatusRunning:
			running++
		}
	}

	return map[string]interface{}{
		"task_id":     pipeline.TaskID,
		"status":      pipeline.Status,
		"total_steps": len(pipeline.Steps),
		"completed":   completed,
		"failed":      failed,
		"pending":     pending,
		"running":     running,
		"progress":    float64(completed) / float64(len(pipeline.Steps)),
		"duration":    pipeline.Duration,
		"error":       pipeline.Error,
	}
}
