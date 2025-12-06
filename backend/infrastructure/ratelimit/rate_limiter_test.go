package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 1.0)
	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
	if rl.maxTokens != 10 {
		t.Errorf("expected maxTokens 10, got %f", rl.maxTokens)
	}
	if rl.refillRate != 1.0 {
		t.Errorf("expected refillRate 1.0, got %f", rl.refillRate)
	}
}

func TestNewOpenAIRateLimiter(t *testing.T) {
	rl := NewOpenAIRateLimiter()
	if rl == nil {
		t.Fatal("NewOpenAIRateLimiter returned nil")
	}
	if rl.maxTokens != 60 {
		t.Errorf("expected maxTokens 60, got %f", rl.maxTokens)
	}
}

func TestRateLimiter_TryAcquire(t *testing.T) {
	rl := NewRateLimiter(3, 1.0)

	// Should acquire 3 tokens
	for i := 0; i < 3; i++ {
		if !rl.TryAcquire() {
			t.Errorf("TryAcquire %d should succeed", i+1)
		}
	}

	// 4th should fail
	if rl.TryAcquire() {
		t.Error("TryAcquire should fail when no tokens available")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(1, 10.0) // 1 token, refill 10/sec

	// Acquire the only token
	if !rl.TryAcquire() {
		t.Fatal("first TryAcquire should succeed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := rl.Wait(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Wait should succeed, got error: %v", err)
	}
	if elapsed < 50*time.Millisecond {
		t.Errorf("Wait should have waited for refill, elapsed: %v", elapsed)
	}
}

func TestRateLimiter_Wait_ContextCancelled(t *testing.T) {
	rl := NewRateLimiter(1, 0.1) // Very slow refill

	// Acquire the only token
	rl.TryAcquire()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := rl.Wait(ctx)
	if err == nil {
		t.Error("Wait should return error when context cancelled")
	}
}

func TestRateLimiter_Available(t *testing.T) {
	rl := NewRateLimiter(5, 1.0)

	available := rl.Available()
	if int(available) != 5 {
		t.Errorf("expected 5 available, got %d", int(available))
	}

	rl.TryAcquire()
	rl.TryAcquire()

	available = rl.Available()
	if int(available) != 3 {
		t.Errorf("expected 3 available, got %d", int(available))
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	rl := NewRateLimiter(5, 100.0) // Fast refill for testing

	// Drain all tokens
	for i := 0; i < 5; i++ {
		rl.TryAcquire()
	}

	// Wait for refill
	time.Sleep(50 * time.Millisecond)

	available := rl.Available()
	if available < 1 {
		t.Errorf("expected at least 1 token after refill, got %f", available)
	}
}

func TestConcurrencyLimiter_New(t *testing.T) {
	cl := NewConcurrencyLimiter(5)
	if cl == nil {
		t.Fatal("NewConcurrencyLimiter returned nil")
	}
}

func TestConcurrencyLimiter_AcquireRelease(t *testing.T) {
	cl := NewConcurrencyLimiter(2)
	ctx := context.Background()

	// Acquire 2 slots
	if err := cl.Acquire(ctx); err != nil {
		t.Errorf("first Acquire failed: %v", err)
	}
	if err := cl.Acquire(ctx); err != nil {
		t.Errorf("second Acquire failed: %v", err)
	}

	// 3rd should block, use TryAcquire
	if cl.TryAcquire() {
		t.Error("TryAcquire should fail when at capacity")
	}

	// Release one
	cl.Release()

	// Now should succeed
	if !cl.TryAcquire() {
		t.Error("TryAcquire should succeed after Release")
	}
}

func TestConcurrencyLimiter_Acquire_ContextCancelled(t *testing.T) {
	cl := NewConcurrencyLimiter(1)
	ctx := context.Background()

	// Acquire the only slot
	cl.Acquire(ctx)

	// Try to acquire with cancelled context
	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := cl.Acquire(ctx2)
	if err == nil {
		t.Error("Acquire should return error when context cancelled")
	}
}

func TestCompositeRateLimiter_New(t *testing.T) {
	rl := NewRateLimiter(10, 1.0)
	cl := NewConcurrencyLimiter(5)
	crl := NewCompositeRateLimiter(rl, cl)

	if crl == nil {
		t.Fatal("NewCompositeRateLimiter returned nil")
	}
}

func TestCompositeRateLimiter_AcquireRelease(t *testing.T) {
	rl := NewRateLimiter(10, 1.0)
	cl := NewConcurrencyLimiter(2)
	crl := NewCompositeRateLimiter(rl, cl)
	ctx := context.Background()

	// Should acquire both rate limit and concurrency
	if err := crl.Acquire(ctx); err != nil {
		t.Errorf("Acquire failed: %v", err)
	}

	crl.Release()
}

func TestRateLimiter_Concurrent(t *testing.T) {
	rl := NewRateLimiter(100, 10.0)
	var wg sync.WaitGroup
	acquired := 0
	var mu sync.Mutex

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.TryAcquire() {
				mu.Lock()
				acquired++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if acquired != 50 {
		t.Errorf("expected 50 acquired, got %d", acquired)
	}
}
