// Package client provides a net/rpc client for AWS Lambda servers.
/*
TODO: documentation
*/
package client

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net/rpc"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

func defaultregion() *string {
	var defaultregion = "us-east-1"
	if s := os.Getenv("AWS_REGION"); len(s) > 0 {
		return &s
	}
	return &defaultregion
}

// FunctionError represents an error returned
// by a net/rpc invocation (the error returned
// by an invoked method).
type FunctionError string

// Error implements error.
func (f FunctionError) Error() string {
	return string(f)
}

// IsFunctionError returns true if
// err returned by Call is error
// returned by net/rpc method, false
// if serialization, network, AWS
// or other error.
func IsFunctionError(err error) bool {
	switch err.(type) {
	case FunctionError:
		return true
	}
	return false
}

// Interface is the client Interface.
type Interface interface {
	// Call dispatches a net/rpc call to serviceMethod via the AWS
	// REST API.
	Call(serviceMethod string, args interface{}, reply interface{}) error
}

type client struct {
	name string
	svc  *lambda.Lambda
}

func (c *client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	req := rpc.Request{ServiceMethod: serviceMethod, Seq: 1}
	buf := new(bytes.Buffer)
	encode := gob.NewEncoder(buf).Encode
	err := encode(req)
	if err != nil {
		return err
	}
	if err = encode(args); err != nil {
		return err
	}
	input := &lambda.InvokeInput{
		FunctionName: &c.name,
	}
	input.Payload, err = json.Marshal(buf.Bytes())
	if err != nil {
		return err
	}
	output, err := c.svc.Invoke(input)
	if err != nil {
		return err
	}
	if output.FunctionError != nil {
		err = FunctionError(*output.FunctionError)
	}
	if len(output.Payload) > 0 {
		decode := gob.NewDecoder(bytes.NewReader(output.Payload)).Decode
		e := decode(reply)
		if err != nil {
			return err
		}
		return e
	}
	return nil
}

// Default returns a client.Interface
// using the region set in AWS_REGION
// or "us-east-1" if not set. An
// error is returned if funcName
// cannot be verified as a Lambda
// function in that region.
func Default(funcName string) (Interface, error) {
	var sess = session.New(&aws.Config{
		Region: defaultregion(),
	})
	svc := lambda.New(sess)
	return New(svc, funcName)
}

// New returns a client.Interface using
// svc. If funcName cannot be verified to
// exist, an error is returned.
func New(svc *lambda.Lambda, funcName string) (Interface, error) {
	req := &lambda.GetFunctionInput{FunctionName: &funcName}
	_, err := svc.GetFunction(req)
	if err != nil {
		return nil, err
	}
	return &client{name: funcName, svc: svc}, nil
}
