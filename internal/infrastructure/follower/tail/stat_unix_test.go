package tail

import (
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Unix Stats Implementation", func() {
	var td *TempDir
	var fileStat, newfileStat, renamedStat *FileStat
	var file, newfile *os.File
	ginkgo.BeforeEach(func() {
		td = CreateTempDir()
		defer td.RemoveAll()
		file, fileStat = td.CreateFile("foo-file")
		file.Close()
		os.Rename(file.Name(), file.Name()+".bk")
		renamedStat = StatByName(file.Name() + ".bk")
		newfile, newfileStat = td.CreateFile("foo-file")
		newfile.Close()
	})
	ginkgo.Context("Comparing two files", func() {
		ginkgo.When("We compare two files", func() {
			ginkgo.It("SameFile has to fail, if files are different", func() {
				ok := SameFile(fileStat, renamedStat)
				gomega.Expect(ok).To(gomega.BeTrue())
			})
			ginkgo.It("SameFile has to succeed, if files are equal", func() {
				ok := SameFile(fileStat, newfileStat)
				gomega.Expect(ok).To(gomega.BeFalse())
			})
		})
	})
})
