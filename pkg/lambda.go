package benthosx

import (
	"context"

	"github.com/Jeffail/benthos/lib/config"
	"github.com/Jeffail/benthos/lib/output"
	"github.com/aws/aws-lambda-go/lambda"
)

// FormatIdentity returns the input value unchanged.
func FormatIdentity(v interface{}) (interface{}, error) {
	return v, nil
}

// FormatStatic always returns the Benthos success message.
func FormatStatic(v interface{}) (interface{}, error) {
	return map[string]interface{}{"message": "request successful"}, nil
}

// LambdaFunc coordinates between a Producer and FormatFN to implement the
// behavior that will run in Lambda. The HandleEvent method is compatible with
// the lambda.NewHandler() method from the Go SDK for Lambda.
type LambdaFunc struct {
	Producer Producer
	FormatFn FormatFn
}

// HandleEvent implements a function signature that is compatible with the Go
// Lambda SDK.
func (f *LambdaFunc) HandleEvent(ctx context.Context, in interface{}) (interface{}, error) {
	out, err := f.Producer.Produce(ctx, in)
	if err != nil {
		return nil, err
	}
	return f.FormatFn(out)
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

	var outputFormatter FormatFn = FormatStatic
	conf, echo := replaceStdout(conf)
	if echo {
		outputFormatter = FormatIdentity
	}

	producer, err := NewProducer(conf)
	if err != nil {
		return nil, nil, err
	}

	f := &LambdaFunc{
		Producer: producer,
		FormatFn: outputFormatter,
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

// replaceStdout converts any instance of STDOUT output to
// a drop. If any instance of STDOUT is found then the boolean
// returns true to indicate that the configuration requests outputting
// the processing result.
func replaceStdout(conf *config.Type) (*config.Type, bool) {
	var echo bool
	if conf.Output.Type == output.TypeSTDOUT {
		conf.Output.Type = output.TypeDrop
		echo = true
	}
	if conf.Output.Type == output.TypeBroker {
		conf.Output.Broker, echo = replaceBroker(conf.Output.Broker)
	}
	return conf, echo
}

// replaceBroker scans a broker configuration tree and replaces
// all instances of STDOUT with DROP.
func replaceBroker(conf output.BrokerConfig) (output.BrokerConfig, bool) {
	var echo bool
	for offset, out := range conf.Outputs {
		if out.Type == output.TypeSTDOUT {
			conf.Outputs[offset].Type = output.TypeDrop
			echo = true
		}
		if out.Type == output.TypeBroker {
			var echoBroker bool
			conf.Outputs[offset].Broker, echoBroker = replaceBroker(out.Broker)
			if echoBroker {
				echo = true
			}
		}
	}
	return conf, echo
}
