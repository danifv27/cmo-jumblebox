package file

import (
	"bufio"
	"context"
	"io"
	"os"

	afollower "fry.org/qft/jumble/internal/application/follower"
	"github.com/speijnik/go-errortree"
)

type Reader struct {
	fileName string
	file     *os.File
	lines    chan string
	errs     chan error
}

func NewFollower(path string) afollower.Follower {

	return &Reader{
		fileName: path,
	}
}

// openFile opens the file for reading.
func (r *Reader) openFile() error {
	var rcerror, err error

	if r.file, err = os.Open(r.fileName); err != nil {
		return errortree.Add(rcerror, "file.openFile", err)
	}

	return nil
}

func (r *Reader) Lines(ctx context.Context) (chan string, chan error, error) {
	var rcerror, err error

	failFinalizer := func() {
		if r.file != nil {
			_ = r.file.Close()
		}
	}

	if err = r.openFile(); err != nil {
		failFinalizer()
		return r.lines, r.errs, errortree.Add(rcerror, "file.Run", err)
	}

	r.lines = make(chan string)
	r.errs = make(chan error)

	go func() {
		defer func() {
			var rcerror error

			if r.file != nil {
				if err := r.file.Close(); err != nil {
					r.errs <- errortree.Add(rcerror, "file.Lines.goroutine", err)
				}
			}
			close(r.lines)
			close(r.errs)
		}()
		reader := bufio.NewReader(r.file)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if line, err := reader.ReadString('\n'); err != nil {
					if err != io.EOF {
						r.errs <- errortree.Add(rcerror, "file.Lines.goroutine", err)
					}
					return
				} else {
					r.lines <- line
					//fmt.Printf("[DBG]Line sended\n")
				}
			}
		}
	}()

	return r.lines, r.errs, nil
}
