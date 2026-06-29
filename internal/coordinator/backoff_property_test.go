package coordinator

import (
	"testing"
	"time"

	"pgregory.net/rapid"
)

// **Validates: Requirement 15.2**

// TestProperty_BackoffNeverExceedsCap verifies that for any retryCount (0-20) and
// any jitter (0-100ms), backoff never exceeds 10,000ms.
func TestProperty_BackoffNeverExceedsCap(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		retryCount := rapid.IntRange(0, 20).Draw(t, "retryCount")
		jitterMs := rapid.IntRange(0, 100).Draw(t, "jitterMs")
		jitter := time.Duration(jitterMs) * time.Millisecond

		backoff := CalculateBackoff(retryCount, jitter)

		if backoff > MaxBackoff {
			t.Fatalf("backoff %v exceeds cap %v for retryCount=%d, jitter=%v",
				backoff, MaxBackoff, retryCount, jitter)
		}
	})
}

// TestProperty_BackoffRetryZeroNoJitterIsExactly100ms verifies that for retryCount=0
// with zero jitter, backoff is exactly 100ms (2^0 * 100ms).
func TestProperty_BackoffRetryZeroNoJitterIsExactly100ms(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		backoff := CalculateBackoff(0, 0)

		expected := 100 * time.Millisecond
		if backoff != expected {
			t.Fatalf("expected backoff of %v for retryCount=0 with zero jitter, got %v",
				expected, backoff)
		}
	})
}

// TestProperty_BackoffAtLeastExponentialWhenBelowCap verifies that backoff is always
// >= (2^retryCount * 100ms) when jitter is 0 and result is below cap.
func TestProperty_BackoffAtLeastExponentialWhenBelowCap(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Limit retryCount to values where 2^retryCount * 100ms < 10000ms (retryCount <= 6)
		retryCount := rapid.IntRange(0, 6).Draw(t, "retryCount")

		backoff := CalculateBackoff(retryCount, 0)

		minExpected := time.Duration(1<<uint(retryCount)) * BaseBackoff
		if backoff < minExpected {
			t.Fatalf("backoff %v is less than minimum expected %v for retryCount=%d with zero jitter",
				backoff, minExpected, retryCount)
		}
	})
}

// TestProperty_BackoffNeverNegative verifies that backoff is always >= 0 (never negative)
// for any retryCount and jitter value.
func TestProperty_BackoffNeverNegative(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		retryCount := rapid.IntRange(0, 20).Draw(t, "retryCount")
		jitterMs := rapid.IntRange(0, 100).Draw(t, "jitterMs")
		jitter := time.Duration(jitterMs) * time.Millisecond

		backoff := CalculateBackoff(retryCount, jitter)

		if backoff < 0 {
			t.Fatalf("backoff %v is negative for retryCount=%d, jitter=%v",
				backoff, retryCount, jitter)
		}
	})
}

// TestProperty_BackoffMonotonicallyIncreases verifies that backoff increases monotonically
// with retryCount (for same jitter value) until hitting the cap.
func TestProperty_BackoffMonotonicallyIncreases(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		jitterMs := rapid.IntRange(0, 100).Draw(t, "jitterMs")
		jitter := time.Duration(jitterMs) * time.Millisecond

		// Check pairs of consecutive retry counts
		retryCount := rapid.IntRange(0, 19).Draw(t, "retryCount")

		backoffLow := CalculateBackoff(retryCount, jitter)
		backoffHigh := CalculateBackoff(retryCount+1, jitter)

		// Backoff should be monotonically non-decreasing (it can be equal when both are at the cap)
		if backoffHigh < backoffLow {
			t.Fatalf("backoff decreased from retryCount=%d (%v) to retryCount=%d (%v) with jitter=%v",
				retryCount, backoffLow, retryCount+1, backoffHigh, jitter)
		}
	})
}
