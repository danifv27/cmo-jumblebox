package pipeline

import (
	"context"
)

type Stager[S Messager] interface {
	Action(ctx context.Context, in <-chan S, prms ...interface{}) (<-chan S, error)
}

type Piper[S Messager] interface {
	// Next simply stores the chain of action steps.
	// f is Stage implementation
	Next(st Stager[S]) Piper[S]
	// Do executes the chain, or cuts it early in case of an error
	Do(ctx context.Context, in <-chan S) (<-chan S, error)
}

// Pipe message interface
type Messager interface {
	// Topic get message topic
	Topic() string
	// Get data by index
	Get(key string) interface{}
	// Set value by key
	Set(key string, val interface{})
	// Add value by key
	Add(key string, val interface{})
	// Data get all data
	Data() map[string]interface{}
	// SetData set data to the event
	SetData(map[string]interface{}) Messager
}
