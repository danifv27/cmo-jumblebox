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

			_, err = Parse("logger:dummy")
			gomega.Expect(err).Should(gomega.HaveOccurred())
			gomega.Expect(err.Error()).Should(gomega.ContainSubstring("unsupported logger implementation"))
			_, err = Parse("foo:logrus")
			gomega.Expect(err).Should(gomega.HaveOccurred())
			gomega.Expect(err.Error()).Should(gomega.ContainSubstring("invalid scheme"))
			lg, err = Parse("logger:logrus")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(lg).NotTo(gomega.BeNil())
		})
		ginkgo.It("Has to be a void based one", func() {
			var err error
			var lg alogger.Logger

			lg, err = Parse("logger:void")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(lg).NotTo(gomega.BeNil())
		})
	})
})
