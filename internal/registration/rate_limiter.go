package registration

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// RateLimiter tracks registration attempts by IP address
type RateLimiter struct {
	mu                sync.RWMutex
	attempts          map[string][]int64 // IP -> timestamps
	maxAttempts       int
	windowDuration    time.Duration
	cleanupInterval   time.Duration
	logger            zerolog.Logger
	stopCleanup       chan struct{}
	cleanupWaitGroup  sync.WaitGroup
}

// NewRateLimiter creates a new rate limiter for registration attempts
func NewRateLimiter(maxAttemptsPerHour int, logger zerolog.Logger) *RateLimiter {
	rl := &RateLimiter{
		attempts:        make(map[string][]int64),
		maxAttempts:     maxAttemptsPerHour,
		windowDuration:  time.Hour,
		cleanupInterval: 10 * time.Minute,
		logger:          logger,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	rl.cleanupWaitGroup.Add(1)
	go rl.cleanupLoop()

	return rl
}

// CheckAndRecord checks if an IP can register and records the attempt
func (rl *RateLimiter) CheckAndRecord(ipAddr string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().Unix()
	windowStart := now - int64(rl.windowDuration.Seconds())

	// Get existing attempts for this IP
	timestamps := rl.attempts[ipAddr]

	// Filter out attempts outside the time window
	validAttempts := make([]int64, 0)
	for _, ts := range timestamps {
		if ts > windowStart {
			validAttempts = append(validAttempts, ts)
		}
	}

	// Check if limit exceeded
	if len(validAttempts) >= rl.maxAttempts {
		rl.logger.Warn().
			Str("remote_addr", ipAddr).
			Int("attempts", len(validAttempts)).
			Int("max_attempts", rl.maxAttempts).
			Msg("registration_rate_limit")

		return fmt.Errorf("registration rate limit exceeded: maximum %d registrations per hour", rl.maxAttempts)
	}

	// Record this attempt
	validAttempts = append(validAttempts, now)
	rl.attempts[ipAddr] = validAttempts

	rl.logger.Debug().
		Str("remote_addr", ipAddr).
		Int("attempts", len(validAttempts)).
		Int("max_attempts", rl.maxAttempts).
		Msg("registration attempt recorded")

	return nil
}

// cleanupLoop periodically removes old entries to prevent memory leaks
func (rl *RateLimiter) cleanupLoop() {
	defer rl.cleanupWaitGroup.Done()

	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

// cleanup removes entries older than the time window
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().Unix()
	windowStart := now - int64(rl.windowDuration.Seconds())

	removedCount := 0
	for ip, timestamps := range rl.attempts {
		// Filter out old attempts
		validAttempts := make([]int64, 0)
		for _, ts := range timestamps {
			if ts > windowStart {
				validAttempts = append(validAttempts, ts)
			}
		}

		// Remove IP if no valid attempts remain
		if len(validAttempts) == 0 {
			delete(rl.attempts, ip)
			removedCount++
		} else {
			rl.attempts[ip] = validAttempts
		}
	}

	if removedCount > 0 {
		rl.logger.Debug().
			Int("removed_ips", removedCount).
			Int("remaining_ips", len(rl.attempts)).
			Msg("rate limiter cleanup completed")
	}
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	select {
	case <-rl.stopCleanup:
		// Already stopped
		return
	default:
		close(rl.stopCleanup)
		rl.cleanupWaitGroup.Wait()
	}
}

// GetAttemptCount returns the number of attempts for an IP in the current window (for testing)
func (rl *RateLimiter) GetAttemptCount(ipAddr string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now().Unix()
	windowStart := now - int64(rl.windowDuration.Seconds())

	timestamps := rl.attempts[ipAddr]
	count := 0
	for _, ts := range timestamps {
		if ts > windowStart {
			count++
		}
	}

	return count
}
