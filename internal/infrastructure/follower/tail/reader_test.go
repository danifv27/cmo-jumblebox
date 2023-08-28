package tail

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	aconfigurable "fry.org/qft/jumble/internal/application/configurable"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Context("Using inMemory Position File", func() {
	ginkgo.Describe("Test No Position File", func() {
		ginkgo.When("Before rotate", func() {
			ginkgo.It("Reading and writing, should increment the file offset", func() {
				var err error
				var td *TempDir
				var fileStat *FileStat
				var file *os.File
				var r *Reader

				td = CreateTempDir()
				file, fileStat = td.CreateFile("foo.log")
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					// 	WithDetectRotateDelay(0),
					// 	WithWatchRotateInterval(100 * time.Millisecond),
				}
				r, err = mustOpenReader(file.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				file.WriteString("Foo")

				err = wantRead(r, "Fo", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, fileStat, 2)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				err = wantRead(r, "o", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantReadAll(r, "")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, fileStat, 3)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				file.WriteString("bar")
				err = wantReadAll(r, "bar")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, fileStat, 6)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				//Cleanup
				r.Close()
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				file.Close()
				td.RemoveAll()
			})
		}) //ginkgo.When("Before rotate", func() {
		ginkgo.When("After rotate", func() {
			ginkgo.It("Reading old file, none old file", func() {
				td := CreateTempDir()
				defer td.RemoveAll()

				old, oldStat := td.CreateFile("test.log")
				oldc := OnceCloser{
					C: old,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(old.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()
				old.WriteString("old")
				oldc.Close()
				err = mustRename(old.Name(), old.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = mustRemoveFile(old.Name() + ".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				current, currentStat := td.CreateFile(filepath.Base(old.Name()))
				defer current.Close()
				current.WriteString("current")
				err = wantRead(r, "ol", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, oldStat, 2)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantRead(r, "d", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, oldStat, 3)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantRead(r, "curr", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, currentStat, 4)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantRead(r, "ent", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, currentStat, 7)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
			ginkgo.It("Exist old file, none remaining bytes", func() {
				td := CreateTempDir()
				defer td.RemoveAll()

				old, _ := td.CreateFile("test.log")
				oldc := OnceCloser{
					C: old,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(old.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				defer r.Close()
				oldc.Close()
				err = mustRename(old.Name(), old.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				current, currentStat := td.CreateFile(filepath.Base(old.Name()))
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
				td := CreateTempDir()
				defer td.RemoveAll()

				old, oldStat := td.CreateFile("test.log")
				oldc := OnceCloser{
					C: old,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(old.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				old.WriteString("old")
				oldc.Close()
				err = mustRename(old.Name(), old.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				current, currentStat := td.CreateFile(filepath.Base(old.Name()))
				defer current.Close()
				current.WriteString("current")

				err = wantRead(r, "ol", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, oldStat, 2)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantRead(r, "d", 10*time.Millisecond, time.Second)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, oldStat, 3)
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
				td := CreateTempDir()
				defer td.RemoveAll()

				f, fileStat := td.CreateFile("test.log")
				fc := OnceCloser{
					C: f,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(f.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				f.WriteString("file")
				fc.Close()
				err = mustRename(f.Name(), f.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantReadAll(r, "file")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				err = wantPositionFile(r, fileStat, 4)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})
			ginkgo.It("Read new file, exist new file", func() {
				td := CreateTempDir()
				defer td.RemoveAll()

				old, _ := td.CreateFile("test.log")
				oldc := OnceCloser{
					C: old,
				}
				configs := []aconfigurable.ConfigurablerFn{
					// WithPositionFile(InMemory(nil, 0)),
					// 	WithRotatedFilePathPatterns([]string{filepath.Join(td.Path, "foo.log.*")}),
					WithDetectRotateDelay(0),
					WithWatchRotateInterval(10 * time.Millisecond),
				}
				r, err := mustOpenReader(old.Name(), configs...)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				old.WriteString("old")
				oldc.Close()
				err = mustRename(old.Name(), old.Name()+".bk")
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				current, currentStat := td.CreateFile(filepath.Base(old.Name()))
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
		}) //ginkgo.When("After rotate", func() {
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

	b, err := ioutil.ReadAll(reader)
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
