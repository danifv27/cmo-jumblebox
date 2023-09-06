package tailor

import (
	"context"
	"io"
	"time"

	afollower "fry.org/qft/jumble/internal/application/follower"
	"github.com/un000/tailor"
)

// Reader is a file reader that behaves like tail -F
type Reader struct {
	t *tailor.Tailor
}

func NewFollower(path string) afollower.Follower {

	ta := tailor.New(
		path,
		tailor.WithSeekOnStartup(0, io.SeekStart),
		tailor.WithPollerTimeout(100*time.Millisecond),
	)

	return &Reader{
		t: ta,
	}
}

func (r *Reader) Lines(ctx context.Context) (chan string, chan error, error) {

	chLines := make(chan string)
	if err := r.t.Run(ctx); err != nil {
		return chLines, r.t.Errors(), err
	}
	go func() {
		for {
			select {
			case entry := <-r.t.Lines():
				chLines <- entry.String()
			case <-ctx.Done():
				return
			}
		}
	}()

	return chLines, r.t.Errors(), nil
}
