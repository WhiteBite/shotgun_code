package tools

import (
"fmt"
"strings"
)

func (e *Executor) registerCallGraphTools() {
e.tools["get_callers"] = e.getCallers
e.tools["get_callees"] = e.getCallees
e.tools["get_impact"] = e.getImpact
}

func (e *Executor) getCallers(args map[string]any, projectRoot string) (string, error) {
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
callers := e.callGraph.GetCallers(functionID)
if len(callers) == 0 {
return fmt.Sprintf("No callers for '%s'", function), nil
}
var lines []string
lines = append(lines, fmt.Sprintf("Callers of '%s':", function))
for _, c := range callers {
lines = append(lines, fmt.Sprintf("  %s in %s:%d", c.Name, c.FilePath, c.Line))
}
return strings.Join(lines, "\n"), nil
}

func (e *Executor) getCallees(args map[string]any, projectRoot string) (string, error) {
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
callees := e.callGraph.GetCallees(functionID)
if len(callees) == 0 {
return fmt.Sprintf("'%s' has no callees", function), nil
}
var lines []string
lines = append(lines, fmt.Sprintf("Callees of '%s':", function))
for _, c := range callees {
lines = append(lines, fmt.Sprintf("  %s in %s:%d", c.Name, c.FilePath, c.Line))
}
return strings.Join(lines, "\n"), nil
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
var lines []string
lines = append(lines, fmt.Sprintf("Impact of '%s':", function))
for _, a := range affected {
lines = append(lines, fmt.Sprintf("  %s in %s:%d", a.Name, a.FilePath, a.Line))
}
return strings.Join(lines, "\n"), nil
}
