package logger_test

import (
	alogger "fry.org/qft/jumble/internal/application/logger"
	ilogger "fry.org/qft/jumble/internal/infrastructure/logger"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Parse Logger Implementation", func() {
	ginkgo.When("We create a new logger", func() {
		ginkgo.It("Has to be a logrus based one", func() {
			var err error
			var lg alogger.Logger

			_, err = ilogger.Parse("logger:dummy")
			gomega.Expect(err).Should(gomega.HaveOccurred())
			gomega.Expect(err.Error()).Should(gomega.ContainSubstring("unsupported logger implementation"))
			_, err = ilogger.Parse("foo:logrus")
			gomega.Expect(err).Should(gomega.HaveOccurred())
			gomega.Expect(err.Error()).Should(gomega.ContainSubstring("invalid scheme"))
			lg, err = ilogger.Parse("logger:logrus")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(lg).NotTo(gomega.BeNil())
		})
	})
})
