package orchestration

import (
	"context"
	"time"
)

// RetryOptions configures withRetry.
type RetryOptions struct {
	Retries      int
	BaseDelayMs  int
	OnRetry      func(attempt, maxAttempts, delayMs int, err error)
}

func withRetry(ctx context.Context, fn func() error, opts RetryOptions) error {
	maxAttempts := opts.Retries + 1
	var last error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fn(); err != nil {
			last = err
			if attempt == opts.Retries {
				break
			}
			delay := time.Duration(opts.BaseDelayMs*(1<<attempt)) * time.Millisecond
			if opts.OnRetry != nil {
				opts.OnRetry(attempt+1, maxAttempts, int(delay/time.Millisecond), err)
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			continue
		}
		return nil
	}
	if last != nil {
		return last
	}
	return context.Canceled
}
