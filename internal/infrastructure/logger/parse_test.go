package logger

import (
	alogger "fry.org/qft/jumble/internal/application/logger"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Parse Logger Implementation", func() {
	ginkgo.When("We create a new logger", func() {
		ginkgo.It("Has to be a logrus based one", func() {
			var err error
			var lg alogger.Logger

			_, _, _, err = Parse("logger:dummy")
			gomega.Expect(err).Should(gomega.HaveOccurred())
			gomega.Expect(err.Error()).Should(gomega.ContainSubstring("unsupported logger implementation"))
			_, _, _, err = Parse("foo:logrus")
			gomega.Expect(err).Should(gomega.HaveOccurred())
			gomega.Expect(err.Error()).Should(gomega.ContainSubstring("invalid scheme"))
			lg, _, _, err = Parse("logger:logrus")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(lg).NotTo(gomega.BeNil())
		})
	})
	ginkgo.When("We create a new logger", func() {
		ginkgo.It("Has to be a void based one", func() {
			var err error
			var pt alogger.Printer

			_, pt, _, err = Parse("printer:void")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(pt).NotTo(gomega.BeNil())
		})
	})
})
