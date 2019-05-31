package benthosx

import (
	"context"
	"testing"

	"github.com/Jeffail/benthos/lib/message"
	"github.com/Jeffail/benthos/lib/types"

	"github.com/stretchr/testify/require"
)

func TestNewResponseContext(t *testing.T) {
	ctx := NewResponseContext(context.Background())
	require.NotNil(t, ctx.Value(ctxKey))
	require.Equal(t, ctx.Value(ctxKey), NewResponseContext(ctx).Value(ctxKey))
}

func TestResponseFromContext(t *testing.T) {
	m := &ResponseMap{}
	ctx := context.WithValue(context.Background(), ctxKey, m)
	require.Equal(t, m, ResponseFromContext(ctx))
	require.NotNil(t, ResponseFromContext(context.Background()))
}

func TestResponseMap(t *testing.T) {
	m := &ResponseMap{}
	msg := message.New(nil)

	m.Append("", msg)
	require.Equal(t, 1, m.Len())
	foundMsg, found := m.Load("")
	require.True(t, found)
	require.Equal(t, msg, foundMsg[0])
	m.Range(func(_ string, response []types.Message) bool {
		require.Equal(t, msg, response[0])
		return true
	})
	m.Delete("")
	require.Equal(t, 0, m.Len())
}
