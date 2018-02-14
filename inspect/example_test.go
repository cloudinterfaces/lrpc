package inspect_test // main

import (
	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/cloudinterfaces/lrpc/inspect"
	"github.com/cloudinterfaces/lrpc/server"
)

type BadIdea struct{}

func (BadIdea) Inspect(args *struct{}, reply *messages.InvokeRequest) error {
	reply = inspect.InvokeRequest(args)
	return nil
}

func Example() { // main()
	server.Register(new(BadIdea))
	server.Serve()
}
