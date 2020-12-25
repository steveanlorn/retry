// Example of retry with initialization and runs it concurently.
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/steveanlorn/retry"
	"golang.org/x/sync/errgroup"
)

func main() {
	retrier := retry.NewRetrier(
		retry.WithMaxRetryAttempts(1),
		retry.WithBackoff(
			retry.NewBackoffTruncatedExponential(
				time.Second,
				time.Millisecond,
				retry.NewRand(time.Now().UnixNano()),
			),
		),
	)

	g := new(errgroup.Group)

	g.Go(func() error {
		return retrier.Do(dummyFuncA)
	})

	g.Go(func() error {
		return retrier.Do(dummyFuncB)
	})

	g.Go(func() error {
		return retrier.Do(dummyFuncC)
	})

	_, _ = fmt.Fprintln(os.Stdout, "waiting")

	if err := g.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "get error: %v\n", err)
	}
}

var errorDummy = errors.New("dummy error")

func dummyFuncA() error {
	time.Sleep(time.Second)
	return errorDummy
}

func dummyFuncB() error {
	time.Sleep(2 * time.Second)
	return errorDummy
}

func dummyFuncC() error {
	time.Sleep(3 * time.Second)
	return errorDummy
}
