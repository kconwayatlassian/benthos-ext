package benthosx

import (
	"context"

	"github.com/Jeffail/benthos/lib/config"
	"github.com/aws/aws-lambda-go/lambda"
)

// LambdaFunc coordinates between a Producer and FormatFN to implement the
// behavior that will run in Lambda. The HandleEvent method is compatible with
// the lambda.NewHandler() method from the Go SDK for Lambda.
type LambdaFunc struct {
	Producer Producer
}

// HandleEvent implements a function signature that is compatible with the Go
// Lambda SDK.
func (f *LambdaFunc) HandleEvent(ctx context.Context, in interface{}) (interface{}, error) {
	out, err := f.Producer.Produce(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NewLambdaFunc generates a LambdaFunc from a given Benthos configuratino.
// The input and buffer sections of the configuration are ignored since they
// are not relevant in a Lambda setting where the only input is the function
// call. If any of the outputs are STDOUT then the resulting LambdaFunc will
// return the processed event to the Lambda caller in JSON form. If none of
// outputs are STDOUT then a static message will be returned instead.
func NewLambdaFunc(conf *config.Type) (*LambdaFunc, func() error, error) {
	// The mainline Benthos binary for Lambda implements a switching behavior
	// based on whether or not an output is configured. If the Benthos configuration
	// contains an output definition then the Lambda will process each event,
	// produce the final form to the output, and return a canned success response.
	// If there is no output configuration then the default value of STDOUT is
	// set in the parsed configuration. The STDOUT output is used then the Lambda
	// will send all events through the processing pipeline and the final, modified
	// payload is returned as JSON from the Lambda function to the caller.
	//
	// This version modifies the behavior to allow for having both configured
	// outputs and returning a JSON version of the event produced to the outputs.
	// This is accomplished by doing the following:
	//
	//	- Always capturing the processing output regardless of mode
	//	- Replacing all instances of STDOUT in a configuration with DROP
	//	- Switching the Lambda return value based on whether STDOUT was used
	//	  anywhere in the original configuration.
	//
	// Using this allows for users to configure the output as a broker that
	// sends to both a stream, such as Kinesis, and STDOUT. This combination
	// would result in both a message being produced to Kinesis and the same
	// message being returned to the Lambda caller.

	producer, err := NewProducer(conf)
	if err != nil {
		return nil, nil, err
	}

	f := &LambdaFunc{
		Producer: producer,
	}
	return f, producer.Close, nil
}

// NewLambdaHandler is a convenience function for converting the LambdaFunc
// output from NewLambdaFunc into a lambda.Handler type.
func NewLambdaHandler(conf *config.Type) (lambda.Handler, func() error, error) {
	f, close, err := NewLambdaFunc(conf)
	if err != nil {
		return nil, nil, err
	}
	return lambda.NewHandler(f.HandleEvent), close, nil
}
