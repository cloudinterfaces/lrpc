package main

import (
	"errors"

	"github.com/cloudinterfaces/lrpc/server"
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

func main() {
	server.Register(new(Arith))
	server.Serve()
}
