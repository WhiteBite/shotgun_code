package ratelimit

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu           sync.Mutex
	tokens       float64
	maxTokens    float64
	refillRate   float64 // tokens per second
	lastRefill   time.Time
	requestQueue chan struct{}
}

// NewRateLimiter creates a new rate limiter
// maxTokens: maximum burst capacity
// refillRate: tokens added per second
func NewRateLimiter(maxTokens float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:       maxTokens,
		maxTokens:    maxTokens,
		refillRate:   refillRate,
		lastRefill:   time.Now(),
		requestQueue: make(chan struct{}, int(maxTokens)),
	}
}

// NewOpenAIRateLimiter creates a rate limiter configured for OpenAI API
// Default: 60 requests per minute for embeddings
func NewOpenAIRateLimiter() *RateLimiter {
	return NewRateLimiter(60, 1.0) // 60 tokens, 1 per second refill
}

// Wait blocks until a token is available or context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.tryAcquire() {
			return nil
		}

		// Calculate wait time
		waitTime := rl.timeUntilToken()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again
		}
	}
}

// TryAcquire attempts to acquire a token without blocking
func (rl *RateLimiter) TryAcquire() bool {
	return rl.tryAcquire()
}

func (rl *RateLimiter) tryAcquire() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now
}

func (rl *RateLimiter) timeUntilToken() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= 1 {
		return 0
	}

	tokensNeeded := 1 - rl.tokens
	secondsNeeded := tokensNeeded / rl.refillRate
	return time.Duration(secondsNeeded*1000) * time.Millisecond
}

// Available returns the current number of available tokens
func (rl *RateLimiter) Available() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.refill()
	return rl.tokens
}

// ConcurrencyLimiter limits concurrent operations
type ConcurrencyLimiter struct {
	sem chan struct{}
}

// NewConcurrencyLimiter creates a new concurrency limiter
func NewConcurrencyLimiter(maxConcurrent int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		sem: make(chan struct{}, maxConcurrent),
	}
}

// Acquire acquires a slot, blocking if necessary
func (cl *ConcurrencyLimiter) Acquire(ctx context.Context) error {
	select {
	case cl.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release releases a slot
func (cl *ConcurrencyLimiter) Release() {
	<-cl.sem
}

// TryAcquire attempts to acquire a slot without blocking
func (cl *ConcurrencyLimiter) TryAcquire() bool {
	select {
	case cl.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

// CompositeRateLimiter combines rate limiting and concurrency limiting
type CompositeRateLimiter struct {
	rateLimiter        *RateLimiter
	concurrencyLimiter *ConcurrencyLimiter
}

// NewCompositeRateLimiter creates a composite limiter
func NewCompositeRateLimiter(rateLimit *RateLimiter, concurrencyLimit *ConcurrencyLimiter) *CompositeRateLimiter {
	return &CompositeRateLimiter{
		rateLimiter:        rateLimit,
		concurrencyLimiter: concurrencyLimit,
	}
}

// Acquire acquires both rate limit token and concurrency slot
func (crl *CompositeRateLimiter) Acquire(ctx context.Context) error {
	// First wait for rate limit
	if err := crl.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	// Then acquire concurrency slot
	return crl.concurrencyLimiter.Acquire(ctx)
}

// Release releases the concurrency slot
func (crl *CompositeRateLimiter) Release() {
	crl.concurrencyLimiter.Release()
}
