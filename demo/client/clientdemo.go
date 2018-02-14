package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/cloudinterfaces/lrpc/client"
	"github.com/cloudinterfaces/lrpc/demo"
)

func main() {
	flag.Parse()
	funcName := flag.Arg(0)
	if len(funcName) == 0 {
		log.Println("Function name required as first argument")
		log.Fatal("demo funcName")
	}
	c, err := client.Default(funcName)
	if err != nil {
		log.Printf("AWS_REGION may need to be set to use function name: %s", funcName)
		log.Fatal(err)
	}
	var out string
	quot := demo.Quotient{}
	args := demo.Args{A: 5, B: 2}
	log.Printf("Calling Arith.Divide with: %#v", args)
	if err = c.Call("Arith.Divide", &args, &quot); err != nil {
		log.Fatal("Unexpected error:", err)
	}
	out = fmt.Sprintf("%v / %v = %v", args.A, args.B, quot.Quo)
	if quot.Rem > 0 {
		out += fmt.Sprintf(" with remainder %v/%v", quot.Rem, args.B)
	}
	log.Println(out)
	log.Println("Calling Arith.Panic")
	if err = c.Call("Arith.Panic", &out, &out); err != nil {
		if client.IsServerPanic(err) {
			log.Println("Expected panic: ", err)
		} else {
			log.Fatal("Not a panic: ", err)
		}
	}
	log.Println("Calling Arith.Error")
	if err = c.Call("Arith.Error", &out, &out); err != nil {
		if client.IsMethodErr(err) {
			log.Println("Expected err:", err)
		} else {
			log.Fatal("Unexpected err:", err)
		}
	}
	log.Println("Calling Arith.BadIdea")
	if err = c.Call("Arith.BadIdea", new(struct{}), &out); err != nil {
		log.Fatal(err)
	}
	log.Printf("Request ID was: %s", out)
	log.Println("Calling Arith.Divide via JSON-RPC", args)
	if lam := c.(interface {
		Lambda() *lambda.Lambda
	}); lam != nil {
		l := lam.Lambda()
		req := lambda.InvokeInput{
			FunctionName: &funcName,
			Payload:      []byte(`{"method":"Arith.Divide","params":{"A":5,"B":2},"id":"one"}`),
		}
		res, err := l.Invoke(&req)
		if err != nil {
			log.Fatal(err)
		}
		if res.FunctionError != nil {
			log.Fatalf("FunctionError: %s", *res.FunctionError)
		}
		log.Printf("Output: %s", string(res.Payload))
	}

}
