package domain

import "time"

// DecompTaskStatus represents decomposed task status
type DecompTaskStatus string

const (
	DecompTaskPending    DecompTaskStatus = "pending"
	DecompTaskInProgress DecompTaskStatus = "in_progress"
	DecompTaskCompleted  DecompTaskStatus = "completed"
	DecompTaskFailed     DecompTaskStatus = "failed"
	DecompTaskSkipped    DecompTaskStatus = "skipped"
)

// DecompTask represents a decomposed subtask
type DecompTask struct {
	ID           string            `json:"id"`
	ParentID     string            `json:"parentId,omitempty"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       DecompTaskStatus  `json:"status"`
	Order        int               `json:"order"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Files        []string          `json:"files,omitempty"`
	Result       string            `json:"result,omitempty"`
	Error        string            `json:"error,omitempty"`
	CreatedAt    time.Time         `json:"createdAt"`
	StartedAt    *time.Time        `json:"startedAt,omitempty"`
	CompletedAt  *time.Time        `json:"completedAt,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// DecompTaskPlan represents a complete task decomposition
type DecompTaskPlan struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Tasks       []*DecompTask      `json:"tasks"`
	CreatedAt   time.Time          `json:"createdAt"`
	Status      DecompTaskStatus   `json:"status"`
}

// DecompTaskManager interface for task decomposition
type DecompTaskManager interface {
	CreatePlan(title, description string) (*DecompTaskPlan, error)
	AddTask(planID string, task *DecompTask) error
	GetPlan(planID string) (*DecompTaskPlan, error)
	UpdateTaskStatus(taskID string, status DecompTaskStatus, result, errMsg string) error
	GetNextTask(planID string) (*DecompTask, error)
	GetPlanProgress(planID string) (completed, total int, err error)
	Close() error
}
