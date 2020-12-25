package retry

import "time"

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// sum two value. Second value indicates
// integer overflow status.
func add64(a, b int64) (int64, bool) {
	c := a + b
	if (c > a) == (b > 0) {
		return c, true
	}
	return c, false
}
