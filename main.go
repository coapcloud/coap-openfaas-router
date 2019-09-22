package main

import (
	"fmt"
	"log"

	coap "github.com/go-ocf/go-coap"
	flag "github.com/spf13/pflag"
)

var port = flag.IntP("port", "p", 5683, "coap port to listen on")

// add flags to register routes + functions too (plumb through to openfaas)

var methods = map[string]coap.COAPCode{
	"GET":    coap.GET,
	"POST":   coap.POST,
	"PUT":    coap.PUT,
	"DELETE": coap.DELETE,
}

type routeTuple struct {
	Method   coap.COAPCode
	Callback coap.Handler
}

// add registration ability for this in a bit
var routes = map[string]routeTuple{
	"/foo": routeTuple{
		methods["GET"],
		faasGetHandler{cbName: "foo.get", method: methods["GET"]},
	},
}

func main() {
	flag.Parse()

	mux := coap.NewServeMux()
	for path, routeTuple := range routes {
		mux.Handle(path, routeTuple.Callback)
	}

	fmt.Printf("serving coap requests on %d\n", *port)
	log.Fatal(coap.ListenAndServe("udp", fmt.Sprintf(":%d", *port), mux))
}
