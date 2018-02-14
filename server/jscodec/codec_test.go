package jscodec

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"testing"
)

var server = rpc.NewServer()

type Service struct{}

func (*Service) Error(args *string, reply *string) error {
	return fmt.Errorf(*args)
}

func (*Service) Add(args *[]float64, reply *float64) error {
	var sum float64
	floats := *args
	for _, f := range floats {
		sum += f
	}
	*reply = sum
	return nil
}

func init() {
	server.Register(new(Service))
}

func TestRequest(t *testing.T) {
	r := []byte(`{"method":"Service.Add","params":[1,2,3],"id":1}`)
	req := JSRequest{}
	err := json.Unmarshal(r, &req)
	if err != nil {
		t.Fatal(err)
	}
	if err = server.ServeRequest(&req); err != nil {
		t.Fatal(err)
	}
	output := req.Output()
	res := JSResponse{}
	if err = json.Unmarshal(output, &res); err != nil {
		t.Fatal(err)
	}
	if res.Result.(float64) != 6 {
		t.Fatalf("want result 6, got %v", res.Result)
	}
	if res.Error != nil {
		t.Fatalf("error not nil")
	}
	if len(res.ID) != 1 || res.ID[0] != '1' {
		t.Fatalf("want id 1 got %v", res.ID)
	}
}

func TestBadRequest(t *testing.T) {
	r := []byte(`{"method":"Service.Add","params":["a","b"],"id":1}`)
	req := JSRequest{}
	err := json.Unmarshal(r, &req)
	if err != nil {
		t.Fatal(err)
	}
	if err = server.ServeRequest(&req); err == nil {
		t.Fatal("error expected")
	}
}

func TestError(t *testing.T) {
	r := []byte(`{"method":"Service.Error","params":"hai","id":1}`)
	req := JSRequest{}
	err := json.Unmarshal(r, &req)
	if err != nil {
		t.Fatal(err)
	}
	if err = server.ServeRequest(&req); err != nil {
		t.Fatal(err)
	}
	output := req.Output()
	res := JSResponse{}
	if err = json.Unmarshal(output, &res); err != nil {
		t.Fatal(err)
	}
	if res.Result != nil {
		t.Fatal("should have nil result")
	}
	if res.Error.(string) != "hai" {
		t.Fatalf("Expect error `hai` got %v", res.Error)
	}
}

func TestNotify(t *testing.T) {
	r := []byte(`{"method":"Service.Error","params":"hai"}`)
	req := JSRequest{}
	err := json.Unmarshal(r, &req)
	if err != nil {
		t.Fatal(err)
	}
	if err = server.ServeRequest(&req); err != nil {
		t.Fatal(err)
	}
	output := req.Output()
	if len(output) != 0 {
		t.Fatalf("Unexpected output, got %s", string(output))
	}
}
