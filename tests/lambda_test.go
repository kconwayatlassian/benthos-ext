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
			for _, st := range res.Error.StackTrace {
				t.Log(st.Path, st.Line, st.Label)
			}
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
	conf.Logger.LogLevel = "OFF"
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
	conf.Logger.LogLevel = "OFF"
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
	conf.Logger.LogLevel = "OFF"
	conf.Output.Type = output.TypeBroker
	conf.Output.Broker.Outputs = append(
		conf.Output.Broker.Outputs,
		output.NewConfig(),
		output.NewConfig(),
	)
	conf.Output.Broker.Outputs[1].Type = output.TypeDrop
	conf.Output.Broker.Outputs[0].Type = benthosx.TypeServerless
	payload := []byte(`{"test":"value"}`)

	h, closeFn, err := benthosx.NewLambdaHandler(&conf)
	require.NoError(t, err)
	defer closeFn()
	go StartHandler(stop, port, h)
	out, err := makeRPCCall(t, port, payload)
	require.NoError(t, err)
	require.Equal(t, payload, out)
}

var yamlConf = `http:
  address: 0.0.0.0:4195
  read_timeout: 5s
  root_path: /benthos
  debug_endpoints: false
pipeline:
  processors: []
  threads: 1
output:
  type: serverless_response
  serverless_response:
    name: lambda
logger:
  prefix: benthos
  level: OFF
  add_timestamp: true
  json_format: true
  static_fields:
    '@service': benthos
metrics:
  type: http_server
  http_server: {}
  prefix: benthos
tracer:
  type: none
  none: {}
shutdown_timeout: 20s
`

func TestLambdaYAMLConfig(t *testing.T) {
	port, err := getPort()
	require.NoError(t, err)
	stop := make(chan interface{})
	defer close(stop)
	conf, err := benthosx.NewConfig([]byte(yamlConf))
	require.Nil(t, err)
	payload := []byte(`{"test":"value"}`)

	h, closeFn, err := benthosx.NewLambdaHandler(conf)
	require.NoError(t, err)
	defer closeFn()
	go StartHandler(stop, port, h)
	out, err := makeRPCCall(t, port, payload)
	require.NoError(t, err)
	require.Equal(t, payload, out)
}
