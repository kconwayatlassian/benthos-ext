package benthosx

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Jeffail/benthos/lib/log"
	"github.com/Jeffail/benthos/lib/manager"
	"github.com/Jeffail/benthos/lib/metrics"
	"github.com/Jeffail/benthos/lib/types"
)

const (
	serverlessMetaKey = "serverlessID"
)

// ServerlessMessage is a Message wrapper that also contains a unique messageID.
type ServerlessMessage struct {
	types.Message
	key string
}

// GetKey returns the unique ID of the message
func (msg *ServerlessMessage) GetKey() string {
	return msg.key
}

// Copy the message and preserve tracking.
func (msg *ServerlessMessage) Copy() types.Message {
	return &ServerlessMessage{
		Message: msg.Message.Copy(),
		key:     msg.key,
	}
}

// DeepCopy the message and preserve tracking.
func (msg *ServerlessMessage) DeepCopy() types.Message {
	return &ServerlessMessage{
		Message: msg.Message.DeepCopy(),
		key:     msg.key,
	}
}

// Manager is an extension to the types.Manager from Benthos that adds some
// coordination functions for serverless use cases.
type Manager interface {
	types.Manager
	AddMessageID(msg types.Message) types.Message
	GetMessageResponse(key types.Message) (interface{}, bool)
	SetMessageResponse(key types.Message, response interface{}) error
}

// ServerlessManager extends the default Benthos Manager implementation
// to add a registry of outputs so functions can return the processed
// messages.
type ServerlessManager struct {
	*manager.Type
	serverlessRegistry *sync.Map
	nextServerlessID   *uint64
}

// NewServerlessManager returns an instance of manager.Type, which can be
// shared amongst components and logical threads of a Benthos service.
func NewServerlessManager(
	conf manager.Config,
	apiReg manager.APIReg,
	log log.Modular,
	stats metrics.Type,
) (*ServerlessManager, error) {
	mgr, err := manager.New(conf, apiReg, log, stats)
	if err != nil {
		return nil, err
	}
	return &ServerlessManager{
		Type:               mgr,
		serverlessRegistry: &sync.Map{},
		nextServerlessID:   new(uint64),
	}, nil
}

// AddMessageID converts any message into one trackable through a unique
// ID.
func (mgr *ServerlessManager) AddMessageID(msg types.Message) types.Message {
	newMsg := &ServerlessMessage{
		Message: msg,
		key:     fmt.Sprintf("%d", atomic.AddUint64(mgr.nextServerlessID, 1)),
	}
	_ = newMsg.Iter(func(_ int, p types.Part) error {
		p.Metadata().Set(serverlessMetaKey, newMsg.key)
		return nil
	})
	return newMsg
}

type keyGetter interface {
	GetKey() string
}

func keyFromMessage(msg types.Message) (string, bool) {
	if kg, ok := msg.(keyGetter); ok {
		return kg.GetKey(), true
	}
	if msg.Len() < 1 {
		return "", false
	}
	if k := msg.Get(0).Metadata().Get(serverlessMetaKey); k != "" {
		return k, true
	}
	return "", false
}

// GetMessageResponse fetches any recorded response for the given message.
// If there is no response recorded, or if the message contains no tracking
// data, then a default message is returned and the boolean value becomes false.
func (mgr *ServerlessManager) GetMessageResponse(key types.Message) (interface{}, bool) {
	k, _ := keyFromMessage(key)
	resp, ok := mgr.serverlessRegistry.Load(k)
	if !ok {
		resp = map[string]interface{}{
			"message": "request successful",
		}
	}
	mgr.serverlessRegistry.Delete(k)
	return resp, true
}

// SetMessageResponse records the outcome of processing for a given input
// if the input has preserved tracking information.
func (mgr *ServerlessManager) SetMessageResponse(key types.Message, response interface{}) error {
	k, ok := keyFromMessage(key)
	if !ok {
		return fmt.Errorf("given message has no tracking data")
	}
	mgr.serverlessRegistry.Store(k, response)
	return nil
}
