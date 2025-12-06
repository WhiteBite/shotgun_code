package initmanager

import (
	"context"
	"sync"
)

// InitStatus represents the status of initialization
type InitStatus struct {
	Completed bool
	Error     error
}

// AsyncInitManager manages asynchronous initialization of services
type AsyncInitManager struct {
	mu       sync.RWMutex
	statuses map[string]*InitStatus
	handlers map[string]func(context.Context) error
	waiters  map[string][]chan struct{}
}

// New creates a new AsyncInitManager
func New() *AsyncInitManager {
	return &AsyncInitManager{
		statuses: make(map[string]*InitStatus),
		handlers: make(map[string]func(context.Context) error),
		waiters:  make(map[string][]chan struct{}),
	}
}

// Register registers an initialization handler for a service
func (m *AsyncInitManager) Register(serviceName string, handler func(context.Context) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.handlers[serviceName] = handler
	m.statuses[serviceName] = &InitStatus{Completed: false, Error: nil}
}

// IsInitialized checks if a service is already initialized
func (m *AsyncInitManager) IsInitialized(serviceName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	status, exists := m.statuses[serviceName]
	if !exists {
		return false
	}
	return status.Completed
}

// GetStatus returns the initialization status for a service
func (m *AsyncInitManager) GetStatus(serviceName string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	status, exists := m.statuses[serviceName]
	if !exists {
		return false, nil
	}
	return status.Completed, status.Error
}

// InitAsync initiates asynchronous initialization for a service
func (m *AsyncInitManager) InitAsync(ctx context.Context, serviceName string) {
	m.mu.Lock()
	
	handler, exists := m.handlers[serviceName]
	if !exists {
		status := &InitStatus{
			Completed: true,
			Error:     nil,
		}
		m.statuses[serviceName] = status
		m.mu.Unlock()
		return
	}
	
	// Check if already initializing
	if status, exists := m.statuses[serviceName]; exists && status.Completed {
		m.mu.Unlock()
		return
	}
	
	// Mark as initializing
	m.statuses[serviceName] = &InitStatus{
		Completed: false,
		Error:     nil,
	}
	
	m.mu.Unlock()
	
	// Run initialization in a goroutine
	go func() {
	err := handler(ctx)
		
		m.mu.Lock()
		defer m.mu.Unlock()
		
		// Update status
		m.statuses[serviceName] = &InitStatus{
			Completed: true,
			Error:     err,
		}
		
		// Notify all waiters
		if waiters, exists := m.waiters[serviceName]; exists {
			for _, waiter := range waiters {
				close(waiter)
			}
			delete(m.waiters, serviceName)
		}
	}()
}

// WaitFor waits for a service to be initialized
func (m *AsyncInitManager) WaitFor(serviceName string) error {
	m.mu.RLock()
	
	status, exists := m.statuses[serviceName]
	if !exists {
		m.mu.RUnlock()
		return nil // No handler registered, consider it initialized
	}
	
	if status.Completed {
		m.mu.RUnlock()
		return status.Error
	}
	
	// Service is not completed yet, need to wait
	waiter := make(chan struct{})
	m.waiters[serviceName] = append(m.waiters[serviceName], waiter)
	m.mu.RUnlock()
	
	// Wait for completion
	<-waiter
	
	// Return the final status
	m.mu.RLock()
	finalStatus := m.statuses[serviceName]
	m.mu.RUnlock()
	
	return finalStatus.Error
}

// MustWaitFor waits for initialization and panics on error
func (m *AsyncInitManager) MustWaitFor(serviceName string) {
	if err := m.WaitFor(serviceName); err != nil {
		panic(err)
	}
}