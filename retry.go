// Package retry offers feature to retry a function.
// It offers different backoff strategies.
//
// Retry means when the first function execution fails, then it will do retry.
// It will try to retry until maximum allowed attempt.
// There will be a delay (backoff) from one attempt to another.
//
// When a function return unretryable error, it will not try to the next attempt.
package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// DefaultRetry default configuration for retrier.
const (
	DefaultRetryMaxAttempts uint = 10
)

var defaultRetryBackoff = NewBackoffConstant(DefaultBackoffBaseInterval)

// Retrier denotes a retrier struct.
type Retrier struct {
	maximumRetryAttempts uint
	backoff              Backoff
}

// UnretryableError denotes an error that should be exit immediately.
type UnretryableError struct {
	error
}

func (u *UnretryableError) Unwrap() error {
	return u.error
}

// Unretryable wraps an error in UnretryableError struct
// When you do this inside your function, it will not retry into the next attempt.
func Unretryable(err error) error {
	return &UnretryableError{
		fmt.Errorf("unretryable error: %w", err),
	}
}

// ContextDoneError denotes an error that triggered when context is done.
type ContextDoneError struct {
	error
}

func (u *ContextDoneError) Unwrap() error {
	return u.error
}

func contextDone(err error) error {
	return &ContextDoneError{
		fmt.Errorf("context done error: %w", err),
	}
}

// NewRetrier initializes Retrier struct.
// This is handy if you want to re-use retrier with the same configuration.
//
// If option is not provided, it will use default configuration.
// Example:
//	retrier := retry.NewRetrier(
// 		retry.WithMaxRetryAttempts(1),
// 		retry.WithBackoff(
// 			retry.NewBackoffTruncatedExponential(
// 				time.Second,
// 				time.Millisecond,
// 				retry.NewRand(time.Now().UnixNano()),
// 			),
// 		),
// 	)
func NewRetrier(options ...Option) *Retrier {
	rr := &Retrier{
		maximumRetryAttempts: DefaultRetryMaxAttempts,
		backoff:              defaultRetryBackoff,
	}

	for _, option := range options {
		option(rr)
	}

	return rr
}

// Do runs f with retry.
func (r *Retrier) Do(f func() error) error {
	return r.do(context.Background(), f)
}

// DoWithContext runs f with retry.
// While waiting for next attempt, it will also check context cancelation.
func (r *Retrier) DoWithContext(ctx context.Context, f func() error) error {
	return r.do(ctx, f)
}

// Do runs f with retry.
// This is handy if you do not want to initialize Retrier.
func Do(f func() error, options ...Option) error {
	r := NewRetrier(options...)
	return r.do(context.Background(), f)
}

// DoWithContext runs f with retry.
// While waiting for next attempt, it will also check context cancelation.
// This is handy if you do not want to initialize Retrier.
func DoWithContext(ctx context.Context, f func() error, options ...Option) error {
	r := NewRetrier(options...)
	return r.do(ctx, f)
}

func (r *Retrier) do(ctx context.Context, f func() error) error {
	var attempt uint
	for {
		err := f()
		if err == nil {
			return nil
		}

		if !r.isRetryable(attempt, err) {
			return err
		}

		backoffDuration := r.backoff.Get(attempt)
		timer := time.NewTimer(backoffDuration)

		select {
		case <-ctx.Done():
			timer.Stop()
			return contextDone(err)
		case <-timer.C:
			timer.Stop()
		}

		attempt++
	}
}

func (r *Retrier) isRetryable(attempt uint, err error) bool {
	if attempt >= r.maximumRetryAttempts {
		return false
	}

	var unretryableError *UnretryableError
	if errors.As(err, &unretryableError) {
		return false
	}

	return true
}
