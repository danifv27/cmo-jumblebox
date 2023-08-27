package tail

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Position File Implementation", func() {
	var td *TempDir
	var fileStat *FileStat
	var file *os.File
	var pfc OnceCloser
	ginkgo.BeforeEach(func() {
		td = CreateTempDir()
		file, fileStat = td.CreateFile("foo.log")
		// fmt.Println("[DBG]", fileStat)
	})
	ginkgo.AfterEach(func() {
		err := pfc.Close()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		file.Close()
		td.RemoveAll()
	})
	ginkgo.When("When we create a posfile from the log file", func() {
		var pf PositionFile
		var pfpath string
		ginkgo.BeforeEach(func() {
			var err error

			pfpath = filepath.Join(td.Path, "posfile")
			pf, err = OpenPositionFile(pfpath)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			pfc = OnceCloser{
				C: pf,
			}
			pf.Set(fileStat, 0)
			pf.IncreaseOffset(2)
		})
		ginkgo.It("If we increase the offset, we still have the same file", func() {
			ok := SameFile(fileStat, pf.FileStat())
			gomega.Expect(ok).To(gomega.BeTrue())
			g := pf.Offset()
			gomega.Expect(g).To(gomega.Equal(int64(2)))
		})
		ginkgo.It("If we open the file after modify the offset, it should be the same", func() {
			pf2, err := OpenPositionFile(pfpath)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			pfc2 := OnceCloser{
				C: pf2,
			}
			defer pfc2.Close()
			ok := SameFile(fileStat, pf2.FileStat())
			gomega.Expect(ok).To(gomega.BeTrue())
			g := pf2.Offset()
			gomega.Expect(g).To(gomega.Equal(int64(2)))
		})
	})
})
