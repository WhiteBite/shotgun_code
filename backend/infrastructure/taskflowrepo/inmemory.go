package taskflowrepo

import (
	"shotgun_code/domain"
	"sync"
)

// InMemoryTaskflowRepository implements TaskflowRepository using in-memory storage
type InMemoryTaskflowRepository struct {
	statuses map[string]domain.TaskState
	mu       sync.RWMutex
}

// NewInMemoryTaskflowRepository creates a new in-memory taskflow repository
func NewInMemoryTaskflowRepository() *InMemoryTaskflowRepository {
	return &InMemoryTaskflowRepository{
		statuses: make(map[string]domain.TaskState),
	}
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

// SaveStatuses saves task statuses to memory
func (r *InMemoryTaskflowRepository) SaveStatuses(statuses map[string]domain.TaskState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Update the map
	for k, v := range statuses {
		r.statuses[k] = v
	}

	return nil
}
