// Package main builds a Lambda function for Benthos.
package main

import (
	"os"

	benthosx "github.com/asecurityteam/benthos-ext/pkg"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	b := []byte(os.Getenv("BENTHOS_CONFIG"))
	conf, err := benthosx.NewConfig(b)
	if err != nil {
		panic(err.Error())
	}
	handler, close, err := benthosx.NewLambdaHandler(conf)
	if err != nil {
		panic(err.Error())
	}
	defer close()
	lambda.StartHandler(handler)
}
