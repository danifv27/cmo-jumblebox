package logrus_test

import (
	"bytes"
	"io"
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	ilogger "fry.org/qft/jumble/internal/infrastructure/logger/logrus"
)

func capture(f func()) string {

	r, w, err := os.Pipe()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	stdout := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = stdout
	}()

	f()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

var _ = ginkgo.Describe("Logger Logrus Implementation", func() {
	var (
	// lgr alogger.Logger
	)
	ginkgo.BeforeEach(func() {
		// l := ilogger.NewLogger(os.Stdout)
		// lgr = l
	})
	ginkgo.When("We print a debug message", func() {
		ginkgo.It("Print it to stdout ", func() {

			tf := func() {
				lg := ilogger.NewLogger(os.Stdout)
				lg.Log("[DBG]", "Hello")
			}
			rc := capture(tf)
			gomega.Expect(rc).Should(gomega.ContainSubstring("[DBG]"))
			gomega.Expect(rc).Should(gomega.ContainSubstring("Hello"))
			gomega.Expect(rc).ShouldNot(gomega.ContainSubstring("World"))
			logf := func() {
				lg := ilogger.NewLogger(os.Stdout)
				lg.Logf("[DBG]%S", "World")
			}
			rc = capture(logf)
			gomega.Expect(rc).Should(gomega.ContainSubstring("[DBG]"))
			gomega.Expect(rc).ShouldNot(gomega.ContainSubstring("Hello"))
			gomega.Expect(rc).Should(gomega.ContainSubstring("World"))
		})
	})
})
