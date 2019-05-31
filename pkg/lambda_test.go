package benthosx

import (
	"context"
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestLambdaFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	p := NewMockProducer(ctrl)
	f := &LambdaFunc{
		Producer: p,
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
}
