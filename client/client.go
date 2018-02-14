// Package client provides a net/rpc client for AWS Lambda servers.
/*
This package is used to call net/rpc servers depolyed to Lambda
with the lrpc/server package. It uses the AWS REST API as the transport.
A two-stage codec is used for call arguments: they are encoded
to a buffer with gob, then the buffer is marshalled to JSON.
This prevents poor type conversions.

In some cases it is nessisary to determine if an error was returned
by the rpc method, the server package, the client, or Lambda itself. Functions
are provided to facilitate this.
*/
package client

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net/rpc"
	"os"
	"strings"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/cloudinterfaces/lrpc"
)

func defaultregion() *string {
	var defaultregion = "us-east-1"
	if s := os.Getenv("AWS_REGION"); len(s) > 0 {
		return &s
	}
	return &defaultregion
}

// Error is an error
// associated with the client.
type Error string

// Error implements error.
func (c Error) Error() string {
	return string(c)
}

// IsClientError returns true
// if err is a ClientError.
// This indicates a problem with
// the call, Lambda itself
// or a bug in the client.
func IsClientError(err error) bool {
	switch err.(type) {
	case Error:
		return true
	}
	return false
}

// IsMethodErr returns true if
// err was returned by an rpc method.
// This is expected to be true for
// most errors returned.
func IsMethodErr(err error) bool {
	switch err.(type) {
	case lrpc.MethodError:
		return true
	}
	return false
}

// IsServerPanic returns true if
// err represents a recovered server panic.
// This usually indicates a defect with
// the net/rpc server.
func IsServerPanic(err error) bool {
	if e, ok := err.(lrpc.LambdaError); ok {
		return e.ErrorType == "panic"
	}
	return false
}

// IsServerError returns true if
// err represents a server error.
// This usually indicates a defect
// with a net/rpc call.
func IsServerError(err error) bool {
	if e, ok := err.(lrpc.LambdaError); ok {
		return e.ErrorType == "error"
	}
	return false
}

func wrap(err error) error {
	if err == nil {
		return nil
	}
	return Error(err.Error())
}

// Interface is the client Interface.
type Interface interface {
	// Call invokes Invoke, discarding the request ID.
	Call(serviceMethod string, args interface{}, reply interface{}) error
	// Invoke dispatches a net/rpc call to serviceMethod via the AWS
	// REST API.
	Invoke(serviceMethod string, args interface{}, reply interface{}) (requestid *string, err error)
}

type client struct {
	name    string
	qual    *string
	svc     *lambda.Lambda
	counter *uint64
}

func (c *client) Lambda() *lambda.Lambda {
	return c.svc
}

func (c *client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	_, err := c.Invoke(serviceMethod, args, reply)
	return err
}

func (c *client) Invoke(serviceMethod string, args interface{}, reply interface{}) (*string, error) {
	req := rpc.Request{ServiceMethod: serviceMethod, Seq: atomic.AddUint64(c.counter, 1)}
	buf := new(bytes.Buffer)
	encode := gob.NewEncoder(buf).Encode
	err := encode(req)
	if err != nil {
		return nil, wrap(err)
	}
	if err = encode(args); err != nil {
		return nil, wrap(err)
	}
	input := &lambda.InvokeInput{
		FunctionName: &c.name,
		Qualifier:    c.qual,
	}
	input.Payload, err = json.Marshal(buf.Bytes())
	if err != nil {
		return nil, wrap(err)
	}
	output, err := c.svc.Invoke(input)
	if err != nil {
		return nil, wrap(err)
	}
	if output.FunctionError != nil {
		e := lrpc.LambdaError{}
		err = json.Unmarshal(output.Payload, &e)
		if err != nil {
			return nil, wrap(err)
		}
		return nil, e
	}
	if len(output.Payload) > 0 {
		decode := gob.NewDecoder(bytes.NewReader(output.Payload)).Decode
		var requestid string
		if err = decode(&requestid); err != nil {
			return nil, wrap(err)
		}
		res := rpc.Response{}
		if err = decode(&res); err != nil {
			return nil, wrap(err)
		}
		if len(res.Error) == 0 {
			if err = decode(reply); err != nil {
				return nil, wrap(err)
			}
		}
		if len(res.Error) > 0 {
			return nil, lrpc.MethodError{
				Err: res.Error,
				ID:  lrpc.ID(requestid),
			}
		}
		return &requestid, nil
	}
	return nil, Error("No response payload")
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
	var qualifier *string
	parts := strings.SplitN(funcName, ":", 2)
	if len(parts) == 2 {
		funcName = parts[0]
		qualifier = &parts[1]
	}
	req := &lambda.GetFunctionInput{FunctionName: &funcName}
	if qualifier != nil {
		req.Qualifier = qualifier
	}
	_, err := svc.GetFunction(req)
	if err != nil {
		return nil, err
	}
	return &client{name: funcName, qual: qualifier, svc: svc, counter: new(uint64)}, nil
}
