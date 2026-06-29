package coordinator

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"federal-payment-processing/internal/models"
)

// MaxRetries is the maximum number of invocation attempts (1 initial + 2 retries = 3 total).
const MaxRetries = 3

// MaxBackoff is the maximum backoff duration cap.
const MaxBackoff = 10000 * time.Millisecond

// BaseBackoff is the base unit for exponential backoff calculation.
const BaseBackoff = 100 * time.Millisecond

// MaxJitter is the upper bound for random jitter added to backoff.
const MaxJitter = 100 * time.Millisecond

// AgentFunc represents the function signature for invoking an agent.
// It takes a context and a payload, returning the result bytes or an error.
type AgentFunc func(ctx context.Context, payload []byte) ([]byte, error)

// JitterFunc is a function that returns a random jitter duration between 0 and MaxJitter.
// It can be replaced in tests for deterministic behavior.
type JitterFunc func() time.Duration

// AgentInvoker manages agent invocations with retry and exponential backoff.
type AgentInvoker struct {
	// Invoke is the function used to call the agent.
	Invoke AgentFunc
	// Jitter returns a random jitter duration. Defaults to random if not set.
	Jitter JitterFunc
}

// NewAgentInvoker creates a new AgentInvoker with the given agent function.
func NewAgentInvoker(fn AgentFunc) *AgentInvoker {
	return &AgentInvoker{
		Invoke: fn,
		Jitter: defaultJitter,
	}
}

// defaultJitter returns a random duration between 0 and MaxJitter.
func defaultJitter() time.Duration {
	return time.Duration(rand.Int63n(int64(MaxJitter)))
}

// CalculateBackoff computes the backoff duration for a given retry count.
// Formula: min((2^retryCount) * 100ms + jitter, 10000ms)
// retryCount starts at 0 for the first retry wait.
func CalculateBackoff(retryCount int, jitter time.Duration) time.Duration {
	exponential := time.Duration(math.Pow(2, float64(retryCount))) * BaseBackoff
	backoff := exponential + jitter
	if backoff > MaxBackoff {
		return MaxBackoff
	}
	return backoff
}

// InvocationError is returned when all retry attempts are exhausted.
// It carries the ESCALATE decision to indicate the payment should be routed to human review.
type InvocationError struct {
	AgentName string
	MessageID string
	Decision  models.Decision
	LastErr   error
}

func (e *InvocationError) Error() string {
	return fmt.Sprintf(
		"agent %q invocation failed after %d attempts (messageID=%s, decision=%s): %v",
		e.AgentName, MaxRetries, e.MessageID, e.Decision, e.LastErr,
	)
}

// InvokeWithRetry attempts to invoke an agent up to MaxRetries times with exponential backoff.
// On success, it returns the agent's response. If all retries are exhausted, it returns an
// InvocationError with decision ESCALATE, indicating the payment should be routed to human review.
func (ai *AgentInvoker) InvokeWithRetry(ctx context.Context, agentName string, messageID string, payload []byte) ([]byte, error) {
	if ai.Invoke == nil {
		return nil, fmt.Errorf("agent function must not be nil")
	}

	jitterFn := ai.Jitter
	if jitterFn == nil {
		jitterFn = defaultJitter
	}

	var lastErr error
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		log.Printf("[AgentInvoker] invoking agent=%q messageID=%s attempt=%d/%d", agentName, messageID, attempt, MaxRetries)

		result, err := ai.Invoke(ctx, payload)
		if err == nil {
			return result, nil
		}

		lastErr = err
		log.Printf("[AgentInvoker] agent=%q messageID=%s attempt=%d failed: %v", agentName, messageID, attempt, err)

		// Don't wait after the last attempt
		if attempt < MaxRetries {
			retryCount := attempt - 1 // retryCount 0 for first wait, 1 for second wait
			backoff := CalculateBackoff(retryCount, jitterFn())

			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during backoff for agent %q: %w", agentName, ctx.Err())
			case <-time.After(backoff):
				// continue to next attempt
			}
		}
	}

	// All retries exhausted: escalate to human review
	return nil, &InvocationError{
		AgentName: agentName,
		MessageID: messageID,
		Decision:  models.DecisionEscalate,
		LastErr:   lastErr,
	}
}
