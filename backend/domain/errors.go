package domain

import "errors"

// Sentinel errors used across the application domain.
var (
	// ErrInvalidAPIKey is returned when an AI provider rejects the API key.
	ErrInvalidAPIKey = errors.New("invalid API key")

	// ErrRateLimitExceeded is returned when an AI provider indicates a rate limit has been hit.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
