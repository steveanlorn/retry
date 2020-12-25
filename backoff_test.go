package retry

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/steveanlorn/retry/mocks"
)

func TestBackoffConstant(t *testing.T) {
	backoff := NewBackoffConstant(0)
	if got := backoff.Get(0); got != DefaultBackoffBaseInterval {
		t.Fatalf("want %v got %v", DefaultBackoffBaseInterval, got)
	}

	if got := backoff.Get(0); got != DefaultBackoffBaseInterval {
		t.Fatalf("want %v got %v", DefaultBackoffBaseInterval, got)
	}
}

func TestBackoffCappedExponential(t *testing.T) {
	type args struct {
		maxInterval  time.Duration
		baseInterval time.Duration
	}

	tests := map[string]struct {
		input uint
		args  args
		// base * 2 ** attempt
		want time.Duration
	}{
		"default value": {
			args: args{
				maxInterval:  0,
				baseInterval: 0,
			},
			input: 0,
			want:  1 * DefaultBackoffBaseInterval,
		},
		"attempt 0": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 0,
			want:  1 * time.Millisecond,
		},
		"attempt 1": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 1,
			want:  2 * time.Millisecond,
		},
		"attempt 63": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 63,
			want:  time.Second,
		},
		"attempt 65": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 65,
			want:  time.Second,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			backoff := NewBackoffCappedExponential(tc.args.maxInterval, tc.args.baseInterval)
			got := backoff.Get(tc.input)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestBackoffFullJitter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRandomizer := mocks.NewMockRandomizer(ctrl)

	type args struct {
		maxInterval  time.Duration
		baseInterval time.Duration
	}

	tests := map[string]struct {
		input   uint
		setMock func()
		args    args
		// random(0, min(cap, base * 2 ** attempt))
		want time.Duration
	}{
		"default value": {
			args: args{
				maxInterval:  0,
				baseInterval: 0,
			},
			input: 0,
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(DefaultBackoffBaseInterval.Nanoseconds()).
					Return(int64(500000))
			},
			want: 500000,
		},
		"attempt 0": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 0,
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(time.Millisecond.Nanoseconds()).
					Return(int64(500000))
			},
			want: 500000,
		},
		"attempt 1": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 1,
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(2 * time.Millisecond.Nanoseconds()).
					Return(int64(750000))
			},
			want: 750000,
		},
		"attempt 63": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 63,
			setMock: func() {
				mockRandomizer.EXPECT().Int63n(gomock.Any()).Times(0)
			},
			want: time.Second,
		},
		"attempt 65": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 65,
			setMock: func() {
				mockRandomizer.EXPECT().Int63n(gomock.Any()).Times(0)
			},
			want: time.Second,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.setMock()
			backoff := NewBackoffFullJitter(tc.args.maxInterval, tc.args.baseInterval, mockRandomizer)
			got := backoff.Get(tc.input)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestBackoffEqualJitter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRandomizer := mocks.NewMockRandomizer(ctrl)

	type args struct {
		maxInterval  time.Duration
		baseInterval time.Duration
	}

	tests := map[string]struct {
		input   uint
		setMock func()
		args    args
		// (exponential * base) / 2 + random(0, ((exponential * base) / 2))
		want time.Duration
	}{
		"default value": {
			args: args{
				maxInterval:  0,
				baseInterval: 0,
			},
			input: 0,
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(DefaultBackoffBaseInterval.Nanoseconds() / 2).
					Return(int64(500000))
			},
			want: (1*DefaultBackoffBaseInterval)/2 + 500000,
		},
		"attempt 0": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 0,
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(time.Millisecond.Nanoseconds() / 2).
					Return(int64(500000))
			},
			want: (1*time.Millisecond)/2 + 500000,
		},
		"attempt 1": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 1,
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n((2 * time.Millisecond.Nanoseconds()) / 2).
					Return(int64(500000))
			},
			want: (2*time.Millisecond)/2 + 500000,
		},
		"attempt 63": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 63,
			setMock: func() {
				mockRandomizer.EXPECT().Int63n(gomock.Any()).Times(0)
			},
			want: time.Second,
		},
		"attempt 65": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 65,
			setMock: func() {
				mockRandomizer.EXPECT().Int63n(gomock.Any()).Times(0)
			},
			want: time.Second,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.setMock()
			backoff := NewBackoffEqualJitter(tc.args.maxInterval, tc.args.baseInterval, mockRandomizer)
			got := backoff.Get(tc.input)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestBackoffDecorrelated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRandomizer := mocks.NewMockRandomizer(ctrl)

	type args struct {
		maxInterval  time.Duration
		baseInterval time.Duration
	}

	tests := map[string]struct {
		setMock func()
		args    args
		// random(base, sleep * 3)
		want     time.Duration
		wantNext time.Duration
	}{
		"default value": {
			args: args{
				maxInterval:  0,
				baseInterval: 0,
			},
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(DefaultBackoffBaseInterval.Nanoseconds()*3 - DefaultBackoffBaseInterval.Nanoseconds()).
					Return(int64(500000))

				mockRandomizer.EXPECT().
					Int63n((500000+DefaultBackoffBaseInterval.Nanoseconds())*3 - DefaultBackoffBaseInterval.Nanoseconds()).
					Return(int64(500000))
			},
			want:     500000 + DefaultBackoffBaseInterval,
			wantNext: 500000 + DefaultBackoffBaseInterval,
		},
		"attempt 1": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(time.Millisecond.Nanoseconds()*3 - time.Millisecond.Nanoseconds()).
					Return(int64(500000))

				mockRandomizer.EXPECT().
					Int63n(1500000*3 - time.Millisecond.Nanoseconds()).
					Return(int64(500000))
			},
			want:     500000 + time.Millisecond,
			wantNext: 500000 + time.Millisecond,
		},
		"attempt overflow": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(time.Millisecond.Nanoseconds()*3 - time.Millisecond.Nanoseconds()).
					Return(int64(9223372036854775807))

				mockRandomizer.EXPECT().
					Int63n(time.Millisecond.Nanoseconds()*3 - time.Millisecond.Nanoseconds()).
					Return(int64(9223372036854775807))
			},
			want:     time.Second,
			wantNext: time.Second,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.setMock()
			backoff := NewBackoffDecorrelated(tc.args.maxInterval, tc.args.baseInterval, mockRandomizer)
			got := backoff.Get(0)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}

			gotNext := backoff.Get(0)
			if gotNext != tc.wantNext {
				t.Fatalf("gotNext %v wantNext %v", gotNext, tc.wantNext)
			}
		})
	}
}

func TestBackoffTruncatedExponential(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRandomizer := mocks.NewMockRandomizer(ctrl)

	type args struct {
		maxInterval  time.Duration
		baseInterval time.Duration
	}

	tests := map[string]struct {
		input   uint
		setMock func()
		args    args
		// exponential * base + random(0,1000)
		want time.Duration
	}{
		"default value": {
			args: args{
				maxInterval:  0,
				baseInterval: 0,
			},
			input: 0,
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(int64(1e+8)).
					Return(int64(500))
			},
			want: (1 * DefaultBackoffBaseInterval) + 500*time.Nanosecond,
		},
		"attempt 0": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 0,
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(int64(1e+6)).
					Return(int64(500))
			},
			want: (1 * time.Millisecond) + 500*time.Nanosecond,
		},
		"attempt 1": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 1,
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(int64(1e+6)).
					Return(int64(500))
			},
			want: (2 * time.Millisecond) + 500*time.Nanosecond,
		},
		"attempt 63": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 63,
			setMock: func() {
				mockRandomizer.EXPECT().
					Int63n(int64(1e+6)).
					Return(int64(500))
			},
			want: time.Second,
		},
		"attempt 65": {
			args: args{
				maxInterval:  time.Second,
				baseInterval: time.Millisecond,
			},
			input: 65,
			setMock: func() {
				mockRandomizer.EXPECT().Int63n(gomock.Any()).Times(0)
			},
			want: time.Second,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.setMock()
			backoff := NewBackoffTruncatedExponential(tc.args.maxInterval, tc.args.baseInterval, mockRandomizer)
			got := backoff.Get(tc.input)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
