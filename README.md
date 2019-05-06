<a id="markdown-benthos-ext---benthos-extensions" name="benthos-ext---benthos-extensions"></a>
# benthos-ext - Benthos extensions

<https://github.com/asecurityteam/benthos-ext>

<!-- TOC -->

- [benthos-ext - Benthos extensions](#benthos-ext---benthos-extensions)
    - [Overview](#overview)
    - [Using The Lambda Function](#using-the-lambda-function)
    - [Using The Code Client](#using-the-code-client)
    - [Status](#status)
    - [Contributing](#contributing)
        - [Building And Testing](#building-and-testing)
        - [Quality Gates](#quality-gates)
        - [License](#license)
        - [Contributing Agreement](#contributing-agreement)

<!-- /TOC -->

<a id="markdown-overview" name="overview"></a>
## Overview

This project contains a set of [Benthos](https://github.com/Jeffail/benthos)
extensions that add or modify some behaviors. Notably, this project contains
an alternative version of
[Benthos Lambda](https://github.com/Jeffail/benthos/blob/master/docs/serverless/lambda.md)
that supports producing events to a stream _and_ returning the processed message
body at the same time. The mainline build offers this functionality as a switch
where the function either returns the processed body _or_ writes it to a stream.

Additionally, this project provides a trimmed down code interface for producing
events into a Benthos configuration. This is to help with specialized cases where
restructuring code as a
[custom Benthos Processor](https://godoc.org/github.com/Jeffail/benthos/lib/stream)
is either too large of an effort or undesirable for one reason or another. This
code component enables consumers to use the same Benthos configuration as would
be used with a Lambda (processor + output) and create a client.

<a id="markdown-using-the-lambda-function" name="using-the-lambda-function"></a>
## Using The Lambda Function

The provided `main.go` is a drop-in replacement for the official Benthos Lambda
binary that supports the additional features. The only difference is that any
use of the `STDOUT` output type triggers the function to return the processed
message body. For example, the output may be defined as a broker with a Kinesis
and STDOUT output. This would result in a message being produced to Kinesis and
the final message produced being returned by the Lambda.

```bash
LAMBDA_ENV=`cat yourconfig.yaml | jq -csR {Variables:{BENTHOS_CONFIG:.}}`
aws lambda create-function \
  --runtime go1.x \
  --handler benthos-lambda \
  --role benthos-example-role \
  --zip-file fileb://benthos-lambda.zip \
  --environment "$LAMBDA_ENV" \
  --function-name benthos-example

aws lambda invoke \
  --function-name benthos-example \
  --payload '{"your":"document"}' \
  out.txt && cat out.txt && rm out.txt
```

<a id="markdown-using-the-code-client" name="using-the-code-client"></a>
## Using The Code Client

The code client is constructed using the same configuration you would provide
for the Lambda function. Any input and buffer sections are ignored.

```golang
configBytes := []byte(os.Getenv("BENTHOS_CONFIG")) // or any other way to load the config
conf, err := benthosx.NewConfig(configBytes)
if err != nil {
    panic(err.Error())
}
p, err := benthosx.NewProducer(conf)
if err != nil {
    panic(err.Error())
}
defer p.Close()
```

The producer interface is a generic:

```golang
type Producer interface {
    Produce(ctx context.Context, input interface{}) (interface{}, error)
    Close() error
}
```

The producer will operate on any data type that can be marshaled to JSON and
will return the final form of the message it sent to any outputs. Note that
the output may actually be a slice of items if any of the processors converted
the message into multiple parts.

```golang
type MyEvent struct {
    String string `json:"string"`
    Integer int `json:"int"`
    Bool bool `json:"bool"`
    Time time.Time `json:"time"`
}
result, err := p.Produce(MyEvent{})
```

<a id="markdown-status" name="status"></a>
## Status

This project is in incubation which means we are not yet operating this tool in production
and the interfaces are subject to change.

<a id="markdown-contributing" name="contributing"></a>
## Contributing

<a id="markdown-building-and-testing" name="building-and-testing"></a>
### Building And Testing

We publish a docker image called [SDCLI](https://github.com/asecurityteam/sdcli) that
bundles all of our build dependencies. It is used by the included Makefile to help make
building and testing a bit easier. The following actions are available through the Makefile:

-   make dep

    Install the project dependencies into a vendor directory

-   make lint

    Run our static analysis suite

-   make test

    Run unit tests and generate a coverage artifact

-   make integration

    Run integration tests and generate a coverage artifact

-   make coverage

    Report the combined coverage for unit and integration tests

-   make build

    Generate a local build of the project (if applicable)

-   make run

    Run a local instance of the project (if applicable)

-   make doc

    Generate the project code documentation and make it viewable
    locally.

<a id="markdown-quality-gates" name="quality-gates"></a>
### Quality Gates

Our build process will run the following checks before going green:

-   make lint
-   make test
-   make integration
-   make coverage (combined result must be 85% or above for the project)

Running these locally, will give early indicators of pass/fail.

<a id="markdown-license" name="license"></a>
### License

This project is licensed under Apache 2.0. See LICENSE.txt for details.

<a id="markdown-contributing-agreement" name="contributing-agreement"></a>
### Contributing Agreement

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
