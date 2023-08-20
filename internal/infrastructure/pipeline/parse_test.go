package pipeline_test

import (
	ipipe "fry.org/qft/jumble/internal/infrastructure/pipeline"
	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Parse Pipeline Implementation", func() {
	ginkgo.When("We create a new pipeline", func() {
		ginkgo.It("Has to have a known scheme and opaque value", func() {
			_, err1 := ipipe.Parse[isplunk.SplunkPipeMsg]("pipeline:dummy")
			gomega.Expect(err1).Should(gomega.HaveOccurred())
			gomega.Expect(err1.Error()).Should(gomega.ContainSubstring("unsupported pipeline implementation"))
			_, err2 := ipipe.Parse[isplunk.SplunkPipeMsg]("foo:splunk")
			gomega.Expect(err2).Should(gomega.HaveOccurred())
			gomega.Expect(err2.Error()).Should(gomega.ContainSubstring("invalid scheme"))
			_, err3 := ipipe.Parse[isplunk.SplunkPipeMsg]("foo:splunk")
			gomega.Expect(err3).Should(gomega.HaveOccurred())
		})
	})
	ginkgo.When("We create a new splunk pipeline", func() {
		ginkgo.It("Has to be a splunk based one", func() {
			ppln, err4 := ipipe.Parse[isplunk.SplunkPipeMsg]("pipeline:splunk")
			gomega.Expect(err4).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(ppln).NotTo(gomega.BeNil())
			expectedPipe := new(isplunk.SplunkPipe[isplunk.SplunkPipeMsg])
			gomega.Expect(ppln).To(gomega.BeAssignableToTypeOf(expectedPipe))
		})
	})
})
