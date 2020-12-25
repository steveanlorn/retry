package retry

import (
	"reflect"
	"testing"
	"time"
)

func TestWithMaxRetryAttempts(t *testing.T) {
	var input uint = 2
	var want uint = 2

	r := NewRetrier(WithMaxRetryAttempts(input))

	if r.maximumRetryAttempts != want {
		t.Fatalf("want maximumRetryAttempts %d got %d", want, r.maximumRetryAttempts)
	}
}

func TestWithBackoff(t *testing.T) {
	input := NewBackoffEqualJitter(time.Second, time.Millisecond, NewRand(time.Now().UnixNano()))
	want := input

	r := NewRetrier(WithBackoff(input))

	if !reflect.DeepEqual(r.backoff, want) {
		t.Fatalf("want backoff %v got %v", want, r.backoff)
	}
}
