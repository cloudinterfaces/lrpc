package demo

import (
	"errors"
	"fmt"

	"github.com/cloudinterfaces/lrpc/inspect"
)

// Args is arguments.
type Args struct {
	A, B int
}

// Quotient is quotient.
type Quotient struct {
	Quo, Rem int
}

// Arith is a net/rpc server.
type Arith int

func (t *Arith) BadIdea(args *struct{}, reply *string) error {
	req := inspect.InvokeRequest(args)
	if req == nil {
		return fmt.Errorf("InvokeRequest appears to be nil")
	}
	*reply = req.RequestId
	return nil
}

func (t *Arith) Error(args *string, reply *string) error {
	return errors.New("this is an error")
}

func (t *Arith) Panic(args *string, reply *string) error {
	panic("this is a panic")
	return nil
}

// Multiply is the multiplication rpc method.
func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

// Divide is the division rpc method.
func (t *Arith) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}
