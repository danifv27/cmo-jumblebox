package splunk

import (
	"context"
	"errors"

	"github.com/speijnik/go-errortree"
	"github.com/splunk/pipelines"

	apipe "fry.org/qft/jumble/internal/application/pipeline"
)

type SplunkFlatMapStage[S apipe.Messager] struct {
	f func(S) []S
}

func NewSplunkFlatMapStage[S apipe.Messager](f func(S) []S) apipe.Stager[S] {

	return &SplunkFlatMapStage[S]{
		f: f,
	}
}

func (st *SplunkFlatMapStage[S]) Action(ctx context.Context, in <-chan S, prms ...interface{}) (<-chan S, error) {
	var rcerror error
	var hf func(S) []S
	var ok bool

	switch len(prms) {
	case 0:
		hf = st.f
	case 1:
		if hf, ok = prms[0].(func(S) []S); !ok {
			return nil, errortree.Add(rcerror, "SplunkFlatMapStage.Action", errors.New("wrong type"))
		}
	default:
		return nil, errortree.Add(rcerror, "SplunkFlatMapStage.Action", errors.New("wrong number of prms"))
	}

	return pipelines.FlatMap[S](ctx, in, hf), nil
}
