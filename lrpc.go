// Package lrpc contains clients and servers for net/rpc on AWS Lambda.
/*
The lrpc/server package is intended to replace net/rpc for types that
implement rpc.Server method sets. Main packages
that call lrpc/server.Serve may be used as Lambda handlers. The
https://github.com/cloudinterfaces/lago tool may be used
to deploy such main packages.

The lrpc/client package contains functions to call net/rpc
services deployed to Lambda functions with the lrpc/server
package.

The lrpc/demo package contains a net/rpc service and client wrapper,
as well as main packages that exercise both.
*/
package lrpc

// ID is a Lambda Request ID.
type ID string

// RequestID returns the request ID.
func (i ID) RequestID() string {
	return string(i)
}

// RequestID returns the request ID
// associated with err or *unknown*
// if it cannot be determined.
func RequestID(err error) string {
	if r, ok := err.(interface {
		RequestID() string
	}); ok {
		if id := r.RequestID(); len(id) > 0 {
			return id
		}
	}
	return "*unknown*"
}

// MethodError is an error returned by an rpc
// method invocation.
type MethodError struct {
	Err string
	ID
}

// Error implements error.
func (m MethodError) Error() string {
	return string(m.Err)
}

// ServerError is an error scoped to the server package.
type ServerError struct {
	Err string
	ID
}

// Error implements error.
func (e ServerError) Error() string {
	return string(e.Err)
}

// LambdaError is a structured error returned by
// a Lambda invocation.
type LambdaError struct {
	ErrorMessage string
	ErrorType    string
}

// Error implements Error.
func (l LambdaError) Error() string {
	return l.ErrorMessage
}
