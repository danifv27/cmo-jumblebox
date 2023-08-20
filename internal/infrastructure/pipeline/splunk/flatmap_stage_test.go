package splunk_test

import (
	"context"

	apipe "fry.org/qft/jumble/internal/application/pipeline"
	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Splunk Flatmap Stage Implementation", func() {
	var stages []apipe.Stager[isplunk.SplunkPipeMsg]
	var incFn func(input isplunk.SplunkPipeMsg) []isplunk.SplunkPipeMsg
	ginkgo.BeforeEach(func() {
		incFn = func(input isplunk.SplunkPipeMsg) []isplunk.SplunkPipeMsg {
			var outMsgs []isplunk.SplunkPipeMsg

			if val, ok := input.Get("value").(int); ok == true {
				outMsg := isplunk.NewSplunkMessage("test", nil)
				outMsg.Add("value", val+1)
				return append(outMsgs, outMsg)
			}
			return outMsgs
		}
		st := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](incFn)
		gomega.Expect(st).ShouldNot(gomega.BeNil())
		stages = append(stages, st)
	})
	ginkgo.Context("Playing with Splunk flatmap based stages", func() {
		var inCh <-chan isplunk.SplunkPipeMsg
		var expectedData []isplunk.SplunkPipeMsg
		ginkgo.When("We instantiate a single stage", func() {
			ginkgo.BeforeEach(func() {
				var inputData []isplunk.SplunkPipeMsg
				for _, v := range []int{2, 4, 6, 8, 10} {
					in := isplunk.NewSplunkMessage("test", nil)
					in.Add("value", v)
					inputData = append(inputData, in)
					expected := isplunk.NewSplunkMessage("test", nil)
					expected.Add("value", v+1)
					expectedData = append(expectedData, expected)
				}
				inCh = isplunk.Chan[isplunk.SplunkPipeMsg](inputData)
			})
			ginkgo.It("Has to return the proper values", func() {
				ctx := context.Background()
				outCh, err1 := stages[0].Action(ctx, inCh, incFn)
				gomega.Expect(err1).ShouldNot(gomega.HaveOccurred())
				res, err2 := isplunk.Drain(ctx, outCh)
				gomega.Expect(err2).ShouldNot(gomega.HaveOccurred())
				gomega.Expect(res).Should(gomega.Equal(expectedData))
			})
		})
	})
})
