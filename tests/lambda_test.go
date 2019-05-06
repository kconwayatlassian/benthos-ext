// +build integration

package tests

import (
	"errors"
	"net/rpc"
	"testing"
	"time"

	"github.com/Jeffail/benthos/lib/config"
	"github.com/Jeffail/benthos/lib/output"
	benthosx "github.com/asecurityteam/benthos-ext/pkg"
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/stretchr/testify/require"
)

// makeRPCCall imitates the internal execution path for the native lambda
// system by using the net/rpc module.
func makeRPCCall(t *testing.T, port string, payload []byte) ([]byte, error) {
	// Ping the server until it is available or until we exceed a timeout
	// value. This is to account for arbitrary start-up time of the server
	// in the background.
	stop := time.Now().Add(5 * time.Second)
	for time.Now().Before(stop) {
		time.Sleep(100 * time.Millisecond)
		client, err := rpc.Dial("tcp", "localhost:"+port)
		if err != nil {
			t.Log(err.Error())
			continue
		}
		req := &messages.InvokeRequest{
			Payload: payload,
			Deadline: messages.InvokeRequest_Timestamp{
				Seconds: time.Now().Add(10 * time.Second).Unix(),
				Nanos:   0,
			},
		}
		res := &messages.InvokeResponse{}
		if err := client.Call("Function.Invoke", req, res); err != nil {
			t.Log(err.Error())
			continue
		}
		if res.Error != nil {
			t.Log(res.Error.Message)
			continue
		}
		return res.Payload, nil
	}
	return nil, errors.New("failed to execute function")
}

func TestLambdaDefault(t *testing.T) {
	port, err := getPort()
	require.NoError(t, err)
	stop := make(chan interface{})
	defer close(stop)
	conf := config.New()
	payload := []byte(`{"test":"value"}`)

	h, closeFn, err := benthosx.NewLambdaHandler(&conf)
	require.NoError(t, err)
	defer closeFn()
	go StartHandler(stop, port, h)
	out, err := makeRPCCall(t, port, payload)
	require.NoError(t, err)
	require.Equal(t, payload, out)
}

func TestLambdaNoOutput(t *testing.T) {
	port, err := getPort()
	require.NoError(t, err)
	stop := make(chan interface{})
	defer close(stop)
	conf := config.New()
	conf.Output.Type = output.TypeDrop
	payload := []byte(`{"test":"value"}`)

	h, closeFn, err := benthosx.NewLambdaHandler(&conf)
	require.NoError(t, err)
	defer closeFn()
	go StartHandler(stop, port, h)
	out, err := makeRPCCall(t, port, payload)
	require.NoError(t, err)
	require.Equal(t, []byte(`{"message":"request successful"}`), out)
}

func TestLambdaBroker(t *testing.T) {
	port, err := getPort()
	require.NoError(t, err)
	stop := make(chan interface{})
	defer close(stop)
	conf := config.New()
	conf.Output.Type = output.TypeBroker
	conf.Output.Broker.Outputs = append(
		conf.Output.Broker.Outputs,
		output.NewConfig(),
		output.NewConfig(),
	)
	conf.Output.Broker.Outputs[1].Type = output.TypeDrop
	payload := []byte(`{"test":"value"}`)

	h, closeFn, err := benthosx.NewLambdaHandler(&conf)
	require.NoError(t, err)
	defer closeFn()
	go StartHandler(stop, port, h)
	out, err := makeRPCCall(t, port, payload)
	require.NoError(t, err)
	require.Equal(t, payload, out)
}
