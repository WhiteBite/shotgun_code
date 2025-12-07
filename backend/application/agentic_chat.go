package application

import (
	"context"
	"encoding/json"
	"fmt"
	"shotgun_code/domain"
	"strings"
)

// AgenticChatService handles AI chat with tool use capabilities
type AgenticChatService struct {
	logger        domain.Logger
	aiService     *AIService
	toolExecutor  *ToolExecutorImpl
	maxIterations int
}

// NewAgenticChatService creates a new agentic chat service
func NewAgenticChatService(
	logger domain.Logger,
	aiService *AIService,
	toolExecutor *ToolExecutorImpl,
) *AgenticChatService {
	return &AgenticChatService{
		logger:        logger,
		aiService:     aiService,
		toolExecutor:  toolExecutor,
		maxIterations: 10,
	}
}

// AgenticChatRequest represents a request for agentic chat
type AgenticChatRequest struct {
	Task        string `json:"task"`
	ProjectRoot string `json:"projectRoot"`
	MaxTokens   int    `json:"maxTokens,omitempty"`
}

// AgenticChatResponse represents the response from agentic chat
type AgenticChatResponse struct {
	Response   string        `json:"response"`
	ToolCalls  []ToolCallLog `json:"toolCalls"`
	Iterations int           `json:"iterations"`
	Context    []string      `json:"context"` // Files that were read
}

// ToolCallLog logs a tool call for transparency
type ToolCallLog struct {
	Tool      string `json:"tool"`
	Arguments string `json:"arguments"`
	Result    string `json:"result"`
}

// Chat performs an agentic chat with tool use
func (s *AgenticChatService) Chat(ctx context.Context, req AgenticChatRequest) (*AgenticChatResponse, error) {
	s.logger.Info(fmt.Sprintf("Starting agentic chat: %s", req.Task))

	tools := s.toolExecutor.GetAvailableTools()
	toolsJSON := s.formatToolsForPrompt(tools)

	// Build system prompt with tools
	systemPrompt := fmt.Sprintf(`You are an expert code assistant. You have access to tools to explore and analyze the codebase.

AVAILABLE TOOLS:
%s

INSTRUCTIONS:
1. When you need information about the codebase, use tools by responding with JSON in this format:
   {"tool_calls": [{"name": "tool_name", "arguments": {"arg1": "value1"}}]}

2. You can call multiple tools at once.

3. After receiving tool results, analyze them and either:
   - Call more tools if you need more information
   - Provide your final answer

4. When you have enough information, provide your answer WITHOUT any tool_calls.

5. Be thorough but efficient - don't read files unnecessarily.

IMPORTANT: Always respond in the user's language (Russian if they write in Russian).`, toolsJSON)

	var messages []domain.ChatMessage
	messages = append(messages, domain.ChatMessage{
		Role:    domain.RoleSystem,
		Content: systemPrompt,
	})
	messages = append(messages, domain.ChatMessage{
		Role:    domain.RoleUser,
		Content: req.Task,
	})

	var toolCallLogs []ToolCallLog
	var readFiles []string
	iterations := 0

	for iterations < s.maxIterations {
		iterations++
		s.logger.Info(fmt.Sprintf("Agentic chat iteration %d", iterations))

		// Get AI response
		response, err := s.callAI(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("AI call failed: %w", err)
		}

		// Check if response contains tool calls
		toolCalls := s.parseToolCalls(response)

		if len(toolCalls) == 0 {
			// No tool calls - this is the final answer
			s.logger.Info("Agentic chat completed with final answer")
			return &AgenticChatResponse{
				Response:   response,
				ToolCalls:  toolCallLogs,
				Iterations: iterations,
				Context:    readFiles,
			}, nil
		}

		// Execute tool calls
		messages = append(messages, domain.ChatMessage{
			Role:    domain.RoleAssistant,
			Content: response,
		})

		var toolResults []string
		for _, call := range toolCalls {
			result := s.toolExecutor.ExecuteTool(call, req.ProjectRoot)

			// Log the call
			argsJSON, _ := json.Marshal(call.Arguments)
			toolCallLogs = append(toolCallLogs, ToolCallLog{
				Tool:      call.Name,
				Arguments: string(argsJSON),
				Result:    truncateStr(result.Content, 500),
			})

			// Track read files
			if call.Name == "read_file" {
				if path, ok := call.Arguments["path"].(string); ok {
					readFiles = append(readFiles, path)
				}
			}

			toolResults = append(toolResults, fmt.Sprintf("Tool: %s\nResult:\n%s", call.Name, result.Content))
		}

		// Add tool results to messages
		messages = append(messages, domain.ChatMessage{
			Role:    domain.RoleTool,
			Content: strings.Join(toolResults, "\n\n---\n\n"),
		})
	}

	return nil, fmt.Errorf("max iterations (%d) reached without final answer", s.maxIterations)
}

// formatToolsForPrompt formats tools for the system prompt
func (s *AgenticChatService) formatToolsForPrompt(tools []domain.Tool) string {
	lines := make([]string, 0, len(tools))
	for _, tool := range tools {
		params := []string{}
		for name, prop := range tool.Parameters.Properties {
			required := ""
			for _, r := range tool.Parameters.Required {
				if r == name {
					required = " (required)"
					break
				}
			}
			params = append(params, fmt.Sprintf("  - %s: %s%s", name, prop.Description, required))
		}
		lines = append(lines, fmt.Sprintf(`
%s: %s
Parameters:
%s`, tool.Name, tool.Description, strings.Join(params, "\n")))
	}
	return strings.Join(lines, "\n")
}

// callAI calls the AI service with proper system/user prompt separation
func (s *AgenticChatService) callAI(ctx context.Context, messages []domain.ChatMessage) (string, error) {
	// Extract system prompt and build conversation
	var systemPrompt string
	var userPrompt strings.Builder

	for _, msg := range messages {
		switch msg.Role {
		case domain.RoleSystem:
			systemPrompt = msg.Content
		case domain.RoleUser:
			userPrompt.WriteString("User: ")
			userPrompt.WriteString(msg.Content)
			userPrompt.WriteString("\n\n")
		case domain.RoleAssistant:
			userPrompt.WriteString("Assistant: ")
			userPrompt.WriteString(msg.Content)
			userPrompt.WriteString("\n\n")
		case domain.RoleTool:
			userPrompt.WriteString("Tool Results:\n")
			userPrompt.WriteString(msg.Content)
			userPrompt.WriteString("\n\n")
		}
	}

	userPrompt.WriteString("Assistant: ")

	// Use the AI service with proper system prompt
	response, err := s.aiService.GenerateCode(ctx, systemPrompt, userPrompt.String())
	if err != nil {
		return "", err
	}

	return response, nil
}

// parseToolCalls extracts tool calls from AI response
func (s *AgenticChatService) parseToolCalls(response string) []domain.ToolCall {
	// Try to find JSON with tool_calls
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 || end <= start {
		return nil
	}

	jsonStr := response[start : end+1]

	// Try to parse as tool calls
	var parsed struct {
		ToolCalls []struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		} `json:"tool_calls"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil
	}

	calls := make([]domain.ToolCall, 0, len(parsed.ToolCalls))
	for i, tc := range parsed.ToolCalls {
		calls = append(calls, domain.ToolCall{
			ID:        fmt.Sprintf("call_%d", i),
			Name:      tc.Name,
			Arguments: tc.Arguments,
		})
	}

	return calls
}

// truncateStr truncates a string to max length
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
