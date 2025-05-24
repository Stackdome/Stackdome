package interfaces

import "context"

type StreamObject interface {
	Data() string
	Error() error
}

type ServerSideStreamable interface {
	Stream(ctx context.Context) (<-chan StreamObject, error)
}
