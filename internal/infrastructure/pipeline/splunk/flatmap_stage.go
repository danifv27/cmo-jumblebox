package splunk

import (
	"context"
	"errors"

	"github.com/speijnik/go-errortree"

	aconfigurable "fry.org/qft/jumble/internal/application/configurable"
	apipe "fry.org/qft/jumble/internal/application/pipeline"
)

type SplunkFlatMapStage[S apipe.Messager] struct {
	f   func(S) []S
	cfg aconfigurable.Configurabler
}

func NewSplunkFlatMapStage[S apipe.Messager](f func(S) []S, configs ...aconfigurable.ConfigurablerFn) apipe.Stager[S] {

	fm := SplunkFlatMapStage[S]{
		f:   f,
		cfg: NewSplunkConfig(configs...),
	}

	return &fm
}

func (st *SplunkFlatMapStage[S]) Action(ctx context.Context, in <-chan S, prms ...interface{}) (<-chan S, error) {
	var rcerror, err error
	var hf func(S) []S
	var ok bool
	var outputs, buffersize, workers int

	if outputs, err = st.cfg.GetInt("outputs"); err != nil {
		return nil, errortree.Add(rcerror, "SplunkFlatMapStage.Action", err)
	}
	if buffersize, err = st.cfg.GetInt("buffersize"); err != nil {
		return nil, errortree.Add(rcerror, "SplunkFlatMapStage.Action", err)
	}
	if workers, err = st.cfg.GetInt("workers"); err != nil {
		return nil, errortree.Add(rcerror, "SplunkFlatMapStage.Action", err)
	}
	outs := makeOutputChannels[S](outputs, buffersize)
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

	go func() {
		defer func() {
			for _, ch := range outs {
				close(ch)
			}
		}()
		// flatmap stage doesn't support pool of workers
		if workers == 1 {
			doFlatMap[S](ctx, in, hf, outs[0])
		}
	}()

	return outs[0], nil
}

func doFlatMap[S apipe.Messager](ctx context.Context, in <-chan S, f func(S) []S, out chan<- S) {

	for {
		select {
		case <-ctx.Done():
			return
		case s, ok := <-in:
			if !ok {
				return
			}
			sendAll(ctx, f(s), out)
		}
	}
}

// sendAll sends all values in a slice to the provided channel.
// It blocks until the channel is closed or the provided context is cancelled.
func sendAll[S any](ctx context.Context, ts []S, ch chan<- S) {

	for _, t := range ts {
		select {
		case <-ctx.Done():
			return
		case ch <- t:
		}
	}
}
