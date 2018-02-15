package client

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/rpc"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var tail = new(string)

func init() {
	*tail = `Tail`
}

func defaultregion() *string {
	var defaultregion = "us-east-1"
	if s := os.Getenv("AWS_REGION"); len(s) > 0 {
		return &s
	}
	return &defaultregion
}

type FunctionError struct {
	ErrorMessage string
	ErrorType    string
}

type response struct {
	req *rpc.Request
	out *lambda.InvokeOutput
	err error
}

type clientcodec struct {
	svc    *lambda.Lambda
	name   *string
	qual   *string
	c      chan response
	dec    *gob.Decoder
	logger func(string, ...interface{})
}

// SetLogger enables logging of CloudWatch invocation
// log tails with logger. If logger is nil, disables
// logging.
func SetLogger(codec rpc.ClientCodec, logger func(string, ...interface{})) error {
	if codec == nil {
		return fmt.Errorf("SetLogger: codec must not be nil")
	}
	c, ok := codec.(*clientcodec)
	if !ok {
		return fmt.Errorf("SetLogger: codec is not from this package")
	}
	c.logger = logger
	return nil
}

func (c *clientcodec) Lambda() *lambda.Lambda {
	return c.svc
}

func (c *clientcodec) ReadResponseHeader(res *rpc.Response) error {
	r := <-c.c
	if r.err != nil {
		res.Error = r.err.Error()
		res.Seq = r.req.Seq
		res.ServiceMethod = r.req.ServiceMethod
		return nil
	}
	if c.logger != nil && r.out.LogResult != nil {
		b, err := base64.StdEncoding.DecodeString(*r.out.LogResult)
		if err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(b))
			for scanner.Scan() {
				c.logger("%s %v: %s", r.req.ServiceMethod, r.req.Seq, scanner.Text())
			}
		}
	}
	if r.out.FunctionError != nil {
		fe := FunctionError{}
		if err := json.Unmarshal(r.out.Payload, &fe); err != nil {
			return err
		}
		res.Error = fe.ErrorMessage
		res.Seq = r.req.Seq
		res.ServiceMethod = r.req.ServiceMethod
		return nil
	}
	dec := gob.NewDecoder(bytes.NewReader(r.out.Payload))
	err := dec.Decode(res)
	if err != nil {
		return err
	}
	c.dec = dec
	return nil
}

func (c *clientcodec) ReadResponseBody(i interface{}) error {
	if i == nil {
		c.dec = nil
		return nil
	}
	return c.dec.Decode(i)
}

func (c *clientcodec) Close() error {
	return nil
}

func (c *clientcodec) invoke(req *rpc.Request, ir *lambda.InvokeInput) {
	res, err := c.svc.Invoke(ir)
	c.c <- response{req: req, out: res, err: err}
}

func (c *clientcodec) WriteRequest(req *rpc.Request, i interface{}) error {
	buf := new(bytes.Buffer)
	encode := gob.NewEncoder(buf).Encode
	err := encode(req)
	if err != nil {
		return err
	}
	if err = encode(i); err != nil {
		return err
	}
	payload, err := json.Marshal(buf.Bytes())
	if err != nil {
		return err
	}
	ir := &lambda.InvokeInput{
		FunctionName: c.name,
		Qualifier:    c.qual,
		Payload:      payload,
	}
	if c.logger != nil {
		ir.LogType = tail
	}
	go c.invoke(req, ir)
	return nil
}

// DefaultCodec calls NewCodec with a default
// Lambda client in the "us-east-1"
// region or AWS_REGION if set.
func DefaultCodec(funcName string) (rpc.ClientCodec, error) {
	var sess = session.New(&aws.Config{
		Region: defaultregion(),
	})
	svc := lambda.New(sess)
	return NewCodec(svc, funcName)
}

// NewCodec returns an rpc.ClientCodec for Lambda
// client svc with funcName. The funcName
// argument may include an alias or version qualifier (for example
// "Function:1" specifies version 1 of Function).
func NewCodec(svc *lambda.Lambda, funcName string) (rpc.ClientCodec, error) {
	if svc == nil {
		return nil, fmt.Errorf("NewCodec: *lambda.Lambda must not be nil")
	}
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
	res, err := svc.GetFunction(req)
	if err != nil {
		return nil, err
	}
	if s := *res.Configuration.Runtime; s != `go1.x` {
		return nil, fmt.Errorf("NewCodec: runtime is not go1.x (%s)", s)
	}
	return &clientcodec{name: &funcName, qual: qualifier, svc: svc, c: make(chan response, 1024)}, nil
}
