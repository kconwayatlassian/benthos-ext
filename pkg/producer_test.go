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
	pipelineOutput := make(chan types.Transaction, 1)
	outputInput := make(chan types.Transaction, 1)
	ctx := context.Background()
	p := &BenthosProducer{
		PipelineInput:  pipelineInput,
		PipelineOutput: pipelineOutput,
		OutputInput:    outputInput,
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

	go func() {
		in := <-pipelineInput
		in.ResponseChan <- response.NewAck()
	}()
	_, err = p.Produce(ctx, event)
	require.Error(t, err)
}

func TestProducerPipelineCancelled(t *testing.T) {
	pipelineInput := make(chan types.Transaction, 1)
	pipelineOutput := make(chan types.Transaction, 1)
	outputInput := make(chan types.Transaction, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := &BenthosProducer{
		PipelineInput:  pipelineInput,
		PipelineOutput: pipelineOutput,
		OutputInput:    outputInput,
	}
	event := map[string]interface{}{
		"test": "value",
	}

	_, err := p.Produce(ctx, event)
	require.Error(t, err)
}

func TestProducerPipelineSingleResult(t *testing.T) {
	pipelineInput := make(chan types.Transaction, 1)
	pipelineOutput := make(chan types.Transaction, 1)
	outputInput := make(chan types.Transaction, 1)
	ctx := context.Background()
	p := &BenthosProducer{
		PipelineInput:  pipelineInput,
		PipelineOutput: pipelineOutput,
		OutputInput:    outputInput,
	}
	event := map[string]interface{}{
		"test": "value",
	}
	result := message.New(nil)
	part := message.NewPart(nil)
	_ = part.SetJSON(map[string]interface{}{
		"part": "one",
	})
	result.Append(part)

	go func() {
		<-pipelineInput
		ackChan := make(chan types.Response, 1)
		tResult := types.NewTransaction(result, ackChan)
		pipelineOutput <- tResult
		out := <-outputInput
		out.ResponseChan <- response.NewAck()
	}()
	out, err := p.Produce(ctx, event)
	require.NoError(t, err)
	expectedOut, _ := part.JSON()
	require.Equal(t, expectedOut, out)
}

func TestProducerPipelineMultiResult(t *testing.T) {
	pipelineInput := make(chan types.Transaction, 1)
	pipelineOutput := make(chan types.Transaction, 1)
	outputInput := make(chan types.Transaction, 1)
	ctx := context.Background()
	p := &BenthosProducer{
		PipelineInput:  pipelineInput,
		PipelineOutput: pipelineOutput,
		OutputInput:    outputInput,
	}
	event := map[string]interface{}{
		"test": "value",
	}
	result := message.New(nil)
	part := message.NewPart(nil)
	_ = part.SetJSON(map[string]interface{}{
		"part": "one",
	})
	result.Append(part)
	part2 := message.NewPart(nil)
	_ = part2.SetJSON(map[string]interface{}{
		"part": "two",
	})
	result.Append(part2)

	go func() {
		<-pipelineInput
		ackChan := make(chan types.Response, 1)
		tResult := types.NewTransaction(result, ackChan)
		pipelineOutput <- tResult
		out := <-outputInput
		out.ResponseChan <- response.NewAck()
	}()
	out, err := p.Produce(ctx, event)
	require.NoError(t, err)
	expectedOut := make([]interface{}, 0, 2)
	_ = result.Iter(func(_ int, p types.Part) error {
		partOut, _ := p.JSON()
		expectedOut = append(expectedOut, partOut)
		return nil
	})
	require.Equal(t, expectedOut, out)
}
