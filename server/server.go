// Package server allows rpc/server implementations
// to be deployed to AWS Lambda.
/*
This package has a default *rpc.Server. Services
may be registered with the Register and RegisterName
functions, as with the net/rpc package. The package's
*rpc.Server may also be set with the RPCServer
function.

Invocations are handled via the builtin server's
ServeRequest method. No assumptions may be made
about the server's lifecycle or state beyond
the assumption a method will be invoked at least
once per container lifetime so it is possible
to do certain types of initialization once. Likewise
init functions and package-scoped variables
may be used with care and consideration.

The Lambda environment receives a payload of an
*rpc.Request and the request body. The response
is an *rpc.Response and the response body.
*/
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
	"runtime/debug"
	"strings"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/cloudinterfaces/lrpc/internal/mapping"
	"github.com/cloudinterfaces/lrpc/server/jscodec"
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
	ir  *messages.InvokeRequest
	req rpc.Request
	*gob.Decoder
	out *bytes.Buffer
}

func (c *codec) ReadRequestHeader(req *rpc.Request) error {
	*req = c.req
	return nil
}

func (c *codec) ReadRequestBody(i interface{}) error {
	if mapping.M != nil {
		mapping.Set(i, c.ir)
	}
	return c.Decode(i)
}

func (c *codec) WriteResponse(res *rpc.Response, i interface{}) error {
	c.out = new(bytes.Buffer)
	enc := gob.NewEncoder(c.out).Encode
	err := enc(c.ir.RequestId)
	if err != nil {
		return fmt.Errorf("error encoding request id: %v", err)
	}
	err = enc(res)
	if err != nil {
		return fmt.Errorf("error encoding rpc response: %v", err)
	}
	if len(res.Error) == 0 {
		err = enc(i)
		if err != nil {
			log.Println("ERROR", err, "while encoding:")
			log.Printf("%#v", i)
			return fmt.Errorf("error encoding rpc body: %v", err)
		}
	}
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

func jserr(req *messages.InvokeRequest, res *messages.InvokeResponse, err error) error {
	res.Error = &messages.InvokeResponse_Error{
		Type:    "error",
		Message: fmt.Sprintf("%s : %s", req.RequestId, err.Error()),
	}
	if strings.HasPrefix(err.Error(), "panic: ") {
		res.Error.Type = "panic"
	}
	return nil
}

func invokejson(req *messages.InvokeRequest, res *messages.InvokeResponse) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic: ", r)
			debug.PrintStack()
			jserr(req, res, fmt.Errorf("panic: %v", r))
		}
	}()
	r := jscodec.JSRequest{}
	err := json.Unmarshal(req.Payload, &r)
	if err != nil {
		return jserr(req, res, err)
	}
	if err = svr.ServeRequest(&r); err != nil {
		return jserr(req, res, err)
	}
	output := r.Output()
	if len(output) == 0 {
		return nil
	}
	res.Payload = output
	return nil
}

func ire(req *messages.InvokeRequest, res *messages.InvokeResponse, err error) error {
	res.Error = &messages.InvokeResponse_Error{
		Message: fmt.Sprintf("%s\n%s", err.Error(), req.RequestId),
		Type:    "error",
	}
	if strings.HasPrefix(err.Error(), "panic: ") {
		res.Error.Type = "panic"
	}
	return nil
}

func (server) Invoke(req *messages.InvokeRequest, res *messages.InvokeResponse) error {
	if j := bytes.TrimSpace(req.Payload); len(j) > 0 && j[0] == '{' {
		return invokejson(req, res)
	}
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic: ", r)
			debug.PrintStack()
			ire(req, res, fmt.Errorf("panic: %s", r))
		}
	}()
	var payload []byte
	err := json.Unmarshal(req.Payload, &payload)
	if err != nil {
		return ire(req, res, err)
	}
	dec := gob.NewDecoder(bytes.NewReader(payload))
	var r rpc.Request
	if err := dec.Decode(&r); err != nil {
		return ire(req, res, err)
	}
	c := &codec{req: r, Decoder: dec, ir: req}
	if mapping.M != nil {
		defer mapping.Delete(req)
	}
	if err = svr.ServeRequest(c); err != nil {
		return ire(req, res, err)
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
