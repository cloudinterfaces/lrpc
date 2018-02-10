package main

import (
	"flag"
	"log"

	"github.com/cloudinterfaces/lrpc/client"
)

// Args is arguments.
type Args struct {
	A, B int
}

// Quotient is quotient.
type Quotient struct {
	Quo, Rem int
}

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
	args := Args{A: 5, B: 4}
	quot := Quotient{}
	err = c.Call("Arith.Divide", &args, &quot)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%v / %v = %v with remainder %v/%v", args.A, args.B, quot.Quo, quot.Rem, args.B)
}
