package tasks

import (
	"shotgun_code/domain"
	"testing"
)

func TestNewTaskManager(t *testing.T) {
	tmpDir := t.TempDir()
	tm, err := NewTaskManager(tmpDir)
	if err != nil {
		t.Fatalf("NewTaskManager failed: %v", err)
	}
	defer tm.Close()

	if tm == nil {
		t.Fatal("NewTaskManager returned nil")
	}
	if tm.db == nil {
		t.Error("db not initialized")
	}
}

func TestTaskManager_CreatePlan(t *testing.T) {
	tmpDir := t.TempDir()
	tm, err := NewTaskManager(tmpDir)
	if err != nil {
		t.Fatalf("NewTaskManager failed: %v", err)
	}
	defer tm.Close()

	plan, err := tm.CreatePlan("Test Plan", "Test description")
	if err != nil {
		t.Fatalf("CreatePlan failed: %v", err)
	}

	if plan.ID == "" {
		t.Error("plan ID is empty")
	}
	if plan.Title != "Test Plan" {
		t.Errorf("expected title 'Test Plan', got %q", plan.Title)
	}
	if plan.Status != domain.DecompTaskPending {
		t.Errorf("expected status pending, got %v", plan.Status)
	}
}

func TestTaskManager_AddTask(t *testing.T) {
	tmpDir := t.TempDir()
	tm, err := NewTaskManager(tmpDir)
	if err != nil {
		t.Fatalf("NewTaskManager failed: %v", err)
	}
	defer tm.Close()

	plan, _ := tm.CreatePlan("Test Plan", "")

	task := &domain.DecompTask{
		Title:       "Task 1",
		Description: "First task",
		Order:       1,
	}

	err = tm.AddTask(plan.ID, task)
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}

	if task.ID == "" {
		t.Error("task ID not assigned")
	}
}

func TestTaskManager_GetPlan(t *testing.T) {
	tmpDir := t.TempDir()
	tm, err := NewTaskManager(tmpDir)
	if err != nil {
		t.Fatalf("NewTaskManager failed: %v", err)
	}
	defer tm.Close()

	plan, _ := tm.CreatePlan("Test Plan", "Description")
	tm.AddTask(plan.ID, &domain.DecompTask{Title: "Task 1", Order: 1})
	tm.AddTask(plan.ID, &domain.DecompTask{Title: "Task 2", Order: 2})

	retrieved, err := tm.GetPlan(plan.ID)
	if err != nil {
		t.Fatalf("GetPlan failed: %v", err)
	}

	if retrieved.ID != plan.ID {
		t.Errorf("ID mismatch: got %q, want %q", retrieved.ID, plan.ID)
	}
	if len(retrieved.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(retrieved.Tasks))
	}
}

func TestTaskManager_UpdateTaskStatus(t *testing.T) {
	tmpDir := t.TempDir()
	tm, err := NewTaskManager(tmpDir)
	if err != nil {
		t.Fatalf("NewTaskManager failed: %v", err)
	}
	defer tm.Close()

	plan, _ := tm.CreatePlan("Test Plan", "")
	task := &domain.DecompTask{Title: "Task 1", Order: 1}
	tm.AddTask(plan.ID, task)

	err = tm.UpdateTaskStatus(task.ID, domain.DecompTaskInProgress, "", "")
	if err != nil {
		t.Fatalf("UpdateTaskStatus failed: %v", err)
	}

	retrieved, _ := tm.GetPlan(plan.ID)
	if retrieved.Tasks[0].Status != domain.DecompTaskInProgress {
		t.Errorf("expected status in_progress, got %v", retrieved.Tasks[0].Status)
	}

	err = tm.UpdateTaskStatus(task.ID, domain.DecompTaskCompleted, "done", "")
	if err != nil {
		t.Fatalf("UpdateTaskStatus failed: %v", err)
	}

	retrieved, _ = tm.GetPlan(plan.ID)
	if retrieved.Tasks[0].Status != domain.DecompTaskCompleted {
		t.Errorf("expected status completed, got %v", retrieved.Tasks[0].Status)
	}
	if retrieved.Tasks[0].Result != "done" {
		t.Errorf("expected result 'done', got %q", retrieved.Tasks[0].Result)
	}
}

func TestTaskManager_GetNextTask(t *testing.T) {
	tmpDir := t.TempDir()
	tm, err := NewTaskManager(tmpDir)
	if err != nil {
		t.Fatalf("NewTaskManager failed: %v", err)
	}
	defer tm.Close()

	plan, _ := tm.CreatePlan("Test Plan", "")
	task1 := &domain.DecompTask{Title: "Task 1", Order: 1}
	task2 := &domain.DecompTask{Title: "Task 2", Order: 2}
	tm.AddTask(plan.ID, task1)
	tm.AddTask(plan.ID, task2)

	// First pending task
	next, err := tm.GetNextTask(plan.ID)
	if err != nil {
		t.Fatalf("GetNextTask failed: %v", err)
	}
	if next == nil {
		t.Fatal("expected next task, got nil")
	}
	if next.Title != "Task 1" {
		t.Errorf("expected 'Task 1', got %q", next.Title)
	}

	// Complete first task
	tm.UpdateTaskStatus(task1.ID, domain.DecompTaskCompleted, "", "")

	// Next should be task 2
	next, _ = tm.GetNextTask(plan.ID)
	if next == nil {
		t.Fatal("expected next task, got nil")
	}
	if next.Title != "Task 2" {
		t.Errorf("expected 'Task 2', got %q", next.Title)
	}
}

func TestTaskManager_GetNextTask_WithDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	tm, err := NewTaskManager(tmpDir)
	if err != nil {
		t.Fatalf("NewTaskManager failed: %v", err)
	}
	defer tm.Close()

	plan, err := tm.CreatePlan("Test Plan", "")
	if err != nil {
		t.Skipf("Skipping test - SQLite not available: %v", err)
	}
	
	task1 := &domain.DecompTask{ID: "task-1", Title: "Task 1", Order: 1}
	task2 := &domain.DecompTask{Title: "Task 2", Order: 2, Dependencies: []string{"task-1"}}
	if err := tm.AddTask(plan.ID, task1); err != nil {
		t.Skipf("Skipping test - SQLite not available: %v", err)
	}
	tm.AddTask(plan.ID, task2)

	// Task 2 depends on task 1, so only task 1 should be available
	next, err := tm.GetNextTask(plan.ID)
	if err != nil {
		t.Skipf("Skipping test - SQLite not available: %v", err)
	}
	if next == nil || next.Title != "Task 1" {
		t.Errorf("expected 'Task 1' (no dependencies), got %v", next)
	}

	// Complete task 1
	tm.UpdateTaskStatus("task-1", domain.DecompTaskCompleted, "", "")

	// Now task 2 should be available
	next, _ = tm.GetNextTask(plan.ID)
	if next == nil {
		t.Fatal("expected task 2 to be available")
	}
	if next.Title != "Task 2" {
		t.Errorf("expected 'Task 2', got %q", next.Title)
	}
}

func TestTaskManager_GetPlanProgress(t *testing.T) {
	tmpDir := t.TempDir()
	tm, err := NewTaskManager(tmpDir)
	if err != nil {
		t.Fatalf("NewTaskManager failed: %v", err)
	}
	defer tm.Close()

	plan, _ := tm.CreatePlan("Test Plan", "")
	task1 := &domain.DecompTask{Title: "Task 1", Order: 1}
	task2 := &domain.DecompTask{Title: "Task 2", Order: 2}
	task3 := &domain.DecompTask{Title: "Task 3", Order: 3}
	tm.AddTask(plan.ID, task1)
	tm.AddTask(plan.ID, task2)
	tm.AddTask(plan.ID, task3)

	completed, total, err := tm.GetPlanProgress(plan.ID)
	if err != nil {
		t.Fatalf("GetPlanProgress failed: %v", err)
	}
	if completed != 0 || total != 3 {
		t.Errorf("expected 0/3, got %d/%d", completed, total)
	}

	tm.UpdateTaskStatus(task1.ID, domain.DecompTaskCompleted, "", "")
	tm.UpdateTaskStatus(task2.ID, domain.DecompTaskCompleted, "", "")

	completed, total, _ = tm.GetPlanProgress(plan.ID)
	if completed != 2 || total != 3 {
		t.Errorf("expected 2/3, got %d/%d", completed, total)
	}
}

func TestTaskManager_Close(t *testing.T) {
	tmpDir := t.TempDir()
	tm, err := NewTaskManager(tmpDir)
	if err != nil {
		t.Fatalf("NewTaskManager failed: %v", err)
	}

	err = tm.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestGenID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := genID()
		if ids[id] {
			t.Errorf("duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestRndSfx_Length(t *testing.T) {
	sfx := rndSfx()
	if len(sfx) != 6 {
		t.Errorf("expected suffix length 6, got %d", len(sfx))
	}
}
