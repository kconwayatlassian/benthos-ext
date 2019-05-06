package benthosx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Jeffail/benthos/lib/config"
	"github.com/Jeffail/benthos/lib/log"
	"github.com/Jeffail/benthos/lib/manager"
	"github.com/Jeffail/benthos/lib/message"
	"github.com/Jeffail/benthos/lib/metrics"
	"github.com/Jeffail/benthos/lib/output"
	"github.com/Jeffail/benthos/lib/pipeline"
	"github.com/Jeffail/benthos/lib/response"
	"github.com/Jeffail/benthos/lib/types"
)

// NewProducer uses the given Benthos configuration to create a Producer
// instance that may be used as either a client in other code or as input
// for constructing a Lambda function.
func NewProducer(conf *config.Type) (*BenthosProducer, error) {
	// Disable logging in favor of creating our own logging decorator for
	// the producer interface as needed.
	logger := log.Noop()
	// Disable metrics in favor of having our own metrics decorator for the
	// producer interfaces as needed.
	stats := metrics.Noop()

	// piplineLayer represents the processing pipeline that a message will
	// pass through.
	var pipelineLayer pipeline.Type
	// outputLayer represents the composite of the various destination/outputs
	// defined by the configuration.
	var outputLayer output.Type
	// pipelineInput is the entry point to the pipeline for all events
	// received by the Lambda.
	var pipelineInput = make(chan types.Transaction, 1)
	// pipelineOutput will be used to extract the final form of the event
	// before shipping it to the output.
	var pipelineOutput <-chan types.Transaction
	// outputInput is the entry point to the output composite.
	var outputInput = make(chan types.Transaction, 1)
	// manager manages statefull resources like caches, rate limits, and system
	// conditions that are shared across the entire runtime.
	var mgr *manager.Type

	var err error

	mgr, err = manager.New(conf.Manager, types.NoopMgr(), logger, stats)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %s", err.Error())
	}

	pipelineLayer, err = pipeline.New(
		conf.Pipeline, mgr,
		logger.NewModule(".pipeline"), metrics.Namespaced(stats, "pipeline"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %s", err.Error())
	}

	outputLayer, err = output.New(
		conf.Output, mgr,
		logger.NewModule(".output"), metrics.Namespaced(stats, "output"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create output: %s", err.Error())
	}

	if err := pipelineLayer.Consume(pipelineInput); err != nil {
		return nil, fmt.Errorf("failed to connect pipeline: %s", err.Error())
	}
	pipelineOutput = pipelineLayer.TransactionChan()
	if err := outputLayer.Consume(outputInput); err != nil {
		return nil, fmt.Errorf("failed to connect output: %s", err.Error())
	}
	closeFn := func() error {
		exitTimeout := time.Second * 30
		timesOut := time.Now().Add(exitTimeout)
		pipelineLayer.CloseAsync()
		outputLayer.CloseAsync()
		_ = outputLayer.WaitForClose(exitTimeout)
		_ = pipelineLayer.WaitForClose(time.Until(timesOut))
		mgr.CloseAsync()
		_ = mgr.WaitForClose(time.Until(timesOut))
		return nil
	}
	return &BenthosProducer{
		PipelineInput:  pipelineInput,
		PipelineOutput: pipelineOutput,
		OutputInput:    outputInput,
		CloseFn:        closeFn,
	}, nil
}

// BenthosProducer uses a set of Benthos transaction channels to coordinate
// processing and outputting an event.
type BenthosProducer struct {
	// PipelineInput will be sent the raw message received from the call to
	// Produce().
	PipelineInput chan<- types.Transaction
	// PipelineOutput will be read from if the PipelineInput transaction
	// results in a non-error Result. The value read from this channel will
	// be treated as the final version of the message to produce and may be
	// any number of parts.
	PipelineOutput <-chan types.Transaction
	// OutputInput will be sent the final version of message before the Lambda
	// returns.
	OutputInput chan<- types.Transaction
	// CloseFn will be called when the producer is closed. This is used to bind
	// shutdown behavior for any long lived resources used to power the producer.
	CloseFn func() error
}

// Close the producer. It is not valid to call Produce after calling Close.
func (p *BenthosProducer) Close() error {
	return p.CloseFn()
}

// Produce an event to one or more outputs. The return is the final version
// of the event produces after being processed or an error if something
// went wrong. The input may be any type that can be marshaled to JSON.
func (p *BenthosProducer) Produce(ctx context.Context, in interface{}) (interface{}, error) {
	// Convert the raw input into a Benthos message.
	msg := message.New(nil)
	part := message.NewPart(nil)
	if err := part.SetJSON(in); err != nil {
		return nil, err
	}
	msg.Append(part)

	// Run the message through the processor and check for errors.
	pipelineErrChan := make(chan types.Response, 1)
	select {
	case p.PipelineInput <- types.NewTransaction(msg, pipelineErrChan):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var pipelineResult types.Transaction
	select {
	case res := <-pipelineErrChan:
		if res.Error() != nil {
			return nil, res.Error()
		}
		return nil, errors.New("unexpected pipeline response")
	case <-ctx.Done():
		return nil, ctx.Err()

	// The pipeline only processes on input at a time which enables us to
	// read from this single pipeline output without worrying about getting
	// the wrong result due to concurrent processing. Other messages will be
	// enqueued until we send the ACK to the pipeline output transaction.
	case pipelineResult = <-p.PipelineOutput:
		go func() {
			pipelineResult.ResponseChan <- response.NewAck()
		}()
	}

	outputErrorChan := make(chan types.Response, 1)
	select {
	case p.OutputInput <- types.NewTransaction(pipelineResult.Payload, outputErrorChan):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case res := <-outputErrorChan:
		if res.Error() != nil {
			return nil, res.Error()
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	m := pipelineResult.Payload
	if m.Len() == 1 {
		jResult, err := m.Get(0).JSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json response: %s", err.Error())
		}
		return jResult, nil
	}

	var results []interface{}
	if err := m.Iter(func(i int, p types.Part) error {
		jResult, err := p.JSON()
		if err != nil {
			return fmt.Errorf("failed to marshal json response: %s", err.Error())
		}
		results = append(results, jResult)
		return nil
	}); err != nil {
		return nil, err
	}
	return results, nil
}
