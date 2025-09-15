package application

import (
	"context"
	"fmt"
	"shotgun_code/domain"
	"strings"
	"time"
)

// ContextAnalyzerImpl implements the ContextAnalyzer interface
type ContextAnalyzerImpl struct {
	logger    domain.Logger
	aiService *AIService
}

// NewContextAnalyzer creates a new ContextAnalyzer implementation
func NewContextAnalyzer(logger domain.Logger, aiService *AIService) *ContextAnalyzerImpl {
	return &ContextAnalyzerImpl{
		logger:    logger,
		aiService: aiService,
	}
}

// AnalyzeTaskAndCollectContext analyzes a task and automatically collects relevant context
func (ca *ContextAnalyzerImpl) AnalyzeTaskAndCollectContext(
	ctx context.Context,
	task string,
	allFiles []*domain.FileNode,
	rootDir string,
) (*domain.ContextAnalysisResult, error) {
	startTime := time.Now()
	ca.logger.Info(fmt.Sprintf("Starting task analysis: %s", task))

	// 1. Analyze task and determine type
	taskAnalysis, err := ca.analyzeTaskType(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("task type analysis error: %w", err)
	}

	// 2. Determine priority files based on analysis
	priorityFiles, err := ca.determinePriorityFiles(ctx, task, taskAnalysis, allFiles)
	if err != nil {
		return nil, fmt.Errorf("priority files determination error: %w", err)
	}

	// 3. Collect context from priority files
	contextContent, err := ca.collectContextFromFiles(ctx, priorityFiles, rootDir, taskAnalysis)
	if err != nil {
		return nil, fmt.Errorf("context collection error: %w", err)
	}

	// 4. Analyze dependencies and add related files
	dependencyFiles, err := ca.analyzeDependencies(ctx, priorityFiles, allFiles, rootDir)
	if err != nil {
		ca.logger.Warning(fmt.Sprintf("Dependency analysis error: %v", err))
	}

	// 5. Form final result
	result := &domain.ContextAnalysisResult{
		Task:            task,
		TaskType:        taskAnalysis.Type,
		Priority:        taskAnalysis.Priority,
		SelectedFiles:   priorityFiles,
		DependencyFiles: dependencyFiles,
		Context:         contextContent,
		AnalysisTime:    time.Since(startTime),
		Recommendations: ca.generateRecommendations(taskAnalysis, priorityFiles),
		EstimatedTokens: ca.estimateTokens(contextContent),
		Confidence:      ca.calculateConfidence(taskAnalysis, priorityFiles),
	}

	ca.logger.Info(fmt.Sprintf("Analysis completed in %v, selected %d files", result.AnalysisTime, len(priorityFiles)))
	return result, nil
}

// analyzeTaskType analyzes task type and determines context collection strategy
func (ca *ContextAnalyzerImpl) analyzeTaskType(ctx context.Context, task string) (*domain.TaskAnalysis, error) {
	if ca.aiService == nil {
		// Fallback to simple analysis if AI service is not available
		return ca.simpleTaskAnalysis(task), nil
	}

	systemPrompt := `You are an expert in development task analysis. Analyze the task and determine:
1. Task type (bug_fix, feature, refactor, test, documentation, optimization)
2. Priority (low, normal, high, critical)
3. Key technologies/frameworks
4. File types to search for
5. Keywords for search

Respond in JSON format:
{
  "type": "task_type",
  "priority": "priority",
  "technologies": ["tech1", "tech2"],
  "fileTypes": [".go", ".ts", ".vue"],
  "keywords": ["keyword1", "keyword2"],
  "reasoning": "explanation of choice"
}`

	response, err := ca.aiService.GenerateCode(ctx, systemPrompt, task)
	if err != nil {
		return nil, fmt.Errorf("task analysis error: %w", err)
	}

	// Parse response (simplified version)
	analysis := &domain.TaskAnalysis{
		Type:         ca.extractTaskType(task, response),
		Priority:     ca.extractPriority(task, response),
		Technologies: ca.extractTechnologies(task, response),
		FileTypes:    ca.extractFileTypes(task, response),
		Keywords:     ca.extractKeywords(task, response),
		Reasoning:    response,
	}

	return analysis, nil
}

// simpleTaskAnalysis provides fallback analysis when AI is not available
func (ca *ContextAnalyzerImpl) simpleTaskAnalysis(task string) *domain.TaskAnalysis {
	taskLower := strings.ToLower(task)
	
	// Simple keyword-based classification
	taskType := "feature"
	if strings.Contains(taskLower, "bug") || strings.Contains(taskLower, "fix") || strings.Contains(taskLower, "error") {
		taskType = "bug_fix"
	} else if strings.Contains(taskLower, "test") {
		taskType = "test"
	} else if strings.Contains(taskLower, "refactor") || strings.Contains(taskLower, "cleanup") {
		taskType = "refactor"
	} else if strings.Contains(taskLower, "doc") {
		taskType = "documentation"
	}
	
	priority := "normal"
	if strings.Contains(taskLower, "critical") || strings.Contains(taskLower, "urgent") {
		priority = "critical"
	} else if strings.Contains(taskLower, "high") {
		priority = "high"
	}
	
	return &domain.TaskAnalysis{
		Type:         taskType,
		Priority:     priority,
		Technologies: []string{"go", "typescript", "vue"},
		FileTypes:    []string{".go", ".ts", ".vue", ".js"},
		Keywords:     strings.Fields(taskLower),
		Reasoning:    "Simple keyword-based analysis",
	}
}

// Helper methods for extracting information from AI response
func (ca *ContextAnalyzerImpl) extractTaskType(task, response string) string {
	// Simple extraction - in real implementation would parse JSON
	if strings.Contains(strings.ToLower(response), "bug_fix") {
		return "bug_fix"
	}
	return "feature"
}

func (ca *ContextAnalyzerImpl) extractPriority(task, response string) string {
	if strings.Contains(strings.ToLower(response), "critical") {
		return "critical"
	}
	return "normal"
}

func (ca *ContextAnalyzerImpl) extractTechnologies(task, response string) []string {
	return []string{"go", "typescript", "vue"}
}

func (ca *ContextAnalyzerImpl) extractFileTypes(task, response string) []string {
	return []string{".go", ".ts", ".vue", ".js"}
}

func (ca *ContextAnalyzerImpl) extractKeywords(task, response string) []string {
	return strings.Fields(strings.ToLower(task))
}

// determinePriorityFiles determines which files are most relevant to the task
func (ca *ContextAnalyzerImpl) determinePriorityFiles(ctx context.Context, task string, analysis *domain.TaskAnalysis, allFiles []*domain.FileNode) ([]*domain.FileNode, error) {
	// Simple implementation - take first few files of matching types
	var priorityFiles []*domain.FileNode
	for _, file := range allFiles {
		if len(priorityFiles) >= 5 {
			break
		}
		for _, fileType := range analysis.FileTypes {
			if strings.HasSuffix(file.RelPath, fileType) {
				priorityFiles = append(priorityFiles, file)
				break
			}
		}
	}
	return priorityFiles, nil
}

// collectContextFromFiles collects context content from the specified files
func (ca *ContextAnalyzerImpl) collectContextFromFiles(ctx context.Context, files []*domain.FileNode, rootDir string, analysis *domain.TaskAnalysis) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files for context collection")
	}

	// For now, we'll need a file reader to collect content
	// In a full implementation, this would be injected as a dependency
	return "", fmt.Errorf("file reading not implemented in this simplified version")
}

// analyzeDependencies analyzes dependencies of the selected files
func (ca *ContextAnalyzerImpl) analyzeDependencies(ctx context.Context, priorityFiles []*domain.FileNode, allFiles []*domain.FileNode, rootDir string) ([]*domain.FileNode, error) {
	// Simple dependency analysis - return empty for now
	return []*domain.FileNode{}, nil
}

// generateRecommendations generates recommendations based on the analysis
func (ca *ContextAnalyzerImpl) generateRecommendations(analysis *domain.TaskAnalysis, files []*domain.FileNode) []string {
	return []string{
		fmt.Sprintf("Task type: %s", analysis.Type),
		fmt.Sprintf("Selected %d files", len(files)),
	}
}

// estimateTokens estimates the number of tokens in the context
func (ca *ContextAnalyzerImpl) estimateTokens(context string) int {
	return len(context) / 4
}

// calculateConfidence calculates the confidence level of the analysis
func (ca *ContextAnalyzerImpl) calculateConfidence(analysis *domain.TaskAnalysis, files []*domain.FileNode) float64 {
	if len(files) > 0 {
		return 0.8
	}
	return 0.5
}