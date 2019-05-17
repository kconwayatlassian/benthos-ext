package benthosx

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/Jeffail/benthos/lib/response"
	"github.com/Jeffail/benthos/lib/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestProducerPipelineError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewMockManager(ctrl)
	pipelineInput := make(chan types.Transaction, 1)
	pipelineOutput := make(chan types.Transaction, 1)
	outputInput := make(chan types.Transaction, 1)
	ctx := context.Background()
	p := &BenthosProducer{
		Manager:        mgr,
		PipelineInput:  pipelineInput,
		PipelineOutput: pipelineOutput,
		OutputInput:    outputInput,
	}
	event := map[string]interface{}{
		"test": "value",
	}

	mgr.EXPECT().AddMessageID(gomock.Any()).DoAndReturn(func(m types.Message) types.Message {
		return &ServerlessMessage{
			Message: m,
			key:     fmt.Sprintf("%d", rand.Uint64()),
		}
	})
	go func() {
		in := <-pipelineInput
		in.ResponseChan <- response.NewError(errors.New("error"))
	}()
	_, err := p.Produce(ctx, event)
	require.Error(t, err)

	mgr.EXPECT().AddMessageID(gomock.Any()).DoAndReturn(func(m types.Message) types.Message {
		return &ServerlessMessage{
			Message: m,
			key:     fmt.Sprintf("%d", rand.Uint64()),
		}
	})
	go func() {
		in := <-pipelineInput
		in.ResponseChan <- response.NewAck()
	}()
	_, err = p.Produce(ctx, event)
	require.Error(t, err)
}

func TestProducerPipelineCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewMockManager(ctrl)
	pipelineInput := make(chan types.Transaction, 1)
	pipelineOutput := make(chan types.Transaction, 1)
	outputInput := make(chan types.Transaction, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := &BenthosProducer{
		Manager:        mgr,
		PipelineInput:  pipelineInput,
		PipelineOutput: pipelineOutput,
		OutputInput:    outputInput,
	}
	event := map[string]interface{}{
		"test": "value",
	}
	mgr.EXPECT().AddMessageID(gomock.Any()).DoAndReturn(func(m types.Message) types.Message {
		return &ServerlessMessage{
			Message: m,
			key:     fmt.Sprintf("%d", rand.Uint64()),
		}
	})

	_, err := p.Produce(ctx, event)
	require.Error(t, err)
}

func TestProducerPipelineResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mgr := NewMockManager(ctrl)
	pipelineInput := make(chan types.Transaction, 1)
	pipelineOutput := make(chan types.Transaction, 1)
	outputInput := make(chan types.Transaction, 1)
	ctx := context.Background()
	p := &BenthosProducer{
		Manager:        mgr,
		PipelineInput:  pipelineInput,
		PipelineOutput: pipelineOutput,
		OutputInput:    outputInput,
	}
	event := map[string]interface{}{
		"test": "value",
	}
	result := map[string]interface{}{
		"test2": "value2",
	}
	messageID := fmt.Sprintf("%d", rand.Uint64())
	mgr.EXPECT().AddMessageID(gomock.Any()).DoAndReturn(func(m types.Message) types.Message {
		return &ServerlessMessage{
			Message: m,
			key:     messageID,
		}
	})
	mgr.EXPECT().GetMessageResponse(gomock.Any()).DoAndReturn(func(m types.Message) (interface{}, bool) {
		smsg, ok := m.(*ServerlessMessage)
		require.True(t, ok)
		require.Equal(t, messageID, smsg.key)
		return result, true
	})

	go func() {
		<-pipelineInput
		ackChan := make(chan types.Response, 1)
		tResult := types.NewTransaction(nil, ackChan)
		pipelineOutput <- tResult
		out := <-outputInput
		out.ResponseChan <- response.NewAck()
	}()
	out, err := p.Produce(ctx, event)
	require.NoError(t, err)
	require.Equal(t, result, out)
}
