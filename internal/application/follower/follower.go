package follower

import "context"

type Follower interface {
	// Lines continuously emits a stream of lines. If an error is found,
	// send it to the error channel and stops emitting strings
	Lines(ctx context.Context) (chan string, chan error)
}
