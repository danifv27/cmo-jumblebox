package tail

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	aconfigurable "fry.org/qft/jumble/internal/application/configurable"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Context("Without Position File", func() {
	ginkgo.Describe("Test No Position File", func() {
		var td *TempDir
		var f1Stat, f2Stat *FileStat
		var f1, f2 *os.File
		ginkgo.BeforeEach(func() {
			td = CreateTempDir()
			f1, f1Stat = td.CreateFile("foo.log")
		})
		ginkgo.AfterEach(func() {
			td.RemoveAll()
			if f1 != nil {
				err := f1.Close()
				if ok := errors.Is(err, os.ErrClosed); !ok {
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
				}
			}
			if f2 != nil {
				err := f2.Close()
				if ok := errors.Is(err, os.ErrClosed); !ok {
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
				}
			}
		})
		ginkgo.When("Before rotate", func() {
			ginkgo.It("Reading and writing, should increment the file offset", func() {
				var err error
				var r *Reader
				r, err = mustOpenReader(f1.Name())
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()
				f1.WriteString("Foo")

				err = wantRead(r, "Fo", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f1Stat, 2)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				err = wantRead(r, "o", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantReadAll(r, "")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f1Stat, 3)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				f1.WriteString("bar")
				err = wantReadAll(r, "bar")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f1Stat, 6)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
		}) //ginkgo.When("Before rotate", func() {
		ginkgo.When("After rotate", func() {
			ginkgo.It("Reading old file, none old file", func() {
				f1c := OnceCloser{
					C: f1,
				}
				configs := []aconfigurable.ConfigurablerFn{
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(f1.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()
				f1.WriteString("old")
				f1c.Close()
				err = mustRename(f1.Name(), f1.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = mustRemoveFile(f1.Name() + ".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				f2, f2Stat = td.CreateFile(filepath.Base(f1.Name()))
				f2.WriteString("current")
				err = wantRead(r, "ol", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f1Stat, 2)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantRead(r, "d", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f1Stat, 3)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantRead(r, "curr", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f2Stat, 4)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantRead(r, "ent", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f2Stat, 7)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
			ginkgo.It("Exist old file, none remaining bytes", func() {
				f1c := OnceCloser{
					C: f1,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(f1.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()
				f1c.Close()
				err = mustRename(f1.Name(), f1.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				current, currentStat := td.CreateFile(filepath.Base(f1.Name()))
				defer current.Close()
				current.WriteString("current")
				err = wantRead(r, "c", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, currentStat, 1)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantRead(r, "urrent", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, currentStat, 7)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
			ginkgo.It("Exist old file, remaining bytes", func() {
				f1c := OnceCloser{
					C: f1,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(f1.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()
				f1.WriteString("old")
				f1c.Close()
				err = mustRename(f1.Name(), f1.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				current, currentStat := td.CreateFile(filepath.Base(f1.Name()))
				defer current.Close()
				current.WriteString("current")

				err = wantRead(r, "ol", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f1Stat, 2)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantRead(r, "d", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f1Stat, 3)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantRead(r, "c", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, currentStat, 1)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantRead(r, "urrent", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, currentStat, 7)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
			ginkgo.It("Read new file, none new file", func() {
				f1c := OnceCloser{
					C: f1,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(f1.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()
				f1.WriteString("file")
				f1c.Close()
				err = mustRename(f1.Name(), f1.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantReadAll(r, "file")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f1Stat, 4)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
			ginkgo.It("Read new file, exist new file", func() {
				f1c := OnceCloser{
					C: f1,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(f1.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()
				f1.WriteString("old")
				f1c.Close()
				err = mustRename(f1.Name(), f1.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				current, currentStat := td.CreateFile(filepath.Base(f1.Name()))
				defer current.Close()
				current.WriteString("current")

				err = wantRead(r, "oldcurrent", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, currentStat, 7)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				current.WriteString("grow")
				err = wantReadAll(r, "grow")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, currentStat, 11)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
			ginkgo.It("Read new file, rotate again", func() {
				f1c := OnceCloser{
					C: f1,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(f1.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()
				f1.WriteString("f1")

				err = wantRead(r, "f", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f1Stat, 1)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				f1c.Close()
				err = mustRename(f1.Name(), f1.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				f2, f2Stat := td.CreateFile(filepath.Base(f1.Name()))
				f2c := OnceCloser{
					C: f2,
				}
				f2.WriteString("f2")

				err = wantRead(r, "1f", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f2Stat, 1)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				f2c.Close()

				err = mustRename(f2.Name(), f2.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				f3, f3Stat := td.CreateFile(filepath.Base(f2.Name()))
				f3.WriteString("f3")
				err = wantRead(r, "2f3", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f3Stat, 2)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				f3.Close()
			})
		}) //ginkgo.When("After rotate"
		ginkgo.When("Check No Follow Rotate", func() {
			ginkgo.It("Nothing should be read", func() {
				f1c := OnceCloser{
					C: f1,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					WithFollowRotate(false),
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(f1.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()
				f1c.Close()

				err = mustRename(f1.Name(), f1.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				current, _ := td.CreateFile(filepath.Base(f1.Name()))
				defer current.Close()

				current.WriteString("foo")
				err = wantReadAll(r, "")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				time.Sleep(100 * time.Millisecond)
				err = wantReadAll(r, "")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				time.Sleep(100 * time.Millisecond)
				err = wantReadAll(r, "")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				time.Sleep(100 * time.Millisecond)
				err = wantReadAll(r, "")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f1Stat, 0)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
			ginkgo.It("Follow Rotate DetectRotateDelay", func() {
				f1c := OnceCloser{
					C: f1,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					// WithFollowRotate(false),
					WithDetectRotateDelay(time.Second),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(f1.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()
				f1.WriteString("foo")
				f1c.Close()
				err = os.Rename(f1.Name(), f1.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				current, currentStat := td.CreateFile(filepath.Base(f1.Name()))
				defer current.Close()
				err = wantReadAll(r, "foo")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f1Stat, 3)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				time.Sleep(100 * time.Millisecond)
				err = wantReadAll(r, "")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				current.WriteString("barbaz")
				err = wantRead(r, "barbaz", 10*time.Millisecond, 3*time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, currentStat, 6)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
		})
	})
})

var _ = ginkgo.Context("With Position File", func() {
	ginkgo.Describe("testing position file", func() {
		var td *TempDir
		var f1Stat *FileStat
		var f1, f2 *os.File
		var f1Name string

		ginkgo.BeforeEach(func() {
			td = CreateTempDir()
			f1Name = "foo.log"
			f1, f1Stat = td.CreateFile(f1Name)

		})
		ginkgo.AfterEach(func() {
			td.RemoveAll()
			if f1 != nil {
				err := f1.Close()
				if ok := errors.Is(err, os.ErrClosed); !ok {
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
				}
			}
			if f2 != nil {
				err := f2.Close()
				if ok := errors.Is(err, os.ErrClosed); !ok {
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
				}
			}
		})
		ginkgo.It("Works", func() {
			f1.WriteString("bar")
			positionFile := InMemory(f1Stat, 2)
			configs := []aconfigurable.ConfigurablerFn{
				WithPositionFile(positionFile),
			}
			r, err := mustOpenReader(f1.Name(), configs...)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			defer r.Close()

			err = wantReadAll(r, "r")
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			err = wantPositionFile(r, f1Stat, 3)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			f1.WriteString("baz")
			err = wantReadAll(r, "baz")
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			err = wantPositionFile(r, f1Stat, 6)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})
		ginkgo.It("Incorrect Offset", func() {
			f1.WriteString("bar")
			positionFile := InMemory(f1Stat, 4)
			configs := []aconfigurable.ConfigurablerFn{
				WithPositionFile(positionFile),
			}
			r, err := mustOpenReader(f1.Name(), configs...)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			defer r.Close()

			err = wantReadAll(r, "")
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			err = wantPositionFile(r, f1Stat, 3)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})
		ginkgo.Describe("Same file not found", func() {
			ginkgo.It("Rotated file not found", func() {
				f1c := OnceCloser{
					C: f1,
				}

				f1.WriteString("foo")
				f1c.Close()
				os.Rename(f1.Name(), f1.Name()+".bk")
				f2, f2Stat := td.CreateFile(filepath.Base(f1.Name()))
				defer f2.Close()

				positionFile := InMemory(f1Stat, 2)
				configs := []aconfigurable.ConfigurablerFn{
					WithPositionFile(positionFile),
				}
				r, err := mustOpenReader(f2.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()

				err = wantReadAll(r, "")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f2Stat, 0)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				f2.WriteString("bar")
				err = wantReadAll(r, "bar")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, f2Stat, 3)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
			ginkgo.It("Rotated file not found", func() {

				foo, _ := td.CreateFile(f1Name + ".foo-1") // want ignore
				foo.WriteString("foo")
				foo.Close()

				bar, barStat := td.CreateFile(f1Name + ".bar-1") // rotated file
				bar.WriteString("bar")
				bar.Close()

				baz, bazStat := td.CreateFile(f1Name) // current file
				baz.WriteString("baz")
				baz.Close()

				globs := []string{
					filepath.Join(td.Path, f1Name+".bar*"),
					filepath.Join(td.Path, f1Name+".foo*"),
				}
				positionFile := InMemory(barStat, 1)
				configs := []aconfigurable.ConfigurablerFn{
					WithPositionFile(positionFile),
					WithRotatedFilePathPatterns(globs),
					WithWatchRotateInterval(10 * time.Millisecond),
					WithDetectRotateDelay(0),
				}
				r, err := mustOpenReader(baz.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()

				err = wantRead(r, "arbaz", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, bazStat, 3)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
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

func wantReadAll(reader io.Reader, want string) error {

	b, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read all: %v", err)
	}
	if g, w := len(b), len(want); g != w {
		return fmt.Errorf("nReadBytes got %v, want %v", g, w)
	}
	if g, w := string(b), want; g != w {
		return fmt.Errorf("byteString got %v, want %v", g, w)
	}

	return nil
}

func mustRemoveFile(name string) error {

	return os.Remove(name)
}

func mustRename(oldname, newname string) error {

	return os.Rename(oldname, newname)
}
