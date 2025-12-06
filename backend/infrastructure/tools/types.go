// Package tools provides tool implementations for agentic AI
package tools

import "shotgun_code/domain"

// Tool represents a tool that AI can use
type Tool = domain.Tool

// ToolCall represents a call to a tool
type ToolCall = domain.ToolCall

// ToolResult represents the result of a tool execution
type ToolResult = domain.ToolResult

// ToolParameters represents tool parameter schema
type ToolParameters = domain.ToolParameters

// ToolProperty represents a tool parameter property
type ToolProperty = domain.ToolProperty
