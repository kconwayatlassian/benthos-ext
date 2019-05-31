package benthosx

import (
	"context"
	"errors"
	"testing"

	"github.com/Jeffail/benthos/lib/message"
	"github.com/Jeffail/benthos/lib/response"
	"github.com/Jeffail/benthos/lib/types"
	"github.com/stretchr/testify/require"
)

func TestProducerPipelineError(t *testing.T) {
	pipelineInput := make(chan types.Transaction, 1)
	ctx := context.Background()
	p := &BenthosProducer{
		CloseFn:       func() error { return nil },
		PipelineInput: pipelineInput,
	}
	event := map[string]interface{}{
		"test": "value",
	}
	go func() {
		in := <-pipelineInput
		in.ResponseChan <- response.NewError(errors.New("error"))
	}()
	_, err := p.Produce(ctx, event)
	require.Error(t, err)
}

func TestProducerPipelineCancelled(t *testing.T) {
	pipelineInput := make(chan types.Transaction, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := &BenthosProducer{
		CloseFn:       func() error { return nil },
		PipelineInput: pipelineInput,
	}
	event := map[string]interface{}{
		"test": "value",
	}

	_, err := p.Produce(ctx, event)
	require.Error(t, err)
}

func TestProducerPipelineResult(t *testing.T) {
	pipelineInput := make(chan types.Transaction, 1)
	testEvent := map[string]interface{}{
		"key": "value",
	}
	testCases := []struct {
		name     string
		m        *ResponseMap
		expected interface{}
	}{
		{
			name: "empty",
			m:    &ResponseMap{},
			expected: map[string]interface{}{
				"message": "request successful",
			},
		},
		{
			name: "one key, one msg, one part",
			m: func() *ResponseMap {
				m := &ResponseMap{}
				msg := message.New(nil)
				part := message.NewPart(nil)
				_ = part.SetJSON(testEvent)
				msg.Append(part)
				m.Append("", msg)
				return m
			}(),
			expected: testEvent,
		},
		{
			name: "one key, one msg, multi part",
			m: func() *ResponseMap {
				m := &ResponseMap{}
				msg := message.New(nil)
				part := message.NewPart(nil)
				_ = part.SetJSON(testEvent)
				msg.Append(part)
				msg.Append(part)
				m.Append("", msg)
				return m
			}(),
			expected: []interface{}{testEvent, testEvent},
		},
		{
			name: "one key, multi msg, any part",
			m: func() *ResponseMap {
				m := &ResponseMap{}
				msg := message.New(nil)
				part := message.NewPart(nil)
				_ = part.SetJSON(testEvent)
				msg.Append(part)
				m.Append("", msg)
				msg = message.New(nil)
				msg.Append(part)
				msg.Append(part)
				m.Append("", msg)
				return m
			}(),
			expected: [][]interface{}{
				[]interface{}{testEvent},
				[]interface{}{testEvent, testEvent},
			},
		},
		{
			name: "multi key, any msg, any part",
			m: func() *ResponseMap {
				m := &ResponseMap{}
				msg := message.New(nil)
				part := message.NewPart(nil)
				_ = part.SetJSON(testEvent)
				msg.Append(part)
				m.Append("a", msg)
				msg = message.New(nil)
				msg.Append(part)
				msg.Append(part)
				m.Append("b", msg)
				return m
			}(),
			expected: map[string][][]interface{}{
				"a": [][]interface{}{
					[]interface{}{testEvent},
				},
				"b": [][]interface{}{
					[]interface{}{testEvent, testEvent},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = context.WithValue(ctx, ctxKey, testCase.m)
			p := &BenthosProducer{
				CloseFn:       func() error { return nil },
				PipelineInput: pipelineInput,
			}
			event := map[string]interface{}{
				"test": "value",
			}
			go func() {
				tx := <-pipelineInput
				tx.ResponseChan <- response.NewAck()
			}()
			out, err := p.Produce(ctx, event)
			require.NoError(t, err)
			require.Equal(t, testCase.expected, out)
		})
	}

}
