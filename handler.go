package main

import (
	"context"
	"fmt"
	"log"
	"time"

	coap "github.com/go-ocf/go-coap"
)

type openFaasCallback string

type faasGetHandler struct {
	cbName openFaasCallback
	method coap.COAPCode
}

func (h faasGetHandler) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	var (
		respBdy string
		err     error
	)

	log.Printf("Got message: path=%q: %#v from %v\n\n", r.Msg.Path(), r.Msg, r.Client.RemoteAddr())

	if r.Msg.Code() != h.method {
		w.SetCode(coap.MethodNotAllowed)
		respBdy = "method not allowed"
	} else {
		// run openfaas function for route + verb
		respBdy, err = openfaasCall(h.cbName, 2, 2)
		if err != nil {
			panic(err) // error invoking openfaas function
		}
	}

	ctx, cancel := context.WithTimeout(r.Ctx, time.Second)
	defer cancel()

	log.Printf("Writing response to %v\n\n", r.Client.RemoteAddr())

	w.SetContentFormat(coap.TextPlain)
	if _, err := w.WriteWithContext(ctx, []byte(respBdy)); err != nil {
		log.Printf("Cannot send response: %v", err)
	}
}

func openfaasCall(name openFaasCallback, args ...interface{}) (string, error) {
	// invoke openfaas function
	// output := invokefaas(name, args)
	output := "4"
	return fmt.Sprintf("Invoked func: %v with args: %v. Result: %s", name, args, output), nil
}
