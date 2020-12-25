// Example of retry with unretryable error.
package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/steveanlorn/retry"
)

var server *httptest.Server

func init() {
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
}

func main() {
	err := retry.Do(getData)
	if err != nil {
		fmt.Fprintln(os.Stderr, "got err ", err)

		var unretryableError *retry.UnretryableError
		if errors.As(err, &unretryableError) {
			fmt.Fprintln(os.Stderr, "error is unretryable")
		}

		os.Exit(1)
	}
}

func getData() error {
	resp, err := http.Get(server.URL)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusBadRequest {
		_ = resp.Body.Close()
		return retry.Unretryable(errors.New("error bad request"))
	}

	return resp.Body.Close()
}
