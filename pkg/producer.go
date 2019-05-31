package benthosx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Jeffail/benthos/lib/config"
	"github.com/Jeffail/benthos/lib/log"
	"github.com/Jeffail/benthos/lib/manager"
	"github.com/Jeffail/benthos/lib/message"
	"github.com/Jeffail/benthos/lib/metrics"
	"github.com/Jeffail/benthos/lib/output"
	"github.com/Jeffail/benthos/lib/pipeline"
	"github.com/Jeffail/benthos/lib/types"
)

// NewProducer uses the given Benthos configuration to create a Producer
// instance that may be used as either a client in other code or as input
// for constructing a Lambda function.
func NewProducer(conf *config.Type) (*BenthosProducer, error) {
	// Logging and stats aggregation.
	logger := log.New(os.Stdout, conf.Logger)

	// Create our metrics type.
	stats, err := metrics.New(conf.Metrics, metrics.OptSetLogger(logger))
	for err != nil {
		logger.Errorf("Failed to connect metrics aggregator: %v\n", err)
		stats = metrics.Noop()
	}

	// piplineLayer represents the processing pipeline that a message will
	// pass through.
	var pipelineLayer pipeline.Type
	// outputLayer represents the composite of the various destination/outputs
	// defined by the configuration.
	var outputLayer output.Type
	// pipelineInput is the entry point to the pipeline for all events
	// received by the Lambda.
	var pipelineInput = make(chan types.Transaction, 1)
	// manager manages statefull resources like caches, rate limits, and system
	// conditions that are shared across the entire runtime.
	var mgr *manager.Type

	// For compatibility we map the default output option to serverless.
	if conf.Output.Type == output.TypeSTDOUT {
		conf.Output.Type = TypeServerless
	}

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
	if err := outputLayer.Consume(pipelineLayer.TransactionChan()); err != nil {
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
		PipelineInput: pipelineInput,
		CloseFn:       closeFn,
	}, nil
}

// BenthosProducer uses a set of Benthos transaction channels to coordinate
// processing and outputting an event.
type BenthosProducer struct {
	// PipelineInput will be sent the raw message received from the call to
	// Produce().
	PipelineInput chan<- types.Transaction
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
	var msg = message.New(nil)
	part := message.NewPart(nil)
	if err := part.SetJSON(in); err != nil {
		return nil, err
	}
	msgPart := message.WithContext(NewResponseContext(ctx), part)
	msg.Append(msgPart)

	resChan := make(chan types.Response, 1)
	select {
	case p.PipelineInput <- types.NewTransaction(msg, resChan):
	case <-ctx.Done():
		return nil, errors.New("request cancelled")
	}

	select {
	case res := <-resChan:
		if res.Error() != nil {
			return nil, res.Error()
		}
	case <-ctx.Done():
		return nil, errors.New("request cancelled")
	}

	r := ResponseFromContext(message.GetContext(msgPart))
	if r.Len() == 0 {
		// nothing set == no serverless outputs. use a canned response.
		return map[string]interface{}{
			"message": "request successful",
		}, nil
	}
	if r.Len() == 1 {
		var responseMsgs []types.Message
		r.Range(func(key string, response []types.Message) bool {
			responseMsgs = response
			return false
		})
		if len(responseMsgs) == 1 {
			if responseMsgs[0].Len() == 1 {
				// one key, one msg, one part == {}
				return responseMsgs[0].Get(0).JSON()
			}
			if responseMsgs[0].Len() > 1 {
				// one key, one msg, multi part == [{},...]
				results := make([]interface{}, 0, responseMsgs[0].Len())
				err := responseMsgs[0].Iter(func(_ int, p types.Part) error {
					jResult, err := p.JSON()
					results = append(results, jResult)
					return err
				})
				return results, err
			}
			// one key, one msg, zero parts == ???
			// TODO: Define this case.
			return nil, nil
		}
		// one key, multi message, any part == [[],[]]
		results := make([][]interface{}, 0, len(responseMsgs))
		for _, responseMsg := range responseMsgs {
			msgParts := make([]interface{}, 0, responseMsg.Len())
			if err := responseMsg.Iter(func(_ int, p types.Part) error {
				jResult, err := p.JSON()
				msgParts = append(msgParts, jResult)
				return err
			}); err != nil {
				return nil, err
			}
			results = append(results, msgParts)
		}
		return results, nil
	}
	// multi-key, any msg, any part == {"k": [[]]}
	results := make(map[string][][]interface{})
	var err error // captures any JSON errors in the range function
	r.Range(func(k string, msgs []types.Message) bool {
		kResults := make([][]interface{}, 0, len(msgs))
		for _, responseMsg := range msgs {
			msgParts := make([]interface{}, 0, responseMsg.Len())
			err = responseMsg.Iter(func(_ int, p types.Part) error {
				jResult, er := p.JSON()
				msgParts = append(msgParts, jResult)
				return er
			})
			if err != nil {
				return false
			}
			kResults = append(kResults, msgParts)
		}
		results[k] = kResults
		return true
	})
	return results, err
}
