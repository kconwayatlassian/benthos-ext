package benthosx

import (
	"context"
	"errors"
	"testing"

	"github.com/Jeffail/benthos/lib/config"
	"github.com/Jeffail/benthos/lib/output"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestReplaceSTDOUT(t *testing.T) {
	conf := config.New()
	newConf, found := replaceStdout(&conf)
	require.Equal(t, newConf.Output.Type, output.TypeDrop)
	require.True(t, found)
}

func TestReplaceSTDOUTBroker(t *testing.T) {
	conf := config.New()
	conf.Output.Type = output.TypeBroker
	conf.Output.Broker.Outputs = append(
		conf.Output.Broker.Outputs,
		output.NewConfig(), // Default output with STDOUT set.
	)
	nestedBroker := output.NewConfig()
	nestedBroker.Type = output.TypeBroker
	nestedBroker.Broker.Outputs = append(
		nestedBroker.Broker.Outputs,
		output.NewConfig(), // Default output with STDOUT set.
	)
	conf.Output.Broker.Outputs = append(
		conf.Output.Broker.Outputs,
		nestedBroker,
	)
	newConf, found := replaceStdout(&conf)
	require.True(t, found)
	require.Equal(t, newConf.Output.Type, output.TypeBroker)
	require.Len(t, newConf.Output.Broker.Outputs, 2)
	require.Equal(t, newConf.Output.Broker.Outputs[0].Type, output.TypeDrop)
	require.Equal(t, newConf.Output.Broker.Outputs[1].Broker.Outputs[0].Type, output.TypeDrop)
}

func TestLambdaFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := NewMockProducer(ctrl)
	f := &LambdaFunc{
		Producer: p,
		FormatFn: FormatIdentity,
	}

	ctx := context.Background()
	event := map[string]interface{}{
		"test": "value",
	}
	finalEvent := map[string]interface{}{
		"modified": "value",
	}

	p.EXPECT().Produce(gomock.Any(), event).Return(nil, errors.New("error"))
	_, err := f.HandleEvent(ctx, event)
	require.Error(t, err)

	p.EXPECT().Produce(gomock.Any(), event).Return(finalEvent, nil)
	out, err := f.HandleEvent(ctx, event)
	require.NoError(t, err)
	require.Equal(t, finalEvent, out)

	f.FormatFn = FormatStatic
	staticEvent := map[string]interface{}{
		"message": "request successful",
	}
	p.EXPECT().Produce(gomock.Any(), event).Return(finalEvent, nil)
	out, err = f.HandleEvent(ctx, event)
	require.NoError(t, err)
	require.Equal(t, staticEvent, out)
}
