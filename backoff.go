package retry

import (
	"time"
)

// Backoff denotes backoff strategy interface
//
//go:generate mockgen -destination=mocks/mock_backoff.go -package=mocks . Backoff
type Backoff interface {
	Get(attempt uint) time.Duration
}

// DefaultBackoff denotes default configuration value for Backoff
const (
	DefaultBackoffMaximumInterval = time.Millisecond * 1000
	DefaultBackoffBaseInterval    = time.Millisecond * 100
)

// BackoffConstant backoff strategy to
// backoff by a constant number for each attempt.
type BackoffConstant struct {
	baseInterval time.Duration
}

var _ Backoff = BackoffConstant{}

// NewBackoffConstant initialize ConstantBackoff.
func NewBackoffConstant(baseInterval time.Duration) BackoffConstant {
	if baseInterval == 0 {
		baseInterval = DefaultBackoffBaseInterval
	}

	return BackoffConstant{
		baseInterval: baseInterval,
	}
}

// Get returns backoff time duration based on the given attempt.
func (c BackoffConstant) Get(_ uint) time.Duration {
	return c.baseInterval
}

// BackoffCappedExponential backoff strategy to multiplied
// backoff by a constant after each attempt,
// up to some maximum value.
type BackoffCappedExponential struct {
	maxInterval  time.Duration
	baseInterval time.Duration
}

var _ Backoff = BackoffCappedExponential{}

// NewBackoffCappedExponential initialize BackoffCappedExponential.
func NewBackoffCappedExponential(maxInterval, baseInterval time.Duration) BackoffCappedExponential {
	if maxInterval == 0 {
		maxInterval = DefaultBackoffMaximumInterval
	}

	if baseInterval == 0 {
		baseInterval = DefaultBackoffBaseInterval
	}

	return BackoffCappedExponential{
		maxInterval:  maxInterval,
		baseInterval: baseInterval,
	}
}

// Get returns backoff time duration based on the given attempt.
func (c BackoffCappedExponential) Get(attempt uint) time.Duration {
	var exponentialFactor int64 = 1 << attempt

	// wrap around to zero value if integer overflow.
	// return maximum interval.
	if exponentialFactor == 0 {
		return c.maxInterval
	}

	exponential := c.baseInterval * time.Duration(exponentialFactor)

	// wrap around to zero value if integer overflow.
	// return maximum interval.
	if exponential == 0 {
		return c.maxInterval
	}

	sleepDuration := minDuration(c.maxInterval, exponential)

	return sleepDuration
}

// BackoffFullJitter backoff strategy to multiplied backoff
// by randomize delay up until current multiplied backoff.
type BackoffFullJitter struct {
	maxInterval  time.Duration
	baseInterval time.Duration
	rand         Randomizer
}

var _ Backoff = BackoffFullJitter{}

// NewBackoffFullJitter initialize BackoffFullJitter.
func NewBackoffFullJitter(maxInterval, baseInterval time.Duration, randomizer Randomizer) BackoffFullJitter {
	if maxInterval == 0 {
		maxInterval = DefaultBackoffMaximumInterval
	}

	if baseInterval == 0 {
		baseInterval = DefaultBackoffBaseInterval
	}

	return BackoffFullJitter{
		maxInterval:  maxInterval,
		baseInterval: baseInterval,
		rand:         randomizer,
	}
}

// Get returns backoff time duration based on the given attempt.
func (c BackoffFullJitter) Get(attempt uint) time.Duration {
	var exponentialFactor int64 = 1 << attempt

	// wrap around to zero value if integer overflow.
	// return maximum interval.
	if exponentialFactor == 0 {
		return c.maxInterval
	}

	exponential := c.baseInterval * time.Duration(exponentialFactor)

	// wrap around to zero value if integer overflow.
	// return maximum interval.
	if exponential == 0 {
		return c.maxInterval
	}

	sleepDuration := minDuration(c.maxInterval, exponential)

	sleepDurationWithJitter := c.rand.Int63n(sleepDuration.Nanoseconds())

	return time.Duration(sleepDurationWithJitter)
}

// BackoffEqualJitter a timed backoff loops
// which keeps some of the backoff and jitter by a smaller amount.
type BackoffEqualJitter struct {
	maxInterval  time.Duration
	baseInterval time.Duration
	rand         Randomizer
}

var _ Backoff = BackoffEqualJitter{}

// NewBackoffEqualJitter initialize BackoffEqualJitter.
func NewBackoffEqualJitter(maxInterval, baseInterval time.Duration, randomizer Randomizer) BackoffEqualJitter {
	if maxInterval == 0 {
		maxInterval = DefaultBackoffMaximumInterval
	}

	if baseInterval == 0 {
		baseInterval = DefaultBackoffBaseInterval
	}

	return BackoffEqualJitter{
		maxInterval:  maxInterval,
		baseInterval: baseInterval,
		rand:         randomizer,
	}
}

// Get returns backoff time duration based on the given attempt.
func (c BackoffEqualJitter) Get(attempt uint) time.Duration {
	var exponentialFactor int64 = 1 << attempt

	// wrap around to zero value if integer overflow.
	// return maximum interval.
	if exponentialFactor == 0 {
		return c.maxInterval
	}

	exponential := c.baseInterval * time.Duration(exponentialFactor)

	// wrap around to zero value if integer overflow.
	// return maximum interval.
	if exponential == 0 {
		return c.maxInterval
	}

	sleepDuration := minDuration(c.maxInterval, exponential)

	sleepDurationWithJitter := sleepDuration.Nanoseconds()/2 + c.rand.Int63n(sleepDuration.Nanoseconds()/2)

	return time.Duration(sleepDurationWithJitter)
}

// BackoffDecorrelated a timed backoff loops
// which is similar to "Full Jitter"
// with increment in the maximum jitter based on the last backoff value.
type BackoffDecorrelated struct {
	maxInterval   time.Duration
	baseInterval  time.Duration
	sleepDuration time.Duration
	rand          Randomizer
}

var _ Backoff = (*BackoffDecorrelated)(nil)

// NewBackoffDecorrelated initialize BackoffDecorrelated.
func NewBackoffDecorrelated(maxInterval, baseInterval time.Duration, randomizer Randomizer) *BackoffDecorrelated {
	if maxInterval == 0 {
		maxInterval = DefaultBackoffMaximumInterval
	}

	if baseInterval == 0 {
		baseInterval = DefaultBackoffBaseInterval
	}

	return &BackoffDecorrelated{
		maxInterval:   maxInterval,
		baseInterval:  baseInterval,
		sleepDuration: baseInterval,
		rand:          randomizer,
	}
}

// Get returns backoff time duration based on the given attempt.
func (c *BackoffDecorrelated) Get(_ uint) time.Duration {
	min := c.baseInterval.Nanoseconds()
	max := c.sleepDuration.Nanoseconds() * 3

	// wrap around to zero value if integer overflow.
	// return maximum interval.
	if max == 0 {
		return c.maxInterval
	}

	sleepDuration, ok := add64(c.rand.Int63n(max-min), min)
	if !ok {
		return c.maxInterval
	}

	c.sleepDuration = minDuration(c.maxInterval, time.Duration(sleepDuration))

	return c.sleepDuration
}

// BackoffTruncatedExponential backoff strategy to multiplied
// backoff by periodically increasing delays
// with additional jitters.
type BackoffTruncatedExponential struct {
	maxInterval  time.Duration
	baseInterval time.Duration
	rand         Randomizer
}

var _ Backoff = BackoffTruncatedExponential{}

// NewBackoffTruncatedExponential initialize BackoffTruncatedExponential.
func NewBackoffTruncatedExponential(maxInterval, baseInterval time.Duration, randomizer Randomizer) BackoffTruncatedExponential {
	if maxInterval == 0 {
		maxInterval = DefaultBackoffMaximumInterval
	}

	if baseInterval == 0 {
		baseInterval = DefaultBackoffBaseInterval
	}

	return BackoffTruncatedExponential{
		maxInterval:  maxInterval,
		baseInterval: baseInterval,
		rand:         randomizer,
	}
}

// Get returns backoff time duration based on the given attempt.
func (c BackoffTruncatedExponential) Get(attempt uint) time.Duration {
	var exponentialFactor int64 = 1 << attempt

	// wrap around to zero value if integer overflow.
	// return maximum interval.
	if exponentialFactor == 0 {
		return c.maxInterval
	}

	randomDurtaion := time.Duration(c.rand.Int63n(int64(c.baseInterval.Nanoseconds())))

	exponential := c.baseInterval*time.Duration(exponentialFactor) + randomDurtaion

	// wrap around to zero value if integer overflow.
	// return maximum interval.
	if exponential == 0 || exponential == randomDurtaion {
		return c.maxInterval
	}

	sleepDuration := minDuration(c.maxInterval, exponential)

	return sleepDuration
}
