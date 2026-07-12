package interfaces

import "context"

type StreamObject interface {
	Data() string
	Error() error
}

// StreamObjectWithID is an optional extension of StreamObject that supplies an
// SSE `id:` for the frame. The stream handler writes the id line before the
// data line when a StreamObject implements it.
type StreamObjectWithID interface {
	StreamObject
	ID() string
}

type ServerSideStreamable interface {
	Stream(ctx context.Context) (<-chan StreamObject, error)
}
