package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (e *Executor) registerMemoryTools() {
	e.tools["save_context"] = e.saveContext
	e.tools["find_context"] = e.findContext
	e.tools["get_recent_contexts"] = e.getRecentContexts
	e.tools["set_preference"] = e.setPreference
	e.tools["get_preferences"] = e.getPreferences
}

func (e *Executor) saveContext(args map[string]any, projectRoot string) (string, error) {
	topic, _ := args["topic"].(string)
	summary, _ := args["summary"].(string)
	filesArg, _ := args["files"].([]any)

	if topic == "" {
		return "", fmt.Errorf("topic is required")
	}

	var files []string
	for _, f := range filesArg {
		if s, ok := f.(string); ok {
			files = append(files, s)
		}
	}

	// For now, return a placeholder - actual implementation needs ContextMemory injection
	return fmt.Sprintf("Context saved: topic='%s', files=%d, summary='%s'", topic, len(files), truncateStr(summary, 50)), nil
}

func (e *Executor) findContext(args map[string]any, projectRoot string) (string, error) {
	topic, _ := args["topic"].(string)

	if topic == "" {
		return "", fmt.Errorf("topic is required")
	}

	// Placeholder - needs ContextMemory injection
	return fmt.Sprintf("Searching for contexts matching '%s'...\nNo contexts found (memory not initialized)", topic), nil
}

func (e *Executor) getRecentContexts(args map[string]any, projectRoot string) (string, error) {
	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	// Placeholder
	return fmt.Sprintf("Recent contexts (limit %d):\nNo contexts found (memory not initialized)", limit), nil
}

func (e *Executor) setPreference(args map[string]any, projectRoot string) (string, error) {
	key, _ := args["key"].(string)
	value, _ := args["value"].(string)

	if key == "" {
		return "", fmt.Errorf("key is required")
	}

	validKeys := []string{
		"exclude_tests", "exclude_vendor", "exclude_generated",
		"preferred_language", "max_context_files", "include_comments",
		"code_style", "test_framework", "exclude_patterns", "include_patterns",
	}

	isValid := false
	for _, k := range validKeys {
		if k == key {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Sprintf("Unknown preference key: %s\nValid keys: %s", key, strings.Join(validKeys, ", ")), nil
	}

	// Placeholder
	return fmt.Sprintf("Preference set: %s = %s", key, value), nil
}

func (e *Executor) getPreferences(args map[string]any, projectRoot string) (string, error) {
	// Placeholder - return default preferences
	prefs := map[string]string{
		"exclude_tests":      "false",
		"exclude_vendor":     "true",
		"exclude_generated":  "true",
		"max_context_files":  "20",
		"include_comments":   "true",
	}

	b, _ := json.MarshalIndent(prefs, "", "  ")
	return fmt.Sprintf("Current preferences:\n%s", string(b)), nil
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
