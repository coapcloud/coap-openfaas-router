package main

import (
	"fmt"
	"log"

	coap "github.com/coapcloud/go-coap"
	flag "github.com/spf13/pflag"
)

var (
	openfaasAddress = "http://127.0.0.1:31112"
	port            = flag.IntP("port", "p", 5683, "coap port to listen on")
)

// add flags to register routes too (plumb through to openfaas)

type routeTuple struct {
	FuncName string
	Verb     coap.COAPCode
}

func main() {
	flag.StringVarP(&openfaasAddress, "addr", "a", openfaasAddress, "openfaas gateway address")
	flag.Parse()

	r := NewRouter()
	registerRoutes(&r, routes)

	mux := coap.NewServeMux()
	mux.Handle("*", r)

	fmt.Println("starting store sync")
	go run(&r)

	fmt.Printf("serving CoAP requests on %d\n", *port)
	log.Fatal(coap.ListenAndServe("udp", fmt.Sprintf(":%d", *port), mux))
}

var routes = map[string][]routeTuple{
	"add": []routeTuple{
		routeTuple{
			Verb:     coap.GET,
			FuncName: "sum",
		},
		routeTuple{
			Verb:     coap.POST,
			FuncName: "add",
		},
	},
	"go-fn": []routeTuple{
		routeTuple{
			Verb:     coap.GET,
			FuncName: "go-fn",
		},
	},
}

func registerRoutes(r *Router, routes map[string][]routeTuple) {
	for k, routeTuples := range routes {
		for _, v := range routeTuples {
			r.registerRoute(v.Verb, k, v.FuncName)
		}
	}
}
