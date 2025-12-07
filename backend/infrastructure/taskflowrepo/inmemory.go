package taskflowrepo

import (
	"shotgun_code/domain"
	"sync"
	"time"
)

// InMemoryTaskflowRepository implements TaskflowRepository using in-memory storage
type InMemoryTaskflowRepository struct {
	statuses  map[string]domain.TaskState
	createdAt map[string]time.Time
	mu        sync.RWMutex
}

const (
	maxTaskStatuses     = 1000
	maxCompletedTaskAge = 7 * 24 * time.Hour
)

// NewInMemoryTaskflowRepository creates a new in-memory taskflow repository
func NewInMemoryTaskflowRepository() *InMemoryTaskflowRepository {
	repo := &InMemoryTaskflowRepository{
		statuses:  make(map[string]domain.TaskState),
		createdAt: make(map[string]time.Time),
	}

	// Start periodic cleanup
	go repo.periodicCleanup()

	return repo
}

// LoadStatuses loads task statuses from memory
func (r *InMemoryTaskflowRepository) LoadStatuses() (map[string]domain.TaskState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy of the map
	statuses := make(map[string]domain.TaskState)
	for k, v := range r.statuses {
		statuses[k] = v
	}

	return statuses, nil
}

// SaveStatuses saves task statuses to memory with limits
func (r *InMemoryTaskflowRepository) SaveStatuses(statuses map[string]domain.TaskState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Update the map
	for k, v := range statuses {
		r.statuses[k] = v
		if _, exists := r.createdAt[k]; !exists {
			r.createdAt[k] = time.Now()
		}
	}

	// Check limits
	if len(r.statuses) > maxTaskStatuses {
		r.cleanupOldCompletedTasksUnlocked(maxCompletedTaskAge)
	}

	return nil
}

// CleanupCompletedTasks удаляет завершенные задачи старше указанного возраста
func (r *InMemoryTaskflowRepository) CleanupCompletedTasks(maxAge time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupOldCompletedTasksUnlocked(maxAge)
}

func (r *InMemoryTaskflowRepository) cleanupOldCompletedTasksUnlocked(maxAge time.Duration) {
	now := time.Now()
	for id, createdTime := range r.createdAt {
		if now.Sub(createdTime) > maxAge {
			if state, exists := r.statuses[id]; exists {
				// Удаляем только завершенные или failed задачи
				if state == "completed" || state == "failed" {
					delete(r.statuses, id)
					delete(r.createdAt, id)
				}
			}
		}
	}
}

// periodicCleanup запускает периодическую очистку старых задач
func (r *InMemoryTaskflowRepository) periodicCleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		r.CleanupCompletedTasks(maxCompletedTaskAge)
	}
}

// GetMemoryUsage возвращает приблизительный размер используемой памяти
func (r *InMemoryTaskflowRepository) GetMemoryUsage() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var size int64
	for id := range r.statuses {
		size += int64(len(id) + 100) // Приблизительный размер
	}
	return size
}

// GetStats возвращает статистику репозитория
func (r *InMemoryTaskflowRepository) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	activeCount := 0
	completedCount := 0

	for _, state := range r.statuses {
		if state == "completed" || state == "failed" {
			completedCount++
		} else {
			activeCount++
		}
	}

	return map[string]interface{}{
		"total_tasks":     len(r.statuses),
		"active_tasks":    activeCount,
		"completed_tasks": completedCount,
		"memory_usage_mb": r.GetMemoryUsage() / (1024 * 1024),
	}
}
