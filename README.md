# Retry

Functionality to retry a function call (usually a network call) with back-off mechanism.

## Overview

When external service calls is failed in our app, we as engineer want to retry it. But retry is selfish because when failures are caused by overload, retries that increase load can make chance of success worse. In the other end, server usually implements rate limitter, therefore bruteforcing retry is not a good solution. To solve this problem, we add some amount of time between retries (back-off).

## Supported Back-Off Strategy
1. **Constant Back-off**  
    Back-off strategy with constant time in every retry attempt. Example below is back-off with 100 millisecond delay:
    ```
    attempt 0: delay 100ms
    attempt 1: delay 100ms
    attempt 2: delay 100ms
    attempt 3: delay 100ms
    attempt 4: delay 100ms
    ```
2. **Capped Exponential Back-Off Strategy**   
    Back-off strategy to multiplied back-off by a constant after each attempt, up to some maximum value. Example below is back-off with 100 millisecond base delay and 1 second maximum delay:
    ```
    attempt 0: delay 100ms
    attempt 1: delay 200ms
    attempt 2: delay 400ms
    attempt 3: delay 800ms
    attempt 4: delay 1s
    ```
3. **Full Jitter Back-Off Strategy**  
    Back-off strategy to multiplied back-off by a random delay up until current multiplied backoff. Example below is back-off with 100 millisecond base delay and 1 second maximum delay:
    ```
    attempt 0: delay 80.374688ms
    attempt 1: delay 52.725529ms
    attempt 2: delay 194.759995ms
    attempt 3: delay 86.285294ms
    attempt 4: delay 371.775302ms
    ```
4. **Equal Jitter Back-Off Strategy**  
    This back-off strategy is a timed backoff loops which keeps some of the backoff and jitter by a smaller amount. Example below is back-off with 100 millisecond base delay and 1 second maximum delay:
    ```
    attempt 0: delay 56.423945ms
    attempt 1: delay 155.708106ms
    attempt 2: delay 325.189706ms
    attempt 3: delay 735.64193ms
    attempt 4: delay 663.795241ms
    ```
5. **Decorrelated Back-Off Strategy**  
    This back-off strategy is a timed backoff loops which is similar to "Full Jitter"
    with increment in the maximum jitter based on the last back-off value. Example below is back-off with 100 millisecond base delay and 1 second maximum delay:
    ```
    attempt 0: delay 235.554134ms
    attempt 1: delay 630.716505ms
    attempt 2: delay 1s
    attempt 3: delay 965.507323ms
    attempt 4: delay 1s
    ```
6. **Truncated Exponential Back-Off Strategy**  
    Back-off strategy to multiplied backoff by periodically increasing delays with additional jitters. Example below is back-off with 100 millisecond base delay and 1 second maximum delay:
    ```
    attempt 0: delay 180.735148ms
    attempt 1: delay 242.153766ms
    attempt 2: delay 475.947205ms
    attempt 3: delay 839.333323ms
    ```

## Default Configuration
Default configurations are applied if configuration is not provided.  
| Configuration                     | Value             | Description                   |
| -----------                       | -----------       | -----------                   |
| DefaultRetryMaxAttempts           | 10                | maximum retry attempt allowed |
| defaultRetryBackoff               | BackoffConstant   | default backoff strategy      |
| DefaultBackoffMaximumInterval     | 1 second          | maximum backoff interval cap  |
| DefaultBackoffBaseInterval        | 100 millisecond   | base backoff interval         | 

## Unretryable Error
When we encountered with error that is not retryable, we do not want to retry to the next attempt. Imagine that we receive status code like `404` or `400` or `401`, then we can assume that service will alywas return this error whenever we try.  

To solve this, you can wrap an error with function `Unretryable`
```go
func() error {
    return Unretryable(errors.New("some error"))
}
```
This will wrap the error into `UnretryableError` struct and will prevent retrying to the next attempt. You can still get the context of the previous error:
```go
if errors.Is(err, context.DeadlineExceeded) {
    // ...
}
```

## Example
Provided two ways to use the retrier. If you just want to use one retrier, you can do like this:
```go
err := Do(yourfunction, optionalOptions)
if err != nil {
    // do something with the error
}
```

If you need to re-use retrier with the same configuration, you can initialize the retrier first.
```go
retrier := NewRetrier(optionalOptions)
err := retrier.Do(yourfunction)
if err != nil {
    // do something with the error
}
```

If you need to limit a whole retry procedure with a context, you can use `DoWithContext` function or method. While waiting for next attempt, it will also listen to context cancelation. It will throw new error that will wrap the last error value.
```go
err := DoWithContext(ctx, yourfunction, optionalOptions)
if err != nil {
    // do something with the error
}
```

Here is an example how to applied different back-off strategy. Function `NewRand` provides concurency save rand seed source.
```go
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
```

More examples are provided in example directory.

## References
- https://aws.amazon.com/builders-library/timeouts-retries-and-backoff-with-jitter/
- https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
- https://cloud.google.com/storage/docs/exponential-backoff