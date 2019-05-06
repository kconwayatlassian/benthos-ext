package benthosx

import "context"

// Producer is referenced by components of this package as an entry
// point for sending messages to an output. The output is expected to be the
// final form of the message(s) sent to the underlying streams. In the case
// that the input results in multiple outputs then the response should be a
// slice of messages.
type Producer interface {
	Produce(ctx context.Context, in interface{}) (interface{}, error)
	Close() error
}

// FormatFn is used by the Lambda function to convert a Producer response
// into a version that it returns to the caller.
type FormatFn func(in interface{}) (interface{}, error)
