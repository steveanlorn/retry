// Example of retry with custom back-off strategy by implementing Backoff interface.
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/steveanlorn/retry"
)

type myCustomBackOff struct{}

func (m *myCustomBackOff) Get(attempt uint) time.Duration {
	return time.Millisecond
}

func main() {
	myCustomBackOff := new(myCustomBackOff)
	err := retry.Do(dummyFunc, retry.WithMaxRetryAttempts(1), retry.WithBackoff(myCustomBackOff))
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not do dummyFunc: %v\n", err)
		os.Exit(1)
	}
}

var errorDummy = errors.New("dummy error")

func dummyFunc() error {
	time.Sleep(time.Second)
	return errorDummy
}
