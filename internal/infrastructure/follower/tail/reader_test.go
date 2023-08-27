package tail

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	aconfigurable "fry.org/qft/jumble/internal/application/configurable"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Context("Using inMemory Position File", func() {
	ginkgo.Describe("Test No Position File", func() {
		var td *TempDir
		var fileStat *FileStat
		var file *os.File
		var r *Reader
		ginkgo.BeforeEach(func() {
			var err error
			td = CreateTempDir()
			file, fileStat = td.CreateFile("foo.log")
			configs := []aconfigurable.ConfigurablerFn{
				WithPositionFile(InMemory(nil, 0)),
				WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
				WithDetectRotateDelay(0),
				WithWatchRotateInterval(100 * time.Millisecond),
			}

			r, err = mustOpenReader(file.Name(), configs...)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			fmt.Println("[DBG]", td)
		})
		ginkgo.AfterEach(func() {
			err := r.Close()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			file.Close()
			td.RemoveAll()
		})
		ginkgo.When("When we create a posfile from the log file", func() {
			ginkgo.BeforeEach(func() {
				file.WriteString("Foo")
			})
			ginkgo.It("If we read two chars, the file offset should be incremented", func() {
				err := wantRead(r, "Fo", 10*time.Millisecond, 5*time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, fileStat, 2)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
			ginkgo.It("If we open the file after modify the offset, it should be the same", func() {

			})
		})
	})
})

func mustOpenReader(name string, opt ...aconfigurable.ConfigurablerFn) (*Reader, error) {

	r, err := OpenTailReader(name, opt...)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func wantRead(r *Reader, want string, interval, timeout time.Duration) error {

	tick := time.NewTicker(interval)
	defer tick.Stop()
	to := time.After(timeout)
	var buf bytes.Buffer

	for {
		b := make([]byte, len(want)-buf.Len())
		n, err := r.Read(b)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read %+v", err)
		}
		if n > 0 {
			buf.Write(b[:n])
		}
		if buf.Len() > len(want) {
			return fmt.Errorf("unexpected read bytes %d, want %s", n, want)
		}
		if buf.Len() == len(want) {
			if g, w := buf.String(), want; g != w {
				return fmt.Errorf("got %v, want %v", g, w)
			}
			return nil
		}
		select {
		case <-to:
			return fmt.Errorf("timeout %s exceeded. got %s, want %s", timeout, buf.String(), want)
		case <-tick.C:
			continue
		}
	}
}

func wantPositionFile(r *Reader, wantFileStat *FileStat, wantOffset int64) error {

	fileStat, offset := r.fu.positionFileInfo()

	if !SameFile(fileStat, wantFileStat) {
		return fmt.Errorf("fileStat not equal")
	}
	if g, w := offset, wantOffset; g != w {
		return fmt.Errorf("offset got %v, want %v", g, w)
	}

	return nil
}
