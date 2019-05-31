package benthosx

import (
	"context"
	"sync"

	"github.com/Jeffail/benthos/lib/types"
)

type ctxKeyType struct{}

var ctxKey = &ctxKeyType{}

// NewResponseContext injects a ResponseMap for tracking serverless responses.
func NewResponseContext(ctx context.Context) context.Context {
	if v := ctx.Value(ctxKey); v != nil {
		// ResponseMap already installed. Keep the existing one.
		// This prevents components from accidentally replacing a
		// populated map with an empty one.
		return ctx
	}
	return context.WithValue(ctx, ctxKey, &ResponseMap{})
}

// ResponseFromContext fetches the ResponseMap. If one is not installed
// then an empty map is returned.
func ResponseFromContext(ctx context.Context) *ResponseMap {
	v := ctx.Value(ctxKey)
	if v == nil {
		return &ResponseMap{}
	}
	return v.(*ResponseMap)
}

// ResponseMap is a thread-safe map for storing serverless responses.
// Each key of the map points to a slice of messages.
type ResponseMap struct {
	m sync.Map
}

// Delete an entire key from the map.
func (r *ResponseMap) Delete(key string) {
	r.m.Delete(key)
}

// Load a response set. Returns false if the key is not found.
func (r *ResponseMap) Load(key string) ([]types.Message, bool) {
	v, ok := r.m.Load(key)
	if !ok {
		return nil, false
	}
	return v.([]types.Message), true
}

// Append a response to a key.
func (r *ResponseMap) Append(key string, response types.Message) {
	v, ok := r.m.Load(key)
	if !ok {
		v = make([]types.Message, 0, 1)
	}
	r.m.Store(key, append(v.([]types.Message), response))
}

// Range over the values in the map.
func (r *ResponseMap) Range(f func(key string, response []types.Message) bool) {
	r.m.Range(func(iKey interface{}, iValue interface{}) bool {
		return f(iKey.(string), iValue.([]types.Message))
	})
}

// Len returns the number of keys in the map.
func (r *ResponseMap) Len() int {
	var length int
	r.m.Range(func(_ interface{}, _ interface{}) bool {
		length = length + 1
		return true
	})
	return length
}
