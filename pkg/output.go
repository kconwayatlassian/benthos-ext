package benthosx

import (
	"sync"
	"time"

	"github.com/Jeffail/benthos/lib/log"
	"github.com/Jeffail/benthos/lib/message"
	"github.com/Jeffail/benthos/lib/metrics"
	"github.com/Jeffail/benthos/lib/output"
	"github.com/Jeffail/benthos/lib/response"
	"github.com/Jeffail/benthos/lib/types"
)

const (
	// TypeServerless selects the serverless response option.
	TypeServerless = "serverless_response"
)

func init() {
	output.RegisterPlugin(
		"serverless_response",
		func() interface{} {
			return NewServerlessResponseConfig()
		},
		func(iconf interface{}, mgr types.Manager, logger log.Modular, stats metrics.Type) (types.Output, error) {
			if iconf == nil {
				iconf = NewServerlessResponseConfig()
			}
			return NewServerlessResponse(iconf.(ServerlessResponseConfig), mgr, logger, stats)
		},
	)

	output.DocumentPlugin(
		"serverless_response",
		`
This plugin enables serverless instances of Benthos to return the processed
message value from the function.`,
		nil, // No need to sanitise the config.
	)
}

// ServerlessResponseConfig contains configuration fields for the
// ServerlessResponse output.
type ServerlessResponseConfig struct {
	Name string
}

// NewServerlessResponseConfig returns a ServerlessResponseConfig with
// default values.
func NewServerlessResponseConfig() ServerlessResponseConfig {
	return ServerlessResponseConfig{}
}

// ServerlessResponse captures the final message value and writes it to a
// store where it can be retrieved by the serverless function.
type ServerlessResponse struct {
	transactionsChan <-chan types.Transaction

	mgr   types.Manager
	log   log.Modular
	stats metrics.Type

	closeOnce  sync.Once
	closeChan  chan struct{}
	closedChan chan struct{}

	name string
}

// NewServerlessResponse creates a new plugin output type.
func NewServerlessResponse(
	conf ServerlessResponseConfig,
	mgr types.Manager,
	log log.Modular,
	stats metrics.Type,
) (output.Type, error) {
	e := &ServerlessResponse{
		mgr:        mgr,
		log:        log,
		stats:      stats,
		name:       conf.Name,
		closeChan:  make(chan struct{}),
		closedChan: make(chan struct{}),
	}

	return e, nil
}

//------------------------------------------------------------------------------

func (e *ServerlessResponse) loop() {
	defer func() {
		close(e.closedChan)
	}()

	for {
		var tran types.Transaction
		var open bool
		select {
		case tran, open = <-e.transactionsChan:
			if !open {
				return
			}
		case <-e.closeChan:
			return
		}

		if tran.Payload.Len() > 0 {
			ResponseFromContext(message.GetContext(tran.Payload.Get(0))).Append(e.name, tran.Payload)
		}

		select {
		case tran.ResponseChan <- response.NewAck():
		case <-e.closeChan:
			return
		}
	}
}

// Connected returns true if this output is currently connected to its target.
func (e *ServerlessResponse) Connected() bool {
	return true // We're always connected
}

// Consume starts this output consuming from a transaction channel.
func (e *ServerlessResponse) Consume(tChan <-chan types.Transaction) error {
	e.transactionsChan = tChan
	go e.loop()
	return nil
}

// CloseAsync shuts down the output and stops processing requests.
func (e *ServerlessResponse) CloseAsync() {
	e.closeOnce.Do(func() {
		close(e.closeChan)
	})
}

// WaitForClose blocks until the output has closed down.
func (e *ServerlessResponse) WaitForClose(timeout time.Duration) error {
	select {
	case <-e.closedChan:
	case <-time.After(timeout):
		return types.ErrTimeout
	}
	return nil
}
