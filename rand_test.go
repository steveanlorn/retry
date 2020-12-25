package retry

import (
	"testing"
	"time"
)

func TestInt63n(t *testing.T) {
	r := NewRand(time.Now().UnixNano())
	got := r.Int63n(100-50) + 50
	if (got < 50) || (got > 100) {
		t.Fatalf("want between 50 and 100, got %d", got)
	}
}
