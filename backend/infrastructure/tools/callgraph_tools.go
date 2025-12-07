package tools

import (
	"fmt"
	"shotgun_code/domain/analysis"
	"strings"
)

func (e *Executor) registerCallGraphTools() {
	e.tools["get_callers"] = e.getCallers
	e.tools["get_callees"] = e.getCallees
	e.tools["get_impact"] = e.getImpact
}

// getCallGraphRelations is a helper for getCallers and getCallees
func (e *Executor) getCallGraphRelations(args map[string]any, projectRoot string, getRelations func(string) []analysis.CallNode, emptyMsg, headerMsg string) (string, error) {
	function, _ := args["function"].(string)
	if function == "" {
		return "", fmt.Errorf("function name is required")
	}
	graph, err := e.callGraph.Build(projectRoot)
	if err != nil {
		return "", err
	}
	var functionID string
	for id, node := range graph.Nodes {
		if node.Name == function {
			functionID = id
			break
		}
	}
	if functionID == "" {
		return fmt.Sprintf("Function '%s' not found", function), nil
	}
	relations := getRelations(functionID)
	if len(relations) == 0 {
		return fmt.Sprintf(emptyMsg, function), nil
	}
	lines := make([]string, 0, len(relations)+1)
	lines = append(lines, fmt.Sprintf(headerMsg, function))
	for _, c := range relations {
		lines = append(lines, fmt.Sprintf("  %s in %s:%d", c.Name, c.FilePath, c.Line))
	}
	return strings.Join(lines, "\n"), nil
}

func (e *Executor) getCallers(args map[string]any, projectRoot string) (string, error) {
	return e.getCallGraphRelations(args, projectRoot, e.callGraph.GetCallers, "No callers for '%s'", "Callers of '%s':")
}

func (e *Executor) getCallees(args map[string]any, projectRoot string) (string, error) {
	return e.getCallGraphRelations(args, projectRoot, e.callGraph.GetCallees, "'%s' has no callees", "Callees of '%s':")
}

func (e *Executor) getImpact(args map[string]any, projectRoot string) (string, error) {
	function, _ := args["function"].(string)
	maxDepth := 5
	if d, ok := args["max_depth"].(float64); ok && d > 0 {
		maxDepth = int(d)
	}
	if function == "" {
		return "", fmt.Errorf("function name is required")
	}
	graph, err := e.callGraph.Build(projectRoot)
	if err != nil {
		return "", err
	}
	var functionID string
	for id, node := range graph.Nodes {
		if node.Name == function {
			functionID = id
			break
		}
	}
	if functionID == "" {
		return fmt.Sprintf("Function '%s' not found", function), nil
	}
	affected := e.callGraph.GetImpact(functionID, maxDepth)
	if len(affected) == 0 {
		return fmt.Sprintf("No impact from '%s'", function), nil
	}
	lines := make([]string, 0, len(affected)+1)
	lines = append(lines, fmt.Sprintf("Impact of '%s':", function))
	for _, a := range affected {
		lines = append(lines, fmt.Sprintf("  %s in %s:%d", a.Name, a.FilePath, a.Line))
	}
	return strings.Join(lines, "\n"), nil
}
