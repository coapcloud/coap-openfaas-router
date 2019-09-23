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

type routeTriple struct {
	FuncName string
	Verb     coap.COAPCode
	Type     coap.COAPType
}

func main() {
	flag.StringVarP(&openfaasAddress, "addr", "a", openfaasAddress, "openfaas gateway address")
	flag.Parse()

	mux := coap.NewServeMux()

	registerRoutes(mux, routes)

	fmt.Printf("serving CoAP requests on %d\n", *port)
	log.Fatal(coap.ListenAndServe("udp", fmt.Sprintf(":%d", *port), mux))
}

var routes = map[string][]routeTriple{
	"add": []routeTriple{
		routeTriple{
			Verb:     coap.POST,
			Type:     coap.NonConfirmable,
			FuncName: "add",
		},
	},
	"go-fn": []routeTriple{
		routeTriple{
			Verb:     coap.GET,
			Type:     coap.NonConfirmable,
			FuncName: "go-fn",
		},
	},
}

func registerRoutes(mux *coap.ServeMux, routes map[string][]routeTriple) {
	// this still does not support multiple verbs at the same route, will update
	for k, rr := range routes {
		for _, v := range rr {
			mux.Handle(k, faasHandler{OpenFaasFuncID: openfaasCallback(v.FuncName), Code: v.Verb, Type: v.Type})
		}
	}

	fmt.Printf("registering routes: %v\n", routes)
}
