package coordinator

import (
	"context"
	"errors"
	"federal-payment-processing/internal/models"
	"testing"
	"time"
)

func TestInvokeWithRetry_SuccessOnFirstTry(t *testing.T) {
	expected := []byte("success-response")
	invoker := &AgentInvoker{
		Invoke: func(ctx context.Context, payload []byte) ([]byte, error) {
			return expected, nil
		},
		Jitter: func() time.Duration { return 0 },
	}

	result, err := invoker.InvokeWithRetry(context.Background(), "extraction", "msg-001", []byte("input"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != string(expected) {
		t.Errorf("expected %q, got %q", string(expected), string(result))
	}
}

func TestInvokeWithRetry_SuccessOnSecondAttempt(t *testing.T) {
	attempts := 0
	expected := []byte("retry-success")
	invoker := &AgentInvoker{
		Invoke: func(ctx context.Context, payload []byte) ([]byte, error) {
			attempts++
			if attempts == 1 {
				return nil, errors.New("transient error")
			}
			return expected, nil
		},
		Jitter: func() time.Duration { return 0 },
	}

	result, err := invoker.InvokeWithRetry(context.Background(), "validation", "msg-002", []byte("input"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != string(expected) {
		t.Errorf("expected %q, got %q", string(expected), string(result))
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestInvokeWithRetry_SuccessOnThirdAttempt(t *testing.T) {
	attempts := 0
	expected := []byte("final-success")
	invoker := &AgentInvoker{
		Invoke: func(ctx context.Context, payload []byte) ([]byte, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("transient error")
			}
			return expected, nil
		},
		Jitter: func() time.Duration { return 0 },
	}

	result, err := invoker.InvokeWithRetry(context.Background(), "compliance", "msg-003", []byte("input"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != string(expected) {
		t.Errorf("expected %q, got %q", string(expected), string(result))
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestInvokeWithRetry_AllRetriesExhausted(t *testing.T) {
	attempts := 0
	invoker := &AgentInvoker{
		Invoke: func(ctx context.Context, payload []byte) ([]byte, error) {
			attempts++
			return nil, errors.New("persistent failure")
		},
		Jitter: func() time.Duration { return 0 },
	}

	result, err := invoker.InvokeWithRetry(context.Background(), "routing", "msg-004", []byte("input"))
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err == nil {
		t.Fatal("expected error when all retries exhausted")
	}

	var invErr *InvocationError
	if !errors.As(err, &invErr) {
		t.Fatalf("expected *InvocationError, got %T: %v", err, err)
	}
	if invErr.AgentName != "routing" {
		t.Errorf("expected agent name 'routing', got %q", invErr.AgentName)
	}
	if invErr.MessageID != "msg-004" {
		t.Errorf("expected message ID 'msg-004', got %q", invErr.MessageID)
	}
	if invErr.Decision != models.DecisionEscalate {
		t.Errorf("expected decision ESCALATE, got %s", invErr.Decision)
	}
	if attempts != MaxRetries {
		t.Errorf("expected %d attempts, got %d", MaxRetries, attempts)
	}
}

func TestInvokeWithRetry_NilInvokeFunc(t *testing.T) {
	invoker := &AgentInvoker{
		Invoke: nil,
	}

	_, err := invoker.InvokeWithRetry(context.Background(), "test", "msg-005", []byte("input"))
	if err == nil {
		t.Fatal("expected error for nil invoke function")
	}
}

func TestInvokeWithRetry_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	invoker := &AgentInvoker{
		Invoke: func(ctx context.Context, payload []byte) ([]byte, error) {
			attempts++
			// Cancel context after first failure to simulate cancellation during backoff
			cancel()
			return nil, errors.New("failure")
		},
		Jitter: func() time.Duration { return 0 },
	}

	_, err := invoker.InvokeWithRetry(ctx, "test-agent", "msg-006", []byte("input"))
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt before cancellation, got %d", attempts)
	}
}

func TestCalculateBackoff_Formula(t *testing.T) {
	tests := []struct {
		name       string
		retryCount int
		jitter     time.Duration
		expected   time.Duration
	}{
		{
			name:       "retry 0 no jitter",
			retryCount: 0,
			jitter:     0,
			expected:   100 * time.Millisecond, // 2^0 * 100ms = 100ms
		},
		{
			name:       "retry 1 no jitter",
			retryCount: 1,
			jitter:     0,
			expected:   200 * time.Millisecond, // 2^1 * 100ms = 200ms
		},
		{
			name:       "retry 2 no jitter",
			retryCount: 2,
			jitter:     0,
			expected:   400 * time.Millisecond, // 2^2 * 100ms = 400ms
		},
		{
			name:       "retry 3 no jitter",
			retryCount: 3,
			jitter:     0,
			expected:   800 * time.Millisecond, // 2^3 * 100ms = 800ms
		},
		{
			name:       "retry 0 with max jitter",
			retryCount: 0,
			jitter:     100 * time.Millisecond,
			expected:   200 * time.Millisecond, // 2^0 * 100ms + 100ms = 200ms
		},
		{
			name:       "retry 1 with jitter 50ms",
			retryCount: 1,
			jitter:     50 * time.Millisecond,
			expected:   250 * time.Millisecond, // 2^1 * 100ms + 50ms = 250ms
		},
		{
			name:       "retry 5 no jitter",
			retryCount: 5,
			jitter:     0,
			expected:   3200 * time.Millisecond, // 2^5 * 100ms = 3200ms
		},
		{
			name:       "retry 10 no jitter caps at 10s",
			retryCount: 10,
			jitter:     0,
			expected:   10000 * time.Millisecond, // 2^10 * 100ms = 102400ms, capped at 10000ms
		},
		{
			name:       "retry 7 with jitter caps at 10s",
			retryCount: 7,
			jitter:     100 * time.Millisecond,
			expected:   10000 * time.Millisecond, // 2^7 * 100ms + 100ms = 12900ms, capped
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateBackoff(tc.retryCount, tc.jitter)
			if got != tc.expected {
				t.Errorf("CalculateBackoff(%d, %v) = %v, want %v", tc.retryCount, tc.jitter, got, tc.expected)
			}
		})
	}
}

func TestCalculateBackoff_NeverExceedsCap(t *testing.T) {
	// Test a range of retry counts to ensure cap is always respected
	for retryCount := 0; retryCount <= 20; retryCount++ {
		backoff := CalculateBackoff(retryCount, MaxJitter)
		if backoff > MaxBackoff {
			t.Errorf("CalculateBackoff(%d, %v) = %v, exceeds cap %v", retryCount, MaxJitter, backoff, MaxBackoff)
		}
	}
}

func TestCalculateBackoff_RetryZeroNoJitter(t *testing.T) {
	// 2^0 * 100ms + 0 = 100ms
	got := CalculateBackoff(0, 0)
	expected := 100 * time.Millisecond
	if got != expected {
		t.Errorf("CalculateBackoff(0, 0) = %v, want %v", got, expected)
	}
}

func TestInvokeWithRetry_PayloadPassedCorrectly(t *testing.T) {
	inputPayload := []byte("test-payload-data")
	var receivedPayload []byte
	invoker := &AgentInvoker{
		Invoke: func(ctx context.Context, payload []byte) ([]byte, error) {
			receivedPayload = payload
			return []byte("ok"), nil
		},
		Jitter: func() time.Duration { return 0 },
	}

	_, err := invoker.InvokeWithRetry(context.Background(), "test", "msg-007", inputPayload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(receivedPayload) != string(inputPayload) {
		t.Errorf("expected payload %q, got %q", string(inputPayload), string(receivedPayload))
	}
}
