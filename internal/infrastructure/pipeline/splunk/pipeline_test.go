package splunk_test

import (
	"context"

	apipe "fry.org/qft/jumble/internal/application/pipeline"
	ipipe "fry.org/qft/jumble/internal/infrastructure/pipeline"
	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Splunk Messager Implementation", func() {
	var msg isplunk.SplunkPipeMsg
	ginkgo.Context("Playing with Splunk Messagers", func() {
		ginkgo.BeforeEach(func() {
			msg = isplunk.NewSplunkMessage("test", nil)
		})
		ginkgo.When("Adding two keys (int,string) to a Messager", func() {
			ginkgo.BeforeEach(func() {
				msg.Add("int", 7)
				msg.Add("string", "foo")
			})
			ginkgo.It("Data has to be retrievable key by key", func() {
				data := msg.Data()
				gomega.Expect(len(data)).To(gomega.Equal(2))
				intValue, ok := msg.Get("int").(int)
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(intValue).To(gomega.Equal(7))
				strValue, ok1 := msg.Get("string").(string)
				gomega.Expect(ok1).To(gomega.BeTrue())
				gomega.Expect(strValue).To(gomega.Equal("foo"))
			})
		})
		ginkgo.When("Adding a three keys (int,string,bool) map to a Messager", func() {
			var expectedMap map[string]interface{}
			ginkgo.BeforeEach(func() {
				expectedMap = make(map[string]interface{})
				expectedMap["int"] = 5
				expectedMap["string"] = "bar"
				expectedMap["bool"] = false
			})
			ginkgo.It("Data has to be retrievable in one go", func() {
				d := msg.SetData(expectedMap)
				data := d.Data()
				gomega.Expect(len(data)).To(gomega.Equal(3))
				gomega.Expect(data).To(gomega.Equal(expectedMap))
			})
		})
		ginkgo.When("When creating a Messager with topic", func() {
			ginkgo.It("topic has to be retrievable", func() {
				gomega.Expect(msg.Topic()).To(gomega.Equal("test"))
			})
		})
	})
	ginkgo.Context("Playing with Splunk flatmap based pipelines", func() {
		var pipe apipe.Piper[isplunk.SplunkPipeMsg]
		var incFn func(input isplunk.SplunkPipeMsg) []isplunk.SplunkPipeMsg
		ginkgo.BeforeEach(func() {
			var err error
			pipe, err = ipipe.Parse[isplunk.SplunkPipeMsg]("pipeline:splunk")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(pipe).NotTo(gomega.BeNil())
			incFn = func(input isplunk.SplunkPipeMsg) []isplunk.SplunkPipeMsg {
				var outMsgs []isplunk.SplunkPipeMsg

				if val, ok := input.Get("value").(int); ok == true {
					outMsg := isplunk.NewSplunkMessage("test", nil)
					outMsg.Add("value", val+1)
					return append(outMsgs, outMsg)
				}
				return outMsgs
			}
		})
		ginkgo.When("We instantiate a single-staged pipeline", func() {
			var expectedData []isplunk.SplunkPipeMsg
			var inCh <-chan isplunk.SplunkPipeMsg
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
			ginkgo.It("Has to execute the stage", func() {
				ctx := context.Background()
				st := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](incFn)
				gomega.Expect(st).ToNot(gomega.BeNil())
				pipe.Next(st)
				outCh, err := pipe.Do(ctx, inCh)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
				gomega.Expect(outCh).ToNot(gomega.BeNil())
				res, err2 := isplunk.Drain(ctx, outCh)
				gomega.Expect(err2).ToNot(gomega.HaveOccurred())
				gomega.Expect(res).Should(gomega.Equal(expectedData))
			})
		})
		ginkgo.When("We instantiate a multi-staged pipeline", func() {
			var expectedData []isplunk.SplunkPipeMsg
			var inCh <-chan isplunk.SplunkPipeMsg
			ginkgo.BeforeEach(func() {
				var inputData []isplunk.SplunkPipeMsg
				for _, v := range []int{2, 4, 6, 8, 10} {
					in := isplunk.NewSplunkMessage("test", nil)
					in.Add("value", v)
					inputData = append(inputData, in)
					expected := isplunk.NewSplunkMessage("test", nil)
					expected.Add("value", v+2)
					expectedData = append(expectedData, expected)
				}
				inCh = isplunk.Chan[isplunk.SplunkPipeMsg](inputData)
			})
			ginkgo.It("Has to execute all stages", func() {
				ctx := context.Background()
				st := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](incFn)
				gomega.Expect(st).ToNot(gomega.BeNil())
				pipe.Next(st)
				st1 := isplunk.NewSplunkFlatMapStage[isplunk.SplunkPipeMsg](incFn)
				gomega.Expect(st).ToNot(gomega.BeNil())
				pipe.Next(st1)
				outCh, err := pipe.Do(ctx, inCh)
				gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
				gomega.Expect(outCh).ToNot(gomega.BeNil())
				res, err2 := isplunk.Drain(ctx, outCh)
				gomega.Expect(err2).ToNot(gomega.HaveOccurred())
				gomega.Expect(res).Should(gomega.Equal(expectedData))
			})
		})
	})
})
