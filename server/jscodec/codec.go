package jscodec

import (
	"encoding/json"
	"log"
	"net/rpc"
	"sync/atomic"
)

var counter = new(uint64)

type JSResponse struct {
	Result interface{}     `json:"result"`
	Error  interface{}     `json:"error"`
	ID     json.RawMessage `json:"id,omitempty"`
}

type JSRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	ID     json.RawMessage `json:"id"`
	out    []byte
}

func (j *JSRequest) Output() []byte {
	return j.out
}

func (j *JSRequest) Close() error {
	return nil
}

func (j *JSRequest) ReadRequestHeader(req *rpc.Request) error {
	req.ServiceMethod = j.Method
	req.Seq = atomic.AddUint64(counter, 1)
	return nil
}

func (j *JSRequest) ReadRequestBody(i interface{}) error {
	return json.Unmarshal([]byte(j.Params), i)
}

func (j *JSRequest) WriteResponse(res *rpc.Response, i interface{}) error {
	if len(j.ID) == 0 {
		return nil
	}
	r := JSResponse{
		ID: j.ID,
	}
	if len(res.Error) > 0 {
		r.Error = res.Error
		o, err := json.Marshal(r)
		j.out = o
		return err
	}
	r.Result = i
	o, err := json.Marshal(r)
	if err != nil {
		log.Println("ERROR", err, "while encoding:")
		log.Printf("%#v", i)
	}
	j.out = o
	return err
}
