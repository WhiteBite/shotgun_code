package testutils

import (
	"github.com/stretchr/testify/mock"
)

// MockLogger is a shared mock implementation of domain.Logger for tests
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Warning(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Fatal(msg string) {
	m.Called(msg)
}
