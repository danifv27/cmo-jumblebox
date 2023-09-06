package tail

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	aconfigurable "fry.org/qft/jumble/internal/application/configurable"
	alogger "fry.org/qft/jumble/internal/application/logger"
)

type fileUnit struct {
	f  *os.File
	pf PositionFile
	mu sync.Mutex
}

// Reader is a file reader that behaves like tail -F
type Reader struct {
	fu             *fileUnit
	state          int32
	followFilePath string
	opt            tailCfg
	closed         chan struct{}
	rotated        chan struct{}
}

const (
	tail_sNormal int32 = iota
	tail_sReadRemaining
	tail_sRotating
)

func newFileUnit(f *os.File, pf PositionFile) *fileUnit {

	return &fileUnit{
		f:  f,
		pf: pf,
	}
}

func (fu *fileUnit) fileInfo() (os.FileInfo, error) {

	fu.mu.Lock()
	defer fu.mu.Unlock()
	return fu.f.Stat()
}

func (fu *fileUnit) fileName() string {

	fu.mu.Lock()
	defer fu.mu.Unlock()
	return fu.f.Name()
}

func (fu *fileUnit) close() error {
	fu.mu.Lock()
	defer fu.mu.Unlock()

	if err := fu.pf.Close(); err != nil {
		fmt.Printf("[DBG]an error occurred while closing the positionFile: %+v", err)
	}

	return fu.f.Close()
}

func (fu *fileUnit) readFile(p []byte) (int, error) {
	fu.mu.Lock()
	defer fu.mu.Unlock()

	n, err := fu.f.Read(p)
	if err != nil {
		return n, err
	}
	if err := fu.pf.IncreaseOffset(n); err != nil {
		return n, err
	}
	return n, nil
}

func (fu *fileUnit) switchFile(next *os.File) error {
	fu.mu.Lock()
	defer fu.mu.Unlock()

	st, err := Stat(next)
	if err != nil {
		return err
	}
	if err := fu.pf.Set(st, 0); err != nil {
		return err
	}
	if err := fu.f.Close(); err != nil {
		fmt.Printf("[DBG]: an error occurred while closing the file: %+v", err)
	}
	fu.f = next

	return nil
}

func (fu *fileUnit) positionFileInfo() (fileStat *FileStat, offset int64) {

	fu.mu.Lock()
	defer fu.mu.Unlock()
	return fu.pf.FileStat(), fu.pf.Offset()
}

func newReader(file *os.File, followFilePath string, positionFile PositionFile, opt tailCfg) *Reader {
	fu := newFileUnit(file, positionFile)
	closed := make(chan struct{})
	rotated := make(chan struct{})
	watchRotate(closed, rotated, fu, followFilePath, opt)
	return &Reader{
		fu:             fu,
		state:          tail_sNormal,
		followFilePath: followFilePath,
		opt:            opt,
		closed:         closed,
		rotated:        rotated,
	}
}

// func NewTailFollower(name string, opts ...aconfigurable.ConfigurablerFn) (afollower.Follower, error) {

// 	return OpenTailReader(name, opts...)
// }

// OpenTailReader opens the named file and returns the follow.Reader
func OpenTailReader(name string, opts ...aconfigurable.ConfigurablerFn) (*Reader, error) {
	var err error
	var opt tailCfg
	var f *os.File

	if opt, err = newTailCfg(); err != nil {
		return nil, err
	}
	opt.Configure(opts...)

	errAndClose := func(err error) (*Reader, error) {
		if f != nil {
			if cErr := f.Close(); cErr != nil {
				opt.printer.Printf("follow: an error occurred while closing the file %s: %+v", name, cErr)
			}
		}
		if opt.positionFile != nil {
			if cErr := opt.positionFile.Close(); cErr != nil {
				opt.printer.Printf("follow: an error occurred while closing the positionFile: %+v", cErr)
			}
		}

		return nil, err
	}

	f, err = OpenFile(name)
	if err != nil {
		return errAndClose(err)
	}
	fileStat, err := Stat(f)
	if err != nil {
		return errAndClose(err)
	}
	fileInfo, err := f.Stat()
	if err != nil {
		return errAndClose(err)
	}

	var initialOffset int64
	if !opt.readFromHead {
		initialOffset = fileInfo.Size()
	}

	positionFile := opt.positionFile
	if positionFile == nil {
		opt.printer.Print("follow: positionFile not specified. use in-memory positionFile.")
		positionFile = InMemory(fileStat, initialOffset)
	}
	if positionFile.FileStat() == nil {
		if err := positionFile.Set(fileStat, initialOffset); err != nil {
			return errAndClose(err)
		}
	}
	if !SameFile(fileStat, positionFile.FileStat()) {
		opt.printer.Printf("follow: file not found that matches fileStat of the positionFile %+v.", positionFile.FileStat())
		sameFile, sameFileStat, sameFileInfo, err := findSameFile(opt.rotatedFilePathPatterns, positionFile.FileStat(), opt.printer)
		if err != nil {
			if !os.IsNotExist(err) {
				return errAndClose(err)
			}
			opt.printer.Printf("follow: reset positionFile %+v.", positionFile.FileStat())
			if err := positionFile.Set(fileStat, initialOffset); err != nil {
				return errAndClose(err)
			}
		} else {
			opt.printer.Printf("follow: %s matches fileStat of the positionFile.", sameFile.Name())
			f = sameFile
			fileStat = sameFileStat
			fileInfo = sameFileInfo
		}
	}

	if fileInfo.Size() < positionFile.Offset() {
		// consider file truncated
		opt.printer.Printf("follow: incorrect positionFile offset %d. file size %d. reset offset to %d.", positionFile.Offset(), fileInfo.Size(), fileInfo.Size())
		if err := positionFile.SetOffset(fileInfo.Size()); err != nil {
			return errAndClose(err)
		}
	}
	offset, err := f.Seek(positionFile.Offset(), 0)
	if err != nil {
		return errAndClose(err)
	}
	if offset != positionFile.Offset() {
		return errAndClose(fmt.Errorf("follow: seems like seek failed. positionFile offset %d. file offset %d", positionFile.Offset(), offset))
	}

	return newReader(f, name, positionFile, opt), nil
}

// Read reads up to len(b) bytes from the File.
func (r *Reader) Read(p []byte) (n int, err error) {
	switch atomic.LoadInt32(&r.state) {
	case tail_sNormal:
		select {
		default:
			return r.fu.readFile(p)
		case <-r.rotated:
			atomic.StoreInt32(&r.state, tail_sReadRemaining)
			return r.Read(p)
		}

	case tail_sReadRemaining:
		n, err := r.fu.readFile(p)
		if err == nil {
			return n, nil
		}
		if err != nil && err != io.EOF {
			return n, err
		}
		// io.EOF (= finish read remaining bytes from rotated file)
		// switch reading to the next file
		if !atomic.CompareAndSwapInt32(&r.state, tail_sReadRemaining, tail_sRotating) {
			// ensure that switching the file is performed by single goroutine
			return 0, io.EOF
		}
		next, err := OpenFile(r.followFilePath)
		if err != nil {
			atomic.StoreInt32(&r.state, tail_sReadRemaining)
			fmt.Printf("[DBG]failed to open the next file. wait for switching the file until next reading: %+v", err)
			return 0, io.EOF
		}
		if err := r.fu.switchFile(next); err != nil {
			atomic.StoreInt32(&r.state, tail_sReadRemaining)
			fmt.Printf("[DBG]failed to switching the file. wait until next reading: %+v", err)
			return 0, io.EOF
		}
		watchRotate(r.closed, r.rotated, r.fu, r.followFilePath, r.opt)
		atomic.StoreInt32(&r.state, tail_sNormal)
		return r.Read(p)

	case tail_sRotating:
		return 0, io.EOF

	default:
		return 0, fmt.Errorf("follow: unexpected state %d", atomic.LoadInt32(&r.state))
	}
}

// Close closes the follow.Reader.
func (r *Reader) Close() error {

	close(r.closed)
	return r.fu.close()
}

func (r *Reader) Lines(ctx context.Context) (chan string, chan error) {

	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	to := time.After(10 * time.Second)
	var buf bytes.Buffer

	ch := make(chan string, 1)
	errCh := make(chan error, 1)

	maxLen := 32
	go func() {
		for {
			b := make([]byte, maxLen-buf.Len())
			n, err := r.Read(b)
			if err != nil && err != io.EOF {
				fmt.Printf("[DBG]Read failure: %v\n", err)
				errCh <- err
				return
			}
			if n > 0 {
				buf.Write(b[:n])
			}

			if buf.Len() > maxLen {
				errCh <- fmt.Errorf("unexpected read bytes %d, want %d", n, maxLen)
				return
			}
			if buf.Len() == maxLen {
				ch <- buf.String()
			}
			select {
			case <-to:
				errCh <- fmt.Errorf("timeout exceeded. got %s", buf.String())
				return
			case <-tick.C:
				continue
			}
		}
	}()

	return ch, errCh

}

func findSameFile(globPatterns []string, findStat *FileStat, printer alogger.Printer) (*os.File, *FileStat, os.FileInfo, error) {
	var f *os.File
	errAndClose := func(tErr error) (*os.File, *FileStat, os.FileInfo, error) {
		if f != nil {
			if cErr := f.Close(); cErr != nil {
				printer.Printf("follow: an error occurred while closing the file %s: %+v", f.Name(), cErr)
			}
		}

		return nil, nil, nil, tErr
	}

	for _, glob := range globPatterns {
		entries, err := filepath.Glob(glob)
		if err != nil {
			return errAndClose(err)
		}

		for _, ent := range entries {
			f, err = OpenFile(ent)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return errAndClose(err)
			}
			fileStat, err := Stat(f)
			if err != nil {
				return errAndClose(err)
			}
			if !SameFile(fileStat, findStat) {
				continue
			}
			// got same file
			fileInfo, err := f.Stat()
			if err != nil {
				return errAndClose(err)
			}

			return f, fileStat, fileInfo, nil
		}
	}

	return nil, nil, nil, os.ErrNotExist
}

func watchRotate(done, notify chan struct{}, fu *fileUnit, followFilePath string, cfg tailCfg) {

	opt := cfg.tailCfgFollowRotate
	if !opt.followRotate {
		return
	}
	fileInfo, err := fu.fileInfo()
	if err != nil {
		cfg.printer.Printf("follow: failed to get FileStat %s on watchRotate: %+v", fu.fileName(), err)
	}

	go func() {
		tick := time.NewTicker(opt.watchRotateInterval)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-tick.C:
				if fileInfo == nil {
					var err error
					fileInfo, err = fu.fileInfo()
					if err != nil {
						cfg.printer.Printf("follow: failed to get FileStat %s on watchRotate: %+v", fu.fileName(), err)
						continue
					}
				}
				currentInfo, err := os.Stat(followFilePath)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					cfg.printer.Printf("follow: failed to get current FileStat %s on watchRotate: %+v", followFilePath, err)
					continue
				}
				if !os.SameFile(fileInfo, currentInfo) {
					<-time.After(opt.detectRotateDelay)
					select {
					case notify <- struct{}{}:
					case <-done:
					}
					return
				}
			}
		}
	}()
}
