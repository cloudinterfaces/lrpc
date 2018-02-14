// Package inspect allows inspection of Lambda function invocations.
/*
Importing and using this function is not recommended; it implies
a close coupling of net/rpc methods and the Lambda environment
and imposes overhead on the function handler.
*/
package inspect

import (
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/cloudinterfaces/lrpc/internal/mapping"
)

func init() {
	mapping.M = make(mapping.Map)
}

// InvokeRequest returns the InvokeRequest message for
// args, which much be the arguments to a net/rpc
// method invocation; therefore this should only
// be called by a net/rpc method running
// in the Lambda environment.
func InvokeRequest(args interface{}) *messages.InvokeRequest {
	return mapping.Get(args)
}
