package registration

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiterBasic(t *testing.T) {
	logger := zerolog.Nop()
	rl := NewRateLimiter(3, logger)
	defer rl.Stop()

	ipAddr := "192.168.1.100"

	// First 3 attempts should succeed
	for i := 0; i < 3; i++ {
		err := rl.CheckAndRecord(ipAddr)
		assert.NoError(t, err, "attempt %d should succeed", i+1)
	}

	// 4th attempt should fail
	err := rl.CheckAndRecord(ipAddr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
}

func TestRateLimiterMultipleIPs(t *testing.T) {
	logger := zerolog.Nop()
	rl := NewRateLimiter(3, logger)
	defer rl.Stop()

	ip1 := "192.168.1.100"
	ip2 := "192.168.1.101"
	ip3 := "192.168.1.102"

	// Each IP should have independent limits
	for i := 0; i < 3; i++ {
		err := rl.CheckAndRecord(ip1)
		assert.NoError(t, err)
		err = rl.CheckAndRecord(ip2)
		assert.NoError(t, err)
		err = rl.CheckAndRecord(ip3)
		assert.NoError(t, err)
	}

	// All IPs should now be at limit
	err := rl.CheckAndRecord(ip1)
	assert.Error(t, err)
	err = rl.CheckAndRecord(ip2)
	assert.Error(t, err)
	err = rl.CheckAndRecord(ip3)
	assert.Error(t, err)
}

func TestRateLimiterGetAttemptCount(t *testing.T) {
	logger := zerolog.Nop()
	rl := NewRateLimiter(5, logger)
	defer rl.Stop()

	ipAddr := "192.168.1.100"

	// Initially no attempts
	count := rl.GetAttemptCount(ipAddr)
	assert.Equal(t, 0, count)

	// Record 3 attempts
	for i := 0; i < 3; i++ {
		err := rl.CheckAndRecord(ipAddr)
		require.NoError(t, err)
	}

	// Should show 3 attempts
	count = rl.GetAttemptCount(ipAddr)
	assert.Equal(t, 3, count)
}

func TestRateLimiterCleanup(t *testing.T) {
	logger := zerolog.Nop()
	rl := NewRateLimiter(3, logger)
	
	// Stop automatic cleanup to test manual cleanup
	rl.Stop()

	// Override window duration for testing
	rl.windowDuration = 100 * time.Millisecond

	ipAddr := "192.168.1.100"

	// Record 2 attempts
	err := rl.CheckAndRecord(ipAddr)
	require.NoError(t, err)
	err = rl.CheckAndRecord(ipAddr)
	require.NoError(t, err)

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Manually trigger cleanup to remove from map
	rl.cleanup()

	// Verify IP was removed from map
	rl.mu.RLock()
	_, exists := rl.attempts[ipAddr]
	rl.mu.RUnlock()
	assert.False(t, exists, "IP should be removed from map after cleanup")

	// Should be able to register again
	err = rl.CheckAndRecord(ipAddr)
	assert.NoError(t, err)
}

func TestRateLimiterConcurrency(t *testing.T) {
	logger := zerolog.Nop()
	rl := NewRateLimiter(10, logger)
	defer rl.Stop()

	ipAddr := "192.168.1.100"
	attempts := 20
	successCount := 0
	errorCount := 0

	// Concurrent attempts
	done := make(chan bool, attempts)
	for i := 0; i < attempts; i++ {
		go func() {
			err := rl.CheckAndRecord(ipAddr)
			if err == nil {
				successCount++
			} else {
				errorCount++
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < attempts; i++ {
		<-done
	}

	// Should have exactly 10 successes (the limit)
	// Note: Due to race conditions, we might have slightly different results
	// but the total should be consistent
	assert.LessOrEqual(t, successCount, 10, "should not exceed limit")
	assert.GreaterOrEqual(t, successCount, 10, "should allow up to limit")
}

func TestRateLimiterZeroLimit(t *testing.T) {
	logger := zerolog.Nop()
	rl := NewRateLimiter(0, logger)
	defer rl.Stop()

	ipAddr := "192.168.1.100"

	// Should immediately fail with 0 limit
	err := rl.CheckAndRecord(ipAddr)
	assert.Error(t, err)
}

func TestRateLimiterHighLimit(t *testing.T) {
	logger := zerolog.Nop()
	rl := NewRateLimiter(1000, logger)
	defer rl.Stop()

	ipAddr := "192.168.1.100"

	// Should allow many attempts
	for i := 0; i < 100; i++ {
		err := rl.CheckAndRecord(ipAddr)
		assert.NoError(t, err, "attempt %d should succeed", i+1)
	}

	count := rl.GetAttemptCount(ipAddr)
	assert.Equal(t, 100, count)
}

func TestRateLimiterCleanupLoop(t *testing.T) {
	logger := zerolog.Nop()
	rl := NewRateLimiter(3, logger)

	// Override intervals for faster testing
	rl.cleanupInterval = 50 * time.Millisecond
	rl.windowDuration = 100 * time.Millisecond

	ipAddr := "192.168.1.100"

	// Record attempts
	err := rl.CheckAndRecord(ipAddr)
	require.NoError(t, err)

	// Wait for cleanup loop to run
	time.Sleep(200 * time.Millisecond)

	// Stop should clean up goroutine
	rl.Stop()

	// Verify cleanup happened (attempts should be gone)
	count := rl.GetAttemptCount(ipAddr)
	assert.Equal(t, 0, count)
}
