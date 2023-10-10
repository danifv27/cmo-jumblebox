package splunk

import "context"

// A sink serves as the last stage of a processing pipeline. All sinks are implemented as blocking calls which don't start any new goroutines.
// Drain receives all values from the provided channel and returns them in a slice.
// Drain blocks the caller until the input channel is closed or the provided context is cancelled.
// An error is returned if and only if the provided context was cancelled before the input channel was closed.
func Drain[T any](ctx context.Context, in <-chan T) ([]T, error) {
	var result []T
	for {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case repo, ok := <-in:
			if !ok {
				return result, nil
			}
			result = append(result, repo)
		}
	}
}

// Chan converts a slice to a channel. The channel returned is a closed, buffered channel containing exactly the same
// values.
func Chan[T any](in []T) <-chan T {
	result := make(chan T, len(in))
	defer close(result) // non-empty buffered channels can be drained even when closed.
	for _, t := range in {
		result <- t
	}
	return result
}
