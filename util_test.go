package retry

import (
	"testing"
	"time"
)

func Test_minDuration(t *testing.T) {
	tests := map[string]struct {
		inputA time.Duration
		inputB time.Duration
		want   time.Duration
	}{
		"A > B": {
			inputA: 2 * time.Second,
			inputB: time.Second,
			want:   time.Second,
		},
		"A < B": {
			inputA: time.Second,
			inputB: 2 * time.Second,
			want:   time.Second,
		},
		"A == B": {
			inputA: time.Second,
			inputB: time.Second,
			want:   time.Second,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := minDuration(tc.inputA, tc.inputB)
			if got != tc.want {
				t.Fatalf("want %v got %v", tc.want, got)
			}
		})
	}
}

func Test_add64(t *testing.T) {
	tests := map[string]struct {
		inputA     int64
		inputB     int64
		want       int64
		wantStatus bool
	}{
		"true": {
			inputA:     9223372036854775806,
			inputB:     1,
			want:       9223372036854775807,
			wantStatus: true,
		},
		"false": {
			inputA:     9223372036854775807,
			inputB:     1,
			want:       -9223372036854775808,
			wantStatus: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotStatus := add64(tc.inputA, tc.inputB)
			if got != tc.want {
				t.Fatalf("want %v got %v", tc.want, got)
			}

			if gotStatus != tc.wantStatus {
				t.Fatalf("want status %v got %v", tc.wantStatus, gotStatus)
			}
		})
	}
}
