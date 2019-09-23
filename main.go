package main

import (
	"fmt"
	"log"

	coap "github.com/go-ocf/go-coap"
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

	mux := coap.NewServeMux()
	registerRoutes(mux, routes)

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

func registerRoutes(mux *coap.ServeMux, routes map[string][]routeTuple) {
	r := NewRouter()

	// this still does not support multiple verbs at the same route, will update
	for k, routeTuples := range routes {
		mux.Handle(k, r) // need to fork to allow a wildcard route so we don't have to do this

		for _, v := range routeTuples {
			r.registerRoute(v.Verb, k, v.FuncName)
		}
	}
}
