// Example of retry with context. We can differentiate if error comming from the context done.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/steveanlorn/retry"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	err := retry.DoWithContext(ctx, dummyFunc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not do dummyFunc: %v\n", err)

		var contextDoneError *retry.ContextDoneError
		if errors.As(err, &contextDoneError) {
			fmt.Fprintln(os.Stderr, "error is context done error")
		}

		if errors.Is(err, errorDummy) {
			fmt.Fprintln(os.Stderr, "error is dummy error")
		}

		os.Exit(1)
	}
}

var errorDummy = errors.New("dummy error")

func dummyFunc() error {
	time.Sleep(time.Second)
	return errorDummy
}
