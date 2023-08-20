package splunk

import (
	"context"

	apipe "fry.org/qft/jumble/internal/application/pipeline"
	"github.com/speijnik/go-errortree"
)

type SplunkPipe[S apipe.Messager] struct {
	steps []apipe.Stager[S]
}

type SplunkPipeMsg struct {
	// message topic
	topic string
	// user data.
	data map[string]interface{}
}

func NewSplunkPipe[S apipe.Messager]() apipe.Piper[S] {

	return &SplunkPipe[S]{}
}

func (p *SplunkPipe[S]) Next(st apipe.Stager[S]) apipe.Piper[S] {

	p.steps = append(p.steps, st)

	return p
}

func (p *SplunkPipe[S]) Do(ctx context.Context, in <-chan S) (<-chan S, error) {
	var rcerror, err error
	var inch, outch <-chan S

	inch = in
	for _, st := range p.steps {
		if outch, err = st.Action(ctx, inch); err != nil {
			return nil, errortree.Add(rcerror, "SplunkPipe.Do", err)
		}
		inch = outch
	}

	return outch, nil
}

// NewSplunkMessage
func NewSplunkMessage(topic string, data map[string]interface{}) SplunkPipeMsg {

	if data == nil {
		data = make(map[string]interface{})
	}

	return SplunkPipeMsg{
		topic: topic,
		data:  data,
	}
}

// Get data by index
func (m SplunkPipeMsg) Get(key string) interface{} {
	if v, ok := m.data[key]; ok {
		return v
	}

	return nil
}

// Add value by key
func (m SplunkPipeMsg) Add(key string, val interface{}) {
	if _, ok := m.data[key]; !ok {
		m.Set(key, val)
	}
}

// Set value by key
func (m SplunkPipeMsg) Set(key string, val interface{}) {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}

	m.data[key] = val
}

// Topic get message name
func (m SplunkPipeMsg) Topic() string {

	return m.topic
}

// Data get all data
func (m SplunkPipeMsg) Data() map[string]interface{} {

	return m.data
}

// SetData set data to the message
func (m SplunkPipeMsg) SetData(data map[string]interface{}) apipe.Messager {

	if data != nil {
		m.data = data
	}

	return m
}
