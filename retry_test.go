package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/steveanlorn/retry/mocks"
)

func Test_isRetryable(t *testing.T) {
	tests := map[string]struct {
		maximumRetryAttempts uint
		attempt              uint
		err                  error
		want                 bool
	}{
		"out of attempt": {
			maximumRetryAttempts: 5,
			attempt:              5,
			err:                  nil,
			want:                 false,
		},
		"retryable error": {
			maximumRetryAttempts: 5,
			attempt:              2,
			err:                  errors.New("some error"),
			want:                 true,
		},
		"unretryable error": {
			maximumRetryAttempts: 5,
			attempt:              2,
			err:                  Unretryable(errors.New("some error")),
			want:                 false,
		},
		"retryable attempt": {
			maximumRetryAttempts: 5,
			attempt:              2,
			err:                  nil,
			want:                 true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			retrier := new(Retrier)
			retrier.maximumRetryAttempts = tc.maximumRetryAttempts
			if got := retrier.isRetryable(tc.attempt, tc.err); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestRetrier_Do(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackoff := mocks.NewMockBackoff(ctrl)

	var counter int

	tests := map[string]struct {
		input   func() error
		setMock func()
		wantErr bool
	}{
		"success": {
			input: func() error {
				return nil
			},
			setMock: func() {},
			wantErr: false,
		},
		"failed until max attempt": {
			input: func() error {
				return errors.New("some error")
			},
			setMock: func() {
				mockBackoff.EXPECT().Get(uint(0)).Return(time.Millisecond)
				mockBackoff.EXPECT().Get(uint(1)).Return(time.Millisecond)
				mockBackoff.EXPECT().Get(uint(2)).Return(time.Millisecond)
			},
			wantErr: true,
		},
		"success in first retry": {
			input: func() error {
				if counter == 1 {
					return nil
				}
				counter++
				return errors.New("some error")
			},
			setMock: func() {
				mockBackoff.EXPECT().Get(uint(0)).Return(time.Millisecond)
			},
			wantErr: false,
		},
		"encounter unretryable error": {
			input: func() error {
				return Unretryable(errors.New("some error"))
			},
			setMock: func() {},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.setMock()
			retrier := NewRetrier(WithBackoff(mockBackoff), WithMaxRetryAttempts(3))
			gotErr := retrier.Do(tc.input)
			if (gotErr != nil) != tc.wantErr {
				t.Fatalf("got err %v want %v", gotErr, tc.wantErr)
			}
		})
	}
}

func TestRetrier_DoWithContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackoff := mocks.NewMockBackoff(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	f := func() error {
		return errors.New("some error")
	}

	mockBackoff.EXPECT().Get(uint(0)).Return(3 * time.Second)

	wantErr := true

	retrier := NewRetrier(WithBackoff(mockBackoff), WithMaxRetryAttempts(3))

	gotErr := retrier.DoWithContext(ctx, f)
	if (gotErr != nil) != wantErr {
		t.Fatalf("got err %v want %v", gotErr, wantErr)
	}
}

func TestDo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackoff := mocks.NewMockBackoff(ctrl)

	var counter int

	tests := map[string]struct {
		input   func() error
		setMock func()
		wantErr bool
	}{
		"success": {
			input: func() error {
				return nil
			},
			setMock: func() {},
			wantErr: false,
		},
		"failed until max attempt": {
			input: func() error {
				return errors.New("some error")
			},
			setMock: func() {
				mockBackoff.EXPECT().Get(uint(0)).Return(time.Millisecond)
				mockBackoff.EXPECT().Get(uint(1)).Return(time.Millisecond)
				mockBackoff.EXPECT().Get(uint(2)).Return(time.Millisecond)
			},
			wantErr: true,
		},
		"success in first retry": {
			input: func() error {
				if counter == 1 {
					return nil
				}
				counter++
				return errors.New("some error")
			},
			setMock: func() {
				mockBackoff.EXPECT().Get(uint(0)).Return(time.Millisecond)
			},
			wantErr: false,
		},
		"encounter unretryable error": {
			input: func() error {
				return Unretryable(errors.New("some error"))
			},
			setMock: func() {},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.setMock()
			gotErr := Do(tc.input, WithBackoff(mockBackoff), WithMaxRetryAttempts(3))
			if (gotErr != nil) != tc.wantErr {
				t.Fatalf("got err %v want %v", gotErr, tc.wantErr)
			}
		})
	}
}

func TestDoWithContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackoff := mocks.NewMockBackoff(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	f := func() error {
		return errors.New("some error")
	}

	mockBackoff.EXPECT().Get(uint(0)).Return(3 * time.Second)

	wantErr := true

	gotErr := DoWithContext(ctx, f, WithBackoff(mockBackoff), WithMaxRetryAttempts(3))
	if (gotErr != nil) != wantErr {
		t.Fatalf("got err %v want %v", gotErr, wantErr)
	}
}

func TestUnretryableError_Unwrap(t *testing.T) {
	f := func() error {
		return Unretryable(context.Canceled)
	}
	gotErr := Do(f)
	if !errors.Is(gotErr, context.Canceled) {
		t.Fatalf("want error %v got %v", context.Canceled, gotErr)
	}

	f2 := func() error {
		return Unretryable(context.DeadlineExceeded)
	}
	gotErr2 := Do(f2)
	if !errors.Is(gotErr2, context.DeadlineExceeded) {
		t.Fatalf("want error %v got %v", context.DeadlineExceeded, gotErr2)
	}
}
