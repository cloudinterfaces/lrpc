package main

import (
	"github.com/cloudinterfaces/lrpc/demo"
	"github.com/cloudinterfaces/lrpc/server"
)

func main() {
	server.Register(new(demo.Arith))
	server.Serve()
}
