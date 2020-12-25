package retry

// Option denotes optional configuration.
type Option func(*Retrier)

// WithMaxRetryAttempts configures maximum attempt to retry.
func WithMaxRetryAttempts(maximumAttempts uint) Option {
	return func(r *Retrier) {
		r.maximumRetryAttempts = maximumAttempts
	}
}

// WithBackoff configures backoff strategy.
func WithBackoff(backoff Backoff) Option {
	return func(r *Retrier) {
		r.backoff = backoff
	}
}
