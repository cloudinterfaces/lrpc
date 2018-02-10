package server

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"

	"github.com/aws/aws-lambda-go/lambda/messages"
)

var svr = rpc.NewServer()

// Register registers i with this
// package's default server.
func Register(i interface{}) error {
	return svr.Register(i)
}

// RegisterName registers i with name
// with this package's default server.
func RegisterName(name string, i interface{}) error {
	return svr.RegisterName(name, i)
}

// RPCServer sets this package's default
// server to s.
func RPCServer(s *rpc.Server) {
	if s == nil {
		panic("lrpc/server: s must not be nil")
	}
	svr = s
}

type server struct{}

type codec struct {
	req rpc.Request
	*gob.Decoder
	out   *bytes.Buffer
	Error string
	err   error
}

func (c *codec) encode(i interface{}) {
	c.out = new(bytes.Buffer)
	enc := gob.NewEncoder(c.out).Encode
	c.err = enc(i)
}

func (c *codec) ReadRequestHeader(req *rpc.Request) error {
	*req = c.req
	return nil
}

func (c *codec) ReadRequestBody(i interface{}) error {
	return c.Decode(i)
}

func (c *codec) WriteResponse(res *rpc.Response, i interface{}) error {
	if len(res.Error) > 0 {
		c.Error = res.Error
		return nil
	}
	c.encode(i)
	return nil
}

func (c *codec) Close() error {
	return nil
}

// Ping is the Lambda keepalive.
func (server) Ping(req *messages.PingRequest, response *messages.PingResponse) error {
	*response = messages.PingResponse{}
	return nil
}

func (server) Invoke(req *messages.InvokeRequest, res *messages.InvokeResponse) error {
	var payload []byte
	err := json.Unmarshal(req.Payload, &payload)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewReader(payload))
	var r rpc.Request
	if err := dec.Decode(&r); err != nil {
		return err
	}
	c := &codec{req: r, Decoder: dec}
	if err = svr.ServeRequest(c); err != nil {
		return err
	}
	if len(c.Error) > 0 {
		return fmt.Errorf(c.Error)
	}
	if c.err != nil {
		return c.err
	}
	res.Payload = c.out.Bytes()
	return nil
}

// Serve begins serving this package's
// default server as set with RPCServer
// or configured with the Register
// functions. In a non-Lambda environment,
// starts a net/rpc server on a random
// port.
func Serve() {
	port := os.Getenv("_LAMBDA_SERVER_PORT")
	if len(port) == 0 {
		port = "0"
	}
	l, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal(err)
	}
	s := rpc.NewServer()
	err = s.RegisterName("Function", &server{})
	if err != nil {
		log.Fatal("failed to register handler function")
	}
	if port == "0" {
		log.Printf("Starting test server on %s", l.Addr().String())
		svr.Accept(l)
		return
	}
	s.Accept(l)
	log.Fatal("accept should not have returned")
}
